package repo

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <workspace/repo>",
		Short: "Delete a repository (irreversible)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepo(f, args)
			if err != nil {
				return err
			}

			// Require --force in non-interactive mode
			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete repo %s/%s", workspace, slug)}
			}

			// Interactive confirmation: user must type the repo name
			if !force {
				var confirm string
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title(fmt.Sprintf("Type %q to confirm deletion", slug)).
							Value(&confirm).
							Validate(func(s string) error {
								if s != slug {
									return errors.New("name doesn't match")
								}
								return nil
							}),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
				if confirm != slug {
					return errors.New("deletion cancelled")
				}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			if err := client.DeleteRepo(workspace, slug); err != nil {
				return fmt.Errorf("deleting repo: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted %s/%s\n", workspace, slug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt (required in --no-tty mode)")
	return cmd
}
