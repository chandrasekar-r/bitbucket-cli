package pr

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view <number>",
		Short: "Show pull request details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parsePRID(args[0])
			if err != nil {
				return err
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
			pr, err := client.GetPR(workspace, slug, id)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, pr, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			w := f.IOStreams.Out
			fmt.Fprintf(w, "PR #%d: %s\n", pr.ID, pr.Title)
			fmt.Fprintf(w, "State:    %s\n", pr.State)
			fmt.Fprintf(w, "Author:   %s\n", pr.Author.Username)
			fmt.Fprintf(w, "Branch:   %s → %s\n",
				pr.Source.Branch.Name, pr.Destination.Branch.Name)
			fmt.Fprintf(w, "Approvals: %d/%d reviewer(s)\n",
				pr.ApprovalCount(), len(pr.Reviewers))
			fmt.Fprintf(w, "Comments: %d\n", pr.CommentCount)
			fmt.Fprintf(w, "URL:      %s\n", pr.Links.HTML.Href)

			if pr.Description != "" {
				fmt.Fprintln(w)
				fmt.Fprintln(w, strings.TrimSpace(pr.Description))
			}
			return nil
		},
	}

	jsonOpts = cmdutil.AddJSONFlags(cmd)
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}
