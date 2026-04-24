package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestListRepoHooks(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/ws/repo/hooks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		encodeJSON(t, w, map[string]any{
			"values": []map[string]any{
				{
					"uuid":        "{a-b-c}",
					"url":         "https://example.com/hook",
					"description": "ci",
					"active":      true,
					"events":      []string{"repo:push"},
				},
			},
			"pagelen": 1,
			"size":    1,
		})
	}))

	hooks, err := client.ListRepoHooks("ws", "repo")
	if err != nil {
		t.Fatalf("ListRepoHooks: %v", err)
	}
	if len(hooks) != 1 {
		t.Fatalf("want 1 hook, got %d", len(hooks))
	}
	if hooks[0].URL != "https://example.com/hook" {
		t.Errorf("url: got %q", hooks[0].URL)
	}
	if len(hooks[0].Events) != 1 || hooks[0].Events[0] != "repo:push" {
		t.Errorf("events: got %v", hooks[0].Events)
	}
}

func TestListWorkspaceHooks_HitsWorkspacePath(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws/hooks" {
			t.Errorf("unexpected path: %s (want /workspaces/ws/hooks)", r.URL.Path)
		}
		encodeJSON(t, w, map[string]any{"values": []any{}, "pagelen": 0, "size": 0})
	}))

	if _, err := client.ListWorkspaceHooks("ws"); err != nil {
		t.Fatalf("ListWorkspaceHooks: %v", err)
	}
}

func TestCreateRepoHook_PostsExpectedBody(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s", r.Method)
		}
		if r.URL.Path != "/repositories/ws/repo/hooks" {
			t.Errorf("path: got %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var in api.WebhookInput
		if err := json.Unmarshal(body, &in); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if in.URL != "https://example.com/hook" {
			t.Errorf("url: got %q", in.URL)
		}
		if in.Description != "ci" {
			t.Errorf("description: got %q", in.Description)
		}
		if len(in.Events) != 2 || in.Events[0] != "repo:push" {
			t.Errorf("events: got %v", in.Events)
		}
		if in.Active == nil || *in.Active != true {
			t.Errorf("active: got %v", in.Active)
		}
		w.WriteHeader(http.StatusCreated)
		encodeJSON(t, w, map[string]any{
			"uuid":        "{new-uuid}",
			"url":         in.URL,
			"description": in.Description,
			"active":      *in.Active,
			"events":      in.Events,
		})
	}))

	active := true
	h, err := client.CreateRepoHook("ws", "repo", api.WebhookInput{
		URL:         "https://example.com/hook",
		Description: "ci",
		Events:      []string{"repo:push", "pullrequest:created"},
		Active:      &active,
	})
	if err != nil {
		t.Fatalf("CreateRepoHook: %v", err)
	}
	if h.UUID != "{new-uuid}" {
		t.Errorf("uuid: got %q", h.UUID)
	}
}

func TestUpdateRepoHook_SendsActiveFalseOnly(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method: got %s", r.Method)
		}
		// UUID should be wrapped with braces.
		if r.URL.Path != "/repositories/ws/repo/hooks/%7Babc%7D" && r.URL.Path != "/repositories/ws/repo/hooks/{abc}" {
			t.Errorf("path: got %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if _, hasURL := decoded["url"]; hasURL {
			t.Errorf("expected url field omitted, got body %s", body)
		}
		if v, ok := decoded["active"].(bool); !ok || v {
			t.Errorf("expected active=false, got body %s", body)
		}
		encodeJSON(t, w, map[string]any{
			"uuid":   "{abc}",
			"url":    "https://unchanged.example.com",
			"active": false,
			"events": []string{"repo:push"},
		})
	}))

	inactive := false
	if _, err := client.UpdateRepoHook("ws", "repo", "abc", api.WebhookInput{Active: &inactive}); err != nil {
		t.Fatalf("UpdateRepoHook: %v", err)
	}
}

func TestDeleteRepoHook(t *testing.T) {
	called := false
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("method: got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.DeleteRepoHook("ws", "repo", "abc"); err != nil {
		t.Fatalf("DeleteRepoHook: %v", err)
	}
	if !called {
		t.Error("handler never hit")
	}
}
