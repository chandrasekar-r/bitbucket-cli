package status

import (
	"errors"
	"fmt"

	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdStatus creates the `bb auth status` command.
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := bbauth.NewTokenStore()
			tf, err := store.Load()
			if err != nil {
				return err
			}
			if len(tf.Accounts) == 0 {
				return errors.New("not logged in. Run: bb auth login")
			}

			for username, acc := range tf.Accounts {
				active := ""
				if username == tf.ActiveAccount {
					active = " (active)"
				}
				fmt.Fprintf(f.IOStreams.Out, "github.com\n")
				fmt.Fprintf(f.IOStreams.Out, "  ✓ Logged in to bitbucket.org account %s%s\n", acc.Username, active)
				fmt.Fprintf(f.IOStreams.Out, "  - Auth type: %s\n", acc.AuthType)

				if !acc.Expiry.IsZero() {
					if acc.IsExpired() {
						// Attempt transparent refresh
						if rfErr := bbauth.RefreshAccessToken(store, username); rfErr != nil {
							fmt.Fprintf(f.IOStreams.Out, "  - Token: expired (run `bb auth login` to re-authenticate)\n")
						} else {
							fmt.Fprintf(f.IOStreams.Out, "  - Token: refreshed\n")
						}
					} else {
						fmt.Fprintf(f.IOStreams.Out, "  - Token: valid (expires %s)\n", acc.Expiry.Format("2006-01-02 15:04"))
					}
				} else {
					fmt.Fprintf(f.IOStreams.Out, "  - Token: valid (no expiry)\n")
				}

				if len(acc.WorkspaceSlugs) > 0 {
					fmt.Fprintf(f.IOStreams.Out, "  - Workspaces: %v\n", acc.WorkspaceSlugs)
				}
			}
			return nil
		},
	}
	return cmd
}
