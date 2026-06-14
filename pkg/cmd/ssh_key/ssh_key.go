package sshkeycmd

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdSSHKey returns the `bb ssh-key` command group.
func NewCmdSSHKey(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ssh-key <subcommand>",
		Aliases: []string{"ssh-keys"},
		Short:   "Manage your SSH keys",
		Long: `Manage SSH public keys for the authenticated Bitbucket user.

  bb ssh-key list
  bb ssh-key add --key-file ~/.ssh/id_ed25519.pub --label "Laptop"
  bb ssh-key delete <uuid>`,
	}
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdAdd(f))
	cmd.AddCommand(newCmdDelete(f))
	return cmd
}