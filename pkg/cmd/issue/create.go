package issue

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var title, body, kind, priority, assignee string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			client, err := issueClient(f, workspace, slug)
			if err != nil {
				return err
			}

			if title == "" {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.FlagError{Err: fmt.Errorf("--title is required in --no-tty mode")}
				}
				if kind == "" {
					kind = "bug"
				}
				if priority == "" {
					priority = "major"
				}
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Issue title").Value(&title),
						huh.NewText().Title("Description (optional)").Value(&body),
						huh.NewSelect[string]().Title("Kind").
							Options(
								huh.NewOption("Bug", "bug"),
								huh.NewOption("Enhancement", "enhancement"),
								huh.NewOption("Proposal", "proposal"),
								huh.NewOption("Task", "task"),
							).Value(&kind),
						huh.NewSelect[string]().Title("Priority").
							Options(
								huh.NewOption("Blocker", "blocker"),
								huh.NewOption("Critical", "critical"),
								huh.NewOption("Major", "major"),
								huh.NewOption("Minor", "minor"),
								huh.NewOption("Trivial", "trivial"),
							).Value(&priority),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
			}

			if kind == "" {
				kind = "bug"
			}
			if priority == "" {
				priority = "major"
			}

			i, err := client.CreateIssue(workspace, slug, api.CreateIssueOptions{
				Title:    title,
				Content:  body,
				Kind:     kind,
				Priority: priority,
				Assignee: assignee,
			})
			if err != nil {
				return fmt.Errorf("creating issue: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Created issue #%d: %s\n", i.ID, i.Title)
			fmt.Fprintf(f.IOStreams.Out, "  %s\n", i.Links.HTML.Href)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Issue title")
	cmd.Flags().StringVar(&body, "body", "", "Issue description")
	cmd.Flags().StringVar(&kind, "kind", "", "Kind: bug, enhancement, proposal, task")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority: trivial, minor, major, critical, blocker")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee username")
	return cmd
}
