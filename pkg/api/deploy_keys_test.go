package api_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestListDeployKeys(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/deploy-keys/") {
			http.NotFound(w, r)
			return
		}
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{"id": 7, "label": "CI", "fingerprint": "fp1"},
			},
		})
	}))

	keys, err := client.ListDeployKeys("ws", "repo", 0)
	if err != nil {
		t.Fatalf("ListDeployKeys: %v", err)
	}
	if len(keys) != 1 || keys[0].ID != 7 {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}