package pr

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdUnapprove(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unapprove <number>",
		Short: "Remove your approval from a pull request",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			id, err := resolvePRID(f, client, workspace, slug, args)
			if err != nil {
				return err
			}
			if err := client.RemoveApproval(workspace, slug, id); err != nil {
				return fmt.Errorf("removing approval: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Removed approval from PR #%d\n", id)
			return nil
		},
	}
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}