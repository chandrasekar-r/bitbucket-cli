package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// CodeSearchHit is a flattened code search match for CLI display and JSON output.
type CodeSearchHit struct {
	Repo    string `json:"repo,omitempty"`
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// CodeSearchDisabledError is returned when code search is not enabled for a workspace.
type CodeSearchDisabledError struct {
	Workspace string
}

func (e *CodeSearchDisabledError) Error() string {
	return fmt.Sprintf(
		"code search is not enabled for workspace %q. Enable it at https://bitbucket.org/%s/workspace/settings/search",
		e.Workspace, e.Workspace,
	)
}

// SearchCode searches source code in a workspace.
func (c *Client) SearchCode(workspace, query string, limit int) ([]CodeSearchHit, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query is required")
	}
	path := fmt.Sprintf(
		"/workspaces/%s/search/code?%s",
		workspace,
		url.Values{
			"search_query": {query},
			"pagelen":      {"50"},
		}.Encode(),
	)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			return nil, &CodeSearchDisabledError{Workspace: workspace}
		}
		return nil, fmt.Errorf("searching code: %w", err)
	}

	var hits []CodeSearchHit
	for _, raw := range items {
		hits = append(hits, flattenCodeSearchItem(raw)...)
	}
	return hits, nil
}

// SearchRepos finds repositories in a workspace by name substring.
func (c *Client) SearchRepos(workspace, query string, limit int) ([]Repository, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query is required")
	}
	return c.ListRepos(workspace, ListReposOptions{
		Query: query,
		Limit: limit,
	})
}

func flattenCodeSearchItem(raw json.RawMessage) []CodeSearchHit {
	var item struct {
		Repository struct {
			Slug string `json:"slug"`
			Name string `json:"name"`
		} `json:"repository"`
		File struct {
			Path string `json:"path"`
		} `json:"file"`
		ContentMatches []struct {
			Lines []struct {
				Line     int `json:"line"`
				Segments []struct {
					Text string `json:"text"`
				} `json:"segments"`
			} `json:"lines"`
		} `json:"content_matches"`
	}
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil
	}
	repo := item.Repository.Slug
	if repo == "" {
		repo = item.Repository.Name
	}
	var hits []CodeSearchHit
	for _, match := range item.ContentMatches {
		for _, line := range match.Lines {
			var content strings.Builder
			for _, seg := range line.Segments {
				content.WriteString(seg.Text)
			}
			text := strings.TrimSpace(content.String())
			if text == "" {
				continue
			}
			hits = append(hits, CodeSearchHit{
				Repo:    repo,
				Path:    item.File.Path,
				Line:    line.Line,
				Content: text,
			})
		}
	}
	return hits
}