package pr

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/browser"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdDiff(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <number>",
		Short: "Show the pull request diff",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parsePRID(args[0])
			if err != nil {
				return err
			}
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			diff, err := client.GetPRDiff(workspace, slug, id)
			if err != nil {
				return fmt.Errorf("getting diff: %w", err)
			}
			fmt.Fprint(f.IOStreams.Out, diff)
			return nil
		},
	}
	return cmd
}

func newCmdBrowse(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "browse <number>",
		Short: "Open pull request in the browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parsePRID(args[0])
			if err != nil {
				return err
			}
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			pr, err := client.GetPR(workspace, slug, id)
			if err != nil {
				return err
			}
			prURL := pr.Links.HTML.Href
			if prURL == "" {
				prURL = fmt.Sprintf("https://bitbucket.org/%s/%s/pull-requests/%d",
					workspace, slug, id)
			}
			fmt.Fprintf(f.IOStreams.Out, "Opening %s\n", prURL)
			return browser.Open(prURL)
		},
	}
	return cmd
}

// openBrowser opens a URL in the system browser (wraps browser.Open for use in comment.go).
func openBrowser(url string) error {
	return browser.Open(url)
}
