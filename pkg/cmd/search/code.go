package search

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdCode(f *cmdutil.Factory) *cobra.Command {
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "code <query>",
		Short: "Search source code in the workspace",
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

			hits, err := client.SearchCode(workspace, query, limit)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, hits, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			if len(hits) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No code matches found")
				return nil
			}

			var lastPath string
			for _, h := range hits {
				path := h.Path
				if h.Repo != "" {
					path = h.Repo + "/" + h.Path
				}
				if path != lastPath {
					if lastPath != "" {
						fmt.Fprintln(f.IOStreams.Out)
					}
					fmt.Fprintln(f.IOStreams.Out, path)
					lastPath = path
				}
				fmt.Fprintf(f.IOStreams.Out, "  line %d:  %s\n", h.Line, h.Content)
			}
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 20, "Maximum number of matches to return (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}