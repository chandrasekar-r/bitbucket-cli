package login

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

// testFactory builds a *cmdutil.Factory wired to an httptest server and
// injectable stdin/stdout buffers.
func testFactory(t *testing.T, srv *httptest.Server, stdinContent string) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	in := bytes.NewBufferString(stdinContent)
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(in),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	return &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
	}, ios.Out.(*bytes.Buffer)
}

func mockServer(t *testing.T, username string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encode := func(v any) {
			if err := json.NewEncoder(w).Encode(v); err != nil {
				t.Errorf("encoding JSON response: %v", err)
			}
		}
		switch r.URL.Path {
		case "/2.0/user":
			encode(api.User{Username: username, DisplayName: "Test User"})
		case "/2.0/workspaces":
			encode(map[string]interface{}{
				"values":  []map[string]interface{}{{"slug": "ws1"}},
				"pagelen": 1, "size": 1,
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestLoginWithToken_Success(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()

	// Redirect token store to temp dir via XDG_CONFIG_HOME
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	f, stdout := testFactory(t, srv, "my-api-token\n")

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--with-token", "--username", "testuser"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("login --with-token: %v", err)
	}
	if !strings.Contains(stdout.String(), "Logged in as testuser") {
		t.Errorf("expected success message, got: %q", stdout.String())
	}
}

func TestLoginWithToken_EmptyToken(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()

	f, _ := testFactory(t, srv, "\n") // empty token line

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--with-token", "--username", "testuser"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if !strings.Contains(err.Error(), "token is empty") {
		t.Errorf("expected 'token is empty' error, got: %v", err)
	}
}

func TestLoginWithToken_NoUsername_NoTTY(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()

	// IOStreams from testFactory has IsStdoutTTY()=false (backed by Buffer, not os.File)
	f, _ := testFactory(t, srv, "my-token\n")

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--with-token"}) // no --username

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --username is missing in no-tty mode")
	}
	if !strings.Contains(err.Error(), "--username is required") {
		t.Errorf("expected '--username is required' error, got: %v", err)
	}
}

func TestLoginToken_DirectFlag_Success(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	f, stdout := testFactory(t, srv, "")
	errBuf := f.IOStreams.ErrOut.(*bytes.Buffer)

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--token", "ATBBXXX", "--username", "testuser"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Logged in as testuser") {
		t.Errorf("expected success message, got: %q", stdout.String())
	}
	if !strings.Contains(errBuf.String(), "Warning") {
		t.Errorf("expected security warning on stderr, got: %q", errBuf.String())
	}
}

func TestLoginToken_DirectFlag_EmptyToken(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()

	f, _ := testFactory(t, srv, "")

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--token", "", "--username", "testuser"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty --token value")
	}
	if !strings.Contains(err.Error(), "--token value cannot be empty") {
		t.Errorf("expected empty token error, got: %v", err)
	}
}

func TestLoginToken_DirectFlag_NoUsername(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()

	f, _ := testFactory(t, srv, "")

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--token", "ATBBXXX"}) // no --username

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --username is absent with --token")
	}
	if !strings.Contains(err.Error(), "--username is required when --token is used") {
		t.Errorf("expected --username required error, got: %v", err)
	}
}

func TestLoginToken_DirectFlag_StdinNotConsumed(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Pipe content in stdin — it must NOT be consumed when --token is used.
	stdinBytes := bytes.NewBufferString("piped-content-should-not-be-read\n")
	f, _ := testFactory(t, srv, "")
	f.IOStreams.In = io.NopCloser(stdinBytes)

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--token", "ATBBXXX", "--username", "testuser"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// stdin buffer should still have all its content (scanner was never called)
	if stdinBytes.Len() == 0 {
		t.Error("stdin was consumed by --token path; expected scanner to be skipped")
	}
}

func TestLoginToken_DirectFlag_BothTokenFlags(t *testing.T) {
	srv := mockServer(t, "testuser")
	defer srv.Close()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// --token wins; stdin is not consumed even when --with-token is also present
	stdinBytes := bytes.NewBufferString("stdin-token-should-not-be-used\n")
	f, stdout := testFactory(t, srv, "")
	f.IOStreams.In = io.NopCloser(stdinBytes)

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--token", "ATBBXXX", "--with-token", "--username", "testuser"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Logged in as testuser") {
		t.Errorf("expected success via --token path, got: %q", stdout.String())
	}
	if stdinBytes.Len() == 0 {
		t.Error("stdin was consumed even though --token was provided")
	}
}

func TestLoginToken_DirectFlag_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "Invalid credentials"},
		}); err != nil {
			t.Errorf("encoding error response: %v", err)
		}
	}))
	defer srv.Close()

	f, _ := testFactory(t, srv, "")

	cmd := NewCmdLogin(f)
	cmd.SetArgs([]string{"--token", "ATBBXXX", "--username", "testuser"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error on 401 response")
	}
	if !strings.Contains(err.Error(), "validating credentials") {
		t.Errorf("expected 'validating credentials' error, got: %v", err)
	}
}
