package configcmd

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
	"github.com/spf13/cobra"
)

func newCmdGet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Print a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if !isKnownKey(key) {
				return fmt.Errorf("unknown config key %q (valid keys: %s)", key, strings.Join(config.KnownKeys, ", "))
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			fmt.Fprintln(f.IOStreams.Out, cfg.Get(key))
			return nil
		},
	}
	return cmd
}

func isKnownKey(key string) bool {
	for _, k := range config.KnownKeys {
		if k == key {
			return true
		}
	}
	return false
}