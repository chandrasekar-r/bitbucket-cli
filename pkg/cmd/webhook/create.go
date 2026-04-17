package webhook

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		workspaceOnly bool
		url           string
		description   string
		events        []string
		inactive      bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, workspaceOnly)
			if err != nil {
				return err
			}

			parsedEvents := parseEventFlags(events)

			// Interactive form when nothing was passed and we have a TTY.
			if url == "" && description == "" && len(parsedEvents) == 0 {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.NoTTYError{Operation: "bb webhook create (requires --url and --event, or a TTY)"}
				}
				var eventsStr string
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Webhook URL").Value(&url).
							Validate(func(s string) error {
								if s == "" {
									return errors.New("url is required")
								}
								return nil
							}),
						huh.NewInput().Title("Description (optional)").Value(&description),
						huh.NewInput().Title("Events (comma-separated, e.g. repo:push,pullrequest:created)").
							Value(&eventsStr).
							Validate(func(s string) error {
								if s == "" {
									return errors.New("at least one event is required")
								}
								return nil
							}),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
				parsedEvents = parseEventFlags([]string{eventsStr})
			}

			if url == "" {
				return errors.New("--url is required")
			}
			if len(parsedEvents) == 0 {
				return errors.New("--event is required (e.g. --event repo:push,pullrequest:created)")
			}

			active := !inactive
			input := api.WebhookInput{
				URL:         url,
				Description: description,
				Events:      parsedEvents,
				Active:      &active,
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			var h *api.Webhook
			if sc.workspaceOn {
				h, err = client.CreateWorkspaceHook(sc.workspace, input)
			} else {
				h, err = client.CreateRepoHook(sc.workspace, sc.repoSlug, input)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Created webhook %s\n  URL:    %s\n  Events: %v\n",
				h.UUID, h.URL, h.Events)
			return nil
		},
	}

	addScopeFlag(cmd, &workspaceOnly)
	cmd.Flags().StringVar(&url, "url", "", "Endpoint URL that receives webhook deliveries (required)")
	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	cmd.Flags().StringSliceVar(&events, "event", nil,
		"Webhook event (repeatable, comma-separated). Example: --event repo:push,pullrequest:created")
	cmd.Flags().BoolVar(&inactive, "inactive", false, "Create the webhook in the inactive state")

	return cmd
}
