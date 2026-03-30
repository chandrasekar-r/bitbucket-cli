package pr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
)

// repoContext resolves the workspace and repo slug for the current context.
func repoContext(f *cmdutil.Factory) (workspace, slug string, err error) {
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return ctx.Workspace, ctx.RepoSlug, nil
	}
	ws, werr := f.Workspace()
	if werr != nil {
		return "", "", werr
	}
	return ws, "", fmt.Errorf("no repo found: run from inside a cloned Bitbucket repository")
}

// parsePRID parses a PR number from a string argument.
func parsePRID(s string) (int, error) {
	s = strings.TrimPrefix(s, "#")
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid PR number %q", s)
	}
	return id, nil
}
