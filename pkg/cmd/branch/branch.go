// Package branch provides the `bb branch` command group.
package branch

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdBranch returns the `bb branch` group command.
func NewCmdBranch(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch <subcommand>",
		Short: "Manage repository branches",
		Long: `Work with branches in a Bitbucket Cloud repository.

  bb branch list              List branches
  bb branch create <name>     Create a new branch
  bb branch delete <name>     Delete a branch
  bb branch rename <old> <new> Rename a branch
  bb branch protect <pattern> Enable branch restrictions`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdDelete(f))
	cmd.AddCommand(newCmdRename(f))
	cmd.AddCommand(newCmdProtect(f))
	return cmd
}
