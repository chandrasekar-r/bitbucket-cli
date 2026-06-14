package aliascmd

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
		Short: "List command aliases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			aliases := cfg.Aliases()
			if len(aliases) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No aliases configured")
				return nil
			}
			keys := make([]string, 0, len(aliases))
			for k := range aliases {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("NAME", "EXPANSION")
			for _, k := range keys {
				t.AddRow(k, aliases[k])
			}
			return t.Render()
		},
	}
	return cmd
}