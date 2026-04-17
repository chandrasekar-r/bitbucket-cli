package runner

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		repoFlag bool
		force    bool
	)
	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete a runner (irreversible)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, repoFlag)
			if err != nil {
				return err
			}
			uid := args[0]

			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete runner %s", uid)}
			}
			if !force {
				var confirm bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete runner %s?", uid)).
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

			if err := deleteRunner(client, sc, uid); err != nil {
				return fmt.Errorf("deleting runner: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted runner %s\n", uid)
			return nil
		},
	}
	addRepoFlag(cmd, &repoFlag)
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}
