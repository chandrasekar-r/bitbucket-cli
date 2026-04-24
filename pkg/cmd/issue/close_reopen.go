package issue

import (
	"fmt"
	"strconv"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdClose(f *cmdutil.Factory) *cobra.Command {
	var status string
	cmd := &cobra.Command{
		Use:   "close <number>",
		Short: "Close an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return &cmdutil.FlagError{Err: fmt.Errorf("invalid issue number %q", args[0])}
			}
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			client, err := issueClient(f, workspace, slug)
			if err != nil {
				return err
			}
			if status == "" {
				status = "resolved"
			}
			if _, err := client.UpdateIssueStatus(workspace, slug, id, status); err != nil {
				return fmt.Errorf("closing issue: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Issue #%d marked as %s\n", id, status)
			return nil
		},
	}
	cmd.Flags().StringVar(&status, "status", "resolved",
		"Close status: resolved, invalid, duplicate, wontfix, closed")
	return cmd
}

func newCmdReopen(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reopen <number>",
		Short: "Reopen a closed issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return &cmdutil.FlagError{Err: fmt.Errorf("invalid issue number %q", args[0])}
			}
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			client, err := issueClient(f, workspace, slug)
			if err != nil {
				return err
			}
			if _, err := client.UpdateIssueStatus(workspace, slug, id, "open"); err != nil {
				return fmt.Errorf("reopening issue: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Issue #%d reopened\n", id)
			return nil
		},
	}
	return cmd
}
