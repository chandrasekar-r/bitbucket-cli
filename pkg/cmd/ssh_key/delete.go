package sshkeycmd

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
		Use:   "delete <uuid>",
		Short: "Delete an SSH key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete SSH key %s", uuid)}
			}
			if !force {
				var confirm bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete SSH key %s?", uuid)).
						Affirmative("Delete").Negative("Cancel").
						Value(&confirm),
				))
				if err := form.Run(); err != nil {
					return err
				}
				if !confirm {
					return errors.New("deletion cancelled")
				}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			if err := client.DeleteSSHKey(uuid); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted SSH key %s\n", uuid)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}