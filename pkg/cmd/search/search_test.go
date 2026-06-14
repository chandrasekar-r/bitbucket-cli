package search

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

func searchFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	return &cmdutil.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
		Workspace:  func() (string, error) { return "myws", nil },
	}, ios.Out.(*bytes.Buffer)
}

func TestSearchCodeCmd_Output(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/workspaces/myws/search/code") {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"file": map[string]string{"path": "pkg/main.go"},
					"content_matches": []map[string]interface{}{
						{
							"lines": []map[string]interface{}{
								{
									"line": float64(1),
									"segments": []map[string]string{{"text": "func main() {}"}},
								},
							},
						},
					},
				},
			},
			"pagelen": 50,
			"size":    1,
		})
	}))
	defer srv.Close()

	f, stdout := searchFactory(t, srv)
	cmd := newCmdCode(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"func main"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "pkg/main.go") {
		t.Errorf("expected file path in output, got: %q", out)
	}
	if !strings.Contains(out, "line 1:") || !strings.Contains(out, "func main") {
		t.Errorf("expected line content in output, got: %q", out)
	}
}

func TestSearchCodeCmd_Disabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"message":"not found"}}`))
	}))
	defer srv.Close()

	f, _ := searchFactory(t, srv)
	cmd := newCmdCode(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"query"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "code search is not enabled") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSearchReposCmd_Output(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/repositories/myws") {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{"slug": "my-cli", "name": "my-cli", "is_private": true},
			},
			"pagelen": 100,
			"size":    1,
		})
	}))
	defer srv.Close()

	f, stdout := searchFactory(t, srv)
	cmd := newCmdRepos(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"cli"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "my-cli") {
		t.Errorf("expected repo slug in output, got: %q", stdout.String())
	}
}

func TestSearchCodeCmd_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []map[string]interface{}{
				{
					"file": map[string]string{"path": "a.go"},
					"content_matches": []map[string]interface{}{
						{
							"lines": []map[string]interface{}{
								{
									"line": float64(3),
									"segments": []map[string]string{{"text": "x"}},
								},
							},
						},
					},
				},
			},
			"pagelen": 50,
			"size":    1,
		})
	}))
	defer srv.Close()

	f, stdout := searchFactory(t, srv)
	cmd := newCmdCode(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"x", "--json", "path,line,content"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"path"`) || !strings.Contains(stdout.String(), "a.go") {
		t.Errorf("expected JSON output, got: %q", stdout.String())
	}
}