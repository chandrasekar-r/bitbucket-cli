package issue

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

// changeToBitbucketRepo creates a temp dir with a git repo whose origin remote
// is a Bitbucket HTTPS URL (workspace=testws, repo=testrepo), then changes the
// test's working directory to it so gitcontext.FromRemote() returns a valid
// RepoContext with workspace="testws" and slug="testrepo".
func changeToBitbucketRepo(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("remote", "add", "origin", "https://bitbucket.org/testws/testrepo.git")
	t.Chdir(dir)
}

// issueServer builds a test HTTP server that handles the repo metadata endpoint
// (needed by guardIssues/HasIssues) and delegates all other paths to extra.
func issueServer(t *testing.T, hasIssues bool, extra http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/2.0/repositories/testws/testrepo":
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"slug":       "testrepo",
				"has_issues": hasIssues,
			}); err != nil {
				t.Errorf("encoding repo response: %v", err)
			}
		default:
			if extra != nil {
				extra(w, r)
			} else {
				http.NotFound(w, r)
			}
		}
	}))
}

// makeFactory returns a *cmdutil.Factory wired to srv and a *bytes.Buffer that
// captures everything written to IOStreams.Out.
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

// ---------------------------------------------------------------------------
// guard / has_issues
// ---------------------------------------------------------------------------

func TestIssueGuard_IssuesDisabled(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, false, nil)
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when issues are disabled")
	}
	if !strings.Contains(err.Error(), "issues are not enabled") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// bb issue list
// ---------------------------------------------------------------------------

func TestIssueList_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"id":     1,
					"title":  "Bug",
					"status": "open",
					"kind":   "bug",
				},
			},
			"pagelen": 1,
			"size":    1,
		}); err != nil {
			t.Errorf("encoding issues response: %v", err)
		}
	})
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Bug") {
		t.Errorf("expected 'Bug' in output, got: %q", out.String())
	}
}

func TestIssueList_Empty(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"values":  []map[string]interface{}{},
			"pagelen": 0,
			"size":    0,
		}); err != nil {
			t.Errorf("encoding empty issues response: %v", err)
		}
	})
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "No issues found") {
		t.Errorf("expected 'No issues found', got: %q", out.String())
	}
}

func TestIssueList_StateFilter(t *testing.T) {
	changeToBitbucketRepo(t)

	var capturedQuery string
	srv := issueServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"values":  []map[string]interface{}{},
			"pagelen": 0,
			"size":    0,
		}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	})
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--state", "resolved"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ListIssues builds q=status%3D%22resolved%22 when state is set.
	if !strings.Contains(capturedQuery, "resolved") {
		t.Errorf("expected 'resolved' in query string, got: %q", capturedQuery)
	}
}

// ---------------------------------------------------------------------------
// bb issue view
// ---------------------------------------------------------------------------

func TestIssueView_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/testws/testrepo/issues/42" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       42,
			"title":    "Critical crash on startup",
			"status":   "open",
			"kind":     "bug",
			"priority": "critical",
			"reporter": map[string]string{"username": "alice", "display_name": "Alice"},
			"links": map[string]interface{}{
				"html": map[string]string{"href": "https://bitbucket.org/testws/testrepo/issues/42"},
			},
		}); err != nil {
			t.Errorf("encoding issue response: %v", err)
		}
	})
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Critical crash on startup") {
		t.Errorf("expected title in output, got: %q", got)
	}
	if !strings.Contains(got, "open") {
		t.Errorf("expected state in output, got: %q", got)
	}
	if !strings.Contains(got, "bug") {
		t.Errorf("expected kind in output, got: %q", got)
	}
}

func TestIssueView_InvalidID(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, true, nil)
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"notanumber"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid issue number")
	}
	var flagErr *cmdutil.FlagError
	if !errors.As(err, &flagErr) {
		t.Errorf("expected *cmdutil.FlagError, got: %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// bb issue close
// ---------------------------------------------------------------------------

func TestIssueClose_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/testws/testrepo/issues/1" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     1,
			"status": "resolved",
		}); err != nil {
			t.Errorf("encoding close response: %v", err)
		}
	})
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdClose(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Issue #1") {
		t.Errorf("expected issue reference in output, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "resolved") {
		t.Errorf("expected 'resolved' in output, got: %q", out.String())
	}
}

func TestIssueClose_InvalidID(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, true, nil)
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdClose(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"notanumber"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid issue number")
	}
	var flagErr *cmdutil.FlagError
	if !errors.As(err, &flagErr) {
		t.Errorf("expected *cmdutil.FlagError, got: %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// bb issue reopen
// ---------------------------------------------------------------------------

func TestIssueReopen_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	var capturedBody map[string]interface{}
	srv := issueServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/testws/testrepo/issues/1" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     1,
			"status": "open",
		}); err != nil {
			t.Errorf("encoding reopen response: %v", err)
		}
	})
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdReopen(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Issue #1 reopened") {
		t.Errorf("expected reopen confirmation, got: %q", out.String())
	}
	if capturedBody["status"] != "open" {
		t.Errorf("expected status=open in PUT body, got: %v", capturedBody["status"])
	}
}

// ---------------------------------------------------------------------------
// bb issue comment
// ---------------------------------------------------------------------------

func TestIssueComment_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	var capturedBody map[string]interface{}
	srv := issueServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2.0/repositories/testws/testrepo/issues/7/comments" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Errorf("decoding comment body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	})
	defer srv.Close()

	f, out := makeFactory(t, srv)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"7", "--body", "LGTM"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Comment added") {
		t.Errorf("expected confirmation in output, got: %q", out.String())
	}
	content, _ := capturedBody["content"].(map[string]interface{})
	if content == nil || content["raw"] != "LGTM" {
		t.Errorf("unexpected comment body: %v", capturedBody)
	}
}

func TestIssueComment_EmptyBodyNoTTY(t *testing.T) {
	// comment.go checks body == "" before repoCtx, so we still need a valid git repo
	// because the ID is parsed first, then the TTY check fires.
	changeToBitbucketRepo(t)

	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP request should not be made when body is missing in no-TTY mode")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer dummy.Close()

	f, _ := makeFactory(t, dummy)
	// IOStreams.Out is a *bytes.Buffer → IsStdoutTTY() returns false.
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"7"}) // no --body

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --body in no-TTY mode")
	}
	var flagErr *cmdutil.FlagError
	if !errors.As(err, &flagErr) {
		t.Errorf("expected *cmdutil.FlagError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "--body is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestIssueComment_InvalidID(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := issueServer(t, true, nil)
	defer srv.Close()

	f, _ := makeFactory(t, srv)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"notanumber", "--body", "hello"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid issue number")
	}
	var flagErr *cmdutil.FlagError
	if !errors.As(err, &flagErr) {
		t.Errorf("expected *cmdutil.FlagError, got: %T: %v", err, err)
	}
}
