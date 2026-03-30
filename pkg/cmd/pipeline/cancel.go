package pipeline

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdCancel(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <uuid>",
		Short: "Cancel a running pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			if err := client.StopPipeline(workspace, slug, uuid); err != nil {
				return fmt.Errorf("cancelling pipeline: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Pipeline %s cancelled\n", uuid)
			return nil
		},
	}
	return cmd
}
