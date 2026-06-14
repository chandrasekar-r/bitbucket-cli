package pr

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdChecks(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "checks <number>",
		Short: "Show commit/build statuses for a pull request",
		Long: `Show CI and external build statuses for a pull request's source commit.

Statuses come from the Bitbucket commit statuses API (pipelines, Jenkins, Sonar, etc.).`,
		Args: cobra.MaximumNArgs(1),
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
			commitHash := pr.Source.Commit.Hash
			if commitHash == "" {
				return fmt.Errorf("PR #%d has no source commit hash; cannot fetch build statuses", id)
			}

			statuses, err := client.ListCommitStatuses(workspace, slug, commitHash, 0)
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, statuses, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			if len(statuses) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No checks found for this pull request")
				return nil
			}

			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("KEY", "STATE", "NAME", "DESCRIPTION", "URL")
			for _, s := range statuses {
				desc := s.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
				t.AddRow(
					s.Key,
					checkStateColor(s.State, f.IOStreams.ColorEnabled()),
					s.Name,
					desc,
					s.URL,
				)
			}
			return t.Render()
		},
	}

	jsonOpts = cmdutil.AddJSONFlags(cmd)
	cmd.ValidArgsFunction = cmdutil.CompletePRIDs(f, "OPEN")
	return cmd
}

func checkStateColor(state string, color bool) string {
	if !color {
		return state
	}
	switch strings.ToUpper(state) {
	case "SUCCESSFUL":
		return "\033[32m" + state + "\033[0m"
	case "FAILED", "STOPPED":
		return "\033[31m" + state + "\033[0m"
	case "INPROGRESS":
		return "\033[36m" + state + "\033[0m"
	default:
		return state
	}
}