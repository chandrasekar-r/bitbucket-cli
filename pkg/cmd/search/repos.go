package search

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdRepos(f *cmdutil.Factory) *cobra.Command {
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "repos <query>",
		Short: "Search repositories by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			if query == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("search query cannot be empty")}
			}

			workspace, err := f.Workspace()
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			repos, err := client.SearchRepos(workspace, query, limit)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, repos, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			if len(repos) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No repositories found")
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("NAME", "DESCRIPTION", "LANGUAGE", "VISIBILITY")
			for _, r := range repos {
				vis := "public"
				if r.IsPrivate {
					vis = "private"
				}
				desc := r.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				lang := r.Language
				if lang == "" {
					lang = "-"
				}
				t.AddRow(r.Slug, desc, lang, vis)
			}
			return t.Render()
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 20, "Maximum number of repos to return (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}