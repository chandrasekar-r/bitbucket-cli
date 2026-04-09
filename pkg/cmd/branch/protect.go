package branch

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdProtect(f *cmdutil.Factory) *cobra.Command {
	var pattern string
	var kind string

	cmd := &cobra.Command{
		Use:   "protect <name>",
		Short: "Enable branch restrictions (push protection)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if pattern == "" {
				pattern = name
			}
			workspace, slug, err := resolveRepoBranch(f)
			if err != nil {
				return err
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			r, err := client.CreateBranchRestriction(workspace, slug, kind, pattern)
			if err != nil {
				return fmt.Errorf("creating branch restriction: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Protection applied: %s on pattern %q (id: %d)\n",
				r.Kind, r.Pattern, r.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&pattern, "pattern", "", "Branch glob pattern (default: exact branch name)")
	cmd.Flags().StringVar(&kind, "kind", "push",
		"Restriction type: push, delete, restrict_merges, force")
	cmd.ValidArgsFunction = cmdutil.CompleteBranchNames(f)
	return cmd
}
