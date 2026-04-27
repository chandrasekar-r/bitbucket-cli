package snippet

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

// makeFactory creates a test Factory that routes HTTP requests to srv.
func makeFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")),
		Out:    out,
		ErrOut: &bytes.Buffer{},
	}
	return &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
		Workspace:  func() (string, error) { return "testws", nil },
	}, out
}

// snippetJSON builds a minimal snippet JSON object for use in handler responses.
func snippetJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":         "AbCdEf",
		"title":      "My Script",
		"is_private": true,
		"updated_on": "2024-01-15T10:00:00Z",
		"owner":      map[string]string{"username": "alice"},
		"files": map[string]interface{}{
			"script.sh": map[string]interface{}{},
		},
		"links": map[string]interface{}{
			"html": map[string]string{
				"href": "https://bitbucket.org/snippets/testws/AbCdEf",
			},
			"clone": []map[string]string{
				{"name": "https", "href": "https://bitbucket.org/snippets/testws/AbCdEf"},
				{"name": "ssh", "href": "git@bitbucket.org:snippets/testws/AbCdEf.git"},
			},
		},
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("writeJSON: %v", err)
	}
}

// ---------------------------------------------------------------------------
// validateSnippetCloneURL
// ---------------------------------------------------------------------------

func TestValidateSnippetCloneURL_HTTPS(t *testing.T) {
	err := validateSnippetCloneURL("https://bitbucket.org/snippets/testws/AbCdEf")
	if err != nil {
		t.Errorf("expected nil error for HTTPS URL, got: %v", err)
	}
}

func TestValidateSnippetCloneURL_LeadingDash(t *testing.T) {
	err := validateSnippetCloneURL("-malicious")
	if err == nil {
		t.Error("expected error for URL starting with '-', got nil")
	}
}

func TestValidateSnippetCloneURL_BadScheme(t *testing.T) {
	err := validateSnippetCloneURL("file:///etc/passwd")
	if err == nil {
		t.Error("expected error for 'file://' scheme, got nil")
	}
}

func TestValidateSnippetCloneURL_SSH(t *testing.T) {
	err := validateSnippetCloneURL("ssh://git@bitbucket.org/snippets/testws/AbCdEf.git")
	if err != nil {
		t.Errorf("expected nil error for SSH URL, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// bb snippet list
// ---------------------------------------------------------------------------

func TestSnippetList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]interface{}{
			"values":  []interface{}{snippetJSON()},
			"pagelen": 1,
			"size":    1,
		})
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "AbCdEf") {
		t.Errorf("output missing snippet ID 'AbCdEf'; got:\n%s", got)
	}
	if !strings.Contains(got, "My Script") {
		t.Errorf("output missing snippet title 'My Script'; got:\n%s", got)
	}
}

func TestSnippetList_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]interface{}{
			"values":  []interface{}{},
			"pagelen": 0,
			"size":    0,
		})
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "No snippets found") {
		t.Errorf("expected 'No snippets found'; got:\n%s", out.String())
	}
}

func TestSnippetList_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(t, w, map[string]interface{}{
			"error": map[string]string{"message": "internal server error"},
		})
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for 500 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// bb snippet view
// ---------------------------------------------------------------------------

func TestSnippetView_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/snippets/testws/AbCdEf") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		writeJSON(t, w, snippetJSON())
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true

	cmd.SetArgs([]string{"AbCdEf"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "My Script") {
		t.Errorf("output missing title; got:\n%s", got)
	}
	if !strings.Contains(got, "alice") {
		t.Errorf("output missing owner; got:\n%s", got)
	}
	if !strings.Contains(got, "script.sh") {
		t.Errorf("output missing filename; got:\n%s", got)
	}
}

func TestSnippetView_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(t, w, map[string]interface{}{
			"error": map[string]string{"message": "snippet not found"},
		})
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true

	cmd.SetArgs([]string{"AbCdEf"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for 404 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// bb snippet create
// ---------------------------------------------------------------------------

func TestSnippetCreate_RequiresTitle(t *testing.T) {
	// Server should never be called; pass a no-op server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when --title is missing")
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdCreate(f)
	cmd.SilenceUsage = true

	// No --title flag → expect a FlagError.
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --title is missing, got nil")
	}
	if !strings.Contains(err.Error(), "--title is required") {
		t.Errorf("expected '--title is required' in error; got: %v", err)
	}
}

func TestSnippetCreate_EmptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called for empty content")
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")), // empty stdin
		Out:    out,
		ErrOut: &bytes.Buffer{},
	}
	f := &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
		Workspace:  func() (string, error) { return "testws", nil },
	}

	cmd := newCmdCreate(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--title", "x"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty stdin content, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' in error; got: %v", err)
	}
}

func TestSnippetCreate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(t, w, snippetJSON())
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("echo hello\n")),
		Out:    out,
		ErrOut: &bytes.Buffer{},
	}
	f := &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
		Workspace:  func() (string, error) { return "testws", nil },
	}

	cmd := newCmdCreate(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--title", "My Script"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "✓ Created snippet") {
		t.Errorf("expected '✓ Created snippet' in output; got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// bb snippet delete
// ---------------------------------------------------------------------------

func TestSnippetDelete_NoTTYRequiresForce(t *testing.T) {
	// Buffer-backed Out → IsStdoutTTY() returns false. Without --force, must get NoTTYError.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when NoTTY without --force")
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv) // uses bytes.Buffer for Out — non-TTY
	cmd := newCmdDelete(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"AbCdEf"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected NoTTYError, got nil")
	}
	var noTTY *cmdutil.NoTTYError
	if !isNoTTYError(err, &noTTY) {
		t.Errorf("expected *cmdutil.NoTTYError; got: %T %v", err, err)
	}
}

func TestSnippetDelete_ForceSuccess(t *testing.T) {
	deleted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/snippets/testws/AbCdEf") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdDelete(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"AbCdEf", "--force"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE request to be made")
	}
	if !strings.Contains(out.String(), "✓ Deleted snippet AbCdEf") {
		t.Errorf("expected success message; got:\n%s", out.String())
	}
}

// isNoTTYError checks whether err is or wraps *cmdutil.NoTTYError.
func isNoTTYError(err error, target **cmdutil.NoTTYError) bool {
	if e, ok := err.(*cmdutil.NoTTYError); ok {
		*target = e
		return true
	}
	return false
}
