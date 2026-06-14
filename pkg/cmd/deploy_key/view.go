package deploykeycmd

import (
	"fmt"
	"strconv"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a deploy key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid deploy key ID %q", args[0])
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			key, err := client.GetDeployKey(workspace, slug, id)
			if err != nil {
				return err
			}
			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, key, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			fmt.Fprintf(f.IOStreams.Out, "ID:          %d\n", key.ID)
			fmt.Fprintf(f.IOStreams.Out, "Label:       %s\n", key.Label)
			fmt.Fprintf(f.IOStreams.Out, "Fingerprint: %s\n", key.Fingerprint)
			fmt.Fprintf(f.IOStreams.Out, "Comment:     %s\n", key.Comment)
			fmt.Fprintf(f.IOStreams.Out, "Created:     %s\n", key.CreatedOn)
			fmt.Fprintf(f.IOStreams.Out, "Last used:   %s\n", key.LastUsed)
			return nil
		},
	}

	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}