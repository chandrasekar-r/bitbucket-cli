package pr

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdCheckout(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout <number>",
		Short: "Check out a PR's source branch locally",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parsePRID(args[0])
			if err != nil {
				return err
			}
			workspace, slug, err := repoContext(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			pr, err := client.GetPR(workspace, slug, id)
			if err != nil {
				return err
			}

			branch := pr.Source.Branch.Name
			fmt.Fprintf(f.IOStreams.Out, "Checking out branch %q...\n", branch)

			// Fetch + checkout
			fetchCmd := exec.Command("git", "fetch", "origin", branch+":"+branch)
			fetchCmd.Stdout = os.Stdout
			fetchCmd.Stderr = os.Stderr
			if err := fetchCmd.Run(); err != nil {
				// Branch might already exist locally; try checkout directly
			}

			checkoutCmd := exec.Command("git", "checkout", branch)
			checkoutCmd.Stdout = os.Stdout
			checkoutCmd.Stderr = os.Stderr
			if err := checkoutCmd.Run(); err != nil {
				return fmt.Errorf("git checkout %q failed: %w", branch, err)
			}
			return nil
		},
	}
	return cmd
}
