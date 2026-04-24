package cmdutil

import "fmt"

// AuthError is returned when the user is not authenticated or their session expired.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "not authenticated. Run: bb auth login"
}

// NotFoundError wraps a 404 API response with a human-readable resource description.
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return "not found"
}

// NoTTYError is returned when an interactive operation is attempted in non-TTY mode
// without the required --force flag.
type NoTTYError struct {
	Operation string
}

func (e *NoTTYError) Error() string {
	return fmt.Sprintf(
		"this operation requires confirmation in interactive mode. Use --force to skip the prompt.\nOperation: %s",
		e.Operation,
	)
}

// FlagError signals incorrect flag usage (wrong type, missing required flag, etc.).
type FlagError struct {
	Err error
}

func (e *FlagError) Error() string {
	return e.Err.Error()
}

func (e *FlagError) Unwrap() error {
	return e.Err
}
