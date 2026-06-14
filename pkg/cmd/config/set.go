package configcmd

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdSet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			if !isKnownKey(key) {
				return fmt.Errorf("unknown config key %q", key)
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			if err := cfg.Set(key, value); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Set %s=%s\n", key, value)
			return nil
		},
	}
	return cmd
}