package cmdutil

import (
	"net/http"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/iostreams"
)

// Factory is the dependency-injection container passed to every command constructor.
// Commands should accept *Factory and extract what they need — never access globals.
//
// Pattern (mirrors github.com/cli/cli):
//
//	func NewCmdList(f *cmdutil.Factory) *cobra.Command {
//	    opts := &ListOptions{IO: f.IOStreams, HttpClient: f.HttpClient}
//	    ...
//	}
type Factory struct {
	// IOStreams provides access to stdin/stdout/stderr with TTY detection.
	IOStreams *iostreams.IOStreams

	// HttpClient returns an authenticated *http.Client for Bitbucket API calls.
	// It transparently handles OAuth token refresh on 401 responses.
	HttpClient func() (*http.Client, error)

	// Config returns the loaded CLI configuration.
	Config func() (*config.Config, error)

	// BaseURL is the Bitbucket API base URL (default: https://api.bitbucket.org/2.0).
	BaseURL string

	// Workspace returns the active workspace slug using the resolution chain:
	// --workspace flag → BITBUCKET_WORKSPACE env → git remote inference → config default.
	// Commands should call this instead of reading config directly.
	Workspace func() (string, error)
}
