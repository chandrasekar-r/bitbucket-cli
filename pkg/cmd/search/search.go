// Package search provides the `bb search` command group.
package search

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdSearch returns the `bb search` group command.
func NewCmdSearch(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <subcommand>",
		Short: "Search code and repositories",
		Long: `Search Bitbucket Cloud workspaces for code or repositories.

  bb search code <query>    Search source code in the workspace
  bb search repos <query>   Search repositories by name`,
	}
	cmd.AddCommand(newCmdCode(f))
	cmd.AddCommand(newCmdRepos(f))
	return cmd
}