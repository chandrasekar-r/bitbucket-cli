package webhook

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	var workspaceOnly bool
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view <uuid>",
		Short: "Show a single webhook",
		Args:  cobra.ExactArgs(1),
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

			var h *api.Webhook
			if sc.workspaceOn {
				h, err = client.GetWorkspaceHook(sc.workspace, args[0])
			} else {
				h, err = client.GetRepoHook(sc.workspace, sc.repoSlug, args[0])
			}
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, h, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			fmt.Fprintf(f.IOStreams.Out, "UUID:        %s\n", h.UUID)
			fmt.Fprintf(f.IOStreams.Out, "URL:         %s\n", h.URL)
			fmt.Fprintf(f.IOStreams.Out, "Description: %s\n", h.Description)
			fmt.Fprintf(f.IOStreams.Out, "Active:      %t\n", h.Active)
			fmt.Fprintf(f.IOStreams.Out, "Events:      %s\n", strings.Join(h.Events, ", "))
			fmt.Fprintf(f.IOStreams.Out, "Created:     %s\n", h.CreatedAt)
			return nil
		},
	}

	addScopeFlag(cmd, &workspaceOnly)
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
