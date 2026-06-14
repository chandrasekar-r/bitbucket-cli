package configcmd

import (
	"fmt"
	"sort"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuration values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			settings := cfg.ListSettings()
			keys := make([]string, 0, len(settings))
			for k := range settings {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("KEY", "VALUE")
			for _, k := range keys {
				t.AddRow(k, settings[k])
			}
			if err := t.Render(); err != nil {
				return err
			}

			aliases := cfg.Aliases()
			if len(aliases) > 0 {
				fmt.Fprintln(f.IOStreams.Out)
				fmt.Fprintln(f.IOStreams.Out, "Aliases:")
				aliasKeys := make([]string, 0, len(aliases))
				for k := range aliases {
					aliasKeys = append(aliasKeys, k)
				}
				sort.Strings(aliasKeys)
				at := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
				at.AddHeader("NAME", "EXPANSION")
				for _, k := range aliasKeys {
					at.AddRow(k, aliases[k])
				}
				if err := at.Render(); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}