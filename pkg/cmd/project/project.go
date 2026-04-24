// Package project provides the `bb project` command group.
package project

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdProject returns the `bb project` group command.
func NewCmdProject(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project <subcommand>",
		Short:   "Manage Bitbucket workspace projects",
		Aliases: []string{"projects", "proj"},
		Long: `Work with Bitbucket Cloud workspace projects (the container
that holds repositories).

All project commands operate on the active workspace.

  bb project list                       List projects
  bb project view <key>                 Show one project
  bb project create --key --name        Create a project
  bb project update <key> [flags]       Update a project
  bb project delete <key>               Delete a project`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdView(f))
	cmd.AddCommand(newCmdCreate(f))
	cmd.AddCommand(newCmdUpdate(f))
	cmd.AddCommand(newCmdDelete(f))
	return cmd
}
