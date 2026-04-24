package branch

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var from string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			workspace, slug, err := resolveRepoBranch(f)
			if err != nil {
				return err
			}

			// Default source: resolve to the repo's default branch HEAD
			source := from
			if source == "" {
				httpClient, err := f.HttpClient()
				if err != nil {
					return err
				}
				client := api.New(httpClient, f.BaseURL)
				repo, err := client.GetRepo(workspace, slug)
				if err != nil {
					return err
				}
				if repo.MainBranch != nil {
					// Use branch name as source; API accepts branch names in hash field
					source = repo.MainBranch.Name
				} else {
					source = "main"
				}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			b, err := client.CreateBranch(workspace, slug, name, source)
			if err != nil {
				return fmt.Errorf("creating branch: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Created branch %s (from %s)\n", b.Name, source)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Source branch name or commit hash (default: repo default branch)")
	return cmd
}
