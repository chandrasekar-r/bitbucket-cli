package pipeline

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdRun(f *cmdutil.Factory) *cobra.Command {
	var branch, tag, commit string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Trigger a new pipeline run",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}

			// Default to current git branch
			if branch == "" && tag == "" && commit == "" {
				branch = currentBranch()
				if branch == "" {
					return &cmdutil.FlagError{Err: fmt.Errorf("could not detect current branch: use --branch, --tag, or --commit")}
				}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			p, err := client.TriggerPipeline(workspace, slug, api.TriggerPipelineOptions{
				Branch: branch,
				Tag:    tag,
				Commit: commit,
			})
			if err != nil {
				return fmt.Errorf("triggering pipeline: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Pipeline #%d triggered on %s\n", p.BuildNumber, refDescription(branch, tag, commit))
			fmt.Fprintf(f.IOStreams.Out, "  UUID: %s\n", p.UUID)
			fmt.Fprintf(f.IOStreams.Out, "  URL:  %s\n", p.Links.HTML.Href)
			fmt.Fprintf(f.IOStreams.Out, "\nRun `bb pipeline watch %s` to follow progress.\n", p.UUID)
			return nil
		},
	}

	cmd.Flags().StringVar(&branch, "branch", "", "Branch to run pipeline on (default: current branch)")
	cmd.Flags().StringVar(&tag, "tag", "", "Tag to run pipeline on")
	cmd.Flags().StringVar(&commit, "commit", "", "Commit hash to run pipeline on")
	return cmd
}

func currentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	b := strings.TrimSpace(string(out))
	if b == "HEAD" {
		return ""
	}
	return b
}

func refDescription(branch, tag, commit string) string {
	switch {
	case branch != "":
		return "branch " + branch
	case tag != "":
		return "tag " + tag
	default:
		return "commit " + commit
	}
}
