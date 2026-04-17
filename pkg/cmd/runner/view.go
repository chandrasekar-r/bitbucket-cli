package runner

import (
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdView(f *cmdutil.Factory) *cobra.Command {
	var repoFlag bool
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "view <uuid>",
		Short: "Show a single runner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := resolveScope(f, repoFlag)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			r, err := getRunner(client, sc, args[0])
			if err != nil {
				return err
			}

			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, r, jsonOpts.Fields, jsonOpts.JQExpr)
			}

			fmt.Fprintf(f.IOStreams.Out, "UUID:     %s\n", r.UUID)
			fmt.Fprintf(f.IOStreams.Out, "Name:     %s\n", r.Name)
			fmt.Fprintf(f.IOStreams.Out, "Status:   %s\n", r.State.Status)
			fmt.Fprintf(f.IOStreams.Out, "Disabled: %t\n", r.Disabled)
			fmt.Fprintf(f.IOStreams.Out, "Version:  %s\n", r.State.Version.Version)
			fmt.Fprintf(f.IOStreams.Out, "Labels:   %s\n", strings.Join(r.Labels, ", "))
			fmt.Fprintf(f.IOStreams.Out, "Created:  %s\n", r.CreatedOn)
			return nil
		},
	}
	addRepoFlag(cmd, &repoFlag)
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
