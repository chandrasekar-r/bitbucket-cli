package auth

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/auth/login"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/auth/logout"
	authstatus "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/auth/status"
	authtoken "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/auth/token"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdAuth returns the `bb auth` group command with all subcommands registered.
func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <subcommand>",
		Short: "Authenticate with Bitbucket Cloud",
		Long: `Manage authentication to Bitbucket Cloud.

  bb auth login       Log in using OAuth or an API token
  bb auth logout      Remove stored credentials
  bb auth status      Show authentication state
  bb auth token       Print the active access token`,
	}

	cmd.AddCommand(login.NewCmdLogin(f))
	cmd.AddCommand(logout.NewCmdLogout(f))
	cmd.AddCommand(authstatus.NewCmdStatus(f))
	cmd.AddCommand(authtoken.NewCmdToken(f))
	return cmd
}
