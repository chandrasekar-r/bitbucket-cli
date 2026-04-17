// Package switchacct implements `bb auth switch` — flips the active account
// among the credentials already stored in tokens.json. (Named switchacct
// because "switch" is a Go reserved keyword and can't be a package name.)
package switchacct

import (
	"errors"
	"fmt"

	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdSwitch creates the `bb auth switch` command.
func NewCmdSwitch(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [account]",
		Short: "Switch the active Bitbucket account",
		Long: `Switch which stored account bb uses for subsequent commands.

If [account] is omitted, the available accounts are listed.
Only accounts already authenticated via 'bb auth login' can be selected.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := bbauth.NewTokenStore()
			usernames, active, err := store.ListAccounts()
			if err != nil {
				return err
			}
			if len(usernames) == 0 {
				return errors.New("not logged in. Run: bb auth login")
			}

			if len(args) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "Stored accounts (active marked with *):")
				for _, u := range usernames {
					marker := " "
					if u == active {
						marker = "*"
					}
					fmt.Fprintf(f.IOStreams.Out, "  %s %s\n", marker, u)
				}
				fmt.Fprintln(f.IOStreams.Out, "\nRun: bb auth switch <account>")
				return nil
			}

			target := args[0]
			if target == active {
				fmt.Fprintf(f.IOStreams.Out, "Already active: %s\n", target)
				return nil
			}
			if err := store.SetActiveAccount(target); err != nil {
				return fmt.Errorf("%w\nAvailable accounts: %v", err, usernames)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Active account: %s\n", target)
			return nil
		},
	}
	return cmd
}
