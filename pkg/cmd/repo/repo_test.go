package repo

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

// repoJSON is a minimal but complete repository JSON fixture.
const repoJSON = `{
	"slug":"myrepo",
	"name":"My Repo",
	"full_name":"testws/myrepo",
	"description":"desc",
	"is_private":true,
	"language":"go",
	"mainbranch":{"name":"main"},
	"links":{
		"html":{"href":"https://bitbucket.org/testws/myrepo"},
		"clone":[{"name":"https","href":"https://bitbucket.org/testws/myrepo.git"}]
	}
}`

// listPageJSON wraps one repo in a Bitbucket paginated response envelope.
var listPageJSON = func() string {
	return `{"pagelen":30,"page":1,"size":1,"values":[` + repoJSON + `]}`
}

// emptyPageJSON is a paginated response with no results.
const emptyPageJSON = `{"pagelen":30,"page":1,"size":0,"values":[]}`

// makeFactory builds a test Factory wired to the given httptest.Server.
// It returns the factory and a buffer capturing stdout.
func makeFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")),
		Out:    out,
		ErrOut: &bytes.Buffer{},
	}
	baseURL := ""
	if srv != nil {
		baseURL = srv.URL + "/2.0"
	}
	f := &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    baseURL,
		Workspace:  func() (string, error) { return "testws", nil },
	}
	return f, out
}

// ── resolveRepo ──────────────────────────────────────────────────────────────

func TestResolveRepo_SlashFormat(t *testing.T) {
	f, _ := makeFactory(t, nil)
	ws, slug, err := resolveRepo(f, []string{"ws/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws != "ws" {
		t.Errorf("workspace: got %q, want %q", ws, "ws")
	}
	if slug != "repo" {
		t.Errorf("slug: got %q, want %q", slug, "repo")
	}
}

func TestResolveRepo_TooManySlashes(t *testing.T) {
	f, _ := makeFactory(t, nil)
	_, _, err := resolveRepo(f, []string{"a/b/c"})
	if err == nil {
		t.Fatal("expected error for arg with too many slashes, got nil")
	}
}

// ── bb repo list ─────────────────────────────────────────────────────────────

func TestRepoList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/2.0/repositories/testws") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, listPageJSON()); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "myrepo") {
		t.Errorf("output %q does not contain slug %q", out.String(), "myrepo")
	}
}

func TestRepoList_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, emptyPageJSON); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error on empty list: %v", err)
	}
}

func TestRepoList_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"message":"internal server error"}}`, http.StatusInternalServerError)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

// ── bb repo view ─────────────────────────────────────────────────────────────

func TestRepoView_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/2.0/repositories/testws/myrepo"
		if r.URL.Path != want {
			t.Errorf("path: got %q, want %q", r.URL.Path, want)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, repoJSON); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"testws/myrepo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "My Repo") && !strings.Contains(got, "testws/myrepo") {
		t.Errorf("output %q missing repo name or full_name", got)
	}
	if !strings.Contains(got, "main") {
		t.Errorf("output %q missing mainbranch name", got)
	}
}

func TestRepoView_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"message":"repository not found"}}`, http.StatusNotFound)
	}))
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"testws/myrepo"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

// ── bb repo rename ───────────────────────────────────────────────────────────

