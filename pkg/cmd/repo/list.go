package repo

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var limit int
	var language string
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories in the active workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := f.Workspace()
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			repos, err := client.ListRepos(workspace, api.ListReposOptions{
				Limit:    limit,
				Language: language,
			})
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, repos, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("NAME", "DESCRIPTION", "LANGUAGE", "VISIBILITY", "UPDATED")
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
				updated := r.UpdatedOn
				if len(updated) > 10 {
					updated = updated[:10]
				}
				t.AddRow(r.Slug, desc, lang, vis, updated)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d repo(s) in workspace %s\n", len(repos), workspace)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of repositories to list (0 = all)")
	cmd.Flags().StringVar(&language, "language", "", "Filter by programming language")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
