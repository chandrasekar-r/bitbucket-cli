package version

import (
	"bytes"
	"io"
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

func TestVersion_DefaultOutput(t *testing.T) {
	f, out := makeFactory(t)
	cmd := NewCmdVersion(f)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "bb version") {
		t.Errorf("expected 'bb version' in output, got: %q", got)
	}
}

func TestVersion_JSONFlag(t *testing.T) {
	f, out := makeFactory(t)
	cmd := NewCmdVersion(f)
	cmd.SetArgs([]string{"--json", "version"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error with --json version flag, got: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "version") {
		t.Errorf("expected 'version' key in JSON output, got: %q", got)
	}
}
