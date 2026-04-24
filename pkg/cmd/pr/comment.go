package pr

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdComment(f *cmdutil.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "comment <number>",
		Short: "Add a comment to a pull request",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			id, err := resolvePRID(f, client, workspace, slug, args)
			if err != nil {
				return err
			}

			if body == "" {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.FlagError{Err: errors.New("--body is required in --no-tty mode")}
				}
				// Open $EDITOR for comment
				body, err = openEditor()
				if err != nil {
					return err
				}
			}

			body = strings.TrimSpace(body)
			if body == "" {
				return errors.New("comment body cannot be empty")
			}

			if err := client.AddPRComment(workspace, slug, id, body); err != nil {
				return fmt.Errorf("adding comment: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Comment added to PR #%d\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Comment text")
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}

// openEditor opens $EDITOR and returns the content the user typed.
func openEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Fallback: read from stdin with a prompt
		fmt.Print("Comment (end with Ctrl+D):\n")
		var sb strings.Builder
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			sb.WriteString(scanner.Text() + "\n")
		}
		return sb.String(), nil
	}

	f, err := os.CreateTemp("", "bb-comment-*.md")
	if err != nil {
		return "", err
	}
	fname := f.Name()
	f.Close()
	defer os.Remove(fname)

	// Split $EDITOR on whitespace so "nano -R" or "code --wait" work correctly.
	// exec.Command does not invoke a shell, so without splitting, EDITOR="/usr/bin/env bash -c payload"
	// would pass "bash -c payload" as arguments to /usr/bin/env — an injection path.
	parts := strings.Fields(editor)
	editorArgs := append(parts[1:], fname)
	cmd := exec.Command(parts[0], editorArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
