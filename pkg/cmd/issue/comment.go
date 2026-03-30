package issue

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdComment(f *cmdutil.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "comment <number>",
		Short: "Add a comment to an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return &cmdutil.FlagError{Err: fmt.Errorf("invalid issue number %q", args[0])}
			}

			if body == "" {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.FlagError{Err: errors.New("--body is required in --no-tty mode")}
				}
				body, err = openEditorForComment()
				if err != nil {
					return err
				}
			}

			body = strings.TrimSpace(body)
			if body == "" {
				return errors.New("comment body cannot be empty")
			}

			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			client, err := issueClient(f, workspace, slug)
			if err != nil {
				return err
			}
			if err := client.AddIssueComment(workspace, slug, id, body); err != nil {
				return fmt.Errorf("adding comment: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Comment added to issue #%d\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Comment text")
	return cmd
}

func openEditorForComment() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		fmt.Print("Comment (end with Ctrl+D):\n")
		var sb strings.Builder
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			sb.WriteString(scanner.Text() + "\n")
		}
		return sb.String(), nil
	}
	f, err := os.CreateTemp("", "bb-issue-comment-*.md")
	if err != nil {
		return "", err
	}
	fname := f.Name()
	f.Close()
	defer os.Remove(fname)

	parts := strings.Fields(editor)
	editorArgs := append(parts[1:], fname)
	editorCmd := exec.Command(parts[0], editorArgs...)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
