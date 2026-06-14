package api

import (
	"encoding/json"
	"fmt"
)

// CommitStatus represents a build/CI status on a commit (Bitbucket commit statuses API).
type CommitStatus struct {
	Key         string `json:"key"`
	State       string `json:"state"` // SUCCESSFUL, FAILED, INPROGRESS, STOPPED
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Refname     string `json:"refname"`
	CreatedOn   string `json:"created_on"`
	UpdatedOn   string `json:"updated_on"`
}

// ListCommitStatuses returns all statuses for the given commit hash.
func (c *Client) ListCommitStatuses(workspace, slug, commitHash string, limit int) ([]CommitStatus, error) {
	if commitHash == "" {
		return nil, fmt.Errorf("commit hash is required")
	}
	path := fmt.Sprintf(
		"/repositories/%s/%s/commit/%s/statuses?pagelen=50",
		workspace, slug, commitHash,
	)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing commit statuses: %w", err)
	}
	statuses := make([]CommitStatus, 0, len(items))
	for _, raw := range items {
		var s CommitStatus
		if err := json.Unmarshal(raw, &s); err == nil && s.Key != "" {
			statuses = append(statuses, s)
		}
	}
	return statuses, nil
}