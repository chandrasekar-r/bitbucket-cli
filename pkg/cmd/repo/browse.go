package repo

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/browser"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

func newCmdBrowse(f *cmdutil.Factory) *cobra.Command {
	var branch string

	cmd := &cobra.Command{
		Use:   "browse [workspace/repo]",
		Short: "Open repository in the browser",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var workspace, slug string

			if len(args) == 1 {
				var err error
				workspace, slug, err = resolveRepo(f, args)
				if err != nil {
					return err
				}
			} else if ctx := gitcontext.FromRemote(); ctx != nil {
				workspace, slug = ctx.Workspace, ctx.RepoSlug
			} else {
				ws, err := f.Workspace()
				if err != nil {
					return err
				}
				return fmt.Errorf("no repo specified. Use: bb repo browse %s/<repo>", ws)
			}

			repoURL := fmt.Sprintf("https://bitbucket.org/%s/%s", workspace, slug)
			if branch != "" {
				repoURL += "/src/" + branch
			}

			fmt.Fprintf(f.IOStreams.Out, "Opening %s\n", repoURL)
			return browser.Open(repoURL)
		},
	}

	cmd.Flags().StringVar(&branch, "branch", "", "Open a specific branch")
	return cmd
}
