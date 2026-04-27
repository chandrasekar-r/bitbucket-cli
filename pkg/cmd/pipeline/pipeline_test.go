package pipeline

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

// changeToBitbucketRepo sets up a temp git repo with a Bitbucket origin remote
// so that repoCtx returns workspace="testws" and slug="testrepo".
func changeToBitbucketRepo(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("remote", "add", "origin", "https://bitbucket.org/testws/testrepo.git")
	t.Chdir(dir)
}

func makeFactory(t *testing.T, srv *httptest.Server) (*cmdutil.Factory, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:     io.NopCloser(bytes.NewBufferString("")),
		Out:    out,
		ErrOut: &bytes.Buffer{},
	}
	return &cmdutil.Factory{
		IOStreams:   ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{}, nil },
		BaseURL:    srv.URL + "/2.0",
	}, out
}

// encodeJSON is a test helper that writes a JSON-encoded value to w.
// Uses t.Errorf (not t.Fatalf) because HTTP handlers run in a separate goroutine.
func encodeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("encoding JSON response: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Pure unit tests for helpers.go
// ---------------------------------------------------------------------------

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "-"},
		{30, "30s"},
		{90, "1m30s"},
		{3661, "61m1s"},
	}
	for _, tc := range tests {
		got := formatDuration(tc.seconds)
		if got != tc.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}

func TestStatusColor_NoColor(t *testing.T) {
	got := statusColor("SUCCESSFUL", "", false)
	if got != "SUCCESSFUL" {
		t.Errorf("statusColor no-color: got %q, want %q", got, "SUCCESSFUL")
	}
	// Must not contain ANSI codes
	if strings.Contains(got, "\033[") {
		t.Errorf("statusColor no-color: unexpected ANSI codes in %q", got)
	}
}

func TestStatusColor_WithColor(t *testing.T) {
	got := statusColor("SUCCESSFUL", "", true)
	if !strings.Contains(got, "\033[32m") {
		t.Errorf("statusColor SUCCESSFUL with color: expected green ANSI code, got %q", got)
	}
}

func TestStatusColor_Result(t *testing.T) {
	// When result is provided, statusColor should use result string
	got := statusColor("COMPLETED", "FAILED", true)
	if !strings.Contains(got, "FAILED") {
		t.Errorf("statusColor result: expected FAILED in output, got %q", got)
	}
	// FAILED should be red
	if !strings.Contains(got, "\033[31m") {
		t.Errorf("statusColor FAILED with color: expected red ANSI code, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// bb pipeline list
// ---------------------------------------------------------------------------

func TestPipelineList_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/pipelines") {
			http.NotFound(w, r)
			return
		}
		encodeJSON(t, w, map[string]any{
			"values": []map[string]any{
				{
					"uuid":                "{pipe-1}",
					"build_number":        7,
					"created_on":          "2026-04-01T12:00:00Z",
					"duration_in_seconds": 120,
					"state": map[string]any{
						"name":   "COMPLETED",
						"result": map[string]string{"name": "SUCCESSFUL"},
					},
					"target": map[string]any{
						"ref_name": "main",
						"ref_type": "branch",
					},
					"creator": map[string]string{"display_name": "Alice"},
				},
			},
			"pagelen": 1,
			"size":    1,
		})
	}))
	t.Cleanup(srv.Close)

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pipeline list: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "7") {
		t.Errorf("output missing build number 7: %q", got)
	}
	if !strings.Contains(got, "main") {
		t.Errorf("output missing branch name 'main': %q", got)
	}
}

func TestPipelineList_Empty(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(t, w, map[string]any{
			"values":  []map[string]any{},
			"pagelen": 0,
			"size":    0,
		})
	}))
	t.Cleanup(srv.Close)

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pipeline list empty: %v", err)
	}

	if !strings.Contains(out.String(), "No pipeline runs found") {
		t.Errorf("expected 'No pipeline runs found', got: %q", out.String())
	}
}

