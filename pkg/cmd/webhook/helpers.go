package webhook

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

// scope is the resolved target for a webhook command.
type scope struct {
	workspace   string
	repoSlug    string // empty when workspaceOnly is true
	workspaceOn bool
}

// addScopeFlag registers the --workspace-only flag on cmd.
func addScopeFlag(cmd *cobra.Command, workspaceOnly *bool) {
	cmd.Flags().BoolVar(workspaceOnly, "workspace-only", false,
		"Target the current workspace instead of the current repository")
}

// resolveScope picks between repo and workspace scope.
//
//   - workspaceOnly=false (default): use git remote for workspace + repo slug;
//     fall back to f.Workspace() for the workspace only when no remote is found,
//     but require a repo slug, so error if none can be inferred.
//   - workspaceOnly=true: use f.Workspace() for the workspace only.
func resolveScope(f *cmdutil.Factory, workspaceOnly bool) (scope, error) {
	if workspaceOnly {
		ws, err := f.Workspace()
		if err != nil {
			return scope{}, err
		}
		return scope{workspace: ws, workspaceOn: true}, nil
	}
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return scope{workspace: ctx.Workspace, repoSlug: ctx.RepoSlug}, nil
	}
	return scope{}, fmt.Errorf(
		"no repo found: run from inside a cloned Bitbucket repo, or pass --workspace-only")
}

// parseEventFlags joins repeated --event flags and splits any comma-separated
// values inside each entry. Returns the de-duplicated list in input order.
func parseEventFlags(flags []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, f := range flags {
		for _, e := range strings.Split(f, ",") {
			e = strings.TrimSpace(e)
			if e == "" {
				continue
			}
			if _, ok := seen[e]; ok {
				continue
			}
			seen[e] = struct{}{}
			out = append(out, e)
		}
	}
	return out
}
