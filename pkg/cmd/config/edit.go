package configcmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCmdEdit(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Open the config file in $EDITOR",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := config.ConfigFilePath()
			if _, err := os.Stat(path); os.IsNotExist(err) {
				cfg, cerr := f.Config()
				if cerr != nil {
					return cerr
				}
				if err := cfg.Set(config.KeyGitProtocol, cfg.Get(config.KeyGitProtocol)); err != nil {
					return fmt.Errorf("creating config file: %w", err)
				}
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = os.Getenv("VISUAL")
			}
			if editor == "" {
				return fmt.Errorf("$EDITOR is not set")
			}
			parts := strings.Fields(editor)
			editCmd := exec.Command(parts[0], append(parts[1:], path)...)
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr
			if err := editCmd.Run(); err != nil {
				return err
			}

			v := viper.New()
			v.SetConfigFile(path)
			if err := v.ReadInConfig(); err != nil {
				return fmt.Errorf("config file is not valid YAML: %w", err)
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Config saved to %s\n", path)
			return nil
		},
	}
	return cmd
}