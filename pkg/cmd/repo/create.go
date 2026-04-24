package repo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	bbauth "github.com/chandrasekar-r/bitbucket-cli/pkg/auth"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		name        string
		description string
		private     bool
		clone       bool
		noClone     bool
	)

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				name = args[0]
			}

			workspace, err := f.Workspace()
			if err != nil {
				return err
			}

			// Interactive mode when TTY and name not supplied
			if name == "" {
				if !f.IOStreams.IsStdoutTTY() {
					return &cmdutil.FlagError{Err: fmt.Errorf("repo name required (use: bb repo create <name>)")}
				}
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Repository name").
							Value(&name),
						huh.NewInput().
							Title("Description (optional)").
							Value(&description),
						huh.NewSelect[bool]().
							Title("Visibility").
							Options(
								huh.NewOption("Private", true),
								huh.NewOption("Public", false),
							).
							Value(&private),
					),
				)
				if err := form.Run(); err != nil {
					return err
				}
			}

			name = strings.TrimSpace(name)
			if name == "" {
				return &cmdutil.FlagError{Err: fmt.Errorf("repository name cannot be empty")}
			}

			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			r, err := client.CreateRepo(workspace, api.CreateRepoOptions{
				Name:        name,
				Description: description,
				IsPrivate:   private,
				HasIssues:   true,
				HasWiki:     false,
			})
			if err != nil {
				return fmt.Errorf("creating repository: %w", err)
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Created repo %s\n", r.FullName)
			fmt.Fprintf(f.IOStreams.Out, "  URL: %s\n", r.Links.HTML.Href)

			// Offer to clone after creation
			if !noClone && f.IOStreams.IsStdoutTTY() && !clone {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Clone the repository locally?").
							Value(&clone),
					),
				)
				_ = form.Run()
			}

			if clone || (f.IOStreams.IsStdoutTTY() && clone) {
				cloneURL := r.CloneURL("https")
				if cloneURL == "" {
					fmt.Fprintln(f.IOStreams.ErrOut, "Could not determine clone URL")
					return nil
				}
				cloneURL, err = bbauth.InjectCloneAuth(cloneURL)
				if err != nil {
					return err
				}
				fmt.Fprintf(f.IOStreams.Out, "Cloning into %s...\n", name)
				gitCmd := exec.Command("git", "clone", "--", cloneURL)
				gitCmd.Stdout = os.Stdout
				gitCmd.Stderr = os.Stderr
				if err := gitCmd.Run(); err != nil {
					return fmt.Errorf("git clone failed: %w", err)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Repository description")
	cmd.Flags().BoolVar(&private, "private", true, "Make repository private")
	cmd.Flags().BoolVar(&clone, "clone", false, "Clone repository after creation")
	cmd.Flags().BoolVar(&noClone, "no-clone", false, "Skip the clone prompt")
	return cmd
}
