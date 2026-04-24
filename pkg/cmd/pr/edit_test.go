package pr

import (
	"testing"
)

func TestNewCmdEdit_AcceptsZeroArgs(t *testing.T) {
	cmd := newCmdEdit(nil)
	// 0 args is valid now (interactive picker); only check cobra arg validation
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("expected 0 args to be accepted for interactive mode: %v", err)
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
