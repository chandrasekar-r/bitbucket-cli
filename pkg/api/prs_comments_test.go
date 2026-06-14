package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestListPRComments(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.Contains(r.URL.Path, "/pullrequests/5/comments") {
			http.NotFound(w, r)
			return
		}
		encodeJSON(t, w, map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"id": 1,
					"content": map[string]string{"raw": "top-level"},
					"user":    map[string]string{"nickname": "alice"},
					"created_on": "2026-06-14T10:00:00Z",
				},
				{
					"id": 2,
					"content": map[string]string{"raw": "inline note"},
					"user":    map[string]string{"nickname": "bob"},
					"parent":  map[string]int{"id": 1},
					"inline":  map[string]interface{}{"path": "main.go", "to": 12},
				},
			},
			"pagelen": 2,
			"size":    2,
		})
	}))

	comments, err := client.ListPRComments("ws", "repo", 5, 0)
	if err != nil {
		t.Fatalf("ListPRComments: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
	if comments[1].Parent == nil || comments[1].Parent.ID != 1 {
		t.Errorf("expected parent id 1, got %+v", comments[1].Parent)
	}
	if comments[1].Inline == nil || comments[1].Inline.Path != "main.go" {
		t.Errorf("expected inline path main.go, got %+v", comments[1].Inline)
	}
}

func TestReplyPRComment(t *testing.T) {
	var body map[string]interface{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/pullrequests/5/comments") {
			http.NotFound(w, r)
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading body: %v", err)
			return
		}
		if err := json.Unmarshal(data, &body); err != nil {
			t.Errorf("decoding body: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))

	if err := client.ReplyPRComment("ws", "repo", 5, 9, "sounds good"); err != nil {
		t.Fatalf("ReplyPRComment: %v", err)
	}
	parent, ok := body["parent"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing parent in body: %#v", body)
	}
	if parent["id"] != float64(9) {
		t.Errorf("parent id: got %v, want 9", parent["id"])
	}
	content, ok := body["content"].(map[string]interface{})
	if !ok || content["raw"] != "sounds good" {
		t.Errorf("content.raw: got %#v", body["content"])
	}
}

func TestReplyPRComment_HTTPError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		encodeJSON(t, w, map[string]interface{}{
			"error": map[string]string{"message": "parent comment not found"},
		})
	}))

	err := client.ReplyPRComment("ws", "repo", 5, 999, "reply")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "parent comment not found") {
		t.Errorf("unexpected error: %v", err)
	}
}