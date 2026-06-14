package api_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestListCommitStatuses(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/commit/abc123def/statuses" {
			http.NotFound(w, r)
			return
		}
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]string{
				{
					"key":         "pipelines",
					"state":       "SUCCESSFUL",
					"name":        "Bitbucket Pipelines",
					"description": "Build #42 passed",
					"url":         "https://bitbucket.org/ws/repo/pipelines/results/1",
				},
				{
					"key":         "sonar",
					"state":       "FAILED",
					"name":        "SonarQube",
					"description": "Quality gate failed",
					"url":         "https://sonar.example.com/dashboard",
				},
			},
			"pagelen": 50,
			"size":    2,
		})
	}))

	statuses, err := client.ListCommitStatuses("ws", "repo", "abc123def", 0)
	if err != nil {
		t.Fatalf("ListCommitStatuses: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	if statuses[0].Key != "pipelines" || statuses[0].State != "SUCCESSFUL" {
		t.Errorf("first status: got key=%q state=%q", statuses[0].Key, statuses[0].State)
	}
	if statuses[1].Key != "sonar" || statuses[1].State != "FAILED" {
		t.Errorf("second status: got key=%q state=%q", statuses[1].Key, statuses[1].State)
	}
}

func TestListCommitStatuses_EmptyHash(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected HTTP request for empty commit hash")
		http.NotFound(w, r)
	}))

	_, err := client.ListCommitStatuses("ws", "repo", "", 0)
	if err == nil {
		t.Fatal("expected error for empty commit hash")
	}
}

func TestListCommitStatuses_Empty(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(t, w, map[string]interface{}{
			"values":  []any{},
			"pagelen": 50,
			"size":    0,
		})
	}))

	statuses, err := client.ListCommitStatuses("ws", "repo", "deadbeef", 0)
	if err != nil {
		t.Fatalf("ListCommitStatuses: %v", err)
	}
	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestListCommitStatuses_HTTPError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"message":"commit not found"}}`))
	}))

	_, err := client.ListCommitStatuses("ws", "repo", "missing", 0)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	var httpErr *api.HTTPError
	if !errors.As(err, &httpErr) {
		t.Errorf("expected *api.HTTPError in chain, got %v", err)
	}
}