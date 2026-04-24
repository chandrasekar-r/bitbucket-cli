package api

import (
	"encoding/json"
	"fmt"
)

// Branch represents a Bitbucket repository branch.
type Branch struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Links struct {
		HTML Link `json:"html"`
	} `json:"links"`
	Target struct {
		Hash    string `json:"hash"`
		Date    string `json:"date"`
		Message string `json:"message"`
		Author  struct {
			User struct {
				DisplayName string `json:"display_name"`
				Username    string `json:"username"`
			} `json:"user"`
			Raw string `json:"raw"`
		} `json:"author"`
	} `json:"target"`
}

// ShortHash returns the first 7 characters of the commit hash.
func (b *Branch) ShortHash() string {
	if len(b.Target.Hash) >= 7 {
		return b.Target.Hash[:7]
	}
	return b.Target.Hash
}

// BranchRestriction represents a branch protection rule.
type BranchRestriction struct {
	ID      int    `json:"id"`
	Kind    string `json:"kind"`   // "push", "delete", "restrict_merges", "force"
	Pattern string `json:"pattern"` // branch name glob
}

// ListBranches returns all branches for a repository.
func (c *Client) ListBranches(workspace, slug string, limit int) ([]Branch, error) {
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches?pagelen=100&sort=-target.date", workspace, slug)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing branches: %w", err)
	}
	branches := make([]Branch, 0, len(items))
	for _, raw := range items {
		var b Branch
		if err := json.Unmarshal(raw, &b); err == nil && b.Name != "" {
			branches = append(branches, b)
		}
	}
	return branches, nil
}

// GetBranch fetches a single branch by name.
func (c *Client) GetBranch(workspace, slug, branch string) (*Branch, error) {
	var b Branch
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", workspace, slug, branch)
	if err := c.Get(path, &b); err != nil {
		return nil, fmt.Errorf("getting branch %q: %w", branch, err)
	}
	return &b, nil
}

// CreateBranch creates a new branch from the given source hash or branch name.
func (c *Client) CreateBranch(workspace, slug, name, source string) (*Branch, error) {
	body := map[string]interface{}{
		"name": name,
		"target": map[string]string{
			"hash": source,
		},
	}
	var b Branch
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches", workspace, slug)
	if err := c.Post(path, body, &b); err != nil {
		return nil, fmt.Errorf("creating branch: %w", err)
	}
	return &b, nil
}

// DeleteBranch deletes a branch by name.
func (c *Client) DeleteBranch(workspace, slug, branch string) error {
	path := fmt.Sprintf("/repositories/%s/%s/refs/branches/%s", workspace, slug, branch)
	return c.Delete(path)
}

// CreateBranchRestriction adds a branch protection rule.
// kind is one of: "push", "delete", "restrict_merges", "force"
func (c *Client) CreateBranchRestriction(workspace, slug, kind, pattern string) (*BranchRestriction, error) {
	body := map[string]string{
		"kind":           kind,
		"branch_match_kind": "glob",
		"pattern":        pattern,
	}
	var r BranchRestriction
	path := fmt.Sprintf("/repositories/%s/%s/branch-restrictions", workspace, slug)
	if err := c.Post(path, body, &r); err != nil {
		return nil, fmt.Errorf("creating branch restriction: %w", err)
	}
	return &r, nil
}
