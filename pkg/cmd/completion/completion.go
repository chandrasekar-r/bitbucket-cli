package completion

import (
	"errors"
	"os"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdCompletion(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion scripts",
		Long: `Generate shell completion scripts for bb.

Bash:
  source <(bb completion bash)
  # Add to ~/.bashrc for persistent completions

Zsh:
  bb completion zsh > "${fpath[1]}/_bb"
  # Or: source <(bb completion zsh)

Fish:
  bb completion fish | source
  # Or: bb completion fish > ~/.config/fish/completions/bb.fish

PowerShell:
  bb completion powershell | Out-String | Invoke-Expression
`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return errors.New("unsupported shell: " + args[0])
		},
	}
	return cmd
}
