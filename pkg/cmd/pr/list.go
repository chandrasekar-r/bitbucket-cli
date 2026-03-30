package pr

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var state string
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pull requests",
		Args:  cobra.NoArgs,
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
			prs, err := client.ListPRs(workspace, slug, api.ListPRsOptions{
				State: state,
				Limit: limit,
			})
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, prs, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			if len(prs) == 0 {
				fmt.Fprintf(f.IOStreams.Out, "No %s pull requests in %s/%s\n",
					strings.ToLower(state), workspace, slug)
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("#", "TITLE", "AUTHOR", "FROM → TO", "APPROVALS")
			for _, pr := range prs {
				title := pr.Title
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				from := pr.Source.Branch.Name
				to := pr.Destination.Branch.Name
				approvals := fmt.Sprintf("%d", pr.ApprovalCount())
				t.AddRow(fmt.Sprintf("%d", pr.ID), title, pr.Author.Username,
					from+" → "+to, approvals)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d pull request(s)\n", len(prs))
			return nil
		},
	}

	cmd.Flags().StringVar(&state, "state", "OPEN", "Filter by state: OPEN, MERGED, DECLINED")
	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of PRs to list (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
