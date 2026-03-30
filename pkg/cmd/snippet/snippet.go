// Package snippet provides the `bb snippet` command group.
package snippet

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdSnippet returns the `bb snippet` group command.
func NewCmdSnippet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snippet <subcommand>",
		Short: "Manage Bitbucket Snippets",
		Long: `Work with Bitbucket Cloud Snippets (like Gists).

  bb snippet list             List your snippets
  bb snippet view <id>        Show snippet content
  bb snippet create           Create a new snippet
  bb snippet edit <id>        Edit a snippet in $EDITOR
  bb snippet delete <id>      Delete a snippet
  bb snippet clone <id>       Clone a snippet repository`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdEdit(f))
	cmd.AddCommand(newCmdDelete(f))
	cmd.AddCommand(newCmdClone(f))
	return cmd
}
