package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestListProjects(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/ws/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		encodeJSON(t, w, map[string]any{
			"values": []map[string]any{
				{"key": "ENG", "name": "Engineering", "is_private": true},
				{"key": "OPS", "name": "Operations", "is_private": false},
			},
			"pagelen": 2, "size": 2,
		})
	}))

	projects, err := client.ListProjects("ws")
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("want 2 projects, got %d", len(projects))
	}
	if projects[0].Key != "ENG" || !projects[0].IsPrivate {
		t.Errorf("first project: got %+v", projects[0])
	}
}

func TestCreateProject_PostsExpectedBody(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s", r.Method)
		}
		if r.URL.Path != "/workspaces/ws/projects" {
			t.Errorf("path: got %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var in api.ProjectCreateInput
		if err := json.Unmarshal(body, &in); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if in.Key != "ENG" || in.Name != "Engineering" {
			t.Errorf("body: got %+v", in)
		}
		if !in.IsPrivate {
			t.Error("expected is_private=true in POST body")
		}
		w.WriteHeader(http.StatusCreated)
		encodeJSON(t, w, map[string]any{
			"key": in.Key, "name": in.Name, "is_private": in.IsPrivate,
		})
	}))

	p, err := client.CreateProject("ws", api.ProjectCreateInput{
		Key: "ENG", Name: "Engineering", IsPrivate: true,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.Key != "ENG" {
		t.Errorf("returned key: got %q", p.Key)
	}
}

func TestUpdateProject_SendsIsPrivateFalseOnly(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method: got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if _, hasName := decoded["name"]; hasName {
			t.Errorf("expected name omitted, got body %s", body)
		}
		if v, ok := decoded["is_private"].(bool); !ok || v {
			t.Errorf("expected is_private=false, got body %s", body)
		}
		encodeJSON(t, w, map[string]any{
			"key": "ENG", "name": "Engineering", "is_private": false,
		})
	}))

	pub := false
	if _, err := client.UpdateProject("ws", "ENG", api.ProjectUpdateInput{IsPrivate: &pub}); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
}

func TestDeleteProject(t *testing.T) {
	called := false
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodDelete {
			t.Errorf("method: got %s", r.Method)
		}
		if r.URL.Path != "/workspaces/ws/projects/ENG" {
			t.Errorf("path: got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	if err := client.DeleteProject("ws", "ENG"); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	if !called {
		t.Error("handler never hit")
	}
}
