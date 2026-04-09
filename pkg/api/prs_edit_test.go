package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestUpdatePR(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		wantPath := "/repositories/ws/repo/pullrequests/42"
		if r.URL.Path != wantPath {
			t.Errorf("path: got %q, want %q", r.URL.Path, wantPath)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request body: %v", err)
			return
		}
		if body["title"] != "Updated Title" {
			t.Errorf("title: got %v, want %q", body["title"], "Updated Title")
		}

		encodeJSON(t, w, map[string]interface{}{
			"id":    42,
			"title": "Updated Title",
			"state": "OPEN",
			"source": map[string]interface{}{
				"branch": map[string]string{"name": "feat/x"},
			},
			"destination": map[string]interface{}{
				"branch": map[string]string{"name": "main"},
			},
			"links": map[string]interface{}{
				"html": map[string]string{"href": "https://bitbucket.org/ws/repo/pull-requests/42"},
			},
		})
	}))

	title := "Updated Title"
	pr, err := client.UpdatePR("ws", "repo", 42, api.UpdatePROptions{
		Title: &title,
	})
	if err != nil {
		t.Fatalf("UpdatePR: %v", err)
	}
	if pr.ID != 42 {
		t.Errorf("PR ID: got %d, want 42", pr.ID)
	}
	if pr.Title != "Updated Title" {
		t.Errorf("title: got %q", pr.Title)
	}
	if pr.Links.HTML.Href != "https://bitbucket.org/ws/repo/pull-requests/42" {
		t.Errorf("link: got %q", pr.Links.HTML.Href)
	}
}

func TestUpdatePR_OmitsNilFields(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request body: %v", err)
			return
		}
		// Only title should be present; description, destination, reviewers should be absent
		if _, ok := body["description"]; ok {
			t.Errorf("description should not be in body when nil")
		}
		if _, ok := body["destination"]; ok {
			t.Errorf("destination should not be in body when nil")
		}
		if _, ok := body["reviewers"]; ok {
			t.Errorf("reviewers should not be in body when nil")
		}

		encodeJSON(t, w, map[string]interface{}{
			"id":    1,
			"title": "Only Title",
			"state": "OPEN",
		})
	}))

	title := "Only Title"
	_, err := client.UpdatePR("ws", "repo", 1, api.UpdatePROptions{
		Title: &title,
	})
	if err != nil {
		t.Fatalf("UpdatePR: %v", err)
	}
}
