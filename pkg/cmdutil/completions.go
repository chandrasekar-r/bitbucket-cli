package cmdutil

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/spf13/cobra"
)

// CompletePRIDs returns a ValidArgsFunction that completes PR numbers with titles.
func CompletePRIDs(f *Factory, state string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		ctx := gitcontext.FromRemote()
		if ctx == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		httpClient, err := f.HttpClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		client := api.New(httpClient, f.BaseURL)
		prs, err := client.ListPRs(ctx.Workspace, ctx.RepoSlug, api.ListPRsOptions{
			State: state, Limit: 30,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, pr := range prs {
			id := fmt.Sprintf("%d", pr.ID)
			if strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%d\t%s", pr.ID, pr.Title))
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteBranchNames returns a ValidArgsFunction that completes branch names.
func CompleteBranchNames(f *Factory) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		ctx := gitcontext.FromRemote()
		if ctx == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		httpClient, err := f.HttpClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		client := api.New(httpClient, f.BaseURL)
		branches, err := client.ListBranches(ctx.Workspace, ctx.RepoSlug, 30)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, b := range branches {
			if strings.HasPrefix(b.Name, toComplete) {
				completions = append(completions, b.Name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompleteRepoNames returns a ValidArgsFunction that completes repository names.
func CompleteRepoNames(f *Factory) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		ws, err := f.Workspace()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		httpClient, err := f.HttpClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		client := api.New(httpClient, f.BaseURL)
		repos, err := client.ListRepos(ws, api.ListReposOptions{Limit: 30})
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, r := range repos {
			if strings.HasPrefix(r.Slug, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", r.Slug, r.Description))
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
