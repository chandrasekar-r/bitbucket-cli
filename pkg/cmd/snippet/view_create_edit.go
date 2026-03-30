package snippet

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "Show snippet content and metadata",
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

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, s, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			w := f.IOStreams.Out
			vis := "public"
			if s.IsPrivate {
				vis = "private"
			}
			fmt.Fprintf(w, "Snippet: %s\n", s.Title)
			fmt.Fprintf(w, "ID:      %s\n", s.ID)
			fmt.Fprintf(w, "Owner:   %s\n", s.Owner.Username)
			fmt.Fprintf(w, "Files:   %d\n", s.FileCount())
			fmt.Fprintf(w, "Visibility: %s\n", vis)
			fmt.Fprintf(w, "URL:     %s\n", s.Links.HTML.Href)
			fmt.Fprintf(w, "Clone:   %s\n", s.CloneURL("https"))

			if len(s.Files) > 0 {
				fmt.Fprintln(w, "\nFiles:")
				for name := range s.Files {
					fmt.Fprintf(w, "  %s\n", name)
				}
			}
			return nil
		},
	}
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var title, filename string
	var private bool
	var fromFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new snippet",
		Long: `Create a new Bitbucket Snippet.

Content is read from stdin or --file:

  echo "hello world" | bb snippet create --title "My Snippet"
  bb snippet create --file ./myscript.sh --title "Deploy script"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("--title is required")}
			}

			var content string
			if fromFile != "" {
				data, err := os.ReadFile(fromFile)
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
				content = string(data)
				if filename == "" {
					// Use the file basename
					parts := strings.Split(fromFile, "/")
					filename = parts[len(parts)-1]
				}
			} else {
				data, err := io.ReadAll(f.IOStreams.In)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				content = string(data)
			}

			if strings.TrimSpace(content) == "" {
				return fmt.Errorf("snippet content cannot be empty: pipe content via stdin or use --file")
			}

			workspace, _ := f.Workspace()
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			s, err := client.CreateSnippet(api.CreateSnippetOptions{
				Title:     title,
				Content:   content,
				Filename:  filename,
				IsPrivate: private,
				Workspace: workspace,
			})
			if err != nil {
				return fmt.Errorf("creating snippet: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Created snippet %s\n", s.ID)
			fmt.Fprintf(f.IOStreams.Out, "  %s\n", s.Links.HTML.Href)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Snippet title (required)")
	cmd.Flags().StringVar(&filename, "filename", "", "Filename within the snippet")
	cmd.Flags().BoolVar(&private, "private", true, "Make snippet private")
	cmd.Flags().StringVar(&fromFile, "file", "", "Read content from file instead of stdin")
	return cmd
}

func newCmdEdit(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a snippet in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			workspace, _ := f.Workspace()

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = os.Getenv("VISUAL")
			}
			if editor == "" {
				return fmt.Errorf("no editor configured: set the $EDITOR environment variable")
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			s, err := client.GetSnippet(workspace, id)
			if err != nil {
				return err
			}

			// Write first file content to temp file for editing
			var firstFile string
			for name := range s.Files {
				firstFile = name
				break
			}
			if firstFile == "" {
				return fmt.Errorf("snippet has no files")
			}

			// Fetch file content via the self link
			resp, err := client.GetRaw(
				fmt.Sprintf("/snippets/%s/%s/files/%s", workspace, id, firstFile),
				nil,
			)
			if err != nil {
				return fmt.Errorf("fetching snippet content: %w", err)
			}
			original, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			tmp, err := os.CreateTemp("", "bb-snippet-*-"+firstFile)
			if err != nil {
				return err
			}
			tmpName := tmp.Name()
			_, _ = tmp.Write(original)
			tmp.Close()
			defer os.Remove(tmpName)

			editorCmd := exec.Command(editor, tmpName)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			if err := editorCmd.Run(); err != nil {
				return err
			}

			updated, err := os.ReadFile(tmpName)
			if err != nil {
				return err
			}
			if string(updated) == string(original) {
				fmt.Fprintln(f.IOStreams.Out, "No changes made")
				return nil
			}

			// PUT updated content
			body := map[string]interface{}{
				"files": map[string]interface{}{
					firstFile: map[string]interface{}{
						"content": string(updated),
					},
				},
			}
			path := fmt.Sprintf("/snippets/%s/%s", workspace, id)
			if err := client.Put(path, body, nil); err != nil {
				return fmt.Errorf("updating snippet: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Snippet %s updated\n", id)
			return nil
		},
	}
	return cmd
}
