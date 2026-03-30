package root

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/completion"
	versioncmd "github.com/chandrasekar-r/bitbucket-cli/pkg/cmd/version"
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

	// Workspace resolution chain:
	// 1. --workspace flag
	// 2. BITBUCKET_WORKSPACE env var (handled by Viper binding in config)
	// 3. git remote inference
	// 4. config default_workspace
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
				"no workspace found. Set one with: bb workspace use <slug>\n" +
					"Or set the BITBUCKET_WORKSPACE environment variable.",
			)
		}
		return ws, nil
	}

	// HTTP client factory — currently returns an unauthenticated client.
	// Phase 2 (auth package) will replace this with a token-aware transport.
	getHTTPClient := func() (*http.Client, error) {
		// Check for env-var based token auth (CI/headless mode)
		username := os.Getenv("BITBUCKET_USERNAME")
		token := os.Getenv("BITBUCKET_TOKEN")
		if username != "" && token != "" {
			return &http.Client{
				Transport: api.NewRetryTransport(&api.TokenTransport{
					Username: username,
					Token:    token,
				}),
			}, nil
		}
		// Unauthenticated for now — Phase 2 wires up OAuth token storage
		return &http.Client{Transport: api.NewRetryTransport(nil)}, nil
	}

	f := &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: getHTTPClient,
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
			// Apply --no-tty flag to IOStreams
			ios.SetNoTTY(noTTY)
			// Bind Cobra flags to Viper after parsing
			viper.BindPFlags(cmd.Flags())
			return nil
		},
	}

	// Persistent flags available on every subcommand
	cmd.PersistentFlags().StringVarP(&workspaceFlag, "workspace", "w", "",
		"Bitbucket workspace slug (overrides BITBUCKET_WORKSPACE env and config)")
	cmd.PersistentFlags().BoolVar(&noTTY, "no-tty", false,
		"Disable interactive prompts; commands requiring confirmation will error unless --force is also set")

	// Register subcommands
	cmd.AddCommand(versioncmd.NewCmdVersion(f))
	cmd.AddCommand(completion.NewCmdCompletion(f))
	// Phase 2+: auth, repo, pr, branch, pipeline, issue, snippet, workspace commands added here

	return cmd, f
}

// Execute runs the root command with os.Args and exits with an appropriate code.
func Execute() {
	cmd, _ := NewCmdRoot()
	if err := cmd.Execute(); err != nil {
		var authErr *cmdutil.AuthError
		if errors.As(err, &authErr) {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
