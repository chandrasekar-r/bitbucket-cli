package repo

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdFork(f *cmdutil.Factory) *cobra.Command {
	var destWorkspace string

	cmd := &cobra.Command{
		Use:   "fork <workspace/repo>",
		Short: "Fork a repository into your workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepo(f, args)
			if err != nil {
				return err
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			fork, err := client.ForkRepo(workspace, slug, destWorkspace)
			if err != nil {
				return fmt.Errorf("forking repo: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Forked %s/%s → %s\n", workspace, slug, fork.FullName)
			fmt.Fprintf(f.IOStreams.Out, "  Clone: %s\n", fork.CloneURL("https"))
			return nil
		},
	}

	cmd.Flags().StringVar(&destWorkspace, "into", "",
		"Fork into a specific workspace slug (defaults to your authenticated workspace)")
	return cmd
}
