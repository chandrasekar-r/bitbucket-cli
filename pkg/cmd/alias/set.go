package aliascmd

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/alias"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdSet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name> <expansion>",
		Short: "Create a command alias",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			expansion := strings.TrimSpace(args[1])
			if name == "" {
				return fmt.Errorf("alias name cannot be empty")
			}
			if err := alias.ValidateExpansion(expansion); err != nil {
				return err
			}
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			if err := cfg.SetAlias(name, expansion); err != nil {
				return fmt.Errorf("saving alias: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Alias %q → %q\n", name, expansion)
			return nil
		},
	}
	return cmd
}