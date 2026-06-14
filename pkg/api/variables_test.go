package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestListRepoVariables(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/pipelines_config/variables/") {
			http.NotFound(w, r)
			return
		}
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{"uuid": "{uuid-1}", "key": "AWS_KEY", "value": "", "secured": true},
				{"uuid": "{uuid-2}", "key": "REGION", "value": "us-east-1", "secured": false},
			},
		})
	}))

	vars, err := client.ListRepoVariables("ws", "repo", 0)
	if err != nil {
		t.Fatalf("ListRepoVariables: %v", err)
	}
	if len(vars) != 2 || !vars[0].Secured {
		t.Fatalf("unexpected vars: %+v", vars)
	}
}

func TestSetRepoVariableSecured(t *testing.T) {
	var body map[string]interface{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		encodeJSON(t, w, map[string]interface{}{"uuid": "{uuid-1}", "key": "TOKEN", "secured": true})
	}))

	if _, err := client.SetRepoVariable("ws", "repo", "TOKEN", "secret", true); err != nil {
		t.Fatalf("SetRepoVariable: %v", err)
	}
	if body["secured"] != true {
		t.Errorf("expected secured=true, got %#v", body["secured"])
	}
}