func TestRepoRename_Success(t *testing.T) {
	// The renamed repo response has the updated full_name.
	renamedJSON := `{
		"slug":"newname",
		"name":"newname",
		"full_name":"testws/newname",
		"description":"desc",
		"is_private":true,
		"language":"go",
		"mainbranch":{"name":"main"},
		"links":{"html":{"href":"https://bitbucket.org/testws/newname"},"clone":[]}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method: got %s, want PUT", r.Method)
		}
		want := "/2.0/repositories/testws/myrepo"
		if r.URL.Path != want {
			t.Errorf("path: got %q, want %q", r.URL.Path, want)
		}
		// Validate body contains new name.
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["name"] != "newname" {
			t.Errorf("body name: got %q, want %q", body["name"], "newname")
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, renamedJSON); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdRename(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"testws/myrepo", "newname"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Renamed") {
		t.Errorf("output %q missing 'Renamed'", out.String())
	}
}

func TestRepoRename_WrongArgCount(t *testing.T) {
	f, _ := makeFactory(t, nil)
	cmd := newCmdRename(f)
	cmd.SilenceUsage = true
	// Only one arg — missing the new name.
	cmd.SetArgs([]string{"testws/myrepo"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing new-name arg, got nil")
	}
}

// ── bb repo fork ─────────────────────────────────────────────────────────────

func TestRepoFork_Success(t *testing.T) {
	forkedJSON := `{
		"slug":"myrepo",
		"name":"myrepo",
		"full_name":"testws/myrepo",
		"description":"fork of myrepo",
		"is_private":true,
		"language":"go",
		"mainbranch":{"name":"main"},
		"links":{"html":{"href":"https://bitbucket.org/testws/myrepo"},"clone":[{"name":"https","href":"https://bitbucket.org/testws/myrepo.git"}]}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		want := "/2.0/repositories/testws/myrepo/forks"
		if r.URL.Path != want {
			t.Errorf("path: got %q, want %q", r.URL.Path, want)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := io.WriteString(w, forkedJSON); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdFork(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"testws/myrepo"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Forked") {
		t.Errorf("output %q missing 'Forked'", out.String())
	}
}

// ── bb repo delete ───────────────────────────────────────────────────────────

func TestRepoDelete_NoTTYRequiresForce(t *testing.T) {
	// No server needed — should fail before any HTTP call.
	f, _ := makeFactory(t, nil)
	// IOStreams backed by bytes.Buffer, so IsStdoutTTY() returns false.
	cmd := newCmdDelete(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"testws/myrepo"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected NoTTYError, got nil")
	}
	var noTTY *cmdutil.NoTTYError
	if !isNoTTYError(err, &noTTY) {
		t.Errorf("expected *cmdutil.NoTTYError, got %T: %v", err, err)
	}
}

func TestRepoDelete_ForceSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method: got %s, want DELETE", r.Method)
		}
		want := "/2.0/repositories/testws/myrepo"
		if r.URL.Path != want {
			t.Errorf("path: got %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdDelete(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"testws/myrepo", "--force"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Deleted") {
		t.Errorf("output %q missing 'Deleted'", out.String())
	}
}

// isNoTTYError checks if err (or an error it wraps) is a *cmdutil.NoTTYError.
func isNoTTYError(err error, target **cmdutil.NoTTYError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*cmdutil.NoTTYError); ok {
		*target = e
		return true
	}
	// errors.As traversal — unwrap manually since we can't import errors here.
	type unwrapper interface{ Unwrap() error }
	if u, ok := err.(unwrapper); ok {
		return isNoTTYError(u.Unwrap(), target)
	}
	// Check string match as last resort (cobra may wrap the error in usage text).
	return strings.Contains(err.Error(), "requires confirmation")
}

// ── validateCloneURL ─────────────────────────────────────────────────────────

func TestValidateCloneURL_Safe(t *testing.T) {
	if err := validateCloneURL("https://bitbucket.org/ws/repo.git"); err != nil {
		t.Errorf("unexpected error for safe URL: %v", err)
	}
}

func TestValidateCloneURL_LeadingDash(t *testing.T) {
	if err := validateCloneURL("-malicious"); err == nil {
		t.Error("expected error for URL starting with '-', got nil")
	}
}

func TestValidateCloneURL_DangerousScheme(t *testing.T) {
	cases := []string{
		"ext::git-remote-ext",
		"file:///etc/passwd",
	}
	for _, u := range cases {
		if err := validateCloneURL(u); err == nil {
			t.Errorf("expected error for dangerous URL %q, got nil", u)
		}
	}
}
