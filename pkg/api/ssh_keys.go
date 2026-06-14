package api

import (
	"encoding/json"
	"fmt"
)

// SSHKey is a user SSH public key.
type SSHKey struct {
	UUID        string `json:"uuid"`
	Comment     string `json:"comment"`
	Label       string `json:"label"`
	Fingerprint string `json:"fingerprint"`
	CreatedOn   string `json:"created_on"`
	Key         string `json:"key"`
}

// ListSSHKeys returns SSH keys for the authenticated user.
func (c *Client) ListSSHKeys(limit int) ([]SSHKey, error) {
	user, err := c.GetUser()
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/users/%s/ssh-keys", user.AccountID)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing SSH keys: %w", err)
	}
	keys := make([]SSHKey, 0, len(items))
	for _, raw := range items {
		var k SSHKey
		if err := json.Unmarshal(raw, &k); err == nil && k.UUID != "" {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// AddSSHKeyOptions holds parameters for adding an SSH key.
type AddSSHKeyOptions struct {
	Key   string
	Label string
}

// AddSSHKey adds an SSH public key for the authenticated user.
func (c *Client) AddSSHKey(opts AddSSHKeyOptions) (*SSHKey, error) {
	user, err := c.GetUser()
	if err != nil {
		return nil, err
	}
	body := map[string]string{
		"key":   opts.Key,
		"label": opts.Label,
	}
	var k SSHKey
	path := fmt.Sprintf("/users/%s/ssh-keys", user.AccountID)
	if err := c.Post(path, body, &k); err != nil {
		return nil, fmt.Errorf("adding SSH key: %w", err)
	}
	return &k, nil
}

// DeleteSSHKey removes an SSH key by UUID.
func (c *Client) DeleteSSHKey(uuid string) error {
	user, err := c.GetUser()
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/users/%s/ssh-keys/%s", user.AccountID, uuid)
	if err := c.Delete(path); err != nil {
		return fmt.Errorf("deleting SSH key: %w", err)
	}
	return nil
}