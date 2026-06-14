package deploykeycmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a deploy key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid deploy key ID %q", args[0])
			}

			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete deploy key %d", id)}
			}
			if !force {
				var confirm bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete deploy key #%d?", id)).
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
			if err := client.DeleteDeployKey(workspace, slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted deploy key #%d\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}