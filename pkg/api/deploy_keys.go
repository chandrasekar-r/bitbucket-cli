package api

import (
	"encoding/json"
	"fmt"
)

// DeployKey is a repository deploy key.
type DeployKey struct {
	ID          int    `json:"id"`
	Label       string `json:"label"`
	Key         string `json:"key"`
	Comment     string `json:"comment"`
	Fingerprint string `json:"fingerprint"`
	CreatedOn   string `json:"created_on"`
	LastUsed    string `json:"last_used"`
}

// ListDeployKeys returns deploy keys for a repository.
func (c *Client) ListDeployKeys(workspace, slug string, limit int) ([]DeployKey, error) {
	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys/", workspace, slug)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing deploy keys: %w", err)
	}
	keys := make([]DeployKey, 0, len(items))
	for _, raw := range items {
		var k DeployKey
		if err := json.Unmarshal(raw, &k); err == nil && k.ID != 0 {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// GetDeployKey fetches a single deploy key by integer ID.
func (c *Client) GetDeployKey(workspace, slug string, id int) (*DeployKey, error) {
	var k DeployKey
	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys/%d", workspace, slug, id)
	if err := c.Get(path, &k); err != nil {
		return nil, fmt.Errorf("getting deploy key %d: %w", id, err)
	}
	return &k, nil
}

// AddDeployKeyOptions holds parameters for adding a deploy key.
type AddDeployKeyOptions struct {
	Key   string
	Label string
}

// AddDeployKey adds a deploy key to a repository.
func (c *Client) AddDeployKey(workspace, slug string, opts AddDeployKeyOptions) (*DeployKey, error) {
	body := map[string]string{
		"key":   opts.Key,
		"label": opts.Label,
	}
	var k DeployKey
	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys/", workspace, slug)
	if err := c.Post(path, body, &k); err != nil {
		return nil, fmt.Errorf("adding deploy key: %w", err)
	}
	return &k, nil
}

// DeleteDeployKey removes a deploy key by integer ID.
func (c *Client) DeleteDeployKey(workspace, slug string, id int) error {
	path := fmt.Sprintf("/repositories/%s/%s/deploy-keys/%d", workspace, slug, id)
	if err := c.Delete(path); err != nil {
		return fmt.Errorf("deleting deploy key %d: %w", id, err)
	}
	return nil
}