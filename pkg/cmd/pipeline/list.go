package pipeline

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var limit int
	var branch string
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent pipeline runs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			pipelines, err := client.ListPipelines(workspace, slug, limit)
			if err != nil {
				return err
			}

			// Filter by branch if specified
			if branch != "" {
				filtered := pipelines[:0]
				for _, p := range pipelines {
					if p.Target.RefName == branch {
						filtered = append(filtered, p)
					}
				}
				pipelines = filtered
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, pipelines, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			if len(pipelines) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No pipeline runs found")
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("#", "STATUS", "BRANCH", "TRIGGERED BY", "DURATION", "CREATED")
			for _, p := range pipelines {
				status := statusColor(p.State.Name, p.ResultName(), f.IOStreams.ColorEnabled())
				created := p.CreatedOn
				if len(created) > 10 {
					created = created[:10]
				}
				t.AddRow(
					fmt.Sprintf("%d", p.BuildNumber),
					status,
					p.Target.RefName,
					p.Creator.DisplayName,
					formatDuration(p.DurationInSeconds),
					created,
				)
			}
			return t.Render()
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 20, "Maximum number of pipeline runs to list (0 = all)")
	cmd.Flags().StringVar(&branch, "branch", "", "Filter by branch name")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
