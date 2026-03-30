package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PipelineState holds state and result for a pipeline or step.
type PipelineState struct {
	Name   string `json:"name"` // PENDING, IN_PROGRESS, COMPLETED
	Result *struct {
		Name string `json:"name"` // SUCCESSFUL, FAILED, ERROR, STOPPED
	} `json:"result"`
}

// Pipeline represents a single Bitbucket Pipelines run.
type Pipeline struct {
	UUID      string        `json:"uuid"`
	BuildNumber int         `json:"build_number"`
	CreatedOn  string       `json:"created_on"`
	CompletedOn string      `json:"completed_on"`
	DurationInSeconds int   `json:"duration_in_seconds"`
	State     PipelineState `json:"state"`
	Target    struct {
		RefName string `json:"ref_name"`
		RefType string `json:"ref_type"` // "branch", "tag", "bookmark"
		Commit  struct {
			Hash string `json:"hash"`
		} `json:"commit"`
	} `json:"target"`
	Creator struct {
		DisplayName string `json:"display_name"`
		Username    string `json:"nickname"`
	} `json:"creator"`
	Links struct {
		HTML Link `json:"html"`
	} `json:"links"`
}

// IsComplete reports whether the pipeline has finished (any terminal state).
func (p *Pipeline) IsComplete() bool {
	return p.State.Name == "COMPLETED"
}

// ResultName returns the result name or empty string if still running.
func (p *Pipeline) ResultName() string {
	if p.State.Result != nil {
		return p.State.Result.Name
	}
	return ""
}

// PipelineStep represents one step within a pipeline run.
type PipelineStep struct {
	UUID  string        `json:"uuid"`
	Name  string        `json:"name"`
	State PipelineState `json:"state"`
	DurationInSeconds int `json:"duration_in_seconds"`
}

// TriggerPipelineOptions holds parameters for triggering a new pipeline run.
type TriggerPipelineOptions struct {
	// Exactly one of Branch, Tag, or Commit should be set.
	Branch string
	Tag    string
	Commit string
}

// ListPipelines returns recent pipeline runs for a repository.
func (c *Client) ListPipelines(workspace, slug string, limit int) ([]Pipeline, error) {
	path := fmt.Sprintf(
		"/repositories/%s/%s/pipelines/?sort=-created_on&pagelen=50",
		workspace, slug,
	)
	items, err := PaginateAll(c, path, limit)
	if err != nil {
		return nil, fmt.Errorf("listing pipelines: %w", err)
	}
	pipelines := make([]Pipeline, 0, len(items))
	for _, raw := range items {
		var p Pipeline
		if err := json.Unmarshal(raw, &p); err == nil && p.UUID != "" {
			pipelines = append(pipelines, p)
		}
	}
	return pipelines, nil
}

// GetPipeline fetches a single pipeline run by UUID.
func (c *Client) GetPipeline(workspace, slug, uuid string) (*Pipeline, error) {
	var p Pipeline
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s", workspace, slug, uuid)
	if err := c.Get(path, &p); err != nil {
		return nil, fmt.Errorf("getting pipeline %s: %w", uuid, err)
	}
	return &p, nil
}

// TriggerPipeline starts a new pipeline run.
func (c *Client) TriggerPipeline(workspace, slug string, opts TriggerPipelineOptions) (*Pipeline, error) {
	var target map[string]interface{}
	switch {
	case opts.Branch != "":
		target = map[string]interface{}{
			"type":     "pipeline_ref_target",
			"ref_type": "branch",
			"ref_name": opts.Branch,
		}
	case opts.Tag != "":
		target = map[string]interface{}{
			"type":     "pipeline_ref_target",
			"ref_type": "tag",
			"ref_name": opts.Tag,
		}
	case opts.Commit != "":
		target = map[string]interface{}{
			"type":   "pipeline_commit_target",
			"commit": map[string]string{"hash": opts.Commit},
		}
	default:
		return nil, fmt.Errorf("one of Branch, Tag, or Commit must be set")
	}

	body := map[string]interface{}{"target": target}
	var p Pipeline
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/", workspace, slug)
	if err := c.Post(path, body, &p); err != nil {
		return nil, fmt.Errorf("triggering pipeline: %w", err)
	}
	return &p, nil
}

// StopPipeline cancels a running pipeline. Uses POST /stopPipeline (not DELETE).
func (c *Client) StopPipeline(workspace, slug, uuid string) error {
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/stopPipeline", workspace, slug, uuid)
	return c.Post(path, nil, nil)
}

// ListSteps returns all steps for a pipeline run.
func (c *Client) ListSteps(workspace, slug, uuid string) ([]PipelineStep, error) {
	var result struct {
		Values []PipelineStep `json:"values"`
	}
	path := fmt.Sprintf("/repositories/%s/%s/pipelines/%s/steps/", workspace, slug, uuid)
	if err := c.Get(path, &result); err != nil {
		return nil, fmt.Errorf("listing steps: %w", err)
	}
	return result.Values, nil
}

// GetStepLog fetches log bytes for a step starting at byteOffset.
// Returns the bytes read and whether the step log is complete (i.e. step finished).
// Uses HTTP Range header to simulate incremental log streaming via polling.
func (c *Client) GetStepLog(workspace, slug, pipelineUUID, stepUUID string, byteOffset int64) ([]byte, error) {
	path := fmt.Sprintf(
		"/repositories/%s/%s/pipelines/%s/steps/%s/log",
		workspace, slug, pipelineUUID, stepUUID,
	)
	headers := map[string]string{}
	if byteOffset > 0 {
		headers["Range"] = fmt.Sprintf("bytes=%d-", byteOffset)
	}
	resp, err := c.GetRaw(path, headers)
	if err != nil {
		return nil, fmt.Errorf("fetching step log: %w", err)
	}
	defer resp.Body.Close()

	// 416 Range Not Satisfiable = no new bytes yet (step still running or offset past EOF)
	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("step log request failed: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading step log: %w", err)
	}
	return data, nil
}
