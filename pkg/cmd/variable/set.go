package variablecmd

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdSet(f *cmdutil.Factory) *cobra.Command {
	var workspaceOnly bool
	var envName string
	var secured bool

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a pipeline variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, workspaceOnly, envName)
			if err != nil {
				return err
			}
			key, value := args[0], args[1]
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			var v *api.PipelineVariable
			switch {
			case sc.workspaceOnly:
				v, err = client.SetWorkspaceVariable(sc.workspace, key, value, secured)
			case sc.envName != "":
				envUUID, eerr := resolveEnvUUID(client, sc)
				if eerr != nil {
					return eerr
				}
				v, err = client.SetDeploymentVariable(sc.workspace, sc.repoSlug, envUUID, key, value, secured)
			default:
				v, err = client.SetRepoVariable(sc.workspace, sc.repoSlug, key, value, secured)
			}
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Set variable %q (%s)\n", v.Key, v.UUID)
			return nil
		},
	}

	addScopeFlags(cmd, &workspaceOnly, &envName)
	cmd.Flags().BoolVar(&secured, "secured", false, "Store as a secured (secret) variable")
	return cmd
}