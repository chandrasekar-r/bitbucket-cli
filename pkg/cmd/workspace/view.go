package workspace

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view [slug]",
		Short: "Show workspace details",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var slug string
			if len(args) == 1 {
				slug = args[0]
			} else {
				var err error
				slug, err = f.Workspace()
				if err != nil {
					return err
				}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			ws, err := client.GetWorkspace(slug)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, ws, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			fmt.Fprintf(f.IOStreams.Out, "Workspace: %s (%s)\n", ws.Name, ws.Slug)
			fmt.Fprintf(f.IOStreams.Out, "Type:      %s\n", ws.Type)
			fmt.Fprintf(f.IOStreams.Out, "UUID:      %s\n", ws.UUID)
			fmt.Fprintf(f.IOStreams.Out, "URL:       https://bitbucket.org/%s\n", ws.Slug)
			return nil
		},
	}

	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
