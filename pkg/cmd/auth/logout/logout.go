package logout

import (
	"errors"
	"fmt"

	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdLogout creates the `bb auth logout` command.
func NewCmdLogout(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored Bitbucket credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := bbauth.NewTokenStore()
			acc, err := store.GetActive()
			if err != nil {
				return err
			}
			if acc == nil {
				return errors.New("not logged in")
			}
			if err := store.RemoveAccount(acc.Username); err != nil {
				return fmt.Errorf("removing credentials: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Logged out as %s\n", acc.Username)
			return nil
		},
	}
	return cmd
}
