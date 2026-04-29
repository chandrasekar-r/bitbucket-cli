package extension

import (
	"testing"

	"github.com/spf13/cobra"
)

// findSubcmd locates a named subcommand among NewCmdExtension's children.
func findSubcmd(t *testing.T, name string) *cobra.Command {
	t.Helper()
	root := NewCmdExtension(nil)
	for _, sub := range root.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	t.Fatalf("subcommand %q not found", name)
	return nil
}

// checkArgs invokes the cobra Args validator function stored on cmd.
// cobra's Args field is a PositionalArgs func, not a method.
func checkArgs(cmd *cobra.Command, args []string) error {
	if cmd.Args == nil {
		return nil // no validator registered — all args accepted
	}
	return cmd.Args(cmd, args)
}

// --------------------------------------------------------------------------
// install
// --------------------------------------------------------------------------

func TestExtensionInstall_RequiresArg(t *testing.T) {
	cmd := findSubcmd(t, "install")

	// Zero args should fail cobra's ExactArgs(1) validator.
	if err := checkArgs(cmd, []string{}); err == nil {
		t.Error("expected error for zero args, got nil")
	}
}

func TestExtensionInstall_TooManyArgs(t *testing.T) {
	cmd := findSubcmd(t, "install")

	// Two args should fail cobra's ExactArgs(1) validator.
	if err := checkArgs(cmd, []string{"a", "b"}); err == nil {
		t.Error("expected error for two args, got nil")
	}
}

// --------------------------------------------------------------------------
// remove
// --------------------------------------------------------------------------

func TestExtensionRemove_RequiresArg(t *testing.T) {
	cmd := findSubcmd(t, "remove")

	if err := checkArgs(cmd, []string{}); err == nil {
		t.Error("expected error for zero args, got nil")
	}
}

// --------------------------------------------------------------------------
// list
// --------------------------------------------------------------------------

func TestExtensionList_AcceptsNoArgs(t *testing.T) {
	cmd := findSubcmd(t, "list")

	// Zero args — cobra's NoArgs validator must accept it.
	if err := checkArgs(cmd, []string{}); err != nil {
		t.Errorf("unexpected arg-validation error for zero args: %v", err)
	}

	// Extra args should be rejected by NoArgs.
	if err := checkArgs(cmd, []string{"extra"}); err == nil {
		t.Error("expected error for extra arg on list, got nil")
	}
}

// --------------------------------------------------------------------------
// top-level wiring
// --------------------------------------------------------------------------

func TestNewCmdExtension_Subcommands(t *testing.T) {
	root := NewCmdExtension(nil)

	want := map[string]bool{
		"install": false,
		"list":    false,
		"remove":  false,
	}
	for _, sub := range root.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

func TestNewCmdExtension_Aliases(t *testing.T) {
	root := NewCmdExtension(nil)
	for _, alias := range root.Aliases {
		if alias == "ext" {
			return
		}
	}
	t.Errorf("expected alias %q, got %v", "ext", root.Aliases)
}
