package repo

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view [workspace/repo]",
		Short: "Show repository details",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepo(f, args)
			if err != nil {
				return err
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			r, err := client.GetRepo(workspace, slug)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, r, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			vis := "public"
			if r.IsPrivate {
				vis = "private"
			}
			branch := "-"
			if r.MainBranch != nil {
				branch = r.MainBranch.Name
			}
			fmt.Fprintf(f.IOStreams.Out, "Repo:        %s\n", r.FullName)
			fmt.Fprintf(f.IOStreams.Out, "Description: %s\n", r.Description)
			fmt.Fprintf(f.IOStreams.Out, "Visibility:  %s\n", vis)
			fmt.Fprintf(f.IOStreams.Out, "Language:    %s\n", r.Language)
			fmt.Fprintf(f.IOStreams.Out, "Default:     %s\n", branch)
			fmt.Fprintf(f.IOStreams.Out, "HTTPS:       %s\n", r.CloneURL("https"))
			fmt.Fprintf(f.IOStreams.Out, "SSH:         %s\n", r.CloneURL("ssh"))
			fmt.Fprintf(f.IOStreams.Out, "URL:         %s\n", r.Links.HTML.Href)
			return nil
		},
	}

	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}

// resolveRepo extracts workspace and repo slug from either:
//   - A "workspace/slug" positional arg
//   - The current directory's git remote
//   - The configured default workspace + the arg as slug
func resolveRepo(f *cmdutil.Factory, args []string) (workspace, slug string, err error) {
	if len(args) == 1 {
		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
		// Bare slug — resolve workspace from context
		ws, werr := f.Workspace()
		if werr != nil {
			return "", "", werr
		}
		return ws, parts[0], nil
	}
	// No arg — infer from git remote
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return ctx.Workspace, ctx.RepoSlug, nil
	}
	ws, werr := f.Workspace()
	if werr != nil {
		return "", "", werr
	}
	return ws, "", fmt.Errorf("no repository specified. Use `bb repo %s workspace/repo`", "view")
}
