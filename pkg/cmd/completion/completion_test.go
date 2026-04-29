package completion

import (
	"bytes"
	"io"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
	"github.com/spf13/cobra"
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

// runCompletion builds a rooted command tree and executes "completion <shell>".
// Completion generation writes to os.Stdout (the real file), so we only verify
// that err is nil and do not check captured output for content.
func runCompletion(t *testing.T, shell string) error {
	t.Helper()
	f, _ := makeFactory(t)
	root := &cobra.Command{Use: "bb"}
	comp := NewCmdCompletion(f)
	root.AddCommand(comp)
	root.SetArgs([]string{"completion", shell})
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}

func TestCompletion_InvalidShell(t *testing.T) {
	f, _ := makeFactory(t)
	root := &cobra.Command{Use: "bb"}
	comp := NewCmdCompletion(f)
	root.AddCommand(comp)
	root.SetArgs([]string{"completion", "unknownshell"})
	root.SilenceUsage = true
	root.SilenceErrors = true

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid shell argument, got nil")
	}
}

func TestCompletion_BashGenerates(t *testing.T) {
	if err := runCompletion(t, "bash"); err != nil {
		t.Fatalf("expected no error generating bash completions, got: %v", err)
	}
}

func TestCompletion_ZshGenerates(t *testing.T) {
	if err := runCompletion(t, "zsh"); err != nil {
		t.Fatalf("expected no error generating zsh completions, got: %v", err)
	}
}

func TestCompletion_FishGenerates(t *testing.T) {
	if err := runCompletion(t, "fish"); err != nil {
		t.Fatalf("expected no error generating fish completions, got: %v", err)
	}
}
