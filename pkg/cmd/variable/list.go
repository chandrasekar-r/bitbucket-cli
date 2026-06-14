package variablecmd

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var workspaceOnly bool
	var envName string
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipeline variables",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, workspaceOnly, envName)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			var vars []api.PipelineVariable
			switch {
			case sc.workspaceOnly:
				vars, err = client.ListWorkspaceVariables(sc.workspace, limit)
			case sc.envName != "":
				envUUID, eerr := resolveEnvUUID(client, sc)
				if eerr != nil {
					return eerr
				}
				vars, err = client.ListDeploymentVariables(sc.workspace, sc.repoSlug, envUUID, limit)
			default:
				vars, err = client.ListRepoVariables(sc.workspace, sc.repoSlug, limit)
			}
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, vars, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(vars) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No variables found")
				return nil
			}
			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("KEY", "VALUE", "SECURED")
			for _, v := range vars {
				secured := "no"
				if v.Secured {
					secured = "yes"
				}
				t.AddRow(v.Key, displayValue(v), secured)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d variable(s)\n", len(vars))
			return nil
		},
	}

	addScopeFlags(cmd, &workspaceOnly, &envName)
	cmd.Flags().IntVarP(&limit, "limit", "L", 50, "Maximum variables to list (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}