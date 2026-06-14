package deploykeycmd

import (
	"fmt"
	"strconv"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var limit int
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deploy keys",
		Args:  cobra.NoArgs,
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
			keys, err := client.ListDeployKeys(workspace, slug, limit)
			if err != nil {
				return err
			}
			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, keys, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(keys) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No deploy keys found")
				return nil
			}
			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("ID", "LABEL", "FINGERPRINT")
			for _, k := range keys {
				t.AddRow(strconv.Itoa(k.ID), k.Label, k.Fingerprint)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d deploy key(s)\n", len(keys))
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 50, "Maximum keys to list (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}