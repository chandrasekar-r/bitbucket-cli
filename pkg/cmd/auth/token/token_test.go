package token

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

func TestToken_NotLoggedIn(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	f, _ := makeFactory(t)
	cmd := NewCmdToken(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when not authenticated, got nil")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("expected 'not authenticated' error, got: %v", err)
	}
}

func TestToken_PrintsToken(t *testing.T) {
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
	cmd := NewCmdToken(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != "tok123" {
		t.Errorf("expected access token 'tok123', got: %q", got)
	}
}
