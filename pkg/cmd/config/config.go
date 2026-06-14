package configcmd

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdConfig returns the `bb config` command group.
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <subcommand>",
		Short: "Manage bb CLI configuration",
		Long: `View and edit bb CLI settings stored in the config file.

  bb config get <key>     Print a config value
  bb config set <key> <value>  Set a config value
  bb config list          List all config values
  bb config edit          Open config file in $EDITOR`,
	}
	cmd.AddCommand(newCmdGet(f))
	cmd.AddCommand(newCmdSet(f))
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdEdit(f))
	return cmd
}