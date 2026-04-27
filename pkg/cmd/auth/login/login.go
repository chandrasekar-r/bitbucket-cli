package login

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
	"github.com/spf13/cobra"
)

// NewCmdLogin creates the `bb auth login` command.
func NewCmdLogin(f *cmdutil.Factory) *cobra.Command {
	var withToken bool
	var username string
	var token string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Bitbucket Cloud",
		Long: `Authenticate with Bitbucket Cloud.

By default bb opens a browser for OAuth 2.0 authentication.
In CI or headless environments, pipe a Bitbucket API token via stdin:

  echo "$BB_TOKEN" | bb auth login --with-token --username myusername

To pass the token directly (less secure — value appears in shell history):

  bb auth login --token "$BB_TOKEN" --username myusername

Note: App passwords are disabled as of June 9, 2026. Use Bitbucket API tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("token") {
				if token == "" {
					return &cmdutil.FlagError{Err: errors.New("--token value cannot be empty")}
				}
				if !cmd.Flags().Changed("username") {
					return &cmdutil.FlagError{Err: errors.New("--username is required when --token is used")}
				}
				fmt.Fprintln(f.IOStreams.ErrOut,
					"Warning: --token value appears in shell history and process listings.\n"+
						"Prefer: echo $TOKEN | bb auth login --with-token --username <user>")
				return runTokenLogin(f, username, token)
			}
			if withToken {
				return runTokenLogin(f, username, "")
			}
			return runOAuthLogin(f)
		},
	}

	cmd.Flags().BoolVar(&withToken, "with-token", false,
		"Read a Bitbucket API token from stdin instead of browser OAuth")
	cmd.Flags().StringVar(&username, "username", "",
		"Bitbucket username (required with --with-token in --no-tty mode)")
	cmd.Flags().StringVar(&token, "token", "",
		"Bitbucket API token (insecure: value visible in shell history; prefer --with-token)")

	return cmd
}

// runOAuthLogin opens the browser for an OAuth 2.0 loopback flow.
func runOAuthLogin(f *cmdutil.Factory) error {
	result, err := bbauth.RunOAuthFlow(context.Background())
	if err != nil {
		return err
	}
	return persistAndReport(f, "", result.AccessToken, result.RefreshToken,
		bbauth.AuthTypeOAuth, result.Expiry)
}

// runTokenLogin validates and stores a Bitbucket API token.
// preToken, when non-empty, is used directly instead of reading from stdin.
func runTokenLogin(f *cmdutil.Factory, username, preToken string) error {
	var token string
	if preToken != "" {
		token = preToken
	} else {
		scanner := bufio.NewScanner(f.IOStreams.In)
		scanner.Scan()
		token = strings.TrimSpace(scanner.Text())
		if token == "" {
			return &cmdutil.FlagError{Err: errors.New("token is empty — pipe a Bitbucket API token via stdin")}
		}
		if username == "" {
			if !f.IOStreams.IsStdoutTTY() {
				return &cmdutil.FlagError{Err: errors.New("--username is required when --with-token is used in non-interactive mode")}
			}
			fmt.Fprint(f.IOStreams.Out, "Bitbucket username: ")
			scanner.Scan()
			username = strings.TrimSpace(scanner.Text())
			if username == "" {
				return errors.New("username cannot be empty")
			}
		}
	}

	return persistAndReport(f, username, token, "", bbauth.AuthTypeToken, time.Time{})
}

// persistAndReport validates the token against GET /user, saves it to the store,
// and prints a success message.
func persistAndReport(f *cmdutil.Factory, username, accessToken, refreshToken string, authType bbauth.AuthType, expiry time.Time) error {
	// Build a one-shot client with just this token for validation
	var transport http.RoundTripper
	switch authType {
	case bbauth.AuthTypeToken:
		transport = &api.TokenTransport{Username: username, Token: accessToken}
	default: // OAuth bearer
		transport = &bearerTransport{token: accessToken}
	}
	httpClient := &http.Client{Transport: api.NewRetryTransport(transport)}
	apiClient := api.New(httpClient, f.BaseURL)

	user, err := apiClient.GetUser()
	if err != nil {
		return fmt.Errorf("validating credentials: %w\nensure the token has the required scopes", err)
	}

	workspaces, _ := apiClient.GetUserWorkspaces() // non-fatal

	acc := &bbauth.Account{
		Username:       user.Username,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		TokenType:      "bearer",
		Expiry:         expiry,
		WorkspaceSlugs: workspaces,
		AuthType:       authType,
	}
	store := bbauth.NewTokenStore()
	if err := store.SetAccount(acc); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}

	// Auto-configure default workspace when user has exactly one
	wsNote := ""
	if len(workspaces) == 1 {
		cfg, cerr := config.Load()
		if cerr == nil {
			_ = cfg.Set(config.KeyDefaultWorkspace, workspaces[0])
		}
		wsNote = fmt.Sprintf(" (%s)", workspaces[0])
	} else if len(workspaces) > 1 {
		wsNote = fmt.Sprintf(" (%d workspaces)", len(workspaces))
	}

	fmt.Fprintf(f.IOStreams.Out, "✓ Logged in as %s%s\n", user.Username, wsNote)
	if len(workspaces) > 1 {
		fmt.Fprintf(f.IOStreams.Out, "  Run `bb workspace use <slug>` to set your default workspace.\n")
	}
	return nil
}

// bearerTransport injects an OAuth Bearer token Authorization header.
type bearerTransport struct{ token string }

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(req)
}
