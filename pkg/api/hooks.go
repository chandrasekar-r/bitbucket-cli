package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Webhook represents a Bitbucket Cloud webhook subscription.
// Same shape is returned by the repository and workspace hook endpoints —
// SubjectType differs ("repository" vs "workspace").
type Webhook struct {
	UUID        string   `json:"uuid"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	SubjectType string   `json:"subject_type"`
	Active      bool     `json:"active"`
	Events      []string `json:"events"`
	CreatedAt   string   `json:"created_at"`
	Links       struct {
		Self Link `json:"self"`
	} `json:"links"`
}

// WebhookInput is the create/update request body.
// Active is a *bool so we can distinguish "unset" from "explicit false" on update.
type WebhookInput struct {
	Description string   `json:"description,omitempty"`
	URL         string   `json:"url,omitempty"`
	Active      *bool    `json:"active,omitempty"`
	Events      []string `json:"events,omitempty"`
}

// --- Repository-scoped ------------------------------------------------------

// ListRepoHooks returns all webhooks on a repo.
func (c *Client) ListRepoHooks(workspace, slug string) ([]Webhook, error) {
	return listHooks(c, fmt.Sprintf("/repositories/%s/%s/hooks?pagelen=100", workspace, slug))
}

// GetRepoHook fetches one webhook by UUID.
func (c *Client) GetRepoHook(workspace, slug, uid string) (*Webhook, error) {
	return getHook(c, fmt.Sprintf("/repositories/%s/%s/hooks/%s", workspace, slug, encodeUID(uid)))
}

// CreateRepoHook creates a webhook on the repo.
func (c *Client) CreateRepoHook(workspace, slug string, in WebhookInput) (*Webhook, error) {
	return createHook(c, fmt.Sprintf("/repositories/%s/%s/hooks", workspace, slug), in)
}

// UpdateRepoHook updates a webhook by UUID.
func (c *Client) UpdateRepoHook(workspace, slug, uid string, in WebhookInput) (*Webhook, error) {
	return updateHook(c, fmt.Sprintf("/repositories/%s/%s/hooks/%s", workspace, slug, encodeUID(uid)), in)
}

// DeleteRepoHook deletes a webhook by UUID.
func (c *Client) DeleteRepoHook(workspace, slug, uid string) error {
	return c.Delete(fmt.Sprintf("/repositories/%s/%s/hooks/%s", workspace, slug, encodeUID(uid)))
}

// --- Workspace-scoped -------------------------------------------------------

// ListWorkspaceHooks returns all webhooks on a workspace.
func (c *Client) ListWorkspaceHooks(workspace string) ([]Webhook, error) {
	return listHooks(c, fmt.Sprintf("/workspaces/%s/hooks?pagelen=100", workspace))
}

// GetWorkspaceHook fetches one workspace webhook by UUID.
func (c *Client) GetWorkspaceHook(workspace, uid string) (*Webhook, error) {
	return getHook(c, fmt.Sprintf("/workspaces/%s/hooks/%s", workspace, encodeUID(uid)))
}

// CreateWorkspaceHook creates a webhook on the workspace.
func (c *Client) CreateWorkspaceHook(workspace string, in WebhookInput) (*Webhook, error) {
	return createHook(c, fmt.Sprintf("/workspaces/%s/hooks", workspace), in)
}

// UpdateWorkspaceHook updates a workspace webhook by UUID.
func (c *Client) UpdateWorkspaceHook(workspace, uid string, in WebhookInput) (*Webhook, error) {
	return updateHook(c, fmt.Sprintf("/workspaces/%s/hooks/%s", workspace, encodeUID(uid)), in)
}

// DeleteWorkspaceHook deletes a workspace webhook by UUID.
func (c *Client) DeleteWorkspaceHook(workspace, uid string) error {
	return c.Delete(fmt.Sprintf("/workspaces/%s/hooks/%s", workspace, encodeUID(uid)))
}

// --- shared internals -------------------------------------------------------

func listHooks(c *Client, path string) ([]Webhook, error) {
	items, err := PaginateAll(c, path, 0)
	if err != nil {
		return nil, fmt.Errorf("listing webhooks: %w", err)
	}
	hooks := make([]Webhook, 0, len(items))
	for _, raw := range items {
		var h Webhook
		if err := json.Unmarshal(raw, &h); err == nil && h.UUID != "" {
			hooks = append(hooks, h)
		}
	}
	return hooks, nil
}

func getHook(c *Client, path string) (*Webhook, error) {
	var h Webhook
	if err := c.Get(path, &h); err != nil {
		return nil, fmt.Errorf("getting webhook: %w", err)
	}
	return &h, nil
}

func createHook(c *Client, path string, in WebhookInput) (*Webhook, error) {
	var h Webhook
	if err := c.Post(path, in, &h); err != nil {
		return nil, fmt.Errorf("creating webhook: %w", err)
	}
	return &h, nil
}

func updateHook(c *Client, path string, in WebhookInput) (*Webhook, error) {
	var h Webhook
	if err := c.Put(path, in, &h); err != nil {
		return nil, fmt.Errorf("updating webhook: %w", err)
	}
	return &h, nil
}

// encodeUID wraps Bitbucket UUIDs that the caller may have pasted without
// braces. The API accepts both "{abc-def}" and "abc-def" as of 2024, but
// historically required braces — we add them when missing for safety.
func encodeUID(uid string) string {
	if uid == "" {
		return uid
	}
	if strings.HasPrefix(uid, "{") {
		return uid
	}
	return "{" + uid + "}"
}
