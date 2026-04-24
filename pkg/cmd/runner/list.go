package runner

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var repoFlag bool
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List runners",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, repoFlag)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			runners, err := listRunners(client, sc)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, runners, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(runners) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No runners found")
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("UUID", "NAME", "STATUS", "LABELS", "VERSION")
			for _, r := range runners {
				status := r.State.Status
				if r.Disabled {
					status = "DISABLED"
				}
				t.AddRow(r.UUID, r.Name, status, strings.Join(r.Labels, ","), r.State.Version.Version)
			}
			return t.Render()
		},
	}
	addRepoFlag(cmd, &repoFlag)
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