func TestPipelineList_BranchFilter(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(t, w, map[string]any{
			"values": []map[string]any{
				{
					"uuid":         "{pipe-main}",
					"build_number": 10,
					"created_on":   "2026-04-01T12:00:00Z",
					"state": map[string]any{
						"name":   "COMPLETED",
						"result": map[string]string{"name": "SUCCESSFUL"},
					},
					"target": map[string]any{
						"ref_name": "main",
						"ref_type": "branch",
					},
					"creator": map[string]string{"display_name": "Alice"},
				},
				{
					"uuid":         "{pipe-dev}",
					"build_number": 11,
					"created_on":   "2026-04-02T12:00:00Z",
					"state": map[string]any{
						"name":   "COMPLETED",
						"result": map[string]string{"name": "FAILED"},
					},
					"target": map[string]any{
						"ref_name": "develop",
						"ref_type": "branch",
					},
					"creator": map[string]string{"display_name": "Bob"},
				},
			},
			"pagelen": 2,
			"size":    2,
		})
	}))
	t.Cleanup(srv.Close)

	f, out := makeFactory(t, srv)
	cmd := newCmdList(f)
	cmd.SetArgs([]string{"--branch", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pipeline list --branch: %v", err)
	}

	got := out.String()
	// Build 10 on main should be present
	if !strings.Contains(got, "10") {
		t.Errorf("expected build 10 (main branch) in output: %q", got)
	}
	// Build 11 on develop should be excluded
	if strings.Contains(got, "11") {
		t.Errorf("expected build 11 (develop branch) to be filtered out: %q", got)
	}
	if strings.Contains(got, "develop") {
		t.Errorf("expected 'develop' branch to be filtered out: %q", got)
	}
}

// ---------------------------------------------------------------------------
// bb pipeline cancel
// ---------------------------------------------------------------------------

func TestPipelineCancel_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	const pipelineUUID = "{test-cancel-uuid}"
	stopped := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/stopPipeline") && r.Method == http.MethodPost {
			stopped = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	f, out := makeFactory(t, srv)
	cmd := newCmdCancel(f)
	cmd.SetArgs([]string{pipelineUUID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pipeline cancel: %v", err)
	}

	if !stopped {
		t.Error("expected stopPipeline endpoint to be called")
	}

	got := out.String()
	if !strings.Contains(got, pipelineUUID) {
		t.Errorf("output missing pipeline UUID %q: %q", pipelineUUID, got)
	}
	if !strings.Contains(got, "cancelled") {
		t.Errorf("output missing 'cancelled': %q", got)
	}
}

func TestPipelineCancel_APIError(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"pipeline not found"}}`, http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	f, _ := makeFactory(t, srv)
	cmd := newCmdCancel(f)
	cmd.SetArgs([]string{"{nonexistent-uuid}"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error from pipeline cancel with 404 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// bb pipeline run
// ---------------------------------------------------------------------------

func TestPipelineRun_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
		encodeJSON(t, w, map[string]any{
			"uuid":         "{triggered-uuid}",
			"build_number": 42,
			"state":        map[string]string{"name": "PENDING"},
			"target": map[string]any{
				"ref_name": "main",
				"ref_type": "branch",
			},
			"links": map[string]any{
				"html": map[string]string{
					"href": "https://bitbucket.org/testws/testrepo/pipelines/42",
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	f, out := makeFactory(t, srv)
	cmd := newCmdRun(f)
	cmd.SetArgs([]string{"--branch", "main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pipeline run: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "42") {
		t.Errorf("expected pipeline number 42 in output: %q", got)
	}
}

// ---------------------------------------------------------------------------
// bb pipeline view
// ---------------------------------------------------------------------------

func TestPipelineView_Success(t *testing.T) {
	changeToBitbucketRepo(t)

	const pipelineUUID = "{view-test-uuid}"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/steps/"):
			// ListSteps response
			encodeJSON(t, w, map[string]any{
				"values": []map[string]any{
					{
						"uuid": "{step-1}",
						"name": "Build",
						"state": map[string]any{
							"name":   "COMPLETED",
							"result": map[string]string{"name": "SUCCESSFUL"},
						},
						"duration_in_seconds": 60,
					},
				},
			})
		case strings.Contains(r.URL.Path, "/pipelines/"):
			// GetPipeline response
			encodeJSON(t, w, map[string]any{
				"uuid":                pipelineUUID,
				"build_number":        99,
				"created_on":          "2026-04-10T09:00:00Z",
				"duration_in_seconds": 75,
				"state": map[string]any{
					"name":   "COMPLETED",
					"result": map[string]string{"name": "SUCCESSFUL"},
				},
				"target": map[string]any{
					"ref_name": "main",
					"ref_type": "branch",
					"commit":   map[string]string{"hash": "abc1234"},
				},
				"links": map[string]any{
					"html": map[string]string{
						"href": "https://bitbucket.org/testws/testrepo/pipelines/99",
					},
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	f, out := makeFactory(t, srv)
	cmd := newCmdView(f)
	cmd.SetArgs([]string{pipelineUUID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pipeline view: %v", err)
	}

	got := out.String()
	// Output should contain UUID or build number
	if !strings.Contains(got, "99") && !strings.Contains(got, pipelineUUID) {
		t.Errorf("expected build number 99 or UUID in output: %q", got)
	}
}
