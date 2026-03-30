package issue

import (
	"fmt"
	"strconv"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view <number>",
		Short: "Show issue details",
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
			i, err := client.GetIssue(workspace, slug, id)
			if err != nil {
				return err
			}
			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, i, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			w := f.IOStreams.Out
			assignee := "-"
			if i.Assignee != nil {
				assignee = i.Assignee.Username
			}
			fmt.Fprintf(w, "Issue #%d: %s\n", i.ID, i.Title)
			fmt.Fprintf(w, "State:    %s\n", i.State)
			fmt.Fprintf(w, "Kind:     %s\n", i.Kind)
			fmt.Fprintf(w, "Priority: %s\n", i.Priority)
			fmt.Fprintf(w, "Reporter: %s\n", i.Reporter.Username)
			fmt.Fprintf(w, "Assignee: %s\n", assignee)
			fmt.Fprintf(w, "Comments: %d\n", i.CommentCount)
			fmt.Fprintf(w, "URL:      %s\n", i.Links.HTML.Href)
			if i.Content.Raw != "" {
				fmt.Fprintf(w, "\n%s\n", i.Content.Raw)
			}
			return nil
		},
	}
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
