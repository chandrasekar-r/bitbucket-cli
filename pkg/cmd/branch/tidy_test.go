package branch

import (
	"testing"
)

func TestNewCmdTidy(t *testing.T) {
	cmd := newCmdTidy(nil)

	if cmd.Use != "tidy" {
		t.Errorf("Use: got %q, want %q", cmd.Use, "tidy")
	}

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("expected --dry-run flag")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("--dry-run default: got %q, want %q", dryRunFlag.DefValue, "false")
	}

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("expected --force flag")
	}
	if forceFlag.DefValue != "false" {
		t.Errorf("--force default: got %q, want %q", forceFlag.DefValue, "false")
	}
}

func TestFilterByState(t *testing.T) {
	candidates := []tidyCandidate{
		{Branch: "feat/a", State: "MERGED", PRID: 1},
		{Branch: "feat/b", State: "DECLINED", PRID: 2},
		{Branch: "feat/c", State: "MERGED", PRID: 3},
	}

	merged := filterByState(candidates, "MERGED")
	if len(merged) != 2 {
		t.Errorf("expected 2 MERGED, got %d", len(merged))
	}

	declined := filterByState(candidates, "DECLINED")
	if len(declined) != 1 {
		t.Errorf("expected 1 DECLINED, got %d", len(declined))
	}

	open := filterByState(candidates, "OPEN")
	if len(open) != 0 {
		t.Errorf("expected 0 OPEN, got %d", len(open))
	}
}
