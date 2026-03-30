// Package issue provides the `bb issue` command group.
package issue

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

// NewCmdIssue returns the `bb issue` group command.
func NewCmdIssue(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue <subcommand>",
		Short: "Manage repository issues",
		Long: `Work with Bitbucket Cloud issues.

  bb issue list              List open issues
  bb issue view <number>     Show issue details
  bb issue create            Create a new issue
  bb issue close <number>    Close an issue
  bb issue reopen <number>   Reopen a closed issue
  bb issue comment <number>  Add a comment to an issue

Note: Issues must be enabled in Repository Settings → Features → Issues.`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdClose(f))
	cmd.AddCommand(newCmdReopen(f))
	cmd.AddCommand(newCmdComment(f))
	return cmd
}

// repoCtx resolves workspace and repo slug.
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

// guardIssues verifies the repository has issues enabled.
// Returns a clear, actionable error if issues are disabled.
func guardIssues(client *api.Client, workspace, slug string) error {
	enabled, err := client.HasIssues(workspace, slug)
	if err != nil {
		return err
	}
	if !enabled {
		return errors.New(
			"issues are not enabled for this repository\n" +
				"enable them under Repository Settings → Features → Issues",
		)
	}
	return nil
}

// issueClient creates an API client and verifies issues are enabled.
// This is the entry point for every issue command — ensures the guard runs exactly once.
func issueClient(f *cmdutil.Factory, workspace, slug string) (*api.Client, error) {
	httpClient, err := f.HttpClient()
	if err != nil {
		return nil, err
	}
	client := api.New(httpClient, f.BaseURL)
	if err := guardIssues(client, workspace, slug); err != nil {
		return nil, err
	}
	return client, nil
}
