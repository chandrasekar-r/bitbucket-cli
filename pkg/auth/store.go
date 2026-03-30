// Package auth handles Bitbucket Cloud authentication: OAuth 2.0 flows,
// token storage, and token refresh with rotation-safe file locking.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/config"
	"github.com/gofrs/flock"
)

// AuthType identifies how credentials were obtained.
type AuthType string

const (
	AuthTypeOAuth AuthType = "oauth"
	AuthTypeToken AuthType = "token" // Bitbucket API Token via Basic auth
)

// Account holds all credentials for a single authenticated Bitbucket account.
type Account struct {
	Username       string    `json:"username"`
	Email          string    `json:"email,omitempty"`
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token,omitempty"` // empty for token auth
	TokenType      string    `json:"token_type"`
	Expiry         time.Time `json:"expiry,omitempty"`
	WorkspaceSlugs []string  `json:"workspace_slugs"`
	AuthType       AuthType  `json:"auth_type"`
}

// IsExpired reports whether the access token has expired (with a 30s buffer).
func (a *Account) IsExpired() bool {
	if a.Expiry.IsZero() {
		return false // token auth has no expiry
	}
	return time.Now().After(a.Expiry.Add(-30 * time.Second))
}

// tokenFile is the on-disk structure of tokens.json.
type tokenFile struct {
	Accounts      map[string]*Account `json:"accounts"` // keyed by username
	ActiveAccount string              `json:"active_account"`
}

// TokenStore manages reading and writing ~/.config/bb/tokens.json.
// All write operations acquire an exclusive file lock to prevent concurrent
// refresh races (critical for Bitbucket rotating refresh tokens, May 4 2026+).
type TokenStore struct {
	path string
}

// NewTokenStore returns a TokenStore pointed at the default tokens file.
func NewTokenStore() *TokenStore {
	return &TokenStore{path: config.TokensFile()}
}

// NewTokenStoreAt returns a TokenStore for a custom path (used in tests).
func NewTokenStoreAt(path string) *TokenStore {
	return &TokenStore{path: path}
}

// Load reads all accounts from disk. Returns an empty store if the file
// doesn't exist yet.
func (s *TokenStore) Load() (*tokenFile, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &tokenFile{Accounts: map[string]*Account{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading token store: %w", err)
	}
	var tf tokenFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parsing token store: %w", err)
	}
	if tf.Accounts == nil {
		tf.Accounts = map[string]*Account{}
	}
	return &tf, nil
}

// Save writes tf to disk atomically (write to temp file, rename) with 0600
// permissions so only the owning user can read credentials.
func (s *TokenStore) Save(tf *tokenFile) error {
	if err := os.MkdirAll(dirOf(s.path), 0700); err != nil {
		return fmt.Errorf("creating token store dir: %w", err)
	}
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token store: %w", err)
	}
	// FINDING-005: use a randomly-named temp file (not a predictable .tmp path)
	// to eliminate the window where tokens are readable at a known location.
	dir := dirOf(s.path)
	tmpFile, err := os.CreateTemp(dir, ".tokens-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp token file: %w", err)
	}
	tmpName := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing token store: %w", err)
	}
	if err := tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("setting token file permissions: %w", err)
	}
	tmpFile.Close()
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("committing token store: %w", err)
	}
	return nil
}

// GetActive returns the active account, or nil if not authenticated.
func (s *TokenStore) GetActive() (*Account, error) {
	tf, err := s.Load()
	if err != nil {
		return nil, err
	}
	if tf.ActiveAccount == "" || len(tf.Accounts) == 0 {
		return nil, nil
	}
	acc, ok := tf.Accounts[tf.ActiveAccount]
	if !ok {
		return nil, nil
	}
	return acc, nil
}

// SetAccount stores or updates an account and sets it as active.
// Acquires an exclusive lock so concurrent logins don't clobber each other.
func (s *TokenStore) SetAccount(acc *Account) error {
	return s.withLock(func(tf *tokenFile) error {
		tf.Accounts[acc.Username] = acc
		tf.ActiveAccount = acc.Username
		return nil
	})
}

// RemoveAccount deletes an account by username.
func (s *TokenStore) RemoveAccount(username string) error {
	return s.withLock(func(tf *tokenFile) error {
		delete(tf.Accounts, username)
		if tf.ActiveAccount == username {
			// Pick a remaining account or clear active
			tf.ActiveAccount = ""
			for u := range tf.Accounts {
				tf.ActiveAccount = u
				break
			}
		}
		return nil
	})
}

// UpdateTokens atomically replaces access + refresh tokens for username.
// This is called after each OAuth token refresh to store the new rotating token.
// Holding the exclusive lock across read→write prevents concurrent refresh races.
func (s *TokenStore) UpdateTokens(username, accessToken, refreshToken string, expiry time.Time) error {
	return s.withLock(func(tf *tokenFile) error {
		acc, ok := tf.Accounts[username]
		if !ok {
			return fmt.Errorf("account %q not found in store", username)
		}
		acc.AccessToken = accessToken
		acc.RefreshToken = refreshToken
		acc.Expiry = expiry
		tf.Accounts[username] = acc
		return nil
	})
}

// withLock acquires an exclusive file lock on tokens.json.lock, loads the
// token file, calls fn, saves if fn returns nil, then releases the lock.
//
// The exclusive lock is the critical safety mechanism for Bitbucket's rotating
// refresh tokens: only one process can read-refresh-write at a time.
func (s *TokenStore) withLock(fn func(*tokenFile) error) error {
	lockPath := s.path + ".lock"
	if err := os.MkdirAll(dirOf(s.path), 0700); err != nil {
		return err
	}
	fl := flock.New(lockPath)
	if err := fl.Lock(); err != nil {
		return fmt.Errorf("acquiring token store lock: %w", err)
	}
	defer fl.Unlock()

	tf, err := s.Load()
	if err != nil {
		return err
	}
	if err := fn(tf); err != nil {
		return err
	}
	return s.Save(tf)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
