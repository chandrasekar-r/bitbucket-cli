package snippet

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func validateSnippetCloneURL(rawURL string) error {
	if strings.HasPrefix(rawURL, "-") {
		return fmt.Errorf("clone URL must not begin with '-': %q", rawURL)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid clone URL: %w", err)
	}
	allowed := map[string]bool{"https": true, "ssh": true, "git": true}
	if !allowed[parsed.Scheme] {
		return fmt.Errorf("clone URL has unsafe scheme %q", parsed.Scheme)
	}
	return nil
}

func newCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !f.IOStreams.IsStdoutTTY() && !force {
				return &cmdutil.NoTTYError{Operation: "delete snippet " + id}
			}

			if !force && f.IOStreams.IsStdoutTTY() {
				var confirmed bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete snippet %q?", id)).
						Value(&confirmed),
				))
				if err := form.Run(); err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(f.IOStreams.Out, "Cancelled")
					return nil
				}
			}

			workspace, _ := f.Workspace()
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			if err := client.DeleteSnippet(workspace, id); err != nil {
				return fmt.Errorf("deleting snippet: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Deleted snippet %s\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}

func newCmdClone(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone <id>",
		Short: "Clone a snippet as a git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			workspace, _ := f.Workspace()

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			s, err := client.GetSnippet(workspace, id)
			if err != nil {
				return err
			}

			cloneURL := s.CloneURL("https")
			if cloneURL == "" {
				return fmt.Errorf("no HTTPS clone URL available for snippet %s", id)
			}

			// Validate URL scheme before passing to git (FINDING-003)
			if err := validateSnippetCloneURL(cloneURL); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Cloning snippet %s...\n", id)
			gitCmd := exec.Command("git", "clone", "--", cloneURL)
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr
			return gitCmd.Run()
		},
	}
	return cmd
}
