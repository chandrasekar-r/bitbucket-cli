// Package workspace provides the `bb workspace` command group.
package workspace

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdWorkspace returns the `bb workspace` group command.
func NewCmdWorkspace(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace <subcommand>",
		Short: "Manage Bitbucket workspaces",
		Long: `Work with Bitbucket Cloud workspaces.

  bb workspace list           List your accessible workspaces
  bb workspace use <slug>     Set the default workspace
  bb workspace view [slug]    Show workspace details`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdUse(f))
	cmd.AddCommand(newCmdView(f))
	return cmd
}
