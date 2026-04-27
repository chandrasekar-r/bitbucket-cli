package status

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

func makeFactory(t *testing.T) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")),
		Out:    out,
		ErrOut: &bytes.Buffer{},
	}
	return &cmdutil.Factory{IOStreams: ios}, out
}

func populateTokenStore(t *testing.T, dir string, accounts map[string]interface{}, activeAccount string) {
	t.Helper()
	tokPath := filepath.Join(dir, "bb", "tokens.json")
	if err := os.MkdirAll(filepath.Dir(tokPath), 0700); err != nil {
		t.Fatalf("creating token store dir: %v", err)
	}
	data, err := json.Marshal(map[string]interface{}{
		"active_account": activeAccount,
		"accounts":       accounts,
	})
	if err != nil {
		t.Fatalf("marshaling token store: %v", err)
	}
	if err := os.WriteFile(tokPath, data, 0600); err != nil {
		t.Fatalf("writing token store: %v", err)
	}
}

func TestStatus_NotLoggedIn(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	f, _ := makeFactory(t)
	cmd := NewCmdStatus(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when not logged in, got nil")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in' error, got: %v", err)
	}
}

func TestStatus_LoggedIn(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	populateTokenStore(t, dir, map[string]interface{}{
		"alice": map[string]interface{}{
			"username":     "alice",
			"access_token": "tok123",
			"auth_type":    "token",
		},
	}, "alice")

	f, out := makeFactory(t)
	cmd := NewCmdStatus(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "alice") {
		t.Errorf("expected output to contain username 'alice', got: %q", got)
	}
	if !strings.Contains(got, "Logged in to bitbucket.org") {
		t.Errorf("expected 'Logged in to bitbucket.org' in output, got: %q", got)
	}
}

func TestStatus_ActiveMarker(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	populateTokenStore(t, dir, map[string]interface{}{
		"alice": map[string]interface{}{
			"username":     "alice",
			"access_token": "tok_alice",
			"auth_type":    "token",
		},
		"bob": map[string]interface{}{
			"username":     "bob",
			"access_token": "tok_bob",
			"auth_type":    "token",
		},
	}, "alice")

	f, out := makeFactory(t)
	cmd := NewCmdStatus(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "(active)") {
		t.Errorf("expected '(active)' marker in output for multi-account status, got: %q", got)
	}
}
