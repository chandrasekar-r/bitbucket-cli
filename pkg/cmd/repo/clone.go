package repo

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// allowedCloneSchemes are the only URL schemes accepted for git clone.
// This prevents git from interpreting ext:: or file:// URLs that could
// execute arbitrary commands (FINDING-003).
var allowedCloneSchemes = map[string]bool{
	"https": true,
	"ssh":   true,
	"git":   true,
}

// validateCloneURL rejects URLs with dangerous schemes or leading dashes
// (git flag injection, same class as CVE-2017-1000117).
func validateCloneURL(rawURL string) error {
	if strings.HasPrefix(rawURL, "-") {
		return fmt.Errorf("clone URL must not begin with '-': %q", rawURL)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid clone URL %q: %w", rawURL, err)
	}
	if !allowedCloneSchemes[parsed.Scheme] {
		return fmt.Errorf("clone URL has unsafe scheme %q (allowed: https, ssh, git)", parsed.Scheme)
	}
	return nil
}

func newCmdClone(f *cmdutil.Factory) *cobra.Command {
	var useSSH bool

	cmd := &cobra.Command{
		Use:   "clone <workspace/repo> [directory]",
		Short: "Clone a repository locally",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepo(f, args[:1])
			if err != nil {
				return err
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			r, err := client.GetRepo(workspace, slug)
			if err != nil {
				return err
			}

			protocol := "https"
			if useSSH {
				protocol = "ssh"
			}
			cloneURL := r.CloneURL(protocol)
			if cloneURL == "" {
				return fmt.Errorf("no %s clone URL for %s/%s", protocol, workspace, slug)
			}

			// FINDING-003: validate URL scheme before passing to git
			if err := validateCloneURL(cloneURL); err != nil {
				return err
			}

			gitArgs := []string{"clone", "--", cloneURL}
			if len(args) == 2 {
				gitArgs = append(gitArgs, args[1])
			}

			fmt.Fprintf(f.IOStreams.Out, "Cloning %s/%s...\n", workspace, slug)
			gitCmd := exec.Command("git", gitArgs...)
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr
			return gitCmd.Run()
		},
	}

	cmd.Flags().BoolVar(&useSSH, "ssh", false, "Clone using SSH instead of HTTPS")
	return cmd
}
