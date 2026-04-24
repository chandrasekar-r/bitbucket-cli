package issue

import (
	"fmt"
	"strconv"

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
		Short: "List issues",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			client, err := issueClient(f, workspace, slug)
			if err != nil {
				return err
			}
			issues, err := client.ListIssues(workspace, slug, api.ListIssuesOptions{
				State: state,
				Limit: limit,
			})
			if err != nil {
				return err
			}
			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, issues, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(issues) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No issues found")
				return nil
			}
			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("#", "TITLE", "STATE", "KIND", "ASSIGNEE")
			for _, i := range issues {
				assignee := "-"
				if i.Assignee != nil {
					assignee = i.Assignee.Username
				}
				title := i.Title
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				t.AddRow(strconv.Itoa(i.ID), title, i.State, i.Kind, assignee)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d issue(s)\n", len(issues))
			return nil
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Filter by state: new, open, resolved, closed (default: open+new)")
	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of issues to list (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
