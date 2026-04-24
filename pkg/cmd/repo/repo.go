// Package repo provides the `bb repo` command group.
package repo

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdRepo returns the `bb repo` group command with all subcommands.
func NewCmdRepo(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo <subcommand>",
		Short: "Manage Bitbucket repositories",
		Long: `Work with Bitbucket Cloud repositories.

  bb repo list              List repositories in the current workspace
  bb repo view [repo]       Show repository details
  bb repo create            Create a new repository
  bb repo clone [repo]      Clone a repository locally
  bb repo fork [repo]       Fork a repository
  bb repo delete [repo]     Delete a repository
  bb repo rename [repo]     Rename a repository
  bb repo browse [repo]     Open repository in browser`,
	}

	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdClone(f))
	cmd.AddCommand(newCmdFork(f))
	cmd.AddCommand(newCmdDelete(f))
	cmd.AddCommand(newCmdRename(f))
	cmd.AddCommand(newCmdBrowse(f))
	return cmd
}
