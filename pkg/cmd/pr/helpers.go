package pr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/gitcontext"
)

// repoContext resolves the workspace and repo slug for the current context.
func repoContext(f *cmdutil.Factory) (workspace, slug string, err error) {
	if ctx := gitcontext.FromRemote(); ctx != nil {
		return ctx.Workspace, ctx.RepoSlug, nil
	}
	ws, werr := f.Workspace()
	if werr != nil {
		return "", "", werr
	}
	return ws, "", fmt.Errorf("no repo found: run from inside a cloned Bitbucket repository")
}

// parsePRID parses a PR number from a string argument.
func parsePRID(s string) (int, error) {
	s = strings.TrimPrefix(s, "#")
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid PR number %q", s)
	}
	return id, nil
}

// resolvePRID returns a PR ID from args or an interactive picker.
func resolvePRID(f *cmdutil.Factory, client *api.Client, workspace, slug string, args []string) (int, error) {
	if len(args) >= 1 {
		return parsePRID(args[0])
	}
	if !f.IOStreams.IsStdoutTTY() {
		return 0, &cmdutil.FlagError{Err: fmt.Errorf("<number> argument required in non-interactive mode")}
	}
	prs, err := client.ListPRs(workspace, slug, api.ListPRsOptions{State: "OPEN", Limit: 30})
	if err != nil {
		return 0, fmt.Errorf("fetching PRs for selection: %w", err)
	}
	if len(prs) == 0 {
		return 0, fmt.Errorf("no open pull requests found")
	}
	items := make([]cmdutil.PickerItem, len(prs))
	for i, pr := range prs {
		items[i] = cmdutil.PickerItem{
			Value: strconv.Itoa(pr.ID),
			Label: fmt.Sprintf("#%-4d %s (%s → %s)", pr.ID, pr.Title, pr.Source.Branch.Name, pr.Destination.Branch.Name),
		}
	}
	selected, err := cmdutil.RunPicker("Select a pull request:", items)
	if err != nil {
		return 0, err
	}
	id, _ := strconv.Atoi(selected)
	return id, nil
}
