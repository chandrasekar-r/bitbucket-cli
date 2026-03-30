package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestListBranches(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"name": "main",
					"target": map[string]interface{}{
						"hash": "abc1234567890",
						"date": "2026-03-30T10:00:00Z",
						"author": map[string]interface{}{
							"user": map[string]string{
								"username": "alice",
							},
						},
					},
				},
				{
					"name": "feat/login",
					"target": map[string]interface{}{
						"hash": "def9876543210",
						"date": "2026-03-29T08:00:00Z",
						"author": map[string]interface{}{
							"user": map[string]string{
								"username": "bob",
							},
						},
					},
				},
			},
			"pagelen": 2,
			"size":    2,
		})
	}))

	branches, err := client.ListBranches("ws", "repo", 10)
	if err != nil {
		t.Fatalf("ListBranches: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(branches))
	}
	if branches[0].Name != "main" {
		t.Errorf("first branch name: got %q", branches[0].Name)
	}
	if branches[0].ShortHash() != "abc1234" {
		t.Errorf("short hash: got %q, want %q", branches[0].ShortHash(), "abc1234")
	}
}

func TestListPRs(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"id":    1,
					"title": "Add login feature",
					"state": "OPEN",
					"author": map[string]string{"username": "alice", "display_name": "Alice"},
					"source": map[string]interface{}{
						"branch": map[string]string{"name": "feat/login"},
					},
					"destination": map[string]interface{}{
						"branch": map[string]string{"name": "main"},
					},
					"participants": []map[string]interface{}{
						{"role": "REVIEWER", "approved": true,
							"user": map[string]string{"username": "bob"}},
					},
				},
			},
			"pagelen": 1,
			"size":    1,
		})
	}))

	prs, err := client.ListPRs("ws", "repo", api.ListPRsOptions{State: "OPEN", Limit: 10})
	if err != nil {
		t.Fatalf("ListPRs: %v", err)
	}
	if len(prs) != 1 {
		t.Errorf("expected 1 PR, got %d", len(prs))
	}
	pr := prs[0]
	if pr.ID != 1 {
		t.Errorf("PR ID: got %d", pr.ID)
	}
	if pr.ApprovalCount() != 1 {
		t.Errorf("ApprovalCount: got %d, want 1", pr.ApprovalCount())
	}
	if pr.Source.Branch.Name != "feat/login" {
		t.Errorf("source branch: got %q", pr.Source.Branch.Name)
	}
}

func TestMergePR(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["merge_strategy"] != "squash" {
			t.Errorf("merge_strategy: got %v", body["merge_strategy"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    42,
			"title": "My PR",
			"state": "MERGED",
		})
	}))

	merged, err := client.MergePR("ws", "repo", 42, api.MergeStrategySquash, "")
	if err != nil {
		t.Fatalf("MergePR: %v", err)
	}
	if merged.State != "MERGED" {
		t.Errorf("state: got %q", merged.State)
	}
}
