package aliascmd

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a command alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			if err := cfg.DeleteAlias(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted alias %q\n", args[0])
			return nil
		},
	}
	return cmd
}