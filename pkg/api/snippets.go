package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Snippet represents a Bitbucket Cloud snippet.
type Snippet struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	IsPrivate bool   `json:"is_private"`
	SCM       string `json:"scm"` // "git" or "hg"
	CreatedOn string `json:"created_on"`
	UpdatedOn string `json:"updated_on"`
	Owner     struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	Files map[string]struct {
		Links struct {
			Self Link `json:"self"`
		} `json:"links"`
	} `json:"files"`
	Links struct {
		HTML  Link `json:"html"`
		Clone []struct {
			Name string `json:"name"`
			Href string `json:"href"`
		} `json:"clone"`
	} `json:"links"`
}

// FileCount returns the number of files in the snippet.
func (s *Snippet) FileCount() int {
	return len(s.Files)
}

// CloneURL returns the clone URL for the given protocol ("https" or "ssh").
func (s *Snippet) CloneURL(protocol string) string {
	for _, c := range s.Links.Clone {
		if strings.EqualFold(c.Name, protocol) {
			return c.Href
		}
	}
	return ""
}

// ListSnippets returns snippets owned by the workspace (or the authenticated user if workspace is empty).
func (c *Client) ListSnippets(workspace string, limit int) ([]Snippet, error) {
	var path string
	if workspace != "" {
		path = fmt.Sprintf("/snippets/%s?pagelen=50", workspace)
	} else {
		path = "/snippets?role=owner&pagelen=50"
	}
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing snippets: %w", err)
	}
	snippets := make([]Snippet, 0, len(items))
	for _, raw := range items {
		var s Snippet
		if err := json.Unmarshal(raw, &s); err == nil && s.ID != "" {
			snippets = append(snippets, s)
		}
	}
	return snippets, nil
}

// GetSnippet fetches a single snippet by workspace and ID.
func (c *Client) GetSnippet(workspace, id string) (*Snippet, error) {
	var s Snippet
	path := fmt.Sprintf("/snippets/%s/%s", workspace, id)
	if err := c.Get(path, &s); err != nil {
		return nil, fmt.Errorf("getting snippet %s: %w", id, err)
	}
	return &s, nil
}

// CreateSnippetOptions holds parameters for a new snippet.
type CreateSnippetOptions struct {
	Title     string
	Content   string
	Filename  string // default filename within the snippet
	IsPrivate bool
	Workspace string
}

// CreateSnippet creates a new snippet with a single file.
func (c *Client) CreateSnippet(opts CreateSnippetOptions) (*Snippet, error) {
	filename := opts.Filename
	if filename == "" {
		filename = "snippet.txt"
	}
	body := map[string]interface{}{
		"title":      opts.Title,
		"is_private": opts.IsPrivate,
		"files": map[string]interface{}{
			filename: map[string]interface{}{
				"content": opts.Content,
			},
		},
		"scm": "git",
	}
	var s Snippet
	workspace := opts.Workspace
	if workspace == "" {
		workspace = "me"
	}
	path := fmt.Sprintf("/snippets/%s", workspace)
	if err := c.Post(path, body, &s); err != nil {
		return nil, fmt.Errorf("creating snippet: %w", err)
	}
	return &s, nil
}

// DeleteSnippet permanently deletes a snippet.
func (c *Client) DeleteSnippet(workspace, id string) error {
	path := fmt.Sprintf("/snippets/%s/%s", workspace, id)
	return c.Delete(path)
}
