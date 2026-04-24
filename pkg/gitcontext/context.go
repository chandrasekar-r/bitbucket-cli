package gitcontext

import (
	"os/exec"
	"regexp"
	"strings"
)

// RepoContext holds the workspace slug and repo slug inferred from the git remote.
type RepoContext struct {
	Workspace string
	RepoSlug  string
}

// FromRemote parses the git remote URL of the current directory and extracts
// the Bitbucket workspace slug and repository slug.
//
// Supports both SSH and HTTPS Bitbucket remote formats:
//   - SSH:   git@bitbucket.org:workspace/repo.git
//   - HTTPS: https://bitbucket.org/workspace/repo.git
//
// Returns nil if not inside a git repo or the remote is not a Bitbucket URL.
func FromRemote() *RepoContext {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return nil
	}
	return ParseRemoteURL(strings.TrimSpace(string(out)))
}

var (
	// git@bitbucket.org:workspace/repo.git
	sshPattern = regexp.MustCompile(`^git@bitbucket\.org[:/]([^/]+)/([^/]+?)(?:\.git)?$`)
	// https://bitbucket.org/workspace/repo.git
	// https://user@bitbucket.org/workspace/repo.git
	httpsPattern = regexp.MustCompile(`^https?://(?:[^@]+@)?bitbucket\.org/([^/]+)/([^/]+?)(?:\.git)?$`)
)

// ParseRemoteURL extracts workspace and repo slug from a Bitbucket remote URL.
// Returns nil if the URL is not a recognized Bitbucket format.
func ParseRemoteURL(remoteURL string) *RepoContext {
	for _, re := range []*regexp.Regexp{sshPattern, httpsPattern} {
		if m := re.FindStringSubmatch(remoteURL); m != nil {
			return &RepoContext{
				Workspace: m[1],
				RepoSlug:  m[2],
			}
		}
	}
	return nil
}
