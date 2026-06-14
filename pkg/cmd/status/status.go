package status

import (
	"fmt"
	"sync"
	"time"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
	"github.com/spf13/cobra"
)

// statusConcurrency caps the number of simultaneous repo→PR fetch goroutines.
// Without this, a workspace with 100+ repos would fire 100+ concurrent API calls,
// quickly exhausting the Bitbucket rate limit of 1,000 req/hr.
const statusConcurrency = 5

const pipelineFailureWindow = 24 * time.Hour

// statusData holds the categorised data for JSON output.
type statusData struct {
	MyPRs            []prEntry            `json:"my_prs"`
	ReviewPRs        []prEntry            `json:"review_prs"`
	AssignedIssues   []issueEntry         `json:"assigned_issues"`
	PipelineFailures []pipelineFailEntry  `json:"pipeline_failures"`
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

type issueEntry struct {
	ID    int    `json:"id"`
	Repo  string `json:"repo"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
	URL   string `json:"url"`
}

type pipelineFailEntry struct {
	BuildNumber int    `json:"build_number"`
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	Result      string `json:"result"`
	CreatedOn   string `json:"created_on"`
	URL         string `json:"url"`
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
  - Pull requests awaiting your review
  - Issues assigned to you (open)
  - Recent pipeline failures (last 24 hours)`,
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

	// 2. List repos the user contributes to (cap at 50 to avoid rate-limit hammering
	// on large workspaces; status is a quick at-a-glance view, not an audit).
	fmt.Fprintf(f.IOStreams.ErrOut, "Scanning workspace %s...\n", workspace)
	repos, err := client.ListRepos(workspace, api.ListReposOptions{
		Role:  "contributor",
		Limit: 50,
	})
	if err != nil {
		return fmt.Errorf("listing repos: %w", err)
	}

	if len(repos) == 0 {
		fmt.Fprintf(f.IOStreams.Out, "No repos found in workspace %s where you are a contributor\n", workspace)
		return nil
	}

	// 3. Fetch data concurrently across all repos, bounded by a semaphore.
	var (
		mu               sync.Mutex
		wg               sync.WaitGroup
		myPRs            []prEntry
		reviewPRs        []prEntry
		assignedIssues   []issueEntry
		pipelineFailures []pipelineFailEntry
		fetchErrs        []error
	)
	sem := make(chan struct{}, statusConcurrency)
	cutoff := time.Now().Add(-pipelineFailureWindow)

	for _, repo := range repos {
		wg.Add(1)
		go func(repoSlug string, hasIssues bool) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			prs, ferr := client.ListPRs(workspace, repoSlug, api.ListPRsOptions{
				State: "OPEN",
				Limit: 0,
			})
			if ferr != nil {
				mu.Lock()
				fetchErrs = append(fetchErrs, fmt.Errorf("%s PRs: %w", repoSlug, ferr))
				mu.Unlock()
			} else {
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
			}

			if hasIssues {
				issues, ierr := client.ListIssues(workspace, repoSlug, api.ListIssuesOptions{
					State:    "open",
					Assignee: user.Username,
					Limit:    10,
				})
				if ierr != nil {
					mu.Lock()
					fetchErrs = append(fetchErrs, fmt.Errorf("%s issues: %w", repoSlug, ierr))
					mu.Unlock()
				} else {
					for _, issue := range issues {
						mu.Lock()
						assignedIssues = append(assignedIssues, issueEntry{
							ID:    issue.ID,
							Repo:  repoSlug,
							Title: issue.Title,
							Kind:  issue.Kind,
							URL:   issue.Links.HTML.Href,
						})
						mu.Unlock()
					}
				}
			}

			pipelines, perr := client.ListPipelines(workspace, repoSlug, 5)
			if perr != nil {
				mu.Lock()
				fetchErrs = append(fetchErrs, fmt.Errorf("%s pipelines: %w", repoSlug, perr))
				mu.Unlock()
			} else {
				for _, p := range pipelines {
					if !isRecentPipelineFailure(p, cutoff) {
						continue
					}
					mu.Lock()
					pipelineFailures = append(pipelineFailures, pipelineFailEntry{
						BuildNumber: p.BuildNumber,
						Repo:        repoSlug,
						Branch:      p.Target.RefName,
						Result:      p.ResultName(),
						CreatedOn:   p.CreatedOn,
						URL:         p.Links.HTML.Href,
					})
					mu.Unlock()
				}
			}
		}(repo.Slug, repo.HasIssues)
	}

	wg.Wait()

	for _, e := range fetchErrs {
		fmt.Fprintf(f.IOStreams.ErrOut, "warning: %s\n", e)
	}

	data := statusData{
		MyPRs:            myPRs,
		ReviewPRs:        reviewPRs,
		AssignedIssues:   assignedIssues,
		PipelineFailures: pipelineFailures,
	}

	if jsonOpts.Enabled() {
		return output.PrintJSON(f.IOStreams.Out, data, jsonOpts.Fields, jsonOpts.JQExpr)
	}

	w := f.IOStreams.Out
	color := f.IOStreams.ColorEnabled()

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

	if len(assignedIssues) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, headerStyle("Assigned Issues", color))
		t := output.NewTable(w, color)
		t.AddHeader("#", "REPO", "TITLE", "KIND")
		for _, issue := range assignedIssues {
			t.AddRow(
				fmt.Sprintf("%d", issue.ID),
				issue.Repo,
				truncate(issue.Title, 50),
				issue.Kind,
			)
		}
		if err := t.Render(); err != nil {
			return err
		}
	}

	if len(pipelineFailures) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, headerStyle("Recent Pipeline Failures (24h)", color))
		t := output.NewTable(w, color)
		t.AddHeader("BUILD", "REPO", "BRANCH", "RESULT")
		for _, p := range pipelineFailures {
			t.AddRow(
				fmt.Sprintf("#%d", p.BuildNumber),
				p.Repo,
				p.Branch,
				p.Result,
			)
		}
		if err := t.Render(); err != nil {
			return err
		}
	}

	fmt.Fprintf(f.IOStreams.Out, "\n%d open PR(s), %d awaiting review, %d assigned issue(s), %d pipeline failure(s) (%d repos scanned)\n",
		len(myPRs), len(reviewPRs), len(assignedIssues), len(pipelineFailures), len(repos))

	return nil
}

func isRecentPipelineFailure(p api.Pipeline, cutoff time.Time) bool {
	result := p.ResultName()
	if result != "FAILED" && result != "ERROR" {
		return false
	}
	if p.CreatedOn == "" {
		return false
	}
	created, err := time.Parse(time.RFC3339, p.CreatedOn)
	if err != nil {
		return false
	}
	return !created.Before(cutoff)
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

	for _, p := range pr.Participants {
		if p.User.Username == username && p.Approved {
			return false
		}
	}
	return true
}

func headerStyle(s string, color bool) string {
	if color {
		return "\033[1;36m" + s + "\033[0m"
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}