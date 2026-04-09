package branch

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
)

// resolveBranchName returns a branch name from args or an interactive picker.
func resolveBranchName(f *cmdutil.Factory, client *api.Client, workspace, slug string, args []string) (string, error) {
	if len(args) >= 1 {
		return args[0], nil
	}
	if !f.IOStreams.IsStdoutTTY() {
		return "", &cmdutil.FlagError{Err: fmt.Errorf("<name> argument required in non-interactive mode")}
	}
	branches, err := client.ListBranches(workspace, slug, 30)
	if err != nil {
		return "", fmt.Errorf("fetching branches: %w", err)
	}
	if len(branches) == 0 {
		return "", fmt.Errorf("no branches found")
	}
	items := make([]cmdutil.PickerItem, len(branches))
	for i, b := range branches {
		items[i] = cmdutil.PickerItem{Value: b.Name, Label: b.Name}
	}
	return cmdutil.RunPicker("Select a branch:", items)
}
