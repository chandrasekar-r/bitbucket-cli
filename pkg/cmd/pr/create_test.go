package pr

import "testing"

func TestFormatCommitsAsDescription(t *testing.T) {
	tests := []struct {
		name    string
		commits []string
		want    string
	}{
		{
			name:    "empty commits",
			commits: nil,
			want:    "",
		},
		{
			name:    "single commit",
			commits: []string{"abc1234 Add login feature"},
			want:    "Add login feature",
		},
		{
			name: "multiple commits",
			commits: []string{
				"abc1234 Add login feature",
				"def5678 Fix password validation",
				"ghi9012 Update tests",
			},
			want: "## Changes\n\n- Add login feature\n- Fix password validation\n- Update tests\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCommitsAsDescription(tt.commits)
			if got != tt.want {
				t.Errorf("formatCommitsAsDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractJiraKey(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"feature/PROJ-123-add-login", "PROJ-123"},
		{"PROJ-456-fix", "PROJ-456"},
		{"bugfix/no-ticket", ""},
		{"main", ""},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			got := extractJiraKey(tt.branch)
			if got != tt.want {
				t.Errorf("extractJiraKey(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}
