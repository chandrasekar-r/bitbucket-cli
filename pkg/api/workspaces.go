package api

import (
	"encoding/json"
	"fmt"
)

// Workspace represents a Bitbucket Cloud workspace.
type Workspace struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Type string `json:"type"`
	UUID string `json:"uuid"`
	Links struct {
		Self   Link `json:"self"`
		Avatar Link `json:"avatar"`
	} `json:"links"`
}

// WorkspaceMember represents a member of a workspace.
type WorkspaceMember struct {
	User struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		UUID        string `json:"uuid"`
	} `json:"user"`
	Permission string `json:"permission"`
}

// Link is a generic Bitbucket API hypermedia link.
type Link struct {
	Href string `json:"href"`
}

// ListWorkspaces returns all workspaces accessible to the authenticated user.
func (c *Client) ListWorkspaces(limit int) ([]Workspace, error) {
	items, err := PaginateAll(c, "/workspaces?pagelen=100", limit)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}
	workspaces := make([]Workspace, 0, len(items))
	for _, raw := range items {
		var ws Workspace
		if err := json.Unmarshal(raw, &ws); err == nil && ws.Slug != "" {
			workspaces = append(workspaces, ws)
		}
	}
	return workspaces, nil
}

// GetWorkspace fetches a single workspace by slug.
func (c *Client) GetWorkspace(slug string) (*Workspace, error) {
	var ws Workspace
	if err := c.Get("/workspaces/"+slug, &ws); err != nil {
		return nil, fmt.Errorf("getting workspace %q: %w", slug, err)
	}
	return &ws, nil
}

// ListWorkspaceMembers returns members of a workspace.
func (c *Client) ListWorkspaceMembers(slug string, limit int) ([]WorkspaceMember, error) {
	path := fmt.Sprintf("/workspaces/%s/members?pagelen=50", slug)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing workspace members: %w", err)
	}
	members := make([]WorkspaceMember, 0, len(items))
	for _, raw := range items {
		var m WorkspaceMember
		if err := json.Unmarshal(raw, &m); err == nil {
			members = append(members, m)
		}
	}
	return members, nil
}
