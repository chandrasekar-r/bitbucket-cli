package alias_test

import (
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/alias"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
)

func TestValidateExpansion(t *testing.T) {
	if err := alias.ValidateExpansion("pr list"); err != nil {
		t.Fatalf("expected valid expansion: %v", err)
	}
	if err := alias.ValidateExpansion("unknown list"); err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestExpandWithAliases(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := cfg.SetAlias("pv", "pr view"); err != nil {
		t.Fatalf("SetAlias: %v", err)
	}
	got := alias.Expand([]string{"pv", "42"}, cfg)
	want := []string{"pr", "view", "42"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}