package extension

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtDir_ReturnsNonEmpty(t *testing.T) {
	dir := ExtDir()
	if dir == "" {
		t.Fatal("ExtDir() returned empty string")
	}
}

func TestExtDir_RespectsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	got := ExtDir()
	want := filepath.Join("/tmp/test-xdg", "bb", "extensions")
	if got != want {
		t.Errorf("ExtDir() = %q, want %q", got, want)
	}
}

func TestInstalled_EmptyWhenDirMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "nonexistent"))
	exts, err := Installed()
	if err != nil {
		t.Fatalf("Installed() error = %v, want nil", err)
	}
	if len(exts) != 0 {
		t.Errorf("Installed() returned %d extensions, want 0", len(exts))
	}
}

func TestInstalled_FindsExtension(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	extDir := filepath.Join(tmp, "bb", "extensions", "bb-hello")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(extDir, "bb-hello")
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\necho hello"), 0755); err != nil {
		t.Fatal(err)
	}

	exts, err := Installed()
	if err != nil {
		t.Fatalf("Installed() error = %v", err)
	}
	if len(exts) != 1 {
		t.Fatalf("Installed() returned %d extensions, want 1", len(exts))
	}
	if exts[0].Name != "hello" {
		t.Errorf("extension Name = %q, want %q", exts[0].Name, "hello")
	}
	if exts[0].Path != binPath {
		t.Errorf("extension Path = %q, want %q", exts[0].Path, binPath)
	}
}

func TestRemove_ErrorOnMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Remove should return a clear error when the extension is not installed,
	// so users know their command had no effect rather than silently succeeding.
	err := Remove("nonexistent")
	if err == nil {
		t.Error("Remove() should return error for non-installed extension, got nil")
	}
}

func TestRemove_RejectsPathTraversal(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, bad := range []string{"../../etc", "../secret", "a/b", `a\b`} {
		if err := Remove(bad); err == nil {
			t.Errorf("Remove(%q) should return error for path-traversal name, got nil", bad)
		}
	}
}
