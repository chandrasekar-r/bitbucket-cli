package browse

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/browser"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

// NewCmdBrowse creates the top-level `bb browse` command.
func NewCmdBrowse(f *cmdutil.Factory) *cobra.Command {
	var (
		branch   string
		prNum    int
		issueNum int
		pipeline string
		settings bool
		commits  bool
	)

	cmd := &cobra.Command{
		Use:   "browse",
		Short: "Open the current repository (or resource) in the browser",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			flagsSet := 0
			if branch != "" {
				flagsSet++
			}
			if prNum > 0 {
				flagsSet++
			}
			if issueNum > 0 {
				flagsSet++
			}
			if pipeline != "" {
				flagsSet++
			}
			if settings {
				flagsSet++
			}
			if commits {
				flagsSet++
			}
			if flagsSet > 1 {
				return &cmdutil.FlagError{Err: errors.New("only one of --branch, --pr, --issue, --pipeline, --settings, --commits may be set")}
			}

			var workspace, slug string
			if ctx := gitcontext.FromRemote(); ctx != nil {
				workspace, slug = ctx.Workspace, ctx.RepoSlug
			} else {
				return errors.New("no repo found: run from inside a cloned Bitbucket repository")
			}

			base := fmt.Sprintf("https://bitbucket.org/%s/%s", workspace, slug)
			var target string
			switch {
			case prNum > 0:
				target = fmt.Sprintf("%s/pull-requests/%d", base, prNum)
			case issueNum > 0:
				target = fmt.Sprintf("%s/issues/%d", base, issueNum)
			case pipeline != "":
				target = fmt.Sprintf("%s/pipelines/results/%s", base, pipeline)
			case settings:
				target = base + "/admin"
			case commits:
				target = base + "/commits/"
			case branch != "":
				target = base + "/src/" + branch
			default:
				target = base
			}

			fmt.Fprintf(f.IOStreams.Out, "Opening %s\n", target)
			return browser.Open(target)
		},
	}

	cmd.Flags().StringVar(&branch, "branch", "", "Open a branch view")
	cmd.Flags().IntVar(&prNum, "pr", 0, "Open a pull request by number")
	cmd.Flags().IntVar(&issueNum, "issue", 0, "Open an issue by number")
	cmd.Flags().StringVar(&pipeline, "pipeline", "", "Open a pipeline run by UUID")
	cmd.Flags().BoolVar(&settings, "settings", false, "Open repository settings")
	cmd.Flags().BoolVar(&commits, "commits", false, "Open commits view")
	return cmd
}