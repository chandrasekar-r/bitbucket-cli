package project

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
		Use:   "delete <key>",
		Short: "Delete a project (must contain no repositories)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete project %s", key)}
			}
			if !force {
				var confirm string
				form := huh.NewForm(huh.NewGroup(
					huh.NewInput().
						Title(fmt.Sprintf("Type %q to confirm deletion", key)).
						Value(&confirm).
						Validate(func(s string) error {
							if s != key {
								return errors.New("key doesn't match")
							}
							return nil
						}),
				))
				if err := form.Run(); err != nil {
					return err
				}
				if confirm != key {
					return errors.New("deletion cancelled")
				}
			}

			workspace, err := f.Workspace()
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			if err := client.DeleteProject(workspace, key); err != nil {
				return fmt.Errorf("deleting project: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted project %s\n", key)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}
