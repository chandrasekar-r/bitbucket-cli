package webhook

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var workspaceOnly bool
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, workspaceOnly)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			var hooks []api.Webhook
			if sc.workspaceOn {
				hooks, err = client.ListWorkspaceHooks(sc.workspace)
			} else {
				hooks, err = client.ListRepoHooks(sc.workspace, sc.repoSlug)
			}
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, hooks, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(hooks) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No webhooks found")
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("UUID", "ACTIVE", "URL", "EVENTS", "DESCRIPTION")
			for _, h := range hooks {
				active := "yes"
				if !h.Active {
					active = "no"
				}
				t.AddRow(h.UUID, active, h.URL, strings.Join(h.Events, ","), h.Description)
			}
			return t.Render()
		},
	}

	addScopeFlag(cmd, &workspaceOnly)
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
