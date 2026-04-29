package pr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

// changeToBitbucketRepo creates a temp dir with a git repo whose origin remote is a
// Bitbucket HTTPS URL (workspace=testws, repo=testrepo), then changes the test's
// working directory to it so gitcontext.FromRemote() returns a valid RepoContext.
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

// commentFactory returns a *cmdutil.Factory wired to srv and injectable IOStreams.
func commentFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	return &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
	}, ios.Out.(*bytes.Buffer)
}

// commentServer returns an httptest.Server that accepts a POST comment request and
// decodes the request body into capturedBody. All other requests get a 404.
func commentServer(t *testing.T, capturedBody *map[string]interface{}, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if capturedBody != nil {
			if err := json.NewDecoder(r.Body).Decode(capturedBody); err != nil {
				t.Errorf("decoding request body: %v", err)
			}
		}
		w.WriteHeader(statusCode)
	}))
}


func TestCommentCmd_LineLevelInline(t *testing.T) {
	changeToBitbucketRepo(t)

	var body map[string]interface{}
	srv := commentServer(t, &body, http.StatusCreated)
	defer srv.Close()

	f, stdout := commentFactory(t, srv)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--file", "pkg/api/prs.go", "--line", "42", "--body", "looks good"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inline, _ := body["inline"].(map[string]interface{})
	if inline == nil {
		t.Fatal("expected 'inline' key in POST body")
	}
	if inline["path"] != "pkg/api/prs.go" {
		t.Errorf("inline.path: got %v, want %q", inline["path"], "pkg/api/prs.go")
	}
	if inline["to"] != float64(42) {
		t.Errorf("inline.to: got %v, want 42", inline["to"])
	}
	if _, hasFrom := inline["from"]; hasFrom {
		t.Error("inline.from should be absent when From is nil")
	}

	out := stdout.String()
	if !strings.Contains(out, "pkg/api/prs.go:42") {
		t.Errorf("expected file:line in output, got: %q", out)
	}
	if !strings.Contains(out, "PR #5") {
		t.Errorf("expected PR number in output, got: %q", out)
	}
}

func TestCommentCmd_FileLevelInline(t *testing.T) {
	changeToBitbucketRepo(t)

	var body map[string]interface{}
	srv := commentServer(t, &body, http.StatusCreated)
	defer srv.Close()

	f, stdout := commentFactory(t, srv)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--file", "pkg/api/prs.go", "--body", "file looks good"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inline, _ := body["inline"].(map[string]interface{})
	if inline == nil {
		t.Fatal("expected 'inline' key in POST body")
	}
	if inline["path"] != "pkg/api/prs.go" {
		t.Errorf("inline.path: got %v, want %q", inline["path"], "pkg/api/prs.go")
	}
	if _, hasTo := inline["to"]; hasTo {
		t.Error("inline.to should be absent for file-level comment")
	}
	if _, hasFrom := inline["from"]; hasFrom {
		t.Error("inline.from should be absent for file-level comment")
	}

	out := stdout.String()
	if strings.Contains(out, ":") && strings.Contains(out, "pkg/api/prs.go:") {
		t.Errorf("file-level message should not contain line number, got: %q", out)
	}
	if !strings.Contains(out, "pkg/api/prs.go") {
		t.Errorf("expected filename in output, got: %q", out)
	}
	if !strings.Contains(out, "PR #5") {
		t.Errorf("expected PR number in output, got: %q", out)
	}
}

func TestCommentCmd_GenericComment(t *testing.T) {
	changeToBitbucketRepo(t)

	var body map[string]interface{}
	srv := commentServer(t, &body, http.StatusCreated)
	defer srv.Close()

	f, stdout := commentFactory(t, srv)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--body", "LGTM"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, hasInline := body["inline"]; hasInline {
		t.Error("generic comment must not contain 'inline' key")
	}

	out := stdout.String()
	if !strings.Contains(out, "PR #5") {
		t.Errorf("expected PR number in output, got: %q", out)
	}
}

func TestCommentCmd_LineWithoutFile(t *testing.T) {
	// Flag validation fires before repoContext — no server needed.
	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP request should not be made for flag validation error")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer dummy.Close()

	f, _ := commentFactory(t, dummy)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--line", "42", "--body", "x"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for --line without --file")
	}
	if !strings.Contains(err.Error(), "--line requires --file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommentCmd_LineZero(t *testing.T) {
	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP request should not be made for flag validation error")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer dummy.Close()

	f, _ := commentFactory(t, dummy)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--file", "main.go", "--line", "0", "--body", "x"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for --line 0")
	}
	if !strings.Contains(err.Error(), "--line must be a positive integer") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "got 0") {
		t.Errorf("error should include the bad value, got: %v", err)
	}
}

func TestCommentCmd_LineNegative(t *testing.T) {
	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP request should not be made for flag validation error")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer dummy.Close()

	f, _ := commentFactory(t, dummy)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	// Use --line=-1 form to avoid pflag ambiguity with negative-value args.
	cmd.SetArgs([]string{"5", "--file", "main.go", "--line=-1", "--body", "x"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for --line -1")
	}
	if !strings.Contains(err.Error(), "--line must be a positive integer") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommentCmd_EmptyFile(t *testing.T) {
	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP request should not be made for flag validation error")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer dummy.Close()

	f, _ := commentFactory(t, dummy)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--file", "", "--body", "x"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for --file with empty value")
	}
	if !strings.Contains(err.Error(), "--file cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommentCmd_EmptyBodyNoTTY(t *testing.T) {
	// Body validation is post-repoContext, so we need a valid Bitbucket remote.
	changeToBitbucketRepo(t)

	dummy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP request should not be made when body is missing")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer dummy.Close()

	f, _ := commentFactory(t, dummy)
	// IOStreams backed by a Buffer → IsStdoutTTY() = false (no-TTY mode).
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	// No --body → triggers body-required error.
	cmd.SetArgs([]string{"5", "--file", "foo.go"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --body in no-TTY mode")
	}
	if !strings.Contains(err.Error(), "--body is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCommentCmd_APIError400(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "This pull request is not open"},
		}); err != nil {
			t.Errorf("encoding error response: %v", err)
		}
	}))
	defer srv.Close()

	f, _ := commentFactory(t, srv)
	cmd := newCmdComment(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--file", "main.go", "--line", "1", "--body", "x"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error on 400 response")
	}
	if !strings.Contains(err.Error(), "This pull request is not open") {
		t.Errorf("expected API message in error, got: %v", err)
	}
}

func TestCommentCmd_InlineVsGenericRouting(t *testing.T) {
	changeToBitbucketRepo(t)

	tests := []struct {
		name        string
		args        []string
		wantInline  bool
	}{
		{
			name:       "inline when --file set",
			args:       []string{"5", "--file", "main.go", "--body", "x"},
			wantInline: true,
		},
		{
			name:       "generic when no --file",
			args:       []string{"5", "--body", "x"},
			wantInline: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body map[string]interface{}
			srv := commentServer(t, &body, http.StatusCreated)
			defer srv.Close()

			f, _ := commentFactory(t, srv)
			cmd := newCmdComment(f)
			cmd.SilenceUsage = true
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, hasInline := body["inline"]
			if hasInline != tt.wantInline {
				t.Errorf("'inline' key present=%v, want %v", hasInline, tt.wantInline)
			}
		})
	}
}
