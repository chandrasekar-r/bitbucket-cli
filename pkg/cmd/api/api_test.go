package api

import (
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

func TestNewCmdAPI_RequiresPath(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}
	cmd := NewCmdAPI(f)

	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no endpoint arg provided, got nil")
	}
}

func TestBuildNestedFields(t *testing.T) {
	fields := []string{
		"title=My PR",
		"source.branch.name=feature",
		"destination.branch.name=main",
	}

	result := buildNestedFields(fields)

	// Check top-level title
	if result["title"] != "My PR" {
		t.Errorf("expected title='My PR', got %v", result["title"])
	}

	// Check nested source.branch.name
	source, ok := result["source"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected source to be a map, got %T", result["source"])
	}
	branch, ok := source["branch"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected source.branch to be a map, got %T", source["branch"])
	}
	if branch["name"] != "feature" {
		t.Errorf("expected source.branch.name='feature', got %v", branch["name"])
	}

	// Check nested destination.branch.name
	dest, ok := result["destination"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected destination to be a map, got %T", result["destination"])
	}
	destBranch, ok := dest["branch"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected destination.branch to be a map, got %T", dest["branch"])
	}
	if destBranch["name"] != "main" {
		t.Errorf("expected destination.branch.name='main', got %v", destBranch["name"])
	}
}

func TestBuildNestedFields_Simple(t *testing.T) {
	fields := []string{"key=value", "another=thing"}
	result := buildNestedFields(fields)

	if result["key"] != "value" {
		t.Errorf("expected key='value', got %v", result["key"])
	}
	if result["another"] != "thing" {
		t.Errorf("expected another='thing', got %v", result["another"])
	}
}

func TestBuildNestedFields_Empty(t *testing.T) {
	result := buildNestedFields(nil)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestBuildNestedFields_MalformedSkipped(t *testing.T) {
	fields := []string{"noequals", "good=value"}
	result := buildNestedFields(fields)

	if len(result) != 1 {
		t.Errorf("expected 1 field, got %d", len(result))
	}
	if result["good"] != "value" {
		t.Errorf("expected good='value', got %v", result["good"])
	}
}

func TestBuildNestedFields_ValueWithEquals(t *testing.T) {
	fields := []string{"query=foo=bar"}
	result := buildNestedFields(fields)

	if result["query"] != "foo=bar" {
		t.Errorf("expected query='foo=bar', got %v", result["query"])
	}
}
