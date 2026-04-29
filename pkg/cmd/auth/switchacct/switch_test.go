package switchacct

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

func TestSwitch_NotLoggedIn(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	f, _ := makeFactory(t)
	cmd := NewCmdSwitch(f)
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

func TestSwitch_ListAccounts(t *testing.T) {
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
	cmd := NewCmdSwitch(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	// No args → list mode

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error listing accounts, got: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "alice") {
		t.Errorf("expected 'alice' in account list, got: %q", got)
	}
	if !strings.Contains(got, "bob") {
		t.Errorf("expected 'bob' in account list, got: %q", got)
	}
	if !strings.Contains(got, "*") {
		t.Errorf("expected active marker '*' in account list, got: %q", got)
	}
}

func TestSwitch_AlreadyActive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	populateTokenStore(t, dir, map[string]interface{}{
		"alice": map[string]interface{}{
			"username":     "alice",
			"access_token": "tok_alice",
			"auth_type":    "token",
		},
	}, "alice")

	f, out := makeFactory(t)
	cmd := NewCmdSwitch(f)
	cmd.SetArgs([]string{"alice"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Already active") {
		t.Errorf("expected 'Already active' message, got: %q", got)
	}
}

func TestSwitch_SwitchSuccess(t *testing.T) {
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
	cmd := NewCmdSwitch(f)
	cmd.SetArgs([]string{"bob"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error switching to bob, got: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Active account") {
		t.Errorf("expected 'Active account' in output, got: %q", got)
	}
	if !strings.Contains(got, "bob") {
		t.Errorf("expected 'bob' in output after switch, got: %q", got)
	}
}

func TestSwitch_UnknownAccount(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	populateTokenStore(t, dir, map[string]interface{}{
		"alice": map[string]interface{}{
			"username":     "alice",
			"access_token": "tok_alice",
			"auth_type":    "token",
		},
	}, "alice")

	f, _ := makeFactory(t)
	cmd := NewCmdSwitch(f)
	cmd.SetArgs([]string{"nobody"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error switching to unknown account, got nil")
	}
	if !strings.Contains(err.Error(), "nobody") {
		t.Errorf("expected error to mention 'nobody', got: %v", err)
	}
}
