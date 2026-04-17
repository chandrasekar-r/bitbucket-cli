package api

import (
	"encoding/json"
	"fmt"
)

// Project represents a Bitbucket Cloud workspace project (the container
// that holds repositories).
type Project struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
	CreatedOn   string `json:"created_on"`
	UpdatedOn   string `json:"updated_on"`
	Links       struct {
		HTML Link `json:"html"`
		Self Link `json:"self"`
	} `json:"links"`
}

// ProjectCreateInput is the POST body.
type ProjectCreateInput struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private"`
}

// ProjectUpdateInput is the PUT body. All fields optional; nil/empty means
// "don't change". For IsPrivate we use *bool so false is distinguishable
// from "unset".
type ProjectUpdateInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsPrivate   *bool  `json:"is_private,omitempty"`
}

// ListProjects returns all projects in a workspace.
func (c *Client) ListProjects(workspace string) ([]Project, error) {
	path := fmt.Sprintf("/workspaces/%s/projects?pagelen=100", workspace)
	items, err := PaginateAll(c, path, 0)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	projects := make([]Project, 0, len(items))
	for _, raw := range items {
		var p Project
		if err := json.Unmarshal(raw, &p); err == nil && p.Key != "" {
			projects = append(projects, p)
		}
	}
	return projects, nil
}

// GetProject fetches a project by key.
func (c *Client) GetProject(workspace, key string) (*Project, error) {
	var p Project
	if err := c.Get(fmt.Sprintf("/workspaces/%s/projects/%s", workspace, key), &p); err != nil {
		return nil, fmt.Errorf("getting project %s: %w", key, err)
	}
	return &p, nil
}

// CreateProject creates a project in the workspace.
func (c *Client) CreateProject(workspace string, in ProjectCreateInput) (*Project, error) {
	var p Project
	if err := c.Post(fmt.Sprintf("/workspaces/%s/projects", workspace), in, &p); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	return &p, nil
}

// UpdateProject updates a project by key.
func (c *Client) UpdateProject(workspace, key string, in ProjectUpdateInput) (*Project, error) {
	var p Project
	if err := c.Put(fmt.Sprintf("/workspaces/%s/projects/%s", workspace, key), in, &p); err != nil {
		return nil, fmt.Errorf("updating project %s: %w", key, err)
	}
	return &p, nil
}

// DeleteProject deletes a project by key.
func (c *Client) DeleteProject(workspace, key string) error {
	return c.Delete(fmt.Sprintf("/workspaces/%s/projects/%s", workspace, key))
}
