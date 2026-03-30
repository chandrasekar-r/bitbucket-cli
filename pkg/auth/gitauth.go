package auth

import (
	"fmt"
	"net/url"
	"os"
)

// InjectCloneAuth takes an HTTPS Bitbucket clone URL and embeds credentials
// so that `git clone` does not prompt for a password.
//
// Bitbucket credential formats:
//   - OAuth token:  https://x-token-auth:{access_token}@bitbucket.org/ws/repo.git
//   - API token:    https://{username}:{token}@bitbucket.org/ws/repo.git
//
// SSH URLs are returned unchanged — SSH uses key-based auth and needs no modification.
// If no credentials are available, the URL is returned unchanged (git may prompt).
func InjectCloneAuth(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, nil // return unchanged; git will handle (or fail)
	}

	// SSH and non-HTTPS protocols don't need credential injection
	if parsed.Scheme != "https" {
		return rawURL, nil
	}

	// 1. Check environment variable token auth (CI/headless — highest priority)
	envUser := os.Getenv("BITBUCKET_USERNAME")
	envToken := os.Getenv("BITBUCKET_TOKEN")
	if envUser != "" && envToken != "" {
		parsed.User = url.UserPassword(envUser, envToken)
		return parsed.String(), nil
	}

	// 2. Check stored credentials
	store := NewTokenStore()
	acc, err := store.GetActive()
	if err != nil || acc == nil {
		return rawURL, nil // no stored credentials; let git handle it
	}

	// Refresh the token if it has expired
	if acc.IsExpired() && acc.AuthType == AuthTypeOAuth {
		if rfErr := RefreshAccessToken(store, acc.Username); rfErr != nil {
			return "", fmt.Errorf("token expired and refresh failed: %w\nRun: bb auth login", rfErr)
		}
		acc, err = store.GetActive()
		if err != nil || acc == nil {
			return rawURL, nil
		}
	}

	switch acc.AuthType {
	case AuthTypeOAuth:
		// Bitbucket OAuth token: x-token-auth is the required username
		parsed.User = url.UserPassword("x-token-auth", acc.AccessToken)
	case AuthTypeToken:
		// Bitbucket API token uses Basic auth: username:token
		parsed.User = url.UserPassword(acc.Username, acc.AccessToken)
	}

	return parsed.String(), nil
}
