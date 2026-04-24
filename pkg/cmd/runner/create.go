package runner

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		repoFlag bool
		name     string
		labels   []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new self-hosted runner",
		Long: `Register a new self-hosted runner. Bitbucket returns one-time OAuth
client credentials that the runner host uses to authenticate. These
credentials are printed ONCE — store them securely before closing the
terminal. If lost, delete the runner and create a new one.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return errors.New("--name is required")
			}
			parsed := parseLabelFlags(labels)
			if len(parsed) == 0 {
				return errors.New("--label is required (e.g. --label self.hosted --label linux)")
			}

			sc, err := resolveScope(f, repoFlag)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			r, err := createRunner(client, sc, api.RunnerInput{Name: name, Labels: parsed})
			if err != nil {
				return err
			}

			out := f.IOStreams.Out
			fmt.Fprintf(out, "✓ Created runner %s (%q)\n", r.UUID, r.Name)
			fmt.Fprintf(out, "  Labels: %s\n", strings.Join(r.Labels, ", "))

			if r.OAuthClient != nil {
				fmt.Fprintln(out, "")
				fmt.Fprintln(out, "──────────── one-time credentials ────────────")
				fmt.Fprintln(out, "  These are shown ONCE. Store them now.")
				fmt.Fprintf(out, "  Client ID:     %s\n", r.OAuthClient.ID)
				fmt.Fprintf(out, "  Client Secret: %s\n", r.OAuthClient.Secret)
				if r.OAuthClient.AudienceID != "" {
					fmt.Fprintf(out, "  Audience ID:   %s\n", r.OAuthClient.AudienceID)
				}
				if r.OAuthClient.TokenEndpoint != "" {
					fmt.Fprintf(out, "  Token URL:     %s\n", r.OAuthClient.TokenEndpoint)
				}
				fmt.Fprintln(out, "──────────────────────────────────────────────")
			}
			return nil
		},
	}

	addRepoFlag(cmd, &repoFlag)
	cmd.Flags().StringVar(&name, "name", "", "Runner name (required)")
	cmd.Flags().StringSliceVar(&labels, "label", nil,
		"Runner label (repeatable, comma-separated). Example: --label self.hosted --label linux")
	return cmd
}
