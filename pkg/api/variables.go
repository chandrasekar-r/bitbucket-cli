package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PipelineVariable is a Bitbucket Pipelines variable.
type PipelineVariable struct {
	UUID    string `json:"uuid"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Secured bool   `json:"secured"`
}

// DeploymentEnvironment is a deployment environment in Pipelines.
type DeploymentEnvironment struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func wrapVariableUUID(uuid string) string {
	if strings.HasPrefix(uuid, "{") {
		return uuid
	}
	return "{" + uuid + "}"
}

func listVariables(c *Client, path string, limit int) ([]PipelineVariable, error) {
	items, err := PaginateAll(c, path+"?pagelen=50", limit)
	if err != nil {
		return nil, err
	}
	vars := make([]PipelineVariable, 0, len(items))
	for _, raw := range items {
		var v PipelineVariable
		if err := json.Unmarshal(raw, &v); err == nil && v.Key != "" {
			vars = append(vars, v)
		}
	}
	return vars, nil
}

// ListWorkspaceVariables lists workspace-scoped pipeline variables.
func (c *Client) ListWorkspaceVariables(workspace string, limit int) ([]PipelineVariable, error) {
	path := fmt.Sprintf("/workspaces/%s/pipelines-config/variables/", workspace)
	vars, err := listVariables(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing workspace variables: %w", err)
	}
	return vars, nil
}

// ListRepoVariables lists repository-scoped pipeline variables.
func (c *Client) ListRepoVariables(workspace, slug string, limit int) ([]PipelineVariable, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables/", workspace, slug)
	vars, err := listVariables(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing repo variables: %w", err)
	}
	return vars, nil
}

// ListDeploymentVariables lists variables for a deployment environment.
func (c *Client) ListDeploymentVariables(workspace, slug, envUUID string, limit int) ([]PipelineVariable, error) {
	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/",
		workspace, slug, wrapVariableUUID(envUUID))
	vars, err := listVariables(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing deployment variables: %w", err)
	}
	return vars, nil
}

type setVariableOptions struct {
	secured bool
}

func setVariable(c *Client, path, key, value string, opts setVariableOptions) (*PipelineVariable, error) {
	body := map[string]interface{}{
		"key":   key,
		"value": value,
	}
	if opts.secured {
		body["secured"] = true
	}
	var v PipelineVariable
	if err := c.Post(path, body, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// SetWorkspaceVariable creates or updates a workspace variable.
func (c *Client) SetWorkspaceVariable(workspace, key, value string, secured bool) (*PipelineVariable, error) {
	path := fmt.Sprintf("/workspaces/%s/pipelines-config/variables/", workspace)
	v, err := setVariable(c, path, key, value, setVariableOptions{secured: secured})
	if err != nil {
		return nil, fmt.Errorf("setting workspace variable: %w", err)
	}
	return v, nil
}

// SetRepoVariable creates or updates a repository variable.
func (c *Client) SetRepoVariable(workspace, slug, key, value string, secured bool) (*PipelineVariable, error) {
	path := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables/", workspace, slug)
	v, err := setVariable(c, path, key, value, setVariableOptions{secured: secured})
	if err != nil {
		return nil, fmt.Errorf("setting repo variable: %w", err)
	}
	return v, nil
}

// SetDeploymentVariable creates or updates a deployment environment variable.
func (c *Client) SetDeploymentVariable(workspace, slug, envUUID, key, value string, secured bool) (*PipelineVariable, error) {
	path := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/",
		workspace, slug, wrapVariableUUID(envUUID))
	v, err := setVariable(c, path, key, value, setVariableOptions{secured: secured})
	if err != nil {
		return nil, fmt.Errorf("setting deployment variable: %w", err)
	}
	return v, nil
}

func deleteVariableByKey(c *Client, listPath, deletePathFmt, key string) error {
	vars, err := listVariables(c, listPath, 0)
	if err != nil {
		return err
	}
	for _, v := range vars {
		if v.Key == key {
			path := fmt.Sprintf(deletePathFmt, wrapVariableUUID(v.UUID))
			return c.Delete(path)
		}
	}
	return fmt.Errorf("variable %q not found", key)
}

// DeleteWorkspaceVariable deletes a workspace variable by key name.
func (c *Client) DeleteWorkspaceVariable(workspace, key string) error {
	listPath := fmt.Sprintf("/workspaces/%s/pipelines-config/variables/", workspace)
	deleteFmt := fmt.Sprintf("/workspaces/%s/pipelines-config/variables/%%s/", workspace)
	if err := deleteVariableByKey(c, listPath, deleteFmt, key); err != nil {
		return fmt.Errorf("deleting workspace variable: %w", err)
	}
	return nil
}

// DeleteRepoVariable deletes a repository variable by key name.
func (c *Client) DeleteRepoVariable(workspace, slug, key string) error {
	listPath := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables/", workspace, slug)
	deleteFmt := fmt.Sprintf("/repositories/%s/%s/pipelines_config/variables/%%s/", workspace, slug)
	if err := deleteVariableByKey(c, listPath, deleteFmt, key); err != nil {
		return fmt.Errorf("deleting repo variable: %w", err)
	}
	return nil
}

// DeleteDeploymentVariable deletes a deployment variable by key name.
func (c *Client) DeleteDeploymentVariable(workspace, slug, envUUID, key string) error {
	listPath := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/",
		workspace, slug, wrapVariableUUID(envUUID))
	deleteFmt := fmt.Sprintf("/repositories/%s/%s/deployments_config/environments/%s/variables/%%s/",
		workspace, slug, wrapVariableUUID(envUUID))
	if err := deleteVariableByKey(c, listPath, deleteFmt, key); err != nil {
		return fmt.Errorf("deleting deployment variable: %w", err)
	}
	return nil
}

// ListDeploymentEnvironments lists deployment environments for a repository.
func (c *Client) ListDeploymentEnvironments(workspace, slug string, limit int) ([]DeploymentEnvironment, error) {
	path := fmt.Sprintf("/repositories/%s/%s/environments/", workspace, slug)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing deployment environments: %w", err)
	}
	envs := make([]DeploymentEnvironment, 0, len(items))
	for _, raw := range items {
		var e DeploymentEnvironment
		if err := json.Unmarshal(raw, &e); err == nil && e.UUID != "" {
			envs = append(envs, e)
		}
	}
	return envs, nil
}

// CreateDeploymentEnvironment creates a deployment environment.
func (c *Client) CreateDeploymentEnvironment(workspace, slug, name string) (*DeploymentEnvironment, error) {
	body := map[string]string{"name": name}
	var e DeploymentEnvironment
	path := fmt.Sprintf("/repositories/%s/%s/environments/", workspace, slug)
	if err := c.Post(path, body, &e); err != nil {
		return nil, fmt.Errorf("creating deployment environment: %w", err)
	}
	return &e, nil
}