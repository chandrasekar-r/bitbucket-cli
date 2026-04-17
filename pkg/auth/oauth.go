package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/chandrasekar-r/bitbucket-cli/internal/version"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/browser"
	"golang.org/x/oauth2"
)

const (
	authURL  = "https://bitbucket.org/site/oauth2/authorize"
	tokenURL = "https://bitbucket.org/site/oauth2/access_token"

	// Scopes needed for full bb functionality
	oauthScopes = "repository repository:write pullrequest pullrequest:write " +
		"issue issue:write pipeline pipeline:write snippet snippet:write " +
		"account team project project:write webhook runner runner:write"
)

// OAuthResult holds the token received after a successful OAuth flow.
type OAuthResult struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
}

// RunOAuthFlow opens the browser for the Bitbucket OAuth 2.0 Authorization Code
// flow with a localhost loopback redirect, waits for the callback, and exchanges
// the code for tokens.
//
// clientID must be set (injected at build time via internal/version.OAuthClientID).
// If empty, returns an informative error directing the user to --with-token.
func RunOAuthFlow(ctx context.Context) (*OAuthResult, error) {
	clientID := version.OAuthClientID
	if clientID == "" {
		return nil, errors.New(
			"no OAuth client ID configured in this build\n" +
				"use `bb auth login --with-token` and provide a Bitbucket API token",
		)
	}

	// Start loopback server on a random available port (RFC 8252 §7.3)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting loopback server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: version.OAuthClientSecret, // used in token exchange Basic auth
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      []string{oauthScopes},
	}

	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	// Build auth URL — standard Authorization Code (no PKCE; not confirmed by Bitbucket docs)
	authCodeURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Printf("Opening browser for authentication...\n%s\n\n", authCodeURL)
	if err := browser.Open(authCodeURL); err != nil {
		// Non-fatal: user can copy-paste the URL
		fmt.Printf("Could not open browser automatically. Please open the URL above manually.\n")
	}

	// Wait for callback with a 5-minute timeout
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{
		Handler: callbackHandler(state, codeCh, errCh),
	}
	go srv.Serve(listener)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		srv.Close()
		return nil, fmt.Errorf("OAuth callback error: %w", err)
	case <-ctx.Done():
		srv.Close()
		return nil, errors.New("authentication timed out after 5 minutes: run `bb auth login` to try again")
	}
	srv.Close()

	// Exchange authorization code for tokens
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code for token: %w", err)
	}

	return &OAuthResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}, nil
}

// callbackHandler handles the OAuth redirect callback.
func callbackHandler(expectedState string, codeCh chan<- string, errCh chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			desc := q.Get("error_description")
			http.Error(w, "Authentication failed", http.StatusBadRequest)
			decoded, _ := url.QueryUnescape(desc)
			errCh <- fmt.Errorf("%s: %s", errParam, decoded)
			return
		}

		if state := q.Get("state"); state != expectedState {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			errCh <- errors.New("state mismatch — possible CSRF")
			return
		}

		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			errCh <- errors.New("no authorization code in callback")
			return
		}

		// Show success page in browser
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;padding:2em">
<h2>&#10003; Authenticated successfully</h2>
<p>You can close this tab and return to your terminal.</p>
</body></html>`)

		codeCh <- code
	})
}

// randomState generates a random hex string for OAuth state parameter (CSRF protection).
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
