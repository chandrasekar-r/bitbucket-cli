package workspace

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

// makeFactory wires up a test Factory pointing at the given httptest server.
// BaseURL is set to srv.URL (no /2.0 suffix) because api.New prepends the path
// directly; the pagination and Get helpers use baseURL+path, so the server URL
// must be the raw root.
func makeFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	ios, _, out, errOut := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL,
		Workspace:  func() (string, error) { return "testws", nil },
	}
	return f, out, errOut
}

// pageJSON wraps a slice of values in Bitbucket's paginated envelope.
func pageJSON(t *testing.T, values any) string {
	t.Helper()
	valBytes, err := json.Marshal(values)
	if err != nil {
		t.Fatalf("marshaling values: %v", err)
	}
	return `{"pagelen":10,"page":1,"size":1,"next":"","values":` + string(valBytes) + `}`
}

// --------------------------------------------------------------------------
// workspace list
// --------------------------------------------------------------------------

func TestWorkspaceList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		body := pageJSON(t, []map[string]any{
			{"slug": "ws1", "name": "Workspace 1", "type": "team"},
		})
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	f, out, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "ws1") {
		t.Errorf("output should contain %q, got: %q", "ws1", out.String())
	}
}

func TestWorkspaceList_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"pagelen":0,"page":1,"size":0,"values":[]}`)
	}))
	defer srv.Close()

	f, _, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on empty list: %v", err)
	}
}

func TestWorkspaceList_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"server error"}}`, http.StatusInternalServerError)
	}))
	defer srv.Close()

	f, _, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err == nil {
		t.Error("expected error from 500 response, got nil")
	}
}

// --------------------------------------------------------------------------
// workspace view
// --------------------------------------------------------------------------

func TestWorkspaceView_WithArg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/myws" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		ws := map[string]any{
			"slug": "myws",
			"name": "My Workspace",
			"type": "team",
			"uuid": "{abc-123}",
		}
		if err := json.NewEncoder(w).Encode(ws); err != nil {
			t.Errorf("encoding workspace: %v", err)
		}
	}))
	defer srv.Close()

	f, out, _ := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"myws"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "myws") {
		t.Errorf("output should contain slug %q, got: %q", "myws", got)
	}
	if !strings.Contains(got, "My Workspace") {
		t.Errorf("output should contain name %q, got: %q", "My Workspace", got)
	}
}

func TestWorkspaceView_DefaultWorkspace(t *testing.T) {
	// No arg — should fall back to f.Workspace() which returns "testws".
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/testws" {
			t.Errorf("unexpected path: got %q, want /workspaces/testws", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		ws := map[string]any{
			"slug": "testws",
			"name": "Test Workspace",
			"type": "team",
			"uuid": "{def-456}",
		}
		if err := json.NewEncoder(w).Encode(ws); err != nil {
			t.Errorf("encoding workspace: %v", err)
		}
	}))
	defer srv.Close()

	f, out, _ := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	// No args — uses default workspace from factory.

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "testws") {
		t.Errorf("output should contain %q, got: %q", "testws", out.String())
	}
}

// --------------------------------------------------------------------------
// workspace use
// --------------------------------------------------------------------------

func TestWorkspaceUse_Success(t *testing.T) {
	// Isolate config writes to a temp directory.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/myws" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		ws := map[string]any{
			"slug": "myws",
			"name": "My Workspace",
			"type": "team",
			"uuid": "{abc-123}",
		}
		if err := json.NewEncoder(w).Encode(ws); err != nil {
			t.Errorf("encoding workspace: %v", err)
		}
	}))
	defer srv.Close()

	f, out, _ := makeFactory(t, srv)
	cmd := newCmdUse(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"myws"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Default workspace set") {
		t.Errorf("output should confirm workspace set, got: %q", out.String())
	}
}

func TestWorkspaceUse_NotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"workspace not found"}}`, http.StatusNotFound)
	}))
	defer srv.Close()

	f, _, _ := makeFactory(t, srv)
	cmd := newCmdUse(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"ghost-ws"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %q", err.Error())
	}
}
