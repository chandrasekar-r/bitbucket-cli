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
			diff, err := client.GetPRDiff(workspace, slug, id)
			if err != nil {
				return fmt.Errorf("getting diff: %w", err)
			}
			fmt.Fprint(f.IOStreams.Out, diff)
			return nil
		},
	}
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}

func newCmdBrowse(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "browse <number>",
		Short: "Open pull request in the browser",
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
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}

