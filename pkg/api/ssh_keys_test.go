package api_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestListSSHKeys(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/user":
			encodeJSON(t, w, map[string]interface{}{"account_id": "acc123", "username": "alice"})
		case strings.Contains(r.URL.Path, "/users/acc123/ssh-keys"):
			encodeJSON(t, w, map[string]interface{}{
				"values": []map[string]interface{}{
					{"uuid": "{key-1}", "label": "laptop", "fingerprint": "fp"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))

	keys, err := client.ListSSHKeys(0)
	if err != nil {
		t.Fatalf("ListSSHKeys: %v", err)
	}
	if len(keys) != 1 || keys[0].UUID != "{key-1}" {
		t.Fatalf("unexpected keys: %+v", keys)
	}
}