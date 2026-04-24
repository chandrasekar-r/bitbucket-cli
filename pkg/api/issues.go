package api

import (
	"encoding/json"
	"fmt"
)

// Issue represents a Bitbucket Cloud issue.
type Issue struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  struct {
		Raw string `json:"raw"`
	} `json:"content"`
	State    string `json:"status"` // new, open, resolved, on hold, invalid, duplicate, wontfix, closed
	Kind     string `json:"kind"`   // bug, enhancement, proposal, task
	Priority string `json:"priority"` // trivial, minor, major, critical, blocker
	Assignee *struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
	} `json:"assignee"`
	Reporter struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
	} `json:"reporter"`
	CreatedOn  string `json:"created_on"`
	UpdatedOn  string `json:"updated_on"`
	CommentCount int   `json:"comment_count"`
	Links struct {
		HTML Link `json:"html"`
	} `json:"links"`
}

// IssueComment represents a comment on an issue.
type IssueComment struct {
	ID      int    `json:"id"`
	Content struct {
		Raw string `json:"raw"`
	} `json:"content"`
	CreatedOn string `json:"created_on"`
	Author    struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"username"`
	} `json:"author"`
}

// CreateIssueOptions holds parameters for creating a new issue.
type CreateIssueOptions struct {
	Title    string `json:"title"`
	Content  string `json:"-"`
	Kind     string `json:"kind"`
	Priority string `json:"priority"`
	Assignee string `json:"-"` // username
}

// ListIssuesOptions filters for listing issues.
type ListIssuesOptions struct {
	State    string // "new", "open", "resolved", etc. (empty = all open)
	Assignee string
	Limit    int
}

// ListIssues returns issues for a repository.
func (c *Client) ListIssues(workspace, slug string, opts ListIssuesOptions) ([]Issue, error) {
	q := "?pagelen=50"
	if opts.State != "" {
		q += "&q=status%3D%22" + opts.State + "%22"
	} else {
		q += `&q=status%3D%22new%22+OR+status%3D%22open%22`
	}
	path := fmt.Sprintf("/repositories/%s/%s/issues%s", workspace, slug, q)
	items, err := PaginateAll(c, path, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("listing issues: %w", err)
	}
	issues := make([]Issue, 0, len(items))
	for _, raw := range items {
		var i Issue
		if err := json.Unmarshal(raw, &i); err == nil {
			issues = append(issues, i)
		}
	}
	return issues, nil
}

// GetIssue fetches a single issue by ID.
func (c *Client) GetIssue(workspace, slug string, id int) (*Issue, error) {
	var i Issue
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d", workspace, slug, id)
	if err := c.Get(path, &i); err != nil {
		return nil, fmt.Errorf("getting issue #%d: %w", id, err)
	}
	return &i, nil
}

// CreateIssue creates a new issue.
func (c *Client) CreateIssue(workspace, slug string, opts CreateIssueOptions) (*Issue, error) {
	body := map[string]interface{}{
		"title":    opts.Title,
		"kind":     opts.Kind,
		"priority": opts.Priority,
		"content":  map[string]string{"raw": opts.Content},
	}
	if opts.Assignee != "" {
		body["assignee"] = map[string]string{"username": opts.Assignee}
	}
	var i Issue
	path := fmt.Sprintf("/repositories/%s/%s/issues", workspace, slug)
	if err := c.Post(path, body, &i); err != nil {
		return nil, fmt.Errorf("creating issue: %w", err)
	}
	return &i, nil
}

// UpdateIssueStatus changes the status of an issue (close, reopen, etc.).
func (c *Client) UpdateIssueStatus(workspace, slug string, id int, status string) (*Issue, error) {
	body := map[string]string{"status": status}
	var i Issue
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d", workspace, slug, id)
	if err := c.Put(path, body, &i); err != nil {
		return nil, fmt.Errorf("updating issue #%d: %w", id, err)
	}
	return &i, nil
}

// AddIssueComment adds a comment to an issue.
func (c *Client) AddIssueComment(workspace, slug string, id int, content string) error {
	body := map[string]interface{}{
		"content": map[string]string{"raw": content},
	}
	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments", workspace, slug, id)
	return c.Post(path, body, nil)
}
