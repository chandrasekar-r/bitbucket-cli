package branch

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// tidyCandidate holds a local branch that has a merged or declined PR.
type tidyCandidate struct {
	Branch string
	State  string // "MERGED" or "DECLINED"
	PRID   int
	PRTitle string
}

func newCmdTidy(f *cmdutil.Factory) *cobra.Command {
	var dryRun, force bool

	cmd := &cobra.Command{
		Use:   "tidy",
		Short: "Delete local branches whose PRs have been merged or declined",
		Long: `Scan local branches and check Bitbucket for associated pull requests.
Branches whose PRs are MERGED or DECLINED are candidates for deletion.

The current branch and the repository's default branch are always skipped.`,
		Args: cobra.NoArgs,
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

			// Get repository default branch.
			repo, err := client.GetRepo(workspace, slug)
			if err != nil {
				return fmt.Errorf("fetching repo info: %w", err)
			}
			defaultBranch := "main"
			if repo.MainBranch != nil && repo.MainBranch.Name != "" {
				defaultBranch = repo.MainBranch.Name
			}

			// Get current branch.
			current, err := currentGitBranch()
			if err != nil {
				return fmt.Errorf("determining current branch: %w", err)
			}

			// List local branches.
			locals, err := localGitBranches()
			if err != nil {
				return fmt.Errorf("listing local branches: %w", err)
			}

			// Check each branch for merged/declined PRs.
			var candidates []tidyCandidate
			for _, b := range locals {
				if b == current || b == defaultBranch {
					continue
				}
				prs, err := client.ListPRsForBranch(workspace, slug, b, "")
				if err != nil {
					fmt.Fprintf(f.IOStreams.ErrOut, "warning: could not check branch %q: %v\n", b, err)
					continue
				}
				for _, pr := range prs {
					if pr.State == "MERGED" || pr.State == "DECLINED" {
						candidates = append(candidates, tidyCandidate{
							Branch:  b,
							State:   pr.State,
							PRID:    pr.ID,
							PRTitle: pr.Title,
						})
						break // one match per branch is enough
					}
				}
			}

			if len(candidates) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No stale branches found.")
				return nil
			}

			// Display candidates grouped by state.
			merged := filterByState(candidates, "MERGED")
			declined := filterByState(candidates, "DECLINED")

			if len(merged) > 0 {
				fmt.Fprintln(f.IOStreams.Out, "\nMerged:")
				for _, c := range merged {
					fmt.Fprintf(f.IOStreams.Out, "  %s (PR #%d: %s)\n", c.Branch, c.PRID, c.PRTitle)
				}
			}
			if len(declined) > 0 {
				fmt.Fprintln(f.IOStreams.Out, "\nDeclined:")
				for _, c := range declined {
					fmt.Fprintf(f.IOStreams.Out, "  %s (PR #%d: %s)\n", c.Branch, c.PRID, c.PRTitle)
				}
			}
			fmt.Fprintln(f.IOStreams.Out)

			if dryRun {
				fmt.Fprintf(f.IOStreams.Out, "Would delete %d branch(es). Use without --dry-run to delete.\n", len(candidates))
				return nil
			}

			// Confirm deletion.
			if !force {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.NoTTYError{Operation: fmt.Sprintf("delete %d local branch(es)", len(candidates))}
				}
				var confirmed bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete %d local branch(es)?", len(candidates))).
						Value(&confirmed),
				))
				if err := form.Run(); err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(f.IOStreams.Out, "Cancelled")
					return nil
				}
			}

			// Delete branches.
			deleted := 0
			for _, c := range candidates {
				if err := deleteLocalBranch(c.Branch); err != nil {
					fmt.Fprintf(f.IOStreams.ErrOut, "✗ Failed to delete %s: %v\n", c.Branch, err)
					continue
				}
				fmt.Fprintf(f.IOStreams.Out, "✓ Deleted %s\n", c.Branch)
				deleted++
			}
			fmt.Fprintf(f.IOStreams.Out, "\n%d branch(es) deleted.\n", deleted)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without deleting")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation (required in --no-tty mode)")
	return cmd
}

// filterByState returns candidates matching the given state.
func filterByState(candidates []tidyCandidate, state string) []tidyCandidate {
	var out []tidyCandidate
	for _, c := range candidates {
		if c.State == state {
			out = append(out, c)
		}
	}
	return out
}

// currentGitBranch returns the name of the currently checked-out branch.
func currentGitBranch() (string, error) {
	out, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// localGitBranches returns all local branch names.
func localGitBranches() ([]string, error) {
	out, err := exec.Command("git", "branch", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var branches []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			branches = append(branches, l)
		}
	}
	return branches, nil
}

// deleteLocalBranch force-deletes a local git branch.
func deleteLocalBranch(name string) error {
	return exec.Command("git", "branch", "-D", "--", name).Run()
}
