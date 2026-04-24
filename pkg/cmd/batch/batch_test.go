package batch

import (
	"path/filepath"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestNewCmdBatch(t *testing.T) {
	cmd := NewCmdBatch(nil)
	if cmd == nil {
		t.Fatal("NewCmdBatch(nil) returned nil")
	}
	if cmd.Use == "" {
		t.Error("command Use is empty")
	}
	if cmd.Short == "" {
		t.Error("command Short is empty")
	}

	// Check flags exist.
	if f := cmd.Flags().Lookup("repos"); f == nil {
		t.Error("missing --repos flag")
	}
	if f := cmd.Flags().Lookup("concurrency"); f == nil {
		t.Error("missing --concurrency flag")
	}
}

func TestFilterRepos(t *testing.T) {
	repos := []api.Repository{
		{Slug: "backend-auth"},
		{Slug: "backend-api"},
		{Slug: "frontend-web"},
		{Slug: "frontend-mobile"},
		{Slug: "infra-terraform"},
	}

	tests := []struct {
		pattern string
		want    []string
		wantErr bool
	}{
		{
			pattern: "backend-*",
			want:    []string{"backend-auth", "backend-api"},
		},
		{
			pattern: "frontend-*",
			want:    []string{"frontend-web", "frontend-mobile"},
		},
		{
			pattern: "infra-terraform",
			want:    []string{"infra-terraform"},
		},
		{
			pattern: "*-api",
			want:    []string{"backend-api"},
		},
		{
			pattern: "no-match-*",
			want:    nil,
		},
		{
			pattern: "[invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got, err := filterRepos(repos, tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotSlugs := make([]string, len(got))
			for i, r := range got {
				gotSlugs[i] = r.Slug
			}

			if len(gotSlugs) != len(tt.want) {
				t.Fatalf("got %v, want %v", gotSlugs, tt.want)
			}
			for i := range tt.want {
				if gotSlugs[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, gotSlugs[i], tt.want[i])
				}
			}
		})
	}
}

func TestFilepathMatchBehavior(t *testing.T) {
	// Verify filepath.Match handles our expected patterns correctly.
	tests := []struct {
		pattern, name string
		want          bool
	}{
		{"backend-*", "backend-auth", true},
		{"backend-*", "frontend-web", false},
		{"*-api", "backend-api", true},
		{"*", "anything", true},
		{"exact", "exact", true},
		{"exact", "other", false},
	}

	for _, tt := range tests {
		got, err := filepath.Match(tt.pattern, tt.name)
		if err != nil {
			t.Fatalf("filepath.Match(%q, %q) error: %v", tt.pattern, tt.name, err)
		}
		if got != tt.want {
			t.Errorf("filepath.Match(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
		}
	}
}
