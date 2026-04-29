package output_test

import (
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/output"
)

func TestBitbucketMarkdownGuide_NonEmpty(t *testing.T) {
	guide := output.BitbucketMarkdownGuide()
	if guide == "" {
		t.Fatal("BitbucketMarkdownGuide returned empty string")
	}
}

func TestBitbucketMarkdownGuide_ContainsExpectedSyntax(t *testing.T) {
	guide := output.BitbucketMarkdownGuide()
	checks := []struct {
		name    string
		snippet string
	}{
		{"bold", "**bold**"},
		{"italic", "*italic*"},
		{"code block open", "```"},
		{"go language hint", "```go"},
		{"table header", "| column |"},
		{"at-mention", "@username"},
		{"issue link", "#123"},
		{"emoji shortcode", ":emoji_name:"},
		{"horizontal rule", "---"},
		{"ordered list", "1. ordered item"},
	}
	for _, c := range checks {
		if !strings.Contains(guide, c.snippet) {
			t.Errorf("guide missing %s: expected to find %q", c.name, c.snippet)
		}
	}
}
