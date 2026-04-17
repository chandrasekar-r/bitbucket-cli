package runner

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

// scope is the resolved target for a runner command.
type scope struct {
	workspace string
	repoSlug  string // empty = workspace-scoped
}

func (s scope) isRepo() bool { return s.repoSlug != "" }

// addRepoFlag registers the --repo/-r flag on cmd.
func addRepoFlag(cmd *cobra.Command, repoFlag *bool) {
	cmd.Flags().BoolVarP(repoFlag, "repo", "r", false,
		"Manage runners for the current repository instead of the workspace")
}

// resolveScope picks between workspace (default) and repo (--repo).
func resolveScope(f *cmdutil.Factory, repoFlag bool) (scope, error) {
	if repoFlag {
		if ctx := gitcontext.FromRemote(); ctx != nil {
			return scope{workspace: ctx.Workspace, repoSlug: ctx.RepoSlug}, nil
		}
		return scope{}, fmt.Errorf(
			"--repo: no repo found (run from inside a cloned Bitbucket repo)")
	}
	ws, err := f.Workspace()
	if err != nil {
		return scope{}, err
	}
	return scope{workspace: ws}, nil
}

// parseLabelFlags joins repeated --label flags and splits comma-separated
// values. Values may be bare tokens (e.g. "linux") or key=value pairs
// (e.g. "os=linux"); Bitbucket's runner API accepts either — we pass through.
func parseLabelFlags(flags []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, f := range flags {
		for _, l := range strings.Split(f, ",") {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			if _, ok := seen[l]; ok {
				continue
			}
			seen[l] = struct{}{}
			out = append(out, l)
		}
	}
	return out
}

// listRunners dispatches to the right API call based on scope.
func listRunners(client *api.Client, sc scope) ([]api.Runner, error) {
	if sc.isRepo() {
		return client.ListRepoRunners(sc.workspace, sc.repoSlug)
	}
	return client.ListWorkspaceRunners(sc.workspace)
}

func getRunner(client *api.Client, sc scope, uid string) (*api.Runner, error) {
	if sc.isRepo() {
		return client.GetRepoRunner(sc.workspace, sc.repoSlug, uid)
	}
	return client.GetWorkspaceRunner(sc.workspace, uid)
}

func createRunner(client *api.Client, sc scope, in api.RunnerInput) (*api.Runner, error) {
	if sc.isRepo() {
		return client.CreateRepoRunner(sc.workspace, sc.repoSlug, in)
	}
	return client.CreateWorkspaceRunner(sc.workspace, in)
}

func updateRunner(client *api.Client, sc scope, uid string, up api.RunnerUpdate) (*api.Runner, error) {
	if sc.isRepo() {
		return client.UpdateRepoRunner(sc.workspace, sc.repoSlug, uid, up)
	}
	return client.UpdateWorkspaceRunner(sc.workspace, uid, up)
}

func deleteRunner(client *api.Client, sc scope, uid string) error {
	if sc.isRepo() {
		return client.DeleteRepoRunner(sc.workspace, sc.repoSlug, uid)
	}
	return client.DeleteWorkspaceRunner(sc.workspace, uid)
}
