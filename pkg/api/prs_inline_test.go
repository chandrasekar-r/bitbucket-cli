package api_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func intPtr(n int) *int { return &n }

// captureInlineCommentBody returns the decoded "inline" sub-object from the
// request body, or nil if the key is absent.
func captureInlineCommentBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("decoding request body: %v", err)
		return nil
	}
	if v, ok := body["inline"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

func TestAddPRInlineComment_LineLevel(t *testing.T) {
	var capturedInline map[string]interface{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		wantPath := "/repositories/ws/repo/pullrequests/5/comments"
		if r.URL.Path != wantPath {
			t.Errorf("path: got %q, want %q", r.URL.Path, wantPath)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		capturedInline = captureInlineCommentBody(t, r)
		w.WriteHeader(http.StatusCreated)
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "looks good", api.InlineComment{
		Path: "main.go",
		To:   intPtr(10),
	})
	if err != nil {
		t.Fatalf("AddPRInlineComment: %v", err)
	}
	if capturedInline == nil {
		t.Fatal("expected 'inline' key in request body, got none")
	}
	if capturedInline["path"] != "main.go" {
		t.Errorf("inline.path: got %v, want %q", capturedInline["path"], "main.go")
	}
	// JSON numbers decode as float64
	if capturedInline["to"] != float64(10) {
		t.Errorf("inline.to: got %v, want 10", capturedInline["to"])
	}
	if _, hasFrom := capturedInline["from"]; hasFrom {
		t.Error("inline.from should be absent when From is nil")
	}
}

func TestAddPRInlineComment_FileLevel(t *testing.T) {
	var capturedInline map[string]interface{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedInline = captureInlineCommentBody(t, r)
		w.WriteHeader(http.StatusCreated)
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "file looks good", api.InlineComment{
		Path: "main.go",
		// To and From are nil → file-level comment
	})
	if err != nil {
		t.Fatalf("AddPRInlineComment: %v", err)
	}
	if capturedInline == nil {
		t.Fatal("expected 'inline' key in request body")
	}
	if capturedInline["path"] != "main.go" {
		t.Errorf("inline.path: got %v, want %q", capturedInline["path"], "main.go")
	}
	if _, hasTo := capturedInline["to"]; hasTo {
		t.Error("inline.to should be absent for file-level comment (To is nil)")
	}
	if _, hasFrom := capturedInline["from"]; hasFrom {
		t.Error("inline.from should be absent for file-level comment (From is nil)")
	}
}

func TestAddPRInlineComment_WithFromAndTo(t *testing.T) {
	var capturedInline map[string]interface{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedInline = captureInlineCommentBody(t, r)
		w.WriteHeader(http.StatusCreated)
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "range comment", api.InlineComment{
		Path: "main.go",
		To:   intPtr(10),
		From: intPtr(5),
	})
	if err != nil {
		t.Fatalf("AddPRInlineComment: %v", err)
	}
	if capturedInline["to"] != float64(10) {
		t.Errorf("inline.to: got %v, want 10", capturedInline["to"])
	}
	if capturedInline["from"] != float64(5) {
		t.Errorf("inline.from: got %v, want 5", capturedInline["from"])
	}
}

// Wire-format pin: non-nil pointer to 1 must appear as "to":1 in JSON,
// and nil must be absent. This pins *int + omitempty behavior so a future
// type change from *int to int (which WOULD omit zero) is caught.
func TestAddPRInlineComment_WireFormatPin(t *testing.T) {
	var capturedBody map[string]interface{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Errorf("decoding body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))

	// Non-nil To=1 must be present in JSON
	if err := client.AddPRInlineComment("ws", "repo", 1, "x", api.InlineComment{Path: "f.go", To: intPtr(1)}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inline, _ := capturedBody["inline"].(map[string]interface{})
	if inline == nil {
		t.Fatal("inline key absent")
	}
	if inline["to"] != float64(1) {
		t.Errorf("to=1 should appear in JSON, got %v", inline["to"])
	}
}

func TestAddPRInlineComment_NilToAbsentFromJSON(t *testing.T) {
	var capturedBody map[string]interface{}
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Errorf("decoding body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))

	// nil To must be absent from JSON (file-level)
	if err := client.AddPRInlineComment("ws", "repo", 1, "x", api.InlineComment{Path: "f.go"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inline, _ := capturedBody["inline"].(map[string]interface{})
	if _, hasTo := inline["to"]; hasTo {
		t.Error("nil To should not appear in JSON body")
	}
}

func TestAddPRInlineComment_InvalidToZero(t *testing.T) {
	httpCalled := false
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
		w.WriteHeader(http.StatusCreated)
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "x", api.InlineComment{
		Path: "main.go",
		To:   intPtr(0),
	})
	if err == nil {
		t.Fatal("expected error for To=0")
	}
	if httpCalled {
		t.Error("HTTP request should not be made when precondition fails")
	}
}

func TestAddPRInlineComment_InvalidToNegative(t *testing.T) {
	httpCalled := false
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
		w.WriteHeader(http.StatusCreated)
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "x", api.InlineComment{
		Path: "main.go",
		To:   intPtr(-3),
	})
	if err == nil {
		t.Fatal("expected error for negative To")
	}
	if httpCalled {
		t.Error("HTTP request should not be made when precondition fails")
	}
}

func TestAddPRInlineComment_InvalidFromZero(t *testing.T) {
	httpCalled := false
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalled = true
		w.WriteHeader(http.StatusCreated)
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "x", api.InlineComment{
		Path: "main.go",
		To:   intPtr(5),
		From: intPtr(0),
	})
	if err == nil {
		t.Fatal("expected error for From=0")
	}
	if httpCalled {
		t.Error("HTTP request should not be made when precondition fails")
	}
}

func TestAddPRInlineComment_APIError400(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "This pull request is not open"},
		}); err != nil {
			t.Errorf("encoding error response: %v", err)
		}
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "x", api.InlineComment{Path: "main.go", To: intPtr(1)})
	if err == nil {
		t.Fatal("expected error on 400 response")
	}
	if !strings.Contains(err.Error(), "This pull request is not open") {
		t.Errorf("error should contain API message, got: %v", err)
	}
}

func TestAddPRInlineComment_APIError403(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "insufficient privileges"},
		}); err != nil {
			t.Errorf("encoding error response: %v", err)
		}
	}))

	err := client.AddPRInlineComment("ws", "repo", 5, "x", api.InlineComment{Path: "main.go", To: intPtr(1)})
	if err == nil {
		t.Fatal("expected error on 403 response")
	}
}

