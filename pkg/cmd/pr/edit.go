package pr

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdEdit(f *cmdutil.Factory) *cobra.Command {
	var (
		title        string
		body         string
		base         string
		addReviewers []string
	)

	cmd := &cobra.Command{
		Use:   "edit <number>",
		Short: "Edit a pull request",
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

			hasFlags := cmd.Flags().Changed("title") ||
				cmd.Flags().Changed("body") ||
				cmd.Flags().Changed("base") ||
				cmd.Flags().Changed("add-reviewer")

			if !hasFlags {
				// Interactive mode
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.NoTTYError{Operation: "bb pr edit"}
				}
				return editPRInteractive(client, f, workspace, slug, id)
			}

			// Flag-based mode
			opts := api.UpdatePROptions{}

			if cmd.Flags().Changed("title") {
				opts.Title = &title
			}
			if cmd.Flags().Changed("body") {
				opts.Description = &body
			}
			if cmd.Flags().Changed("base") {
				opts.DestBranch = &base
			}

			// --add-reviewer appends to existing reviewers
			if cmd.Flags().Changed("add-reviewer") {
				existing, err := client.GetPR(workspace, slug, id)
				if err != nil {
					return err
				}
				seen := make(map[string]bool)
				var all []string
				for _, r := range existing.Reviewers {
					seen[r.Username] = true
					all = append(all, r.Username)
				}
				for _, u := range addReviewers {
					if !seen[u] {
						all = append(all, u)
					}
				}
				opts.ReviewerUsernames = all
			}

			updated, err := client.UpdatePR(workspace, slug, id, opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Updated PR #%d: %s\n", updated.ID, updated.Title)
			fmt.Fprintf(f.IOStreams.Out, "  %s\n", updated.Links.HTML.Href)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title for the pull request")
	cmd.Flags().StringVar(&body, "body", "", "New description for the pull request")
	cmd.Flags().StringVar(&base, "base", "", "New target branch")
	cmd.Flags().StringArrayVar(&addReviewers, "add-reviewer", nil, "Add reviewer (can be specified multiple times)")
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}

// editPRInteractive opens a huh form pre-filled with the current PR values.
// Only fields that the user actually changes are sent in the update.
func editPRInteractive(client *api.Client, f *cmdutil.Factory, workspace, slug string, id int) error {
	pr, err := client.GetPR(workspace, slug, id)
	if err != nil {
		return err
	}

	// Pre-fill from current PR
	newTitle := pr.Title
	newDescription := pr.Description
	newBase := pr.Destination.Branch.Name

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&newTitle),
			huh.NewText().
				Title("Description").
				Value(&newDescription),
			huh.NewInput().
				Title("Base branch").
				Value(&newBase),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	// Only send changed fields
	opts := api.UpdatePROptions{}
	changed := false
	if newTitle != pr.Title {
		opts.Title = &newTitle
		changed = true
	}
	if newDescription != pr.Description {
		opts.Description = &newDescription
		changed = true
	}
	if newBase != pr.Destination.Branch.Name {
		opts.DestBranch = &newBase
		changed = true
	}

	if !changed {
		fmt.Fprintln(f.IOStreams.Out, "No changes made.")
		return nil
	}

	updated, err := client.UpdatePR(workspace, slug, id, opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "✓ Updated PR #%d: %s\n", updated.ID, updated.Title)
	fmt.Fprintf(f.IOStreams.Out, "  %s\n", updated.Links.HTML.Href)
	return nil
}
