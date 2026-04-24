package project

import (
	"errors"
	"fmt"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		name        string
		description string
		public      bool
		private     bool
	)

	cmd := &cobra.Command{
		Use:   "update <key>",
		Short: "Update a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if public && private {
				return errors.New("--public and --private are mutually exclusive")
			}

			input := api.ProjectUpdateInput{}
			if name != "" {
				input.Name = name
			}
			if cmd.Flags().Changed("description") {
				input.Description = description
			}
			switch {
			case private:
				t := true
				input.IsPrivate = &t
			case public:
				flt := false
				input.IsPrivate = &flt
			}

			if input.Name == "" && input.Description == "" && input.IsPrivate == nil {
				return errors.New("nothing to update: pass --name, --description, --public, or --private")
			}

			workspace, err := f.Workspace()
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			p, err := client.UpdateProject(workspace, args[0], input)
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Updated project %s\n", p.Key)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New project name")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().BoolVar(&private, "private", false, "Mark the project private")
	cmd.Flags().BoolVar(&public, "public", false, "Mark the project public")

	return cmd
}
