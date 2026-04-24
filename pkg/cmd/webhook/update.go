package webhook

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		workspaceOnly bool
		url           string
		description   string
		events        []string
		active        bool
		inactive      bool
	)

	cmd := &cobra.Command{
		Use:   "update <uuid>",
		Short: "Update a webhook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if active && inactive {
				return errors.New("--active and --inactive are mutually exclusive")
			}

			sc, err := resolveScope(f, workspaceOnly)
			if err != nil {
				return err
			}

			input := api.WebhookInput{}
			if url != "" {
				input.URL = url
			}
			if cmd.Flags().Changed("description") {
				input.Description = description
			}
			if len(events) > 0 {
				input.Events = parseEventFlags(events)
			}
			switch {
			case active:
				t := true
				input.Active = &t
			case inactive:
				f := false
				input.Active = &f
			}

			if input.URL == "" && input.Description == "" && input.Events == nil && input.Active == nil {
				return errors.New("nothing to update: pass --url, --description, --event, --active, or --inactive")
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			var h *api.Webhook
			if sc.workspaceOn {
				h, err = client.UpdateWorkspaceHook(sc.workspace, args[0], input)
			} else {
				h, err = client.UpdateRepoHook(sc.workspace, sc.repoSlug, args[0], input)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Updated webhook %s\n", h.UUID)
			return nil
		},
	}

	addScopeFlag(cmd, &workspaceOnly)
	cmd.Flags().StringVar(&url, "url", "", "New endpoint URL")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringSliceVar(&events, "event", nil, "New event list (replaces existing)")
	cmd.Flags().BoolVar(&active, "active", false, "Mark the webhook active")
	cmd.Flags().BoolVar(&inactive, "inactive", false, "Mark the webhook inactive")

	return cmd
}
