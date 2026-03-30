package pr

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		title, body, base string
		reviewers         []string
		draft             bool
		closeSourceBranch bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Infer repo from git context
			ctx := gitcontext.FromRemote()
			if ctx == nil {
				return fmt.Errorf("not inside a git repository with a Bitbucket remote")
			}

			// Detect current branch
			currentBranch, _ := currentGitBranch()

			// Fork detection: compare remote workspace to authenticated workspace
			destWorkspace := ctx.Workspace
			destSlug := ctx.RepoSlug
			srcWorkspace := ctx.Workspace
			srcSlug := ctx.RepoSlug

			store := bbauth.NewTokenStore()
			if acc, err := store.GetActive(); err == nil && acc != nil {
				if acc.Username != "" && ctx.Workspace != acc.Username {
					// Likely a fork — ask which destination
					if f.IOStreams.IsStdoutTTY() {
						var useUpstream bool
						form := huh.NewForm(huh.NewGroup(
							huh.NewConfirm().
								Title(fmt.Sprintf("This appears to be a fork. Create PR in upstream %s/%s?",
									ctx.Workspace, ctx.RepoSlug)).
								Value(&useUpstream),
						))
						if ferr := form.Run(); ferr == nil && !useUpstream {
							// PR within the fork
							destWorkspace = acc.Username
						}
					}
				}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			// Default base branch from repo
			if base == "" {
				repo, err := client.GetRepo(destWorkspace, destSlug)
				if err == nil && repo.MainBranch != nil {
					base = repo.MainBranch.Name
				} else {
					base = "main"
				}
			}

			// Interactive form when TTY and title not supplied
			if title == "" {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.FlagError{Err: fmt.Errorf("--title is required in --no-tty mode")}
				}
				// Pre-fill title from branch name
				defaultTitle := strings.ReplaceAll(currentBranch, "-", " ")
				defaultTitle = strings.ReplaceAll(defaultTitle, "_", " ")

				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("PR title").
							Value(&title).
							Placeholder(defaultTitle),
						huh.NewText().
							Title("Description (optional)").
							Value(&body),
						huh.NewInput().
							Title(fmt.Sprintf("Base branch (default: %s)", base)).
							Value(&base),
						huh.NewConfirm().
							Title("Close source branch after merge?").
							Value(&closeSourceBranch),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
				if title == "" {
					title = defaultTitle
				}
			}

			_ = draft // Bitbucket doesn't have a draft PR concept (as of v2 API)

			prOpts := api.CreatePROptions{
				Title:             title,
				Description:       body,
				SourceBranch:      currentBranch,
				SourceWorkspace:   srcWorkspace,
				SourceRepoSlug:    srcSlug,
				DestBranch:        base,
				DestWorkspace:     destWorkspace,
				DestRepoSlug:      destSlug,
				ReviewerUsernames: reviewers,
				CloseSourceBranch: closeSourceBranch,
			}

			created, err := client.CreatePR(destWorkspace, destSlug, prOpts)
			if err != nil {
				return fmt.Errorf("creating PR: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Created PR #%d: %s\n", created.ID, created.Title)
			fmt.Fprintf(f.IOStreams.Out, "  %s\n", created.Links.HTML.Href)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "PR title")
	cmd.Flags().StringVar(&body, "body", "", "PR description")
	cmd.Flags().StringVar(&base, "base", "", "Target branch (default: repo default branch)")
	cmd.Flags().StringArrayVar(&reviewers, "reviewer", nil, "Reviewer username (can be specified multiple times)")
	cmd.Flags().BoolVar(&draft, "draft", false, "Mark as draft (note: Bitbucket API does not support draft PRs)")
	cmd.Flags().BoolVar(&closeSourceBranch, "delete-branch", false, "Close source branch after merge")
	return cmd
}

// currentGitBranch returns the current git branch name.
func currentGitBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		return "", fmt.Errorf("not on a named branch (detached HEAD)")
	}
	return branch, nil
}

