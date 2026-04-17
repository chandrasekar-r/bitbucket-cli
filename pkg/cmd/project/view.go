package project

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
		Use:   "view <key>",
		Short: "Show a single project",
		Args:  cobra.ExactArgs(1),
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

			p, err := client.GetProject(workspace, args[0])
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, p, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			fmt.Fprintf(f.IOStreams.Out, "Key:         %s\n", p.Key)
			fmt.Fprintf(f.IOStreams.Out, "Name:        %s\n", p.Name)
			fmt.Fprintf(f.IOStreams.Out, "Private:     %t\n", p.IsPrivate)
			fmt.Fprintf(f.IOStreams.Out, "Description: %s\n", p.Description)
			fmt.Fprintf(f.IOStreams.Out, "Created:     %s\n", p.CreatedOn)
			fmt.Fprintf(f.IOStreams.Out, "URL:         %s\n", p.Links.HTML.Href)
			return nil
		},
	}
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
