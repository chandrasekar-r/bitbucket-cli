package branch

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// newCmdRename implements bb branch rename — Bitbucket has no rename API,
// so we get the source branch commit, create the new branch from it, then delete the old.
func newCmdRename(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a branch (create new + delete old)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName, newName := args[0], args[1]
			workspace, slug, err := resolveRepoBranch(f)
			if err != nil {
				return err
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			// Get source commit hash
			old, err := client.GetBranch(workspace, slug, oldName)
			if err != nil {
				return fmt.Errorf("branch %q not found: %w", oldName, err)
			}

			// Create new branch
			if _, err := client.CreateBranch(workspace, slug, newName, old.Target.Hash); err != nil {
				return fmt.Errorf("creating %q: %w", newName, err)
			}

			// Delete old branch
			if err := client.DeleteBranch(workspace, slug, oldName); err != nil {
				return fmt.Errorf("created %q but could not delete %q: %w", newName, oldName, err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Renamed branch %s → %s\n", oldName, newName)
			return nil
		},
	}
	return cmd
}
