package pr

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdMerge(f *cmdutil.Factory) *cobra.Command {
	var strategy string
	var force bool
	var message string

	cmd := &cobra.Command{
		Use:   "merge <number>",
		Short: "Merge a pull request",
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

			// Strategy selection
			if strategy == "" {
				if !f.IOStreams.IsStdoutTTY() || force {
					strategy = "merge_commit"
				} else {
					form := huh.NewForm(huh.NewGroup(
						huh.NewSelect[string]().
							Title(fmt.Sprintf("Merge strategy for PR #%d", id)).
							Options(
								huh.NewOption("Merge commit", "merge_commit"),
								huh.NewOption("Squash", "squash"),
								huh.NewOption("Fast-forward", "fast_forward"),
							).
							Value(&strategy),
					))
					if err := form.Run(); err != nil {
						return err
					}
				}
			}

			// Confirm in TTY unless --force
			if !force && f.IOStreams.IsStdoutTTY() {
				var confirmed bool
				form := huh.NewForm(huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Merge PR #%d using %s?", id, strategy)).
						Value(&confirmed),
				))
				if err := form.Run(); err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(f.IOStreams.Out, "Cancelled")
					return nil
				}
			} else if !force && !f.IOStreams.IsStdoutTTY() {
				return &cmdutil.NoTTYError{Operation: fmt.Sprintf("merge PR #%d", id)}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			merged, err := client.MergePR(workspace, slug, id, api.MergeStrategy(strategy), message)
			if err != nil {
				return fmt.Errorf("merging PR: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Merged PR #%d: %s\n", merged.ID, merged.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&strategy, "strategy", "",
		"Merge strategy: merge_commit, squash, fast_forward (interactive when omitted)")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation")
	cmd.Flags().StringVar(&message, "message", "", "Custom merge commit message")
	return cmd
}
