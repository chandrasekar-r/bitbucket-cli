package pr

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdApprove(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve <number>",
		Short: "Approve a pull request",
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
			if err := client.ApprovePR(workspace, slug, id); err != nil {
				return fmt.Errorf("approving PR: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Approved PR #%d\n", id)
			return nil
		},
	}
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}

func newCmdDecline(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decline <number>",
		Short: "Decline a pull request",
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
			declined, err := client.DeclinePR(workspace, slug, id)
			if err != nil {
				return fmt.Errorf("declining PR: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Declined PR #%d: %s\n", declined.ID, declined.Title)
			return nil
		},
	}
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}
