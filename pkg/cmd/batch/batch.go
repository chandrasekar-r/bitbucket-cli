package batch

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// BatchOptions holds the configuration for a batch run.
type BatchOptions struct {
	IO          *cmdutil.Factory
	HttpClient  func() (*api.Client, error)
	Workspace   func() (string, error)
	Repos       string
	Concurrency int
	Args        []string
}

// NewCmdBatch creates the `bb batch` command.
func NewCmdBatch(f *cmdutil.Factory) *cobra.Command {
	opts := &BatchOptions{}
	if f != nil {
		opts.IO = f
		opts.Workspace = f.Workspace
		opts.HttpClient = func() (*api.Client, error) {
			httpClient, err := f.HttpClient()
			if err != nil {
				return nil, err
			}
			return api.New(httpClient, f.BaseURL), nil
		}
	}

	cmd := &cobra.Command{
		Use:   "batch -- <command> [flags]",
		Short: "Run a bb command across multiple repos in a workspace",
		Long: `Run any bb command across all (or filtered) repositories in a workspace.

The command after -- is executed once per repository with BITBUCKET_WORKSPACE and
BITBUCKET_REPO environment variables set. This works best with commands that read
those env vars (e.g. bb api) or accept --workspace/--repo flags.

Commands that infer the repo from git context (e.g. bb pr list without flags)
will fail gracefully for repos not cloned locally.`,
		Example: `  # List open PRs across all repos
  bb batch -- pr list --state OPEN

  # Filter to repos matching a glob
  bb batch --repos "backend-*" -- pipeline list --limit 5

  # View repo metadata as JSON
  bb batch --workspace myteam -- repo view --json full_name,language`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			return runBatch(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Repos, "repos", "", "Glob pattern to filter repository slugs (e.g. \"backend-*\")")
	cmd.Flags().IntVar(&opts.Concurrency, "concurrency", 5, "Max parallel command executions")

	return cmd
}

func runBatch(opts *BatchOptions) error {
	if len(opts.Args) == 0 {
		return fmt.Errorf("no command specified; use: bb batch -- <command> [flags]")
	}

	ws, err := opts.Workspace()
	if err != nil {
		return fmt.Errorf("resolving workspace: %w", err)
	}

	client, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("creating API client: %w", err)
	}

	repos, err := client.ListRepos(ws, api.ListReposOptions{Limit: 0})
	if err != nil {
		return fmt.Errorf("listing repos: %w", err)
	}

	// Filter repos if pattern is given.
	if opts.Repos != "" {
		repos, err = filterRepos(repos, opts.Repos)
		if err != nil {
			return err
		}
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories matched")
	}

	concurrency := max(1, opts.Concurrency)
	sem := make(chan struct{}, concurrency)

	type result struct {
		slug   string
		output string
		err    error
	}

	results := make([]result, len(repos))
	var wg sync.WaitGroup

	bbBinary := os.Args[0]

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, slug string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			out, execErr := execForRepo(bbBinary, ws, slug, opts.Args)
			results[idx] = result{slug: slug, output: out, err: execErr}
		}(i, repo.Slug)
	}

	wg.Wait()

	// Print results in order.
	ios := opts.IO.IOStreams
	for _, r := range results {
		header := fmt.Sprintf("── %s/%s ", ws, r.slug)
		header += strings.Repeat("─", max(0, 60-len(header)))
		fmt.Fprintln(ios.Out, header)

		if r.err != nil {
			fmt.Fprintf(ios.ErrOut, "  error: %s\n", r.err)
		}
		if r.output != "" {
			fmt.Fprint(ios.Out, r.output)
			if !strings.HasSuffix(r.output, "\n") {
				fmt.Fprintln(ios.Out)
			}
		}
	}

	return nil
}

// execForRepo runs the bb binary as a subprocess for a single repo.
func execForRepo(binary, workspace, repoSlug string, args []string) (string, error) {
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(),
		"BITBUCKET_WORKSPACE="+workspace,
		"BITBUCKET_REPO="+repoSlug,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return output, fmt.Errorf("%s", errMsg)
	}
	return output, nil
}

// filterRepos returns repos whose slugs match the given glob pattern.
func filterRepos(repos []api.Repository, pattern string) ([]api.Repository, error) {
	var matched []api.Repository
	for _, r := range repos {
		ok, err := filepath.Match(pattern, r.Slug)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		if ok {
			matched = append(matched, r)
		}
	}
	return matched, nil
}
