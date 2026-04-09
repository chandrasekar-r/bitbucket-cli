// Package notify provides cross-platform desktop notifications.
package notify

import (
	"os/exec"
	"runtime"
)

// Send fires a desktop notification. Best-effort: errors are silently ignored.
func Send(title, message, subtitle string) {
	name, args := buildCommand(title, message, subtitle)
	if name == "" {
		return
	}
	cmd := exec.Command(name, args...)
	_ = cmd.Start() // fire and forget
}

// buildCommand returns the OS-specific notification command and arguments.
func buildCommand(title, message, subtitle string) (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		script := `display notification "` + escapeAS(message) + `" with title "` + escapeAS(title) + `"`
		if subtitle != "" {
			script += ` subtitle "` + escapeAS(subtitle) + `"`
		}
		return "osascript", []string{"-e", script}
	case "linux":
		return "notify-send", []string{title, message}
	case "windows":
		ps := `[System.Reflection.Assembly]::LoadWithPartialName('System.Windows.Forms') | Out-Null; ` +
			`$n = New-Object System.Windows.Forms.NotifyIcon; ` +
			`$n.Icon = [System.Drawing.SystemIcons]::Information; ` +
			`$n.Visible = $true; ` +
			`$n.ShowBalloonTip(5000, '` + title + `', '` + message + `', 'Info')`
		return "powershell", []string{"-Command", ps}
	default:
		return "", nil
	}
}

// escapeAS escapes double quotes for AppleScript strings.
func escapeAS(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			result = append(result, '\\')
		}
		result = append(result, s[i])
	}
	return string(result)
}
