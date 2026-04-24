// Package webhook provides the `bb webhook` command group.
package webhook

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdWebhook returns the `bb webhook` group command.
func NewCmdWebhook(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhook <subcommand>",
		Short:   "Manage repository and workspace webhooks",
		Aliases: []string{"webhooks", "hook", "hooks"},
		Long: `Work with Bitbucket Cloud webhooks.

By default commands target the current repository (inferred from the git
remote). Pass --workspace-only to target the current workspace instead.

  bb webhook list                           List webhooks
  bb webhook view <uuid>                    Show one webhook
  bb webhook create --url URL --event E     Create a webhook
  bb webhook update <uuid> [flags]          Update a webhook
  bb webhook delete <uuid>                  Delete a webhook`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdUpdate(f))
	cmd.AddCommand(newCmdDelete(f))
	return cmd
}
