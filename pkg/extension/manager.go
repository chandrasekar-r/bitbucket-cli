// Package extension handles discovering, installing, and running bb extensions.
// Extensions are executables named bb-<name> in ~/.config/bb/extensions/.
package extension

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtDir returns the directory where extensions are installed.
func ExtDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "bb", "extensions")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "bb", "extensions")
}

// Installed returns all installed extensions (directories containing a bb-* executable).
func Installed() ([]Extension, error) {
	dir := ExtDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var exts []Extension
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		binPath := filepath.Join(dir, name, name)
		if _, err := os.Stat(binPath); err == nil {
			exts = append(exts, Extension{Name: strings.TrimPrefix(name, "bb-"), Path: binPath})
		}
	}
	return exts, nil
}

// Extension represents an installed bb extension.
type Extension struct {
	Name string // command name (without "bb-" prefix)
	Path string // path to executable
}

// Install clones a git repository into the extensions directory.
func Install(repo string) (Extension, error) {
	dir := ExtDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Extension{}, err
	}

	name := filepath.Base(repo)
	name = strings.TrimSuffix(name, ".git")
	destDir := filepath.Join(dir, name)

	if _, err := os.Stat(destDir); err == nil {
		return Extension{}, fmt.Errorf("extension %q already installed at %s", name, destDir)
	}

	cmd := exec.Command("git", "clone", "--depth=1", repo, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return Extension{}, fmt.Errorf("cloning %s: %w", repo, err)
	}

	// Try to find the extension binary: prefer <name>/<name>, then <name>/<name-without-bb-prefix>.
	binPath := filepath.Join(destDir, name)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		// Fallback: try the name with the "bb-" prefix stripped (e.g. "ext-jira" inside "bb-ext-jira/")
		trimmed := strings.TrimPrefix(name, "bb-")
		if trimmed != name {
			binPath = filepath.Join(destDir, trimmed)
		}
	}

	return Extension{Name: strings.TrimPrefix(name, "bb-"), Path: binPath}, nil
}

// Remove deletes an installed extension.
func Remove(name string) error {
	// Guard against path traversal: names must not contain separators or dot sequences.
	if strings.ContainsAny(name, "/\\") || name == ".." || strings.Contains(name, "..") {
		return fmt.Errorf("invalid extension name %q", name)
	}
	dir := filepath.Join(ExtDir(), "bb-"+name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = filepath.Join(ExtDir(), name)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("extension %q is not installed", name)
	}
	return os.RemoveAll(dir)
}

// Run executes an extension with the given arguments.
func (e Extension) Run(args []string) error {
	cmd := exec.Command(e.Path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
