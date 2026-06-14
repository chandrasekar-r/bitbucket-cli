package variablecmd

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

func newCmdEnvCreate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a deployment environment",
		Args:  cobra.ExactArgs(1),
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
			env, err := client.CreateDeploymentEnvironment(ctx.Workspace, ctx.RepoSlug, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Created environment %q (%s)\n", env.Name, env.UUID)
			return nil
		},
	}
	return cmd
}