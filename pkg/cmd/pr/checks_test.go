package pr

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

func checksFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	return &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
	}, ios.Out.(*bytes.Buffer)
}

func checksServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/pullrequests/5"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    5,
				"title": "Test PR",
				"source": map[string]interface{}{
					"branch": map[string]string{"name": "feature"},
					"commit": map[string]string{"hash": "abc123def"},
				},
				"destination": map[string]interface{}{
					"branch": map[string]string{"name": "main"},
				},
			})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/commit/abc123def/statuses"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"values": []map[string]string{
					{
						"key":         "pipelines",
						"state":       "SUCCESSFUL",
						"name":        "Bitbucket Pipelines",
						"description": "Build passed",
						"url":         "https://bitbucket.org/ws/repo/pipelines/results/1",
					},
				},
				"pagelen": 50,
				"size":    1,
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestChecksCmd_TableOutput(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := checksServer(t)
	defer srv.Close()

	f, stdout := checksFactory(t, srv)
	cmd := newCmdChecks(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "pipelines") {
		t.Errorf("expected pipelines key in output, got: %q", out)
	}
	if !strings.Contains(out, "SUCCESSFUL") {
		t.Errorf("expected SUCCESSFUL state in output, got: %q", out)
	}
}

func TestChecksCmd_JSONOutput(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := checksServer(t)
	defer srv.Close()

	f, stdout := checksFactory(t, srv)
	cmd := newCmdChecks(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5", "--json", "key,state,url"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, `"key"`) || !strings.Contains(out, "pipelines") {
		t.Errorf("expected JSON with key field, got: %q", out)
	}
}

func TestChecksCmd_NoStatuses(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/pullrequests/5") {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 5,
				"source": map[string]interface{}{
					"commit": map[string]string{"hash": "abc123def"},
				},
			})
			return
		}
		if strings.Contains(r.URL.Path, "/statuses") {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"values":  []any{},
				"pagelen": 50,
				"size":    0,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	f, stdout := checksFactory(t, srv)
	cmd := newCmdChecks(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "No checks found") {
		t.Errorf("expected no checks message, got: %q", stdout.String())
	}
}

func TestChecksCmd_NoCommitHash(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 5,
			"source": map[string]interface{}{
				"branch": map[string]string{"name": "feature"},
			},
		})
	}))
	defer srv.Close()

	f, _ := checksFactory(t, srv)
	cmd := newCmdChecks(f)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"5"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when PR has no commit hash")
	}
	if !strings.Contains(err.Error(), "no source commit hash") {
		t.Errorf("unexpected error: %v", err)
	}
}