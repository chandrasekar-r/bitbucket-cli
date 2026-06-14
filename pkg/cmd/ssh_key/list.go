package sshkeycmd

import (
	"fmt"

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
		Short: "List SSH keys",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			keys, err := client.ListSSHKeys(limit)
			if err != nil {
				return err
			}
			if jsonOpts.Enabled() {
				return output.PrintJSON(f.IOStreams.Out, keys, jsonOpts.Fields, jsonOpts.JQExpr)
			}
			if len(keys) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No SSH keys found")
				return nil
			}
			t := output.NewTable(f.IOStreams.Out, f.IOStreams.ColorEnabled())
			t.AddHeader("UUID", "LABEL", "FINGERPRINT", "CREATED")
			for _, k := range keys {
				t.AddRow(k.UUID, k.Label, k.Fingerprint, k.CreatedOn)
			}
			if err := t.Render(); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "\n%d SSH key(s)\n", len(keys))
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 50, "Maximum keys to list (0 = all)")
	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}