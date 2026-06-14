package variablecmd

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdEnvList(f *cmdutil.Factory) *cobra.Command {
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployment environments",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := gitcontext.FromRemote()
			if ctx == nil {
				return fmt.Errorf("no repo found: run from inside a cloned Bitbucket repository")
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			envs, err := client.ListDeploymentEnvironments(ctx.Workspace, ctx.RepoSlug, limit)
			if err != nil {
				return err
			}
			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, envs, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(envs) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No deployment environments found")
				return nil
			}
			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("NAME", "SLUG", "UUID")
			for _, e := range envs {
				t.AddRow(e.Name, e.Slug, e.UUID)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d environment(s)\n", len(envs))
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 50, "Maximum environments to list (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}