package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func newTestClient(t *testing.T, handler http.Handler) (*api.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	httpClient := &http.Client{}
	return api.New(httpClient, srv.URL), srv
}

func TestListWorkspaces(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]string{
				{"slug": "ws1", "name": "Workspace One", "type": "team"},
				{"slug": "ws2", "name": "Workspace Two", "type": "user"},
			},
			"pagelen": 2,
			"size":    2,
		})
	}))

	workspaces, err := client.ListWorkspaces(10)
	if err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if len(workspaces) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(workspaces))
	}
	if workspaces[0].Slug != "ws1" {
		t.Errorf("first workspace slug: got %q, want %q", workspaces[0].Slug, "ws1")
	}
}

func TestListRepos(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{"slug": "my-repo", "name": "my-repo", "is_private": false, "language": "go"},
				{"slug": "other-repo", "name": "other-repo", "is_private": true, "language": "python"},
			},
			"pagelen": 2,
			"size":    2,
		})
	}))

	repos, err := client.ListRepos("myworkspace", api.ListReposOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].Slug != "my-repo" {
		t.Errorf("first repo slug: got %q", repos[0].Slug)
	}
	if repos[1].IsPrivate != true {
		t.Error("expected second repo to be private")
	}
}

func TestGetRepo(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"slug":      "my-repo",
			"full_name": "myws/my-repo",
			"language":  "go",
			"links": map[string]interface{}{
				"clone": []map[string]string{
					{"name": "https", "href": "https://bitbucket.org/myws/my-repo.git"},
					{"name": "ssh", "href": "git@bitbucket.org:myws/my-repo.git"},
				},
			},
		})
	}))

	repo, err := client.GetRepo("myws", "my-repo")
	if err != nil {
		t.Fatalf("GetRepo: %v", err)
	}
	if repo.Slug != "my-repo" {
		t.Errorf("slug: got %q", repo.Slug)
	}
	if got := repo.CloneURL("https"); got != "https://bitbucket.org/myws/my-repo.git" {
		t.Errorf("HTTPS clone URL: got %q", got)
	}
	if got := repo.CloneURL("ssh"); got != "git@bitbucket.org:myws/my-repo.git" {
		t.Errorf("SSH clone URL: got %q", got)
	}
}

func TestPagination(t *testing.T) {
	page := 0
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		switch page {
		case 1:
			baseURL := "http://" + r.Host
			json.NewEncoder(w).Encode(map[string]interface{}{
				"values":  []map[string]string{{"slug": "repo1"}},
				"pagelen": 1,
				"size":    2,
				"next":    baseURL + "/repositories/ws?page=2",
			})
		case 2:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"values":  []map[string]string{{"slug": "repo2"}},
				"pagelen": 1,
				"size":    2,
			})
		}
	}))

	repos, err := client.ListRepos("ws", api.ListReposOptions{Limit: 100})
	if err != nil {
		t.Fatalf("ListRepos with pagination: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("expected 2 repos across pages, got %d", len(repos))
	}
}
