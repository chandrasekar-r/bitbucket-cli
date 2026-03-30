package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chandrasekar-r/bitbucket-cli/internal/version"
)

// ErrSessionExpired is returned when the refresh token is invalid or expired.
// The caller should direct the user to run `bb auth login`.
var ErrSessionExpired = errors.New("session expired. Run: bb auth login")

// tokenResponse is the JSON shape returned by Bitbucket's token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scopes       string `json:"scopes"`
}

// RefreshAccessToken uses the stored refresh token to obtain a new access token.
//
// IMPORTANT: Bitbucket enforces rotating refresh tokens from May 4, 2026 —
// each refresh issues a new refresh token that MUST be stored immediately.
// This function holds the exclusive file lock across the full read→POST→write
// cycle to prevent concurrent refresh races that would invalidate the session.
func RefreshAccessToken(store *TokenStore, username string) error {
	return store.withLock(func(tf *tokenFile) error {
		acc, ok := tf.Accounts[username]
		if !ok {
			return ErrSessionExpired
		}
		if acc.AuthType == AuthTypeToken {
			// API tokens don't expire; nothing to refresh
			return nil
		}
		if acc.RefreshToken == "" {
			return ErrSessionExpired
		}

		newAccess, newRefresh, expiry, err := exchangeRefreshToken(acc.RefreshToken)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrSessionExpired, err)
		}

		// Store new tokens atomically before releasing the lock
		acc.AccessToken = newAccess
		acc.RefreshToken = newRefresh
		acc.Expiry = expiry
		tf.Accounts[username] = acc
		return nil
	})
}

// exchangeRefreshToken calls the Bitbucket token endpoint to exchange a
// refresh token for a new access + refresh token pair.
func exchangeRefreshToken(refreshToken string) (accessToken, newRefreshToken string, expiry time.Time, err error) {
	clientID := version.OAuthClientID
	if clientID == "" {
		return "", "", time.Time{}, errors.New("no OAuth client ID in this build")
	}

	body := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// 400/401 means the refresh token is invalid or rotated away
		var apiErr struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if json.Unmarshal(raw, &apiErr) == nil && apiErr.Error != "" {
			return "", "", time.Time{}, fmt.Errorf("%s: %s", apiErr.Error, apiErr.ErrorDescription)
		}
		return "", "", time.Time{}, fmt.Errorf("refresh failed (HTTP %d)", resp.StatusCode)
	}

	var tr tokenResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", "", time.Time{}, fmt.Errorf("parsing refresh response: %w", err)
	}
	if tr.AccessToken == "" {
		return "", "", time.Time{}, errors.New("empty access token in refresh response")
	}

	exp := time.Time{}
	if tr.ExpiresIn > 0 {
		exp = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	return tr.AccessToken, tr.RefreshToken, exp, nil
}
