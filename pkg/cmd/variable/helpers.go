package variablecmd

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

type scope struct {
	workspace    string
	repoSlug     string
	envName      string
	envUUID      string
	workspaceOnly bool
}

func addScopeFlags(cmd *cobra.Command, workspaceOnly *bool, envName *string) {
	cmd.Flags().BoolVar(workspaceOnly, "workspace", false, "Target workspace variables instead of repository variables")
	cmd.Flags().StringVar(envName, "env", "", "Target deployment environment variables by environment name")
}

func resolveScope(f *cmdutil.Factory, workspaceOnly bool, envName string) (scope, error) {
	if workspaceOnly && envName != "" {
		return scope{}, fmt.Errorf("--workspace and --env are mutually exclusive")
	}
	if workspaceOnly {
		ws, err := f.Workspace()
		if err != nil {
			return scope{}, err
		}
		return scope{workspace: ws, workspaceOnly: true}, nil
	}
	if envName != "" {
		if ctx := gitcontext.FromRemote(); ctx != nil {
			return scope{workspace: ctx.Workspace, repoSlug: ctx.RepoSlug, envName: envName}, nil
		}
		return scope{}, fmt.Errorf("--env requires a cloned Bitbucket repository")
	}
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return scope{workspace: ctx.Workspace, repoSlug: ctx.RepoSlug}, nil
	}
	return scope{}, fmt.Errorf("no repo found: run from inside a cloned Bitbucket repository, or pass --workspace")
}

func resolveEnvUUID(client *api.Client, sc scope) (string, error) {
	if sc.envName == "" {
		return "", nil
	}
	envs, err := client.ListDeploymentEnvironments(sc.workspace, sc.repoSlug, 0)
	if err != nil {
		return "", err
	}
	for _, e := range envs {
		if strings.EqualFold(e.Name, sc.envName) || strings.EqualFold(e.Slug, sc.envName) {
			return e.UUID, nil
		}
	}
	return "", fmt.Errorf("deployment environment %q not found", sc.envName)
}

func displayValue(v api.PipelineVariable) string {
	if v.Secured {
		return "[secured]"
	}
	return v.Value
}