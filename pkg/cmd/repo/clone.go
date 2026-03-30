package repo

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdClone(f *cmdutil.Factory) *cobra.Command {
	var useSSH bool

	cmd := &cobra.Command{
		Use:   "clone <workspace/repo> [directory]",
		Short: "Clone a repository locally",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, slug, err := resolveRepo(f, args[:1])
			if err != nil {
				return err
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)
			r, err := client.GetRepo(workspace, slug)
			if err != nil {
				return err
			}

			protocol := "https"
			if useSSH {
				protocol = "ssh"
			}
			cloneURL := r.CloneURL(protocol)
			if cloneURL == "" {
				return fmt.Errorf("no %s clone URL for %s/%s", protocol, workspace, slug)
			}

			gitArgs := []string{"clone", cloneURL}
			if len(args) == 2 {
				gitArgs = append(gitArgs, args[1]) // custom directory
			}

			fmt.Fprintf(f.IOStreams.Out, "Cloning %s/%s...\n", workspace, slug)
			gitCmd := exec.Command("git", gitArgs...)
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr
			return gitCmd.Run()
		},
	}

	cmd.Flags().BoolVar(&useSSH, "ssh", false, "Clone using SSH instead of HTTPS")
	return cmd
}
