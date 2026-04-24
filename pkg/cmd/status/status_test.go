package status

import (
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestNewCmdStatus(t *testing.T) {
	cmd := NewCmdStatus(nil)

	if cmd.Use != "status" {
		t.Errorf("expected Use=%q, got %q", "status", cmd.Use)
	}

	// Verify --json flag is registered
	f := cmd.Flags().Lookup("json")
	if f == nil {
		t.Fatal("expected --json flag to be registered")
	}

	// Verify --jq flag is registered
	f = cmd.Flags().Lookup("jq")
	if f == nil {
		t.Fatal("expected --jq flag to be registered")
	}
}

func TestIsReviewerPending(t *testing.T) {
	tests := []struct {
		name     string
		pr       api.PullRequest
		username string
		want     bool
	}{
		{
			name: "reviewer who has not approved",
			pr: api.PullRequest{
				Reviewers: []struct {
					DisplayName string `json:"display_name"`
					Username    string `json:"username"`
				}{
					{Username: "alice"},
				},
				Participants: []struct {
					User struct {
						DisplayName string `json:"display_name"`
						Username    string `json:"username"`
					} `json:"user"`
					Role     string `json:"role"`
					Approved bool   `json:"approved"`
				}{
					{User: struct {
						DisplayName string `json:"display_name"`
						Username    string `json:"username"`
					}{Username: "alice"}, Approved: false},
				},
			},
			username: "alice",
			want:     true,
		},
		{
			name: "reviewer who has approved",
			pr: api.PullRequest{
				Reviewers: []struct {
					DisplayName string `json:"display_name"`
					Username    string `json:"username"`
				}{
					{Username: "alice"},
				},
				Participants: []struct {
					User struct {
						DisplayName string `json:"display_name"`
						Username    string `json:"username"`
					} `json:"user"`
					Role     string `json:"role"`
					Approved bool   `json:"approved"`
				}{
					{User: struct {
						DisplayName string `json:"display_name"`
						Username    string `json:"username"`
					}{Username: "alice"}, Approved: true},
				},
			},
			username: "alice",
			want:     false,
		},
		{
			name: "not a reviewer",
			pr: api.PullRequest{
				Reviewers: []struct {
					DisplayName string `json:"display_name"`
					Username    string `json:"username"`
				}{
					{Username: "bob"},
				},
			},
			username: "alice",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReviewerPending(tt.pr, tt.username)
			if got != tt.want {
				t.Errorf("isReviewerPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"this is a very long title", 15, "this is a ve..."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestHeaderStyle(t *testing.T) {
	plain := headerStyle("Test", false)
	if plain != "Test" {
		t.Errorf("expected plain %q, got %q", "Test", plain)
	}

	colored := headerStyle("Test", true)
	if colored == "Test" {
		t.Error("expected ANSI codes in colored output")
	}
	if colored != "\033[1;36mTest\033[0m" {
		t.Errorf("unexpected colored output: %q", colored)
	}
}
