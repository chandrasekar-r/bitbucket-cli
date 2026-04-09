package api_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestListPRsForBranch(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			t.Error("expected q query parameter to be present")
		}
		if !strings.Contains(q, `source.branch.name="feat/login"`) {
			t.Errorf("q param missing branch filter, got: %s", q)
		}
		if !strings.Contains(q, `state="MERGED"`) {
			t.Errorf("q param missing state filter, got: %s", q)
		}
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"id":    10,
					"title": "Login feature",
					"state": "MERGED",
					"source": map[string]interface{}{
						"branch": map[string]string{"name": "feat/login"},
					},
					"destination": map[string]interface{}{
						"branch": map[string]string{"name": "main"},
					},
				},
			},
			"pagelen": 1,
			"size":    1,
		})
	}))

	prs, err := client.ListPRsForBranch("ws", "repo", "feat/login", "MERGED")
	if err != nil {
		t.Fatalf("ListPRsForBranch: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(prs))
	}
	if prs[0].ID != 10 {
		t.Errorf("PR ID: got %d, want 10", prs[0].ID)
	}
	if prs[0].State != "MERGED" {
		t.Errorf("PR state: got %q, want MERGED", prs[0].State)
	}
}

func TestListPRsForBranch_NoState(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if strings.Contains(q, "state") {
			t.Errorf("expected no state filter when state is empty, got: %s", q)
		}
		encodeJSON(t, w, map[string]interface{}{
			"values":  []map[string]interface{}{},
			"pagelen": 0,
			"size":    0,
		})
	}))

	prs, err := client.ListPRsForBranch("ws", "repo", "feat/login", "")
	if err != nil {
		t.Fatalf("ListPRsForBranch: %v", err)
	}
	if len(prs) != 0 {
		t.Errorf("expected 0 PRs, got %d", len(prs))
	}
}
