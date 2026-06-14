package api_test

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestSearchCode(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/myws/search/code" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query().Get("search_query")
		if q != "func main" {
			t.Errorf("search_query: got %q, want %q", q, "func main")
		}
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"file": map[string]string{"path": "main.go"},
					"content_matches": []map[string]interface{}{
						{
							"lines": []map[string]interface{}{
								{
									"line": float64(10),
									"segments": []map[string]string{
										{"text": "func main() {}"},
									},
								},
							},
						},
					},
				},
			},
			"pagelen": 10,
			"size":    1,
		})
	}))

	hits, err := client.SearchCode("myws", "func main", 0)
	if err != nil {
		t.Fatalf("SearchCode: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Path != "main.go" || hits[0].Line != 10 {
		t.Errorf("hit: path=%q line=%d", hits[0].Path, hits[0].Line)
	}
	if !strings.Contains(hits[0].Content, "func main") {
		t.Errorf("content: got %q", hits[0].Content)
	}
}

func TestSearchCode_Disabled(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"message":"not found"}}`))
	}))

	_, err := client.SearchCode("myws", "query", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	var disabled *api.CodeSearchDisabledError
	if !errors.As(err, &disabled) {
		t.Fatalf("expected CodeSearchDisabledError, got %T: %v", err, err)
	}
	if disabled.Workspace != "myws" {
		t.Errorf("workspace: got %q", disabled.Workspace)
	}
}

func TestSearchRepos(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/myws" {
			http.NotFound(w, r)
			return
		}
		q, _ := url.QueryUnescape(r.URL.Query().Get("q"))
		if !strings.Contains(q, `name ~ "cli"`) {
			t.Errorf("q filter: got %q", q)
		}
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{"slug": "bitbucket-cli", "name": "bitbucket-cli"},
			},
			"pagelen": 100,
			"size":    1,
		})
	}))

	repos, err := client.SearchRepos("myws", "cli", 0)
	if err != nil {
		t.Fatalf("SearchRepos: %v", err)
	}
	if len(repos) != 1 || repos[0].Slug != "bitbucket-cli" {
		t.Errorf("repos: %+v", repos)
	}
}