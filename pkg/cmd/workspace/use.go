package workspace

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
	"github.com/spf13/cobra"
)

func newCmdUse(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <slug>",
		Short: "Set the default workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			// Validate the workspace exists
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			if _, err := client.GetWorkspace(slug); err != nil {
				return fmt.Errorf("workspace %q not found: %w", slug, err)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := cfg.Set(config.KeyDefaultWorkspace, slug); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Default workspace set to %q\n", slug)
			return nil
		},
	}
	return cmd
}
