package branch

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a branch",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepoBranch(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			name, err := resolveBranchName(f, client, workspace, slug, args)
			if err != nil {
				return err
			}

			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: "delete branch " + name}
			}

			if !force && f.IOStreams.IsStdoutTTY() {
				var confirmed bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete branch %q?", name)).
						Value(&confirmed),
				))
				if err := form.Run(); err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(f.IOStreams.Out, "Cancelled")
					return nil
				}
			}

			if err := client.DeleteBranch(workspace, slug, name); err != nil {
				return fmt.Errorf("deleting branch: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted branch %s\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	cmd.ValidArgsFunction = cmdutil.CompleteBranchNames(f)
	return cmd
}
