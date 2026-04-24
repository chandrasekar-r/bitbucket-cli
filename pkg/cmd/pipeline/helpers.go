package pipeline

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
)

// repoCtx resolves workspace and repo slug from git context.
func repoCtx(f *cmdutil.Factory) (workspace, slug string, err error) {
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return ctx.Workspace, ctx.RepoSlug, nil
	}
	ws, werr := f.Workspace()
	if werr != nil {
		return "", "", werr
	}
	return ws, "", fmt.Errorf("no repo found: run from inside a cloned Bitbucket repository")
}

// statusColor returns an ANSI-colored status string when color is enabled.
func statusColor(state, result string, color bool) string {
	text := state
	if result != "" {
		text = result
	}
	if !color {
		return text
	}
	switch strings.ToUpper(text) {
	case "SUCCESSFUL":
		return "\033[32m" + text + "\033[0m" // green
	case "FAILED", "ERROR":
		return "\033[31m" + text + "\033[0m" // red
	case "STOPPED":
		return "\033[33m" + text + "\033[0m" // yellow
	case "IN_PROGRESS":
		return "\033[36m" + text + "\033[0m" // cyan
	default:
		return text
	}
}

// formatDuration formats seconds into a human-readable string like "2m30s".
func formatDuration(seconds int) string {
	if seconds == 0 {
		return "-"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dm%ds", seconds/60, seconds%60)
}
