package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestListWorkspaceRunners(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/workspaces/ws/pipelines-config/runners") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		encodeJSON(t, w, map[string]any{
			"values": []map[string]any{
				{
					"uuid":   "{r1}",
					"name":   "ci-1",
					"labels": []string{"self.hosted", "linux"},
					"state":  map[string]any{"status": "ONLINE"},
				},
			},
			"pagelen": 1, "size": 1,
		})
	}))

	runners, err := client.ListWorkspaceRunners("ws")
	if err != nil {
		t.Fatalf("ListWorkspaceRunners: %v", err)
	}
	if len(runners) != 1 || runners[0].Name != "ci-1" {
		t.Fatalf("unexpected runners: %+v", runners)
	}
	if runners[0].State.Status != "ONLINE" {
		t.Errorf("state: got %q", runners[0].State.Status)
	}
}

func TestCreateWorkspaceRunner_PostsNameAndLabels(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var in api.RunnerInput
		if err := json.Unmarshal(body, &in); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if in.Name != "ci-1" {
			t.Errorf("name: got %q", in.Name)
		}
		if len(in.Labels) != 2 {
			t.Errorf("labels: got %v", in.Labels)
		}
		encodeJSON(t, w, map[string]any{
			"uuid":   "{new}",
			"name":   in.Name,
			"labels": in.Labels,
			"oauth_client": map[string]string{
				"id":     "CLIENT_ID",
				"secret": "SUPER_SECRET",
			},
		})
	}))

	r, err := client.CreateWorkspaceRunner("ws", api.RunnerInput{
		Name:   "ci-1",
		Labels: []string{"self.hosted", "linux"},
	})
	if err != nil {
		t.Fatalf("CreateWorkspaceRunner: %v", err)
	}
	if r.OAuthClient == nil || r.OAuthClient.Secret != "SUPER_SECRET" {
		t.Errorf("oauth_client not parsed: %+v", r.OAuthClient)
	}
}

func TestUpdateWorkspaceRunner_SendsDisabledOnly(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method: got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(decoded) != 1 {
			t.Errorf("body should carry only 'disabled'; got %v", decoded)
		}
		if v, ok := decoded["disabled"].(bool); !ok || !v {
			t.Errorf("disabled not true: got %v", decoded["disabled"])
		}
		encodeJSON(t, w, map[string]any{"uuid": "{r1}", "name": "ci-1", "disabled": true})
	}))

	on := true
	if _, err := client.UpdateWorkspaceRunner("ws", "r1", api.RunnerUpdate{Disabled: &on}); err != nil {
		t.Fatalf("UpdateWorkspaceRunner: %v", err)
	}
}

func TestListRepoRunners_HitsRepoPath(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/repositories/ws/repo/pipelines-config/runners"
		if !strings.HasPrefix(r.URL.Path, want) {
			t.Errorf("path: got %s, want prefix %s", r.URL.Path, want)
		}
		encodeJSON(t, w, map[string]any{"values": []any{}, "pagelen": 0, "size": 0})
	}))
	if _, err := client.ListRepoRunners("ws", "repo"); err != nil {
		t.Fatalf("ListRepoRunners: %v", err)
	}
}
