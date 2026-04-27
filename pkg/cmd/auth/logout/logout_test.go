package logout

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
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

func TestLogout_NotLoggedIn(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	f, _ := makeFactory(t)
	cmd := NewCmdLogout(f)
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

func TestLogout_Success(t *testing.T) {
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
	cmd := NewCmdLogout(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(out.String(), "alice") {
		t.Errorf("expected output to contain 'alice', got: %q", out.String())
	}
	if !strings.Contains(out.String(), "Logged out") {
		t.Errorf("expected 'Logged out' in output, got: %q", out.String())
	}
}

func TestLogout_ClearsCredentials(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	populateTokenStore(t, dir, map[string]interface{}{
		"alice": map[string]interface{}{
			"username":     "alice",
			"access_token": "tok123",
			"auth_type":    "token",
		},
	}, "alice")

	f, _ := makeFactory(t)
	cmd := NewCmdLogout(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify store is empty after logout
	store := bbauth.NewTokenStore()
	acc, err := store.GetActive()
	if err != nil {
		t.Fatalf("loading store after logout: %v", err)
	}
	if acc != nil {
		t.Errorf("expected no active account after logout, got: %v", acc.Username)
	}

	usernames, _, err := store.ListAccounts()
	if err != nil {
		t.Fatalf("listing accounts after logout: %v", err)
	}
	if len(usernames) != 0 {
		t.Errorf("expected no accounts after logout, got: %v", usernames)
	}
}
