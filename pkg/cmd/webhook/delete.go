package webhook

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
		workspaceOnly bool
		force         bool
	)

	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete a webhook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, workspaceOnly)
			if err != nil {
				return err
			}

			uid := args[0]
			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete webhook %s", uid)}
			}
			if !force {
				var confirm bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete webhook %s?", uid)).
						Affirmative("Delete").
						Negative("Cancel").
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

			if sc.workspaceOn {
				err = client.DeleteWorkspaceHook(sc.workspace, uid)
			} else {
				err = client.DeleteRepoHook(sc.workspace, sc.repoSlug, uid)
			}
			if err != nil {
				return fmt.Errorf("deleting webhook: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted webhook %s\n", uid)
			return nil
		},
	}

	addScopeFlag(cmd, &workspaceOnly)
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}
