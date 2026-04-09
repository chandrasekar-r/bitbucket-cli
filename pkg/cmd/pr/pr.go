// Package pr provides the `bb pr` command group.
package pr

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdPR returns the `bb pr` group command.
func NewCmdPR(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr <subcommand>",
		Short: "Manage pull requests",
		Long: `Work with Bitbucket Cloud pull requests.

  bb pr list                List open pull requests
  bb pr view <number>       Show pull request details
  bb pr create              Open a new pull request
  bb pr merge <number>      Merge a pull request
  bb pr approve <number>    Approve a pull request
  bb pr decline <number>    Decline a pull request
  bb pr checkout <number>   Check out a PR's source branch
  bb pr comment <number>    Add a comment to a pull request
  bb pr diff <number>       Show the pull request diff
  bb pr edit <number>       Edit a pull request
  bb pr browse <number>     Open pull request in browser`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdEdit(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdMerge(f))
	cmd.AddCommand(newCmdApprove(f))
	cmd.AddCommand(newCmdDecline(f))
	cmd.AddCommand(newCmdCheckout(f))
	cmd.AddCommand(newCmdComment(f))
	cmd.AddCommand(newCmdDiff(f))
	cmd.AddCommand(newCmdBrowse(f))
	return cmd
}
