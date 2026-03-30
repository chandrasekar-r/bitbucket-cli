package branch

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List branches in the current repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepoBranch(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			branches, err := client.ListBranches(workspace, slug, limit)
			if err != nil {
				return err
			}
			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, branches, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("NAME", "COMMIT", "DATE", "AUTHOR")
			for _, b := range branches {
				date := b.Target.Date
				if len(date) > 10 {
					date = date[:10]
				}
				author := b.Target.Author.User.Username
				if author == "" {
					author = b.Target.Author.Raw
					if len(author) > 30 {
						author = author[:27] + "..."
					}
				}
				t.AddRow(b.Name, b.ShortHash(), date, author)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d branch(es)\n", len(branches))
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of branches to list")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}

// resolveRepoBranch resolves workspace and repo slug from git context or flags.
func resolveRepoBranch(f *cmdutil.Factory) (workspace, slug string, err error) {
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return ctx.Workspace, ctx.RepoSlug, nil
	}
	ws, werr := f.Workspace()
	if werr != nil {
		return "", "", werr
	}
	return ws, "", fmt.Errorf("could not determine repository. Run from inside a cloned Bitbucket repo.")
}
