// Package runner provides the `bb runner` command group.
package runner

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdRunner returns the `bb runner` group command.
func NewCmdRunner(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "runner <subcommand>",
		Short:   "Manage self-hosted Bitbucket Pipelines runners",
		Aliases: []string{"runners"},
		Long: `Work with Bitbucket Cloud self-hosted Pipelines runners.

Commands target the current workspace by default. Pass --repo to manage
runners scoped to a single repository.

  bb runner list                          List runners
  bb runner view <uuid>                   Show one runner
  bb runner create --name --label k=v     Register a new runner (prints one-time credentials)
  bb runner disable <uuid>                Disable a runner
  bb runner enable <uuid>                 Re-enable a disabled runner
  bb runner delete <uuid>                 Delete a runner`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdDisable(f))
	cmd.AddCommand(newCmdEnable(f))
	cmd.AddCommand(newCmdDelete(f))
	return cmd
}
