package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Repository represents a Bitbucket Cloud repository.
type Repository struct {
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"is_private"`
	Language    string    `json:"language"`
	HasIssues   bool      `json:"has_issues"`
	HasWiki     bool      `json:"has_wiki"`
	Size        int64     `json:"size"`
	CreatedOn   string    `json:"created_on"`
	UpdatedOn   string    `json:"updated_on"`
	Scm         string    `json:"scm"`
	MainBranch  *struct {
		Name string `json:"name"`
	} `json:"mainbranch"`
	Owner Workspace `json:"owner"`
	Links struct {
		Self  Link `json:"self"`
		Clone []struct {
			Href string `json:"href"`
			Name string `json:"name"` // "https" or "ssh"
		} `json:"clone"`
		HTML Link `json:"html"`
	} `json:"links"`
	Parent *struct {
		FullName string `json:"full_name"`
	} `json:"parent"`
}

// CloneURL returns the clone URL for the given protocol ("https" or "ssh").
func (r *Repository) CloneURL(protocol string) string {
	for _, c := range r.Links.Clone {
		if c.Name == protocol {
			return c.Href
		}
	}
	return ""
}

// CreateRepoOptions holds parameters for creating a new repository.
type CreateRepoOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private"`
	HasIssues   bool   `json:"has_issues"`
	HasWiki     bool   `json:"has_wiki"`
	Scm         string `json:"scm"` // "git"
}

// ListReposOptions filters for listing repositories.
type ListReposOptions struct {
	Limit    int
	Language string
	Role     string // "owner", "member", "contributor"
}

// ListRepos returns repositories in a workspace.
func (c *Client) ListRepos(workspace string, opts ListReposOptions) ([]Repository, error) {
	q := url.Values{"pagelen": {"100"}}
	if opts.Language != "" {
		q.Set("q", fmt.Sprintf(`language="%s"`, opts.Language))
	}
	if opts.Role != "" {
		q.Set("role", opts.Role)
	}
	path := fmt.Sprintf("/repositories/%s?%s", workspace, q.Encode())

	items, err := PaginateAll(c, path, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("listing repos: %w", err)
	}
	repos := make([]Repository, 0, len(items))
	for _, raw := range items {
		var r Repository
		if err := json.Unmarshal(raw, &r); err == nil && r.Slug != "" {
			repos = append(repos, r)
		}
	}
	return repos, nil
}

// GetRepo fetches a single repository.
func (c *Client) GetRepo(workspace, slug string) (*Repository, error) {
	var r Repository
	path := fmt.Sprintf("/repositories/%s/%s", workspace, slug)
	if err := c.Get(path, &r); err != nil {
		return nil, fmt.Errorf("getting repo %s/%s: %w", workspace, slug, err)
	}
	return &r, nil
}

// CreateRepo creates a new repository in the given workspace.
func (c *Client) CreateRepo(workspace string, opts CreateRepoOptions) (*Repository, error) {
	if opts.Scm == "" {
		opts.Scm = "git"
	}
	var r Repository
	path := fmt.Sprintf("/repositories/%s/%s", workspace, opts.Name)
	if err := c.Post(path, opts, &r); err != nil {
		return nil, fmt.Errorf("creating repo: %w", err)
	}
	return &r, nil
}

// ForkRepo forks a repository into the authenticated user's workspace (or destWorkspace).
func (c *Client) ForkRepo(srcWorkspace, srcSlug, destWorkspace string) (*Repository, error) {
	body := map[string]interface{}{}
	if destWorkspace != "" {
		body["workspace"] = map[string]string{"slug": destWorkspace}
	}
	var r Repository
	path := fmt.Sprintf("/repositories/%s/%s/forks", srcWorkspace, srcSlug)
	if err := c.Post(path, body, &r); err != nil {
		return nil, fmt.Errorf("forking repo: %w", err)
	}
	return &r, nil
}

// DeleteRepo permanently deletes a repository. This is irreversible.
func (c *Client) DeleteRepo(workspace, slug string) error {
	path := fmt.Sprintf("/repositories/%s/%s", workspace, slug)
	return c.Delete(path)
}

// RenameRepo renames a repository by updating its slug and name.
func (c *Client) RenameRepo(workspace, slug, newName string) (*Repository, error) {
	body := map[string]string{
		"name": newName,
		"slug": newName,
	}
	var r Repository
	path := fmt.Sprintf("/repositories/%s/%s", workspace, slug)
	if err := c.Put(path, body, &r); err != nil {
		return nil, fmt.Errorf("renaming repo: %w", err)
	}
	return &r, nil
}

// HasIssues reports whether the repository has the issue tracker enabled.
// Issue commands must call this before making any /issues API calls to return
// a friendly error instead of a confusing 404.
func (c *Client) HasIssues(workspace, slug string) (bool, error) {
	r, err := c.GetRepo(workspace, slug)
	if err != nil {
		return false, err
	}
	return r.HasIssues, nil
}
