package status

import (
	"fmt"
	"sync"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

// statusConcurrency caps the number of simultaneous repo→PR fetch goroutines.
// Without this, a workspace with 100+ repos would fire 100+ concurrent API calls,
// quickly exhausting the Bitbucket rate limit of 1,000 req/hr.
const statusConcurrency = 5

// statusData holds the categorised PR data for JSON output.
type statusData struct {
	MyPRs     []prEntry `json:"my_prs"`
	ReviewPRs []prEntry `json:"review_prs"`
}

// prEntry is a flattened PR record for display and JSON output.
type prEntry struct {
	ID        int    `json:"id"`
	Repo      string `json:"repo"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Approvals int    `json:"approvals"`
	URL       string `json:"url"`
}

// NewCmdStatus creates the `bb status` command.
func NewCmdStatus(f *cmdutil.Factory) *cobra.Command {
	jsonOpts := &cmdutil.JSONOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show your open pull requests and pending reviews across the workspace",
		Long: `Show a personal dashboard of your Bitbucket activity across all repos
in the current workspace:

  - Pull requests you authored (open)
  - Pull requests awaiting your review`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(f, jsonOpts)
		},
	}

	jsonOpts = cmdutil.AddJSONFlags(cmd)
	return cmd
}

func runStatus(f *cmdutil.Factory, jsonOpts *cmdutil.JSONOptions) error {
	workspace, err := f.Workspace()
	if err != nil {
		return err
	}

	httpClient, err := f.HttpClient()
	if err != nil {
		return err
	}
	client := api.New(httpClient, f.BaseURL)

	// 1. Get current user
	user, err := client.GetUser()
	if err != nil {
		return fmt.Errorf("fetching current user: %w", err)
	}

	// 2. List repos the user contributes to
	repos, err := client.ListRepos(workspace, api.ListReposOptions{
		Role:  "contributor",
		Limit: 0, // all
	})
	if err != nil {
		return fmt.Errorf("listing repos: %w", err)
	}

	if len(repos) == 0 {
		fmt.Fprintf(f.IOStreams.Out, "No repos found in workspace %s where you are a contributor\n", workspace)
		return nil
	}

	// 3. Fetch open PRs concurrently across all repos, bounded by a semaphore.
	var (
		mu        sync.Mutex
		wg        sync.WaitGroup
		myPRs     []prEntry
		reviewPRs []prEntry
		fetchErrs []error
	)
	sem := make(chan struct{}, statusConcurrency)

	for _, repo := range repos {
		wg.Add(1)
		go func(repoSlug string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			prs, ferr := client.ListPRs(workspace, repoSlug, api.ListPRsOptions{
				State: "OPEN",
				Limit: 0, // all open PRs
			})
			if ferr != nil {
				mu.Lock()
				fetchErrs = append(fetchErrs, fmt.Errorf("%s: %w", repoSlug, ferr))
				mu.Unlock()
				return
			}

			for _, pr := range prs {
				entry := prEntry{
					ID:        pr.ID,
					Repo:      repoSlug,
					Title:     pr.Title,
					Author:    pr.Author.Username,
					Approvals: pr.ApprovalCount(),
					URL:       pr.Links.HTML.Href,
				}

				mu.Lock()
				if pr.Author.Username == user.Username {
					myPRs = append(myPRs, entry)
				} else if isReviewerPending(pr, user.Username) {
					reviewPRs = append(reviewPRs, entry)
				}
				mu.Unlock()
			}
		}(repo.Slug)
	}

	wg.Wait()

	// Report non-fatal fetch errors to stderr
	for _, e := range fetchErrs {
		fmt.Fprintf(f.IOStreams.ErrOut, "warning: %s\n", e)
	}

	// 4. Output
	data := statusData{MyPRs: myPRs, ReviewPRs: reviewPRs}

	if jsonOpts.Enabled() {
		return output.PrintJSON(f.IOStreams.Out, data, jsonOpts.Fields, jsonOpts.JQExpr)
	}

	w := f.IOStreams.Out
	color := f.IOStreams.ColorEnabled()

	// My open PRs
	fmt.Fprintln(w, headerStyle("Your Open Pull Requests", color))
	if len(myPRs) == 0 {
		fmt.Fprintln(w, "  None")
	} else {
		t := output.NewTable(w, color)
		t.AddHeader("#", "REPO", "TITLE", "APPROVALS")
		for _, pr := range myPRs {
			t.AddRow(
				fmt.Sprintf("%d", pr.ID),
				pr.Repo,
				truncate(pr.Title, 50),
				fmt.Sprintf("%d", pr.Approvals),
			)
		}
		if err := t.Render(); err != nil {
			return err
		}
	}

	fmt.Fprintln(w)

	// Reviews awaiting
	fmt.Fprintln(w, headerStyle("Awaiting Your Review", color))
	if len(reviewPRs) == 0 {
		fmt.Fprintln(w, "  None")
	} else {
		t := output.NewTable(w, color)
		t.AddHeader("#", "REPO", "TITLE", "AUTHOR")
		for _, pr := range reviewPRs {
			t.AddRow(
				fmt.Sprintf("%d", pr.ID),
				pr.Repo,
				truncate(pr.Title, 50),
				pr.Author,
			)
		}
		if err := t.Render(); err != nil {
			return err
		}
	}

	fmt.Fprintf(f.IOStreams.Out, "\n%d open PR(s), %d awaiting review (%d repos scanned)\n",
		len(myPRs), len(reviewPRs), len(repos))

	return nil
}

// isReviewerPending checks if the user is a reviewer on the PR but has not yet approved.
func isReviewerPending(pr api.PullRequest, username string) bool {
	isReviewer := false
	for _, r := range pr.Reviewers {
		if r.Username == username {
			isReviewer = true
			break
		}
	}
	if !isReviewer {
		return false
	}

	// Check if they have already approved via participants
	for _, p := range pr.Participants {
		if p.User.Username == username && p.Approved {
			return false
		}
	}
	return true
}

// headerStyle returns a bold cyan string when color is enabled.
func headerStyle(s string, color bool) string {
	if color {
		return "\033[1;36m" + s + "\033[0m"
	}
	return s
}

// truncate shortens a string to max characters with an ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
