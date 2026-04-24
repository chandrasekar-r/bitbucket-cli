package api

import "encoding/json"

// User represents the Bitbucket Cloud authenticated user resource.
// Returned by GET /user.
type User struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AccountID   string `json:"account_id"`
	Links       struct {
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
	} `json:"links"`
}

// GetUser fetches the authenticated user's profile.
// Used during `bb auth login` to validate credentials and retrieve the username.
func (c *Client) GetUser() (*User, error) {
	var u User
	if err := c.Get("/user", &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// workspaceEntry is the shape of each item in GET /workspaces values.
type workspaceEntry struct {
	Workspace struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	} `json:"workspace"`
	// Bitbucket returns the workspace directly (not nested) in the /workspaces endpoint
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// GetUserWorkspaces returns all workspace slugs accessible to the authenticated user.
func (c *Client) GetUserWorkspaces() ([]string, error) {
	items, err := PaginateAll(c, "/workspaces?pagelen=100", 0)
	if err != nil {
		return nil, err
	}

	var slugs []string
	for _, raw := range items {
		var entry workspaceEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		slug := entry.Slug
		if slug == "" {
			slug = entry.Workspace.Slug
		}
		if slug != "" {
			slugs = append(slugs, slug)
		}
	}
	return slugs, nil
}
