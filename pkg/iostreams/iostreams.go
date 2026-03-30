package iostreams

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/term"
)

// IOStreams wraps stdin/stdout/stderr and provides TTY detection.
// Pass IOStreams into every command via the Factory — never use os.Stdout directly.
type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	colorEnabled bool
	noTTY        bool
}

// System returns IOStreams wired to the real OS file descriptors.
func System() *IOStreams {
	return &IOStreams{
		In:           os.Stdin,
		Out:          os.Stdout,
		ErrOut:       os.Stderr,
		colorEnabled: colorEnabled(),
	}
}

// Test returns IOStreams backed by byte buffers for use in unit tests.
// Returns the streams plus separate buffers for in, out, and errOut.
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &IOStreams{
		In:     io.NopCloser(in),
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

// SetNoTTY forces non-interactive mode regardless of actual terminal state.
// Used when --no-tty flag is passed.
func (s *IOStreams) SetNoTTY(v bool) {
	s.noTTY = v
}

// IsStdoutTTY reports whether stdout is an interactive terminal.
func (s *IOStreams) IsStdoutTTY() bool {
	if s.noTTY {
		return false
	}
	if f, ok := s.Out.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// IsStderrTTY reports whether stderr is an interactive terminal.
func (s *IOStreams) IsStderrTTY() bool {
	if s.noTTY {
		return false
	}
	if f, ok := s.ErrOut.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// ColorEnabled reports whether ANSI color output should be used.
func (s *IOStreams) ColorEnabled() bool {
	return s.colorEnabled && s.IsStdoutTTY()
}

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}
