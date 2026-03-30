package repo

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdRename(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <workspace/repo> <new-name>",
		Short: "Rename a repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepo(f, args[:1])
			if err != nil {
				return err
			}
			newName := args[1]

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			r, err := client.RenameRepo(workspace, slug, newName)
			if err != nil {
				return fmt.Errorf("renaming repo: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Renamed %s/%s → %s\n", workspace, slug, r.FullName)
			return nil
		},
	}
	return cmd
}
