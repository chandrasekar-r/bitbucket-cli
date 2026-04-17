package api

import (
	"encoding/json"
	"fmt"
)

// Runner represents a self-hosted Bitbucket Pipelines runner.
type Runner struct {
	UUID    string   `json:"uuid"`
	Name    string   `json:"name"`
	State   struct {
		Status  string `json:"status"`   // ONLINE, OFFLINE, UNREGISTERED, DISABLED
		Version struct {
			Version string `json:"version"`
		} `json:"version"`
		UpdatedOn string `json:"updated_on"`
	} `json:"state"`
	Labels    []string `json:"labels"`
	CreatedOn string   `json:"created_on"`
	UpdatedOn string   `json:"updated_on"`
	Disabled  bool     `json:"disabled"`

	// OAuthClient is populated only on Create (Bitbucket returns the one-time
	// client credentials used to register the runner on the host machine).
	OAuthClient *struct {
		ID             string `json:"id"`
		AudienceID     string `json:"audience_id,omitempty"`
		TokenEndpoint  string `json:"token_endpoint,omitempty"`
		Secret         string `json:"secret"`
	} `json:"oauth_client,omitempty"`
}

// RunnerInput is the create body.
type RunnerInput struct {
	Name   string   `json:"name"`
	Labels []string `json:"labels,omitempty"`
}

// RunnerUpdate is the state-toggle body (enable/disable).
type RunnerUpdate struct {
	Disabled *bool `json:"disabled,omitempty"`
}

// --- Workspace-scoped -------------------------------------------------------

func (c *Client) ListWorkspaceRunners(workspace string) ([]Runner, error) {
	return listRunners(c, fmt.Sprintf("/workspaces/%s/pipelines-config/runners?pagelen=100", workspace))
}

func (c *Client) GetWorkspaceRunner(workspace, uid string) (*Runner, error) {
	return getRunner(c, fmt.Sprintf("/workspaces/%s/pipelines-config/runners/%s", workspace, encodeUID(uid)))
}

func (c *Client) CreateWorkspaceRunner(workspace string, in RunnerInput) (*Runner, error) {
	return createRunner(c, fmt.Sprintf("/workspaces/%s/pipelines-config/runners", workspace), in)
}

func (c *Client) UpdateWorkspaceRunner(workspace, uid string, up RunnerUpdate) (*Runner, error) {
	return updateRunner(c, fmt.Sprintf("/workspaces/%s/pipelines-config/runners/%s", workspace, encodeUID(uid)), up)
}

func (c *Client) DeleteWorkspaceRunner(workspace, uid string) error {
	return c.Delete(fmt.Sprintf("/workspaces/%s/pipelines-config/runners/%s", workspace, encodeUID(uid)))
}

// --- Repository-scoped ------------------------------------------------------

func (c *Client) ListRepoRunners(workspace, slug string) ([]Runner, error) {
	return listRunners(c, fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners?pagelen=100", workspace, slug))
}

func (c *Client) GetRepoRunner(workspace, slug, uid string) (*Runner, error) {
	return getRunner(c, fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners/%s", workspace, slug, encodeUID(uid)))
}

func (c *Client) CreateRepoRunner(workspace, slug string, in RunnerInput) (*Runner, error) {
	return createRunner(c, fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners", workspace, slug), in)
}

func (c *Client) UpdateRepoRunner(workspace, slug, uid string, up RunnerUpdate) (*Runner, error) {
	return updateRunner(c, fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners/%s", workspace, slug, encodeUID(uid)), up)
}

func (c *Client) DeleteRepoRunner(workspace, slug, uid string) error {
	return c.Delete(fmt.Sprintf("/repositories/%s/%s/pipelines-config/runners/%s", workspace, slug, encodeUID(uid)))
}

// --- shared internals -------------------------------------------------------

func listRunners(c *Client, path string) ([]Runner, error) {
	items, err := PaginateAll(c, path, 0)
	if err != nil {
		return nil, fmt.Errorf("listing runners: %w", err)
	}
	runners := make([]Runner, 0, len(items))
	for _, raw := range items {
		var r Runner
		if err := json.Unmarshal(raw, &r); err == nil && r.UUID != "" {
			runners = append(runners, r)
		}
	}
	return runners, nil
}

func getRunner(c *Client, path string) (*Runner, error) {
	var r Runner
	if err := c.Get(path, &r); err != nil {
		return nil, fmt.Errorf("getting runner: %w", err)
	}
	return &r, nil
}

func createRunner(c *Client, path string, in RunnerInput) (*Runner, error) {
	var r Runner
	if err := c.Post(path, in, &r); err != nil {
		return nil, fmt.Errorf("creating runner: %w", err)
	}
	return &r, nil
}

func updateRunner(c *Client, path string, up RunnerUpdate) (*Runner, error) {
	var r Runner
	if err := c.Put(path, up, &r); err != nil {
		return nil, fmt.Errorf("updating runner: %w", err)
	}
	return &r, nil
}
