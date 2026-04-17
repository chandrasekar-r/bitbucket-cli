package project

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects in the active workspace",
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

			projects, err := client.ListProjects(workspace)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, projects, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(projects) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No projects found")
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("KEY", "NAME", "PRIVATE", "DESCRIPTION")
			for _, p := range projects {
				vis := "no"
				if p.IsPrivate {
					vis = "yes"
				}
				t.AddRow(p.Key, p.Name, vis, p.Description)
			}
			return t.Render()
		},
	}
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
