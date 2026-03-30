package pipeline

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	var stepFlag string
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view <uuid>",
		Short: "Show pipeline details and step status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			p, err := client.GetPipeline(workspace, slug, uuid)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, p, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			w := f.IOStreams.Out
			status := statusColor(p.State.Name, p.ResultName(), f.IOStreams.ColorEnabled())
			fmt.Fprintf(w, "Pipeline #%d\n", p.BuildNumber)
			fmt.Fprintf(w, "Status:   %s\n", status)
			fmt.Fprintf(w, "Branch:   %s\n", p.Target.RefName)
			fmt.Fprintf(w, "Commit:   %s\n", p.Target.Commit.Hash)
			fmt.Fprintf(w, "Duration: %s\n", formatDuration(p.DurationInSeconds))
			fmt.Fprintf(w, "Created:  %s\n", p.CreatedOn)
			fmt.Fprintf(w, "URL:      %s\n", p.Links.HTML.Href)

			// Show steps
			steps, err := client.ListSteps(workspace, slug, uuid)
			if err != nil {
				return err
			}

			fmt.Fprintln(w)
			t := output.NewTable(w, f.IOStreams.ColorEnabled())
			t.AddHeader("STEP", "STATUS", "DURATION")
			for _, s := range steps {
				name := s.Name
				if name == "" {
					name = s.UUID
				}
				stepStatus := statusColor(s.State.Name, "", f.IOStreams.ColorEnabled())
				if s.State.Result != nil {
					stepStatus = statusColor(s.State.Name, s.State.Result.Name, f.IOStreams.ColorEnabled())
				}
				t.AddRow(name, stepStatus, formatDuration(s.DurationInSeconds))
			}
			if err := t.Render(); err != nil {
				return err
			}

			// Optionally print a specific step's log
			if stepFlag != "" {
				for _, s := range steps {
					if s.UUID == stepFlag || s.Name == stepFlag {
						fmt.Fprintf(w, "\n--- Log: %s ---\n", s.Name)
						logData, lerr := client.GetStepLog(workspace, slug, uuid, s.UUID, 0)
						if lerr != nil {
							return lerr
						}
						fmt.Fprint(w, string(logData))
						break
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&stepFlag, "step", "", "Show log for a specific step (name or UUID)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
