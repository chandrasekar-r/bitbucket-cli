package token

import (
	"errors"
	"fmt"

	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdToken creates the `bb auth token` command.
// Prints the active access token to stdout — useful for scripting.
func NewCmdToken(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print the active Bitbucket access token",
		Long: `Print the active Bitbucket access token to stdout.

Useful for scripting and tooling that needs the token directly:

  export BB_TOKEN=$(bb auth token)
  curl -H "Authorization: Bearer $BB_TOKEN" https://api.bitbucket.org/2.0/user`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := bbauth.NewTokenStore()
			acc, err := store.GetActive()
			if err != nil {
				return err
			}
			if acc == nil {
				return errors.New("not authenticated. Run: bb auth login")
			}

			// Refresh if expired
			if acc.IsExpired() && acc.RefreshToken != "" {
				if rfErr := bbauth.RefreshAccessToken(store, acc.Username); rfErr != nil {
					return fmt.Errorf("session expired: %w\nRun: bb auth login", rfErr)
				}
				// Reload after refresh
				acc, err = store.GetActive()
				if err != nil || acc == nil {
					return errors.New("could not retrieve refreshed token. Run: bb auth login")
				}
			}

			fmt.Fprintln(f.IOStreams.Out, acc.AccessToken)
			return nil
		},
	}
	return cmd
}
