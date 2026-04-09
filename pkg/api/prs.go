package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// PullRequest represents a Bitbucket Cloud pull request.
type PullRequest struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"` // OPEN, MERGED, DECLINED, SUPERSEDED
	CreatedOn   string `json:"created_on"`
	UpdatedOn   string `json:"updated_on"`
	CloseSourceBranch bool `json:"close_source_branch"`
	Author      struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
	} `json:"author"`
	Source struct {
		Branch     struct{ Name string `json:"name"` }     `json:"branch"`
		Repository *Repository `json:"repository"`
	} `json:"source"`
	Destination struct {
		Branch     struct{ Name string `json:"name"` }     `json:"branch"`
		Repository *Repository `json:"repository"`
	} `json:"destination"`
	Reviewers []struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
	} `json:"reviewers"`
	Participants []struct {
		User struct {
			DisplayName string `json:"display_name"`
			Username    string `json:"username"`
		} `json:"user"`
		Role     string `json:"role"`
		Approved bool   `json:"approved"`
	} `json:"participants"`
	Links struct {
		HTML Link `json:"html"`
	} `json:"links"`
	CommentCount int `json:"comment_count"`
	TaskCount    int `json:"task_count"`
}

// ApprovalCount returns the number of approved participants.
func (pr *PullRequest) ApprovalCount() int {
	count := 0
	for _, p := range pr.Participants {
		if p.Approved {
			count++
		}
	}
	return count
}

// CreatePROptions holds fields for creating a pull request.
type CreatePROptions struct {
	Title             string
	Description       string
	SourceBranch      string
	SourceWorkspace   string // empty = same repo
	SourceRepoSlug    string // empty = same repo
	DestBranch        string
	DestWorkspace     string
	DestRepoSlug      string
	ReviewerUsernames []string
	CloseSourceBranch bool
}

// MergeStrategy is the PR merge strategy.
type MergeStrategy string

const (
	MergeStrategyMergeCommit MergeStrategy = "merge_commit"
	MergeStrategySquash      MergeStrategy = "squash"
	MergeStrategyFastForward MergeStrategy = "fast_forward"
)

// ListPRsOptions filters for listing pull requests.
type ListPRsOptions struct {
	State  string // "OPEN", "MERGED", "DECLINED" — empty = OPEN
	Author string
	Limit  int
}

// ListPRs returns pull requests for a repository.
func (c *Client) ListPRs(workspace, slug string, opts ListPRsOptions) ([]PullRequest, error) {
	state := opts.State
	if state == "" {
		state = "OPEN"
	}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests?state=%s&pagelen=50", workspace, slug, state)
	items, err := PaginateAll(c, path, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("listing PRs: %w", err)
	}
	prs := make([]PullRequest, 0, len(items))
	for _, raw := range items {
		var pr PullRequest
		if err := json.Unmarshal(raw, &pr); err == nil {
			prs = append(prs, pr)
		}
	}
	return prs, nil
}

// ListPRsForBranch returns PRs that have the given branch as source.
// state filters by PR state; empty string means all states.
func (c *Client) ListPRsForBranch(workspace, slug, branch, state string) ([]PullRequest, error) {
	q := fmt.Sprintf(`source.branch.name="%s"`, branch)
	if state != "" {
		q += fmt.Sprintf(` AND state="%s"`, state)
	}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests?q=%s&pagelen=50",
		workspace, slug, url.QueryEscape(q))
	items, err := PaginateAll(c, path, 0)
	if err != nil {
		return nil, fmt.Errorf("listing PRs for branch %q: %w", branch, err)
	}
	prs := make([]PullRequest, 0, len(items))
	for _, raw := range items {
		var pr PullRequest
		if err := json.Unmarshal(raw, &pr); err == nil {
			prs = append(prs, pr)
		}
	}
	return prs, nil
}

// GetPR fetches a single pull request by ID.
func (c *Client) GetPR(workspace, slug string, id int) (*PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", workspace, slug, id)
	if err := c.Get(path, &pr); err != nil {
		return nil, fmt.Errorf("getting PR #%d: %w", id, err)
	}
	return &pr, nil
}

