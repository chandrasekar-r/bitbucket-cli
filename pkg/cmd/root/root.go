package root

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/completion"
	authcmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/auth"
	branchcmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/branch"
	issuecmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/issue"
	pipelinecmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/pipeline"
	prcmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/pr"
	repocmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/repo"
	snippetcmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/snippet"
	versioncmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/version"
	workspacecmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/workspace"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmdRoot builds the root cobra.Command and wires up the Factory.
// All subcommands are registered here.
func NewCmdRoot() (*cobra.Command, *cmdutil.Factory) {
	ios := iostreams.System()

	var (
		workspaceFlag string
		noTTY         bool
	)

	// Lazy-load config (cached after first call)
	var cachedConfig *config.Config
	getConfig := func() (*config.Config, error) {
		if cachedConfig != nil {
			return cachedConfig, nil
		}
		cfg, err := config.Load()
		if err != nil {
			return nil, err
		}
		cachedConfig = cfg
		return cachedConfig, nil
	}

	// Workspace resolution chain (highest → lowest priority):
	// 1. --workspace flag
	// 2. BITBUCKET_WORKSPACE env var
	// 3. git remote URL inference
	// 4. default_workspace in config file
	resolveWorkspace := func() (string, error) {
		if workspaceFlag != "" {
			return workspaceFlag, nil
		}
		if env := os.Getenv("BITBUCKET_WORKSPACE"); env != "" {
			return env, nil
		}
		if ctx := gitcontext.FromRemote(); ctx != nil {
			return ctx.Workspace, nil
		}
		cfg, err := getConfig()
		if err != nil {
			return "", err
		}
		ws := cfg.DefaultWorkspace()
		if ws == "" {
			return "", errors.New(
				"no workspace found: set one with `bb workspace use <slug>`\n" +
					"or set the BITBUCKET_WORKSPACE environment variable",
			)
		}
		return ws, nil
	}

	// buildHTTPClient returns an authenticated http.Client.
	// Priority:
	//  1. BITBUCKET_USERNAME + BITBUCKET_TOKEN env vars (CI/headless)
	//  2. Stored OAuth token from ~/.config/bb/tokens.json
	//     — transparently refreshes expired tokens (with file lock)
	//  3. Unauthenticated (for `bb auth login` and `bb version` which don't need auth)
	buildHTTPClient := func() (*http.Client, error) {
		// 1. Environment variable token auth
		envUsername := os.Getenv("BITBUCKET_USERNAME")
		envToken := os.Getenv("BITBUCKET_TOKEN")
		if envUsername != "" && envToken != "" {
			return &http.Client{
				Transport: api.NewRetryTransport(&api.TokenTransport{
					Username: envUsername,
					Token:    envToken,
				}),
			}, nil
		}

		// 2. Stored credentials
		store := bbauth.NewTokenStore()
		acc, err := store.GetActive()
		if err != nil || acc == nil {
			// Not authenticated — return bare client; commands that need auth will fail with clear errors
			return &http.Client{Transport: api.NewRetryTransport(nil)}, nil
		}

		// Refresh expired OAuth tokens (transparent to the caller)
		if acc.IsExpired() && acc.AuthType == bbauth.AuthTypeOAuth {
			if rfErr := bbauth.RefreshAccessToken(store, acc.Username); rfErr != nil {
				return nil, &cmdutil.AuthError{Message: "session expired. Run: bb auth login"}
			}
			acc, err = store.GetActive()
			if err != nil || acc == nil {
				return nil, &cmdutil.AuthError{Message: "could not load refreshed token. Run: bb auth login"}
			}
		}

		var transport http.RoundTripper
		if acc.AuthType == bbauth.AuthTypeToken {
			transport = &api.TokenTransport{Username: acc.Username, Token: acc.AccessToken}
		} else {
			transport = &bearerTransport{token: acc.AccessToken}
		}
		return &http.Client{Transport: api.NewRetryTransport(transport)}, nil
	}

	f := &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: buildHTTPClient,
		Config:     getConfig,
		BaseURL:    api.DefaultBaseURL,
		Workspace:  resolveWorkspace,
	}

	cmd := &cobra.Command{
		Use:   "bb <command> <subcommand> [flags]",
		Short: "Bitbucket CLI — work with Bitbucket Cloud from the terminal",
		Long: `bb is a CLI for Bitbucket Cloud, modeled on GitHub's gh.

Manage pull requests, repositories, branches, pipelines, issues, and more — all without leaving your terminal.

Start with: bb auth login`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ios.SetNoTTY(noTTY)
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return fmt.Errorf("binding flags: %w", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&workspaceFlag, "workspace", "w", "",
		"Bitbucket workspace slug (overrides BITBUCKET_WORKSPACE env and config)")
	cmd.PersistentFlags().BoolVar(&noTTY, "no-tty", false,
		"Disable interactive prompts; use --force on destructive operations")

	// Register all subcommand groups
	cmd.AddCommand(authcmd.NewCmdAuth(f))
	cmd.AddCommand(workspacecmd.NewCmdWorkspace(f))
	cmd.AddCommand(repocmd.NewCmdRepo(f))
	cmd.AddCommand(branchcmd.NewCmdBranch(f))
	cmd.AddCommand(prcmd.NewCmdPR(f))
	cmd.AddCommand(pipelinecmd.NewCmdPipeline(f))
	cmd.AddCommand(issuecmd.NewCmdIssue(f))
	cmd.AddCommand(snippetcmd.NewCmdSnippet(f))
	cmd.AddCommand(versioncmd.NewCmdVersion(f))
	cmd.AddCommand(completion.NewCmdCompletion(f))

	return cmd, f
}

// Execute runs the root command and exits with an appropriate code.
func Execute() {
	cmd, _ := NewCmdRoot()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// bearerTransport adds an OAuth Bearer token Authorization header.
type bearerTransport struct{ token string }

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(req)
}
