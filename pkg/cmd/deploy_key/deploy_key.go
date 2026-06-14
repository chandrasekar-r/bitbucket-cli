package deploykeycmd

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdDeployKey returns the `bb deploy-key` command group.
func NewCmdDeployKey(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deploy-key <subcommand>",
		Aliases: []string{"deploy-keys"},
		Short:   "Manage repository deploy keys",
		Long: `Manage SSH deploy keys for the current repository.

  bb deploy-key list
  bb deploy-key add --key-file ~/.ssh/deploy_key.pub --label "CI"
  bb deploy-key view <id>
  bb deploy-key delete <id>`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdAdd(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdDelete(f))
	return cmd
}