// CreatePR opens a new pull request.
func (c *Client) CreatePR(workspace, slug string, opts CreatePROptions) (*PullRequest, error) {
	body := map[string]interface{}{
		"title":               opts.Title,
		"description":         opts.Description,
		"close_source_branch": opts.CloseSourceBranch,
		"source": map[string]interface{}{
			"branch": map[string]string{"name": opts.SourceBranch},
		},
		"destination": map[string]interface{}{
			"branch": map[string]string{"name": opts.DestBranch},
		},
	}

	// Fork scenario: source is in a different repo
	if opts.SourceWorkspace != "" && opts.SourceRepoSlug != "" {
		body["source"] = map[string]interface{}{
			"branch":     map[string]string{"name": opts.SourceBranch},
			"repository": map[string]string{"full_name": opts.SourceWorkspace + "/" + opts.SourceRepoSlug},
		}
	}

	if len(opts.ReviewerUsernames) > 0 {
		reviewers := make([]map[string]string, len(opts.ReviewerUsernames))
		for i, u := range opts.ReviewerUsernames {
			reviewers[i] = map[string]string{"username": u}
		}
		body["reviewers"] = reviewers
	}

	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests", workspace, slug)
	if err := c.Post(path, body, &pr); err != nil {
		return nil, fmt.Errorf("creating PR: %w", err)
	}
	return &pr, nil
}

// MergePR merges a pull request with the given strategy.
func (c *Client) MergePR(workspace, slug string, id int, strategy MergeStrategy, message string) (*PullRequest, error) {
	body := map[string]interface{}{
		"type":           "pullrequest",
		"merge_strategy": string(strategy),
	}
	if message != "" {
		body["message"] = message
	}
	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge", workspace, slug, id)
	if err := c.Post(path, body, &pr); err != nil {
		return nil, fmt.Errorf("merging PR #%d: %w", id, err)
	}
	return &pr, nil
}

// ApprovePR approves a pull request.
func (c *Client) ApprovePR(workspace, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve", workspace, slug, id)
	return c.Post(path, nil, nil)
}

// RemoveApproval removes an approval from a pull request.
func (c *Client) RemoveApproval(workspace, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/approve", workspace, slug, id)
	return c.Delete(path)
}

// DeclinePR declines (closes) a pull request.
func (c *Client) DeclinePR(workspace, slug string, id int) (*PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/decline", workspace, slug, id)
	if err := c.Post(path, nil, &pr); err != nil {
		return nil, fmt.Errorf("declining PR #%d: %w", id, err)
	}
	return &pr, nil
}

// UpdatePROptions holds optional fields for updating a pull request.
// Nil pointer fields are omitted from the request (not changed).
type UpdatePROptions struct {
	Title             *string
	Description       *string
	DestBranch        *string
	ReviewerUsernames []string // replaces all reviewers; nil = don't change
	CloseSourceBranch *bool
}

// UpdatePR modifies an existing pull request.
func (c *Client) UpdatePR(workspace, slug string, id int, opts UpdatePROptions) (*PullRequest, error) {
	body := map[string]interface{}{}
	if opts.Title != nil {
		body["title"] = *opts.Title
	}
	if opts.Description != nil {
		body["description"] = *opts.Description
	}
	if opts.DestBranch != nil {
		body["destination"] = map[string]interface{}{
			"branch": map[string]string{"name": *opts.DestBranch},
		}
	}
	if opts.ReviewerUsernames != nil {
		reviewers := make([]map[string]string, len(opts.ReviewerUsernames))
		for i, u := range opts.ReviewerUsernames {
			reviewers[i] = map[string]string{"username": u}
		}
		body["reviewers"] = reviewers
	}
	if opts.CloseSourceBranch != nil {
		body["close_source_branch"] = *opts.CloseSourceBranch
	}

	var pr PullRequest
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", workspace, slug, id)
	if err := c.Put(path, body, &pr); err != nil {
		return nil, fmt.Errorf("updating PR #%d: %w", id, err)
	}
	return &pr, nil
}

// AddPRComment adds a general comment to a pull request.
func (c *Client) AddPRComment(workspace, slug string, id int, content string) error {
	body := map[string]interface{}{
		"content": map[string]string{"raw": content},
	}
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments", workspace, slug, id)
	return c.Post(path, body, nil)
}

// GetPRDiff returns the unified diff for a pull request as plain text.
func (c *Client) GetPRDiff(workspace, slug string, id int) (string, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diff", workspace, slug, id)
	resp, err := c.GetRaw(path, map[string]string{"Accept": "text/plain"})
	if err != nil {
		return "", fmt.Errorf("getting PR diff: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
