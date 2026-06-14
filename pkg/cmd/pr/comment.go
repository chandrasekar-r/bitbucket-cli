package pr

import (
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdComment(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment <subcommand>",
		Short: "Manage pull request comments",
		Long: `Add, list, and reply to pull request comments.

  bb pr comment add <number>    Add a comment (general or inline)
  bb pr comment list <number>   List comments on a pull request
  bb pr comment reply <number>  Reply to an existing comment`,
	}
	cmd.AddCommand(newCmdCommentAdd(f))
	cmd.AddCommand(newCmdCommentList(f))
	cmd.AddCommand(newCmdCommentReply(f))
	return cmd
}