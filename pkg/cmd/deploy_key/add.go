package deploykeycmd

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
		Short: "Add a deploy key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			keyMaterial, err := readKeyMaterial(key, keyFile)
			if err != nil {
				return &cmdutil.FlagError{Err: err}
			}
			if label == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("--label is required")}
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			created, err := client.AddDeployKey(workspace, slug, api.AddDeployKeyOptions{
				Key:   keyMaterial,
				Label: label,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Added deploy key #%d (%s)\n", created.ID, created.Label)
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "SSH public key material")
	cmd.Flags().StringVar(&keyFile, "key-file", "", "Path to SSH public key file")
	cmd.Flags().StringVar(&label, "label", "", "Label for the deploy key")
	return cmd
}