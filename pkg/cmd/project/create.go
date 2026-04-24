package project

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
		key         string
		name        string
		description string
		public      bool
		private     bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project in the active workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if public && private {
				return errors.New("--public and --private are mutually exclusive")
			}

			// Interactive fallback
			if key == "" && name == "" && description == "" && !public && !private {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.NoTTYError{Operation: "bb project create (requires --key and --name, or a TTY)"}
				}
				var visibility string
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Project key (e.g. ENG)").Value(&key).
							Validate(func(s string) error {
								if s == "" {
									return errors.New("key is required")
								}
								return nil
							}),
						huh.NewInput().Title("Project name").Value(&name).
							Validate(func(s string) error {
								if s == "" {
									return errors.New("name is required")
								}
								return nil
							}),
						huh.NewInput().Title("Description (optional)").Value(&description),
						huh.NewSelect[string]().Title("Visibility").
							Options(
								huh.NewOption("Private", "private"),
								huh.NewOption("Public", "public"),
							).Value(&visibility),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
				private = visibility == "private"
				public = visibility == "public"
			}

			if key == "" || name == "" {
				return errors.New("--key and --name are required")
			}

			isPrivate := !public

			workspace, err := f.Workspace()
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			p, err := client.CreateProject(workspace, api.ProjectCreateInput{
				Key:         key,
				Name:        name,
				Description: description,
				IsPrivate:   isPrivate,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Created project %s (%q)\n", p.Key, p.Name)
			if p.Links.HTML.Href != "" {
				fmt.Fprintf(f.IOStreams.Out, "  %s\n", p.Links.HTML.Href)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Project key, e.g. ENG (required)")
	cmd.Flags().StringVar(&name, "name", "", "Project name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Project description")
	cmd.Flags().BoolVar(&private, "private", false, "Create as private (default)")
	cmd.Flags().BoolVar(&public, "public", false, "Create as public")

	return cmd
}
