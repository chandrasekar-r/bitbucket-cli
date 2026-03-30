package api_test

import (
	"net/http"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

// TestHasIssues_Enabled verifies the guard returns true when has_issues is set.
func TestHasIssues_Enabled(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(t, w, map[string]interface{}{
			"slug":       "my-repo",
			"has_issues": true,
		})
	}))

	ok, err := client.HasIssues("ws", "my-repo")
	if err != nil {
		t.Fatalf("HasIssues: %v", err)
	}
	if !ok {
		t.Error("expected has_issues=true")
	}
}

// TestHasIssues_Disabled verifies the guard returns false when issues are off.
func TestHasIssues_Disabled(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(t, w, map[string]interface{}{
			"slug":       "my-repo",
			"has_issues": false,
		})
	}))

	ok, err := client.HasIssues("ws", "my-repo")
	if err != nil {
		t.Fatalf("HasIssues: %v", err)
	}
	if ok {
		t.Error("expected has_issues=false")
	}
}

// TestListIssues verifies issue list parsing.
func TestListIssues(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"id":     1,
					"title":  "Login is broken",
					"status": "open",
					"kind":   "bug",
					"priority": "critical",
					"reporter": map[string]string{"username": "alice"},
				},
			},
			"pagelen": 1,
			"size":    1,
		})
	}))

	issues, err := client.ListIssues("ws", "repo", api.ListIssuesOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].ID != 1 {
		t.Errorf("ID: got %d", issues[0].ID)
	}
	if issues[0].State != "open" {
		t.Errorf("state: got %q", issues[0].State)
	}
}

// TestListSnippets verifies snippet list parsing and FileCount helper.
func TestListSnippets(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"id":         "abc123",
					"title":      "My deploy script",
					"is_private": true,
					"updated_on": "2026-03-30T10:00:00Z",
					"files": map[string]interface{}{
						"deploy.sh":  map[string]string{},
						"config.yml": map[string]string{},
					},
				},
			},
			"pagelen": 1,
			"size":    1,
		})
	}))

	snippets, err := client.ListSnippets("ws", 10)
	if err != nil {
		t.Fatalf("ListSnippets: %v", err)
	}
	if len(snippets) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(snippets))
	}
	s := snippets[0]
	if s.ID != "abc123" {
		t.Errorf("ID: got %q", s.ID)
	}
	if s.FileCount() != 2 {
		t.Errorf("FileCount: got %d, want 2", s.FileCount())
	}
	if !s.IsPrivate {
		t.Error("expected private snippet")
	}
}

// TestSnippetCloneURL verifies the CloneURL helper picks the right protocol.
func TestSnippetCloneURL(t *testing.T) {
	s := &api.Snippet{}
	s.Links.Clone = []struct {
		Name string `json:"name"`
		Href string `json:"href"`
	}{
		{Name: "https", Href: "https://bitbucket.org/ws/abc123"},
		{Name: "ssh", Href: "git@bitbucket.org:ws/abc123"},
	}

	if got := s.CloneURL("https"); got != "https://bitbucket.org/ws/abc123" {
		t.Errorf("HTTPS clone URL: got %q", got)
	}
	if got := s.CloneURL("ssh"); got != "git@bitbucket.org:ws/abc123" {
		t.Errorf("SSH clone URL: got %q", got)
	}
	if got := s.CloneURL("unknown"); got != "" {
		t.Errorf("unknown protocol should return empty string, got %q", got)
	}
}

// TestCreateIssue verifies the request body for issue creation.
func TestCreateIssue(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		encodeJSON(t, w, map[string]interface{}{
			"id":    5,
			"title": "New feature request",
			"status": "new",
			"kind": "enhancement",
			"links": map[string]interface{}{
				"html": map[string]string{
					"href": "https://bitbucket.org/ws/repo/issues/5",
				},
			},
		})
	}))

	i, err := client.CreateIssue("ws", "repo", api.CreateIssueOptions{
		Title:    "New feature request",
		Kind:     "enhancement",
		Priority: "minor",
	})
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if i.ID != 5 {
		t.Errorf("ID: got %d", i.ID)
	}
}
