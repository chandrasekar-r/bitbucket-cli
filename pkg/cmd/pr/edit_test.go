package pr

import (
	"testing"
)

func TestNewCmdEdit_RequiresArg(t *testing.T) {
	cmd := newCmdEdit(nil)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestNewCmdEdit_AcceptsNumber(t *testing.T) {
	cmd := newCmdEdit(nil)
	// Validate cobra arg parsing only; RunE will fail because factory is nil
	if err := cmd.Args(cmd, []string{"42"}); err != nil {
		t.Errorf("expected arg '42' to be accepted: %v", err)
	}
}

func TestNewCmdEdit_RejectsExtraArgs(t *testing.T) {
	cmd := newCmdEdit(nil)
	if err := cmd.Args(cmd, []string{"1", "2"}); err == nil {
		t.Error("expected error for extra args")
	}
}
