package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

// makeFactory wires up a test Factory pointing at the given httptest server.
// BaseURL is set to srv.URL (no /2.0 suffix) because api.New uses baseURL+path
// directly; the project API paths begin with /workspaces/... .
func makeFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	ios, _, out, _ := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL,
		Workspace:  func() (string, error) { return "testws", nil },
	}
	return f, out
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

// sampleProject returns a minimal project JSON object.
func sampleProject() map[string]any {
	return map[string]any{
		"key":        "PROJ",
		"name":       "My Project",
		"description": "desc",
		"is_private": true,
		"uuid":       "{uuid-here}",
		"links": map[string]any{
			"html": map[string]any{
				"href": "https://bitbucket.org/testws/PROJ",
			},
		},
	}
}

// --------------------------------------------------------------------------
// bb project list
// --------------------------------------------------------------------------

func TestProjectList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/testws/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		body := pageJSON(t, []map[string]any{sampleProject()})
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "PROJ") {
		t.Errorf("output should contain key %q, got: %q", "PROJ", got)
	}
	if !strings.Contains(got, "My Project") {
		t.Errorf("output should contain name %q, got: %q", "My Project", got)
	}
}

func TestProjectList_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"pagelen":0,"page":1,"size":0,"values":[]}`)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on empty list: %v", err)
	}
}

func TestProjectList_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"internal server error"}}`, http.StatusInternalServerError)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err == nil {
		t.Error("expected error from 500 response, got nil")
	}
}

// --------------------------------------------------------------------------
// bb project view
// --------------------------------------------------------------------------

func TestProjectView_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/testws/projects/PROJ" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sampleProject()); err != nil {
			t.Errorf("encoding project: %v", err)
		}
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"PROJ"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "PROJ") {
		t.Errorf("output should contain key %q, got: %q", "PROJ", got)
	}
	if !strings.Contains(got, "My Project") {
		t.Errorf("output should contain name %q, got: %q", "My Project", got)
	}
}

func TestProjectView_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"project not found"}}`, http.StatusNotFound)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"GHOST"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for 404, got nil")
	}
}

// --------------------------------------------------------------------------
// bb project create
// --------------------------------------------------------------------------

func TestProjectCreate_MutuallyExclusiveFlags(t *testing.T) {
	// Server should not be called; use a no-op server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called, got request: %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdCreate(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--key", "PROJ", "--name", "My Project", "--public", "--private"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive --public/--private, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention 'mutually exclusive', got: %q", err.Error())
	}
}

func TestProjectCreate_NoTTY_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workspaces/testws/projects" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(sampleProject()); err != nil {
			t.Errorf("encoding project: %v", err)
		}
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdCreate(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--key", "PROJ", "--name", "My Project"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "PROJ") {
		t.Errorf("success message should contain project key %q, got: %q", "PROJ", got)
	}
}

// --------------------------------------------------------------------------
// bb project update
// --------------------------------------------------------------------------

func TestProjectUpdate_MutuallyExclusiveFlags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called, got request: %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdUpdate(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"PROJ", "--name", "New Name", "--public", "--private"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive --public/--private, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention 'mutually exclusive', got: %q", err.Error())
	}
}

func TestProjectUpdate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/workspaces/testws/projects/PROJ" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		updated := sampleProject()
		updated["name"] = "New Name"
		if err := json.NewEncoder(w).Encode(updated); err != nil {
			t.Errorf("encoding project: %v", err)
		}
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdUpdate(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"PROJ", "--name", "New Name"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "PROJ") {
		t.Errorf("success message should contain project key %q, got: %q", "PROJ", got)
	}
}

// --------------------------------------------------------------------------
// bb project delete
// --------------------------------------------------------------------------

func TestProjectDelete_NoTTYRequiresForce(t *testing.T) {
	// Buffer-backed IOStreams are not a TTY, so delete without --force must fail.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called, got request: %s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdDelete(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"PROJ"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected NoTTYError when running without --force in non-TTY mode, got nil")
	}

	var noTTY *cmdutil.NoTTYError
	if !errors.As(err, &noTTY) {
		t.Errorf("expected *cmdutil.NoTTYError, got: %T: %v", err, err)
	}
}

func TestProjectDelete_ForceFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/workspaces/testws/projects/PROJ" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdDelete(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"PROJ", "--force"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "PROJ") {
		t.Errorf("success message should mention project key %q, got: %q", "PROJ", got)
	}
}
