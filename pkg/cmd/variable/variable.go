package variablecmd

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdVariable returns the `bb variable` command group.
func NewCmdVariable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable <subcommand>",
		Short: "Manage Bitbucket Pipelines variables",
		Long: `Manage pipeline variables at workspace, repository, or deployment scopes.

  bb variable list [--workspace | --env NAME]
  bb variable set KEY VALUE [--secured] [--workspace | --env NAME]
  bb variable delete KEY [--workspace | --env NAME]
  bb variable env list
  bb variable env create NAME`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdSet(f))
	cmd.AddCommand(newCmdDelete(f))
	cmd.AddCommand(newCmdEnv(f))
	return cmd
}

func newCmdEnv(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env <subcommand>",
		Short: "Manage deployment environments",
	}
	cmd.AddCommand(newCmdEnvList(f))
	cmd.AddCommand(newCmdEnvCreate(f))
	return cmd
}