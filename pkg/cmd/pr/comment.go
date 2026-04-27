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
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdComment(f *cmdutil.Factory) *cobra.Command {
	var body string
	var formatHelp bool
	var filePath string
	var lineNum int

	cmd := &cobra.Command{
		Use:   "comment <number>",
		Short: "Add a comment to a pull request",
		Long: `Add a comment to a pull request.

Comment body supports Bitbucket Markdown. Run --format-help to see available syntax.

To add an inline diff comment anchored to a file or specific line:

  bb pr comment 42 --body "why?" --file pkg/api/prs.go --line 10
  bb pr comment 42 --body "whole file" --file pkg/api/prs.go`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if formatHelp {
				fmt.Fprint(f.IOStreams.Out, output.BitbucketMarkdownGuide())
				return nil
			}

			// Resolve inline flag presence using Changed() — value comparison
			// is unreliable (int default 0, string default "").
			fileSet := cmd.Flags().Changed("file")
			lineSet := cmd.Flags().Changed("line")

			// Validate flag combinations before touching the network.
			if fileSet && filePath == "" {
				return &cmdutil.FlagError{Err: errors.New("--file cannot be empty")}
			}
			if lineSet && !fileSet {
				return &cmdutil.FlagError{Err: errors.New("--line requires --file")}
			}
			if lineSet && lineNum < 1 {
				return &cmdutil.FlagError{Err: fmt.Errorf("--line must be a positive integer (1-based diff line), got %d", lineNum)}
			}

			isInline := fileSet

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
					return &cmdutil.FlagError{Err: errors.New("--body is required; use --body or run without --no-tty to open an editor")}
				}
				body, err = openEditor()
				if err != nil {
					return err
				}
			}

			body = strings.TrimSpace(body)
			if body == "" {
				return errors.New("comment body cannot be empty")
			}

			if isInline {
				inline := api.InlineComment{Path: filePath}
				if lineSet {
					inline.To = &lineNum
				}
				if err := client.AddPRInlineComment(workspace, slug, id, body, inline); err != nil {
					return fmt.Errorf("adding inline comment: %w", err)
				}
				if lineSet {
					fmt.Fprintf(f.IOStreams.Out, "✓ Inline comment added to %s:%d in PR #%d\n", filePath, lineNum, id)
				} else {
					fmt.Fprintf(f.IOStreams.Out, "✓ Inline comment added to %s in PR #%d\n", filePath, id)
				}
				return nil
			}

			if err := client.AddPRComment(workspace, slug, id, body); err != nil {
				return fmt.Errorf("adding comment: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Comment added to PR #%d\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Comment text (supports Bitbucket Markdown)")
	cmd.Flags().BoolVar(&formatHelp, "format-help", false, "Show Bitbucket Markdown formatting reference and exit")
	cmd.Flags().StringVar(&filePath, "file", "", "Repo-relative file path for an inline diff comment")
	cmd.Flags().IntVar(&lineNum, "line", 0, "Line number on the destination/head side of the diff (1-based, requires --file)")
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
