package notify

import (
	"runtime"
	"testing"
)

func TestBuildCommand_CurrentPlatform(t *testing.T) {
	name, args := buildCommand("Test Title", "Test message", "")
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		if name == "" {
			t.Fatalf("expected non-empty command name on %s", runtime.GOOS)
		}
		if len(args) == 0 {
			t.Fatalf("expected non-empty args on %s", runtime.GOOS)
		}
	default:
		t.Skipf("unsupported platform %s", runtime.GOOS)
	}
}

func TestBuildCommand_DarwinSubtitle(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only test")
	}
	name, args := buildCommand("Title", "Body", "Sub")
	if name != "osascript" {
		t.Fatalf("expected osascript, got %s", name)
	}
	// The script should contain the subtitle clause
	script := args[len(args)-1]
	if got := `subtitle "Sub"`; !contains(script, got) {
		t.Errorf("expected script to contain %q, got %q", got, script)
	}
}

func TestEscapeAS(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{`say "hi"`, `say \"hi\"`},
		{`""`, `\"\"`},
		{"", ""},
		{"no quotes here", "no quotes here"},
	}
	for _, tt := range tests {
		got := escapeAS(tt.input)
		if got != tt.want {
			t.Errorf("escapeAS(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
