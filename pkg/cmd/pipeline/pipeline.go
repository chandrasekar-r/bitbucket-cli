// Package pipeline provides the `bb pipeline` command group.
package pipeline

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdPipeline returns the `bb pipeline` group command.
func NewCmdPipeline(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline <subcommand>",
		Short: "Manage Bitbucket Pipelines CI/CD",
		Long: `Work with Bitbucket Cloud Pipelines.

  bb pipeline list              List recent pipeline runs
  bb pipeline view <uuid>       Show pipeline details and step status
  bb pipeline run               Trigger a new pipeline run
  bb pipeline cancel <uuid>     Cancel a running pipeline
  bb pipeline watch [uuid]      Stream pipeline logs (polling)`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdRun(f))
	cmd.AddCommand(newCmdCancel(f))
	cmd.AddCommand(newCmdWatch(f))
	return cmd
}
