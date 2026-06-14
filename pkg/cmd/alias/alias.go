package aliascmd

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdAlias returns the `bb alias` command group.
func NewCmdAlias(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias <subcommand>",
		Short: "Manage command aliases",
		Long: `Define shortcuts for frequently used bb commands.

  bb alias set pv 'pr view'
  bb alias list
  bb alias delete pv`,
	}
	cmd.AddCommand(newCmdSet(f))
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdDelete(f))
	return cmd
}