package pr

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdCommentList(f *cmdutil.Factory) *cobra.Command {
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list <number>",
		Short: "List comments on a pull request",
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

			comments, err := client.ListPRComments(workspace, slug, id, limit)
			if err != nil {
				return fmt.Errorf("listing comments: %w", err)
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, comments, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			if len(comments) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No comments found")
				return nil
			}

			w := f.IOStreams.Out
			for _, cmt := range comments {
				prefix := ""
				if cmt.Parent != nil {
					prefix = "  ↳ "
				}
				author := cmt.User.Username
				if author == "" {
					author = cmt.User.DisplayName
				}
				line := fmt.Sprintf("%s#%d %s (%s)", prefix, cmt.ID, author, cmt.CreatedOn)
				if cmt.Inline != nil {
					if cmt.Inline.To > 0 {
						line += fmt.Sprintf(" [%s:%d]", cmt.Inline.Path, cmt.Inline.To)
					} else {
						line += fmt.Sprintf(" [%s]", cmt.Inline.Path)
					}
				}
				fmt.Fprintln(w, line)
				body := cmt.Content.Raw
				if len(body) > 120 {
					body = body[:117] + "..."
				}
				for _, textLine := range splitLines(body) {
					fmt.Fprintf(w, "%s  %s\n", prefix, textLine)
				}
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d comment(s)\n", len(comments))
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 50, "Maximum number of comments to list (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}