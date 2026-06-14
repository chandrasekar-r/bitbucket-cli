package pr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

const requestChangesPrefix = "[request-changes] "

func newCmdReview(f *cmdutil.Factory) *cobra.Command {
	var approve bool
	var requestChanges bool
	var body string

	cmd := &cobra.Command{
		Use:   "review <number>",
		Short: "Approve or request changes on a pull request",
		Long: `Submit a review on a pull request.

  bb pr review 42 --approve
  bb pr review 42 --request-changes --body "Please add tests"

Request-changes removes any existing approval and posts a comment prefixed with [request-changes].`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (approve && requestChanges) || (!approve && !requestChanges) {
				return &cmdutil.FlagError{Err: errors.New("exactly one of --approve or --request-changes is required")}
			}

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

			if approve {
				if err := client.ApprovePR(workspace, slug, id); err != nil {
					return fmt.Errorf("approving PR: %w", err)
				}
				fmt.Fprintf(f.IOStreams.Out, "✓ Approved PR #%d\n", id)
				return nil
			}

			if err := client.RemoveApproval(workspace, slug, id); err != nil {
				return fmt.Errorf("removing approval: %w", err)
			}

			if body == "" {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.FlagError{Err: errors.New("--body is required with --request-changes in non-interactive mode")}
				}
				body, err = openEditor()
				if err != nil {
					return err
				}
			}
			body = strings.TrimSpace(body)
			if body == "" {
				return errors.New("review body cannot be empty for request-changes")
			}
			if !strings.HasPrefix(body, requestChangesPrefix) {
				body = requestChangesPrefix + body
			}
			if err := client.AddPRComment(workspace, slug, id, body); err != nil {
				return fmt.Errorf("posting review comment: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Requested changes on PR #%d\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&approve, "approve", false, "Approve the pull request")
	cmd.Flags().BoolVar(&requestChanges, "request-changes", false, "Request changes (removes approval and posts a comment)")
	cmd.Flags().StringVar(&body, "body", "", "Review comment body (required for --request-changes in --no-tty mode)")
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}