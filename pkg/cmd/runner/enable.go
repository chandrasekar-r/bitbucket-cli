package runner

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdEnable(f *cmdutil.Factory) *cobra.Command {
	var repoFlag bool
	cmd := &cobra.Command{
		Use:   "enable <uuid>",
		Short: "Re-enable a disabled runner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, repoFlag)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			off := false
			if _, err := updateRunner(client, sc, args[0], api.RunnerUpdate{Disabled: &off}); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Enabled runner %s\n", args[0])
			return nil
		},
	}
	addRepoFlag(cmd, &repoFlag)
	return cmd
}
