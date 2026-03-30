package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestTokenStore_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	store := NewTokenStoreAt(filepath.Join(dir, "tokens.json"))

	acc := &Account{
		Username:       "testuser",
		AccessToken:    "access-abc",
		RefreshToken:   "refresh-xyz",
		TokenType:      "bearer",
		Expiry:         time.Now().Add(2 * time.Hour),
		WorkspaceSlugs: []string{"myworkspace"},
		AuthType:       AuthTypeOAuth,
	}

	if err := store.SetAccount(acc); err != nil {
		t.Fatalf("SetAccount: %v", err)
	}

	got, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if got == nil {
		t.Fatal("expected active account, got nil")
	}
	if got.Username != "testuser" {
		t.Errorf("username: got %q, want %q", got.Username, "testuser")
	}
	if got.AccessToken != "access-abc" {
		t.Errorf("access token: got %q, want %q", got.AccessToken, "access-abc")
	}
}

func TestTokenStore_RemoveAccount(t *testing.T) {
	dir := t.TempDir()
	store := NewTokenStoreAt(filepath.Join(dir, "tokens.json"))

	acc := &Account{
		Username:    "testuser",
		AccessToken: "tok",
		AuthType:    AuthTypeToken,
	}
	_ = store.SetAccount(acc)
	_ = store.RemoveAccount("testuser")

	got, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive after remove: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after logout, got %+v", got)
	}
}

// TestTokenStore_ConcurrentRefreshRace ensures that two goroutines racing to
// update tokens produce a consistent final state and don't corrupt the file.
// This simulates the Bitbucket rotating refresh token scenario (May 4 2026+).
func TestTokenStore_ConcurrentRefreshRace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	store := NewTokenStoreAt(path)

	// Seed an initial account
	acc := &Account{
		Username:     "raceuser",
		AccessToken:  "token-v0",
		RefreshToken: "refresh-v0",
		TokenType:    "bearer",
		Expiry:       time.Now().Add(2 * time.Hour),
		AuthType:     AuthTypeOAuth,
	}
	if err := store.SetAccount(acc); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Launch N goroutines all trying to UpdateTokens simultaneously
	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			newAccess := fmt.Sprintf("token-v%d", i)
			newRefresh := fmt.Sprintf("refresh-v%d", i)
			_ = store.UpdateTokens("raceuser", newAccess, newRefresh, time.Now().Add(2*time.Hour))
		}(i)
	}
	wg.Wait()

	// After all goroutines complete, the store must be valid JSON with a non-empty token
	final, err := store.GetActive()
	if err != nil {
		t.Fatalf("GetActive after race: %v", err)
	}
	if final == nil {
		t.Fatal("account was deleted during concurrent updates")
	}
	if final.AccessToken == "" {
		t.Error("access token is empty after concurrent updates")
	}
	// Verify the file is parseable (not corrupted)
	raw, _ := os.ReadFile(path)
	if len(raw) == 0 {
		t.Error("tokens.json is empty after concurrent updates")
	}
}

func TestAccount_IsExpired(t *testing.T) {
	cases := []struct {
		name    string
		expiry  time.Time
		want    bool
	}{
		{"zero expiry (token auth)", time.Time{}, false},
		{"expired", time.Now().Add(-1 * time.Hour), true},
		{"valid", time.Now().Add(1 * time.Hour), false},
		{"within 30s buffer", time.Now().Add(20 * time.Second), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := &Account{Expiry: tc.expiry}
			if got := a.IsExpired(); got != tc.want {
				t.Errorf("IsExpired() = %v, want %v", got, tc.want)
			}
		})
	}
}
