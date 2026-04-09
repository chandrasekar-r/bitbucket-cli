// Package extension provides the `bb extension` command group for managing CLI plugins.
package extension

import (
	"fmt"

	ext "github.com/chandrasekar-r/bitbucket-cli/pkg/extension"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdExtension returns the `bb extension` group command.
func NewCmdExtension(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension <subcommand>",
		Short: "Manage bb CLI extensions",
		Long: `Install, list, and remove bb CLI extensions.

Extensions are git repositories containing an executable named bb-<name>.
Once installed, they appear as top-level bb commands.

  bb extension install <repo>   Install an extension from a git repository
  bb extension list             List installed extensions
  bb extension remove <name>    Remove an installed extension`,
		Aliases: []string{"ext"},
	}

	cmd.AddCommand(newCmdInstall(f))
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdRemove(f))
	return cmd
}

func newCmdInstall(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "install <repository>",
		Short: "Install an extension from a git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			installed, err := ext.Install(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Installed extension %q\n", installed.Name)
			return nil
		},
	}
}

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed extensions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			exts, err := ext.Installed()
			if err != nil {
				return err
			}
			if len(exts) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No extensions installed")
				return nil
			}
			for _, e := range exts {
				fmt.Fprintf(f.IOStreams.Out, "%s\t%s\n", e.Name, e.Path)
			}
			return nil
		},
	}
}

func newCmdRemove(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ext.Remove(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Removed extension %q\n", args[0])
			return nil
		},
	}
}
