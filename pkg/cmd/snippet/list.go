package snippet

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your Bitbucket Snippets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve workspace for scoping (optional — empty lists authenticated user's snippets)
			workspace, _ := f.Workspace()

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			snippets, err := client.ListSnippets(workspace, limit)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, snippets, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			if len(snippets) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No snippets found")
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("ID", "TITLE", "FILES", "VISIBILITY", "UPDATED")
			for _, s := range snippets {
				vis := "public"
				if s.IsPrivate {
					vis = "private"
				}
				updated := s.UpdatedOn
				if len(updated) > 10 {
					updated = updated[:10]
				}
				t.AddRow(s.ID, s.Title, fmt.Sprintf("%d", s.FileCount()), vis, updated)
			}
			return t.Render()
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of snippets to list")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
