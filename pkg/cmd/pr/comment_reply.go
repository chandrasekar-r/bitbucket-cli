package pr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdCommentReply(f *cmdutil.Factory) *cobra.Command {
	var body string
	var parentID int

	cmd := &cobra.Command{
		Use:   "reply <number>",
		Short: "Reply to a pull request comment",
		Long: `Reply to an existing comment on a pull request.

  bb pr comment reply 42 --parent 9 --body "Thanks, fixed."`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if parentID <= 0 {
				return &cmdutil.FlagError{Err: errors.New("--parent is required and must be a positive comment ID")}
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

			if body == "" {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.FlagError{Err: errors.New("--body is required in non-interactive mode")}
				}
				body, err = openEditor()
				if err != nil {
					return err
				}
			}

			body = strings.TrimSpace(body)
			if body == "" {
				return errors.New("comment body cannot be empty")
			}

			if err := client.ReplyPRComment(workspace, slug, id, parentID, body); err != nil {
				return fmt.Errorf("replying to comment: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Replied to comment #%d on PR #%d\n", parentID, id)
			return nil
		},
	}

	cmd.Flags().IntVar(&parentID, "parent", 0, "Parent comment ID to reply to")
	cmd.Flags().StringVar(&body, "body", "", "Reply text")
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}