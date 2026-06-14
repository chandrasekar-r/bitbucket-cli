package sshkeycmd

import (
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdAdd(f *cmdutil.Factory) *cobra.Command {
	var key string
	var keyFile string
	var label string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SSH key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyMaterial, err := readKeyMaterial(key, keyFile)
			if err != nil {
				return &cmdutil.FlagError{Err: err}
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			created, err := client.AddSSHKey(api.AddSSHKeyOptions{
				Key:   keyMaterial,
				Label: label,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Added SSH key %s\n", created.UUID)
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "SSH public key material")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "Path to SSH public key file")
	cmd.Flags().StringVar(&label, "label", "", "Label for the SSH key")
	return cmd
}