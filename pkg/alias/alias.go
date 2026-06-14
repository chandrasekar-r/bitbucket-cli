package alias

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
)

// KnownCommands is the allowlist of top-level bb subcommands aliases may expand to.
var KnownCommands = []string{
	"api", "auth", "batch", "branch", "browse", "completion", "config",
	"deploy-key", "extension", "ext", "issue", "pipeline", "pr", "project",
	"repo", "runner", "search", "snippet", "ssh-key", "status", "variable",
	"version", "webhook", "workspace",
}

// ValidateExpansion ensures an alias expansion starts with a known bb subcommand.
func ValidateExpansion(expansion string) error {
	expansion = strings.TrimSpace(expansion)
	if expansion == "" {
		return fmt.Errorf("alias expansion cannot be empty")
	}
	first := strings.Fields(expansion)[0]
	for _, cmd := range KnownCommands {
		if first == cmd {
			return nil
		}
	}
	return fmt.Errorf("alias expansion must start with a known bb command (%s)", strings.Join(KnownCommands, ", "))
}

// Expand rewrites args when the first token matches a configured alias.
func Expand(args []string, cfg *config.Config) []string {
	if len(args) == 0 || cfg == nil {
		return args
	}
	aliases := cfg.Aliases()
	exp, ok := aliases[args[0]]
	if !ok {
		return args
	}
	expanded := strings.Fields(exp)
	return append(expanded, args[1:]...)
}