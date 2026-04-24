package version

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/internal/version"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func NewCmdVersion(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print bb version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOpts.Enabled() {
				data := map[string]string{
					"version":   version.Version,
					"commit":    version.Commit,
					"buildDate": version.BuildDate,
				}
				return output.PrintJSON(f.IOStreams.Out, data, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			fmt.Fprintf(f.IOStreams.Out, "bb version %s (%s) %s\n",
				version.Version, version.Commit, version.BuildDate)
			return nil
		},
	}

	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}
