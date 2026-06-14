package variablecmd

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var workspaceOnly bool
	var envName string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a pipeline variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, workspaceOnly, envName)
			if err != nil {
				return err
			}
			key := args[0]

			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete variable %s", key)}
			}
			if !force {
				var confirm bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete variable %q?", key)).
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

			switch {
			case sc.workspaceOnly:
				err = client.DeleteWorkspaceVariable(sc.workspace, key)
			case sc.envName != "":
				envUUID, eerr := resolveEnvUUID(client, sc)
				if eerr != nil {
					return eerr
				}
				err = client.DeleteDeploymentVariable(sc.workspace, sc.repoSlug, envUUID, key)
			default:
				err = client.DeleteRepoVariable(sc.workspace, sc.repoSlug, key)
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted variable %q\n", key)
			return nil
		},
	}

	addScopeFlags(cmd, &workspaceOnly, &envName)
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}