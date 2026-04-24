package root

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
)

func TestNewCmdRoot_RegistersExpectedCommands(t *testing.T) {
	cmd, _ := NewCmdRoot()
	want := map[string]bool{
		"auth":      true,
		"workspace": true,
		"repo":      true,
		"branch":    true,
		"pr":        true,
		"pipeline":  true,
		"issue":     true,
		"snippet":   true,
		"status":    true,
		"webhook":   true, // v0.4.0
		"runner":    true, // v0.4.0
		"project":   true, // v0.4.0
	}
	got := map[string]bool{}
	for _, sub := range cmd.Commands() {
		got[sub.Name()] = true
	}
	for name := range want {
		if !got[name] {
			t.Errorf("command %q not registered on root", name)
		}
	}
}

// stubAccounts returns a listAccountsFunc that yields the given state.
func stubAccounts(usernames []string, active string, err error) listAccountsFunc {
	return func() ([]string, string, error) {
		return usernames, active, err
	}
}

func TestAuthHint(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		usernames  []string
		active     string
		listErr    error
		wantSubstr string // "" means hint must be empty
	}{
		{
			name:       "404 with multiple accounts suggests switch",
			err:        &api.HTTPError{StatusCode: 404, Message: "Not Found"},
			usernames:  []string{"alice", "bob"},
			active:     "alice",
			wantSubstr: "bb auth switch bob",
		},
		{
			name:       "403 with multiple accounts suggests switch",
			err:        &api.HTTPError{StatusCode: 403, Message: "Forbidden"},
			usernames:  []string{"alice", "bob"},
			active:     "bob",
			wantSubstr: "bb auth switch alice",
		},
		{
			name:       "403 with single account suggests login-for-scope",
			err:        &api.HTTPError{StatusCode: 403, Message: "Forbidden"},
			usernames:  []string{"alice"},
			active:     "alice",
			wantSubstr: "bb auth login",
		},
		{
			name:       "404 with single account gives no hint",
			err:        &api.HTTPError{StatusCode: 404, Message: "Not Found"},
			usernames:  []string{"alice"},
			active:     "alice",
			wantSubstr: "",
		},
		{
			name:       "403 with zero accounts gives no hint",
			err:        &api.HTTPError{StatusCode: 403, Message: "Forbidden"},
			usernames:  nil,
			active:     "",
			wantSubstr: "",
		},
		{
			name:       "500 never produces a hint",
			err:        &api.HTTPError{StatusCode: 500, Message: "Server Error"},
			usernames:  []string{"alice", "bob"},
			active:     "alice",
			wantSubstr: "",
		},
		{
			name:       "non-HTTP error produces no hint",
			err:        errors.New("boom"),
			usernames:  []string{"alice", "bob"},
			active:     "alice",
			wantSubstr: "",
		},
		{
			name:       "wrapped HTTPError still unwraps",
			err:        fmt.Errorf("listing PRs: %w", &api.HTTPError{StatusCode: 404, Message: "Not Found"}),
			usernames:  []string{"alice", "bob"},
			active:     "alice",
			wantSubstr: "bb auth switch bob",
		},
		{
			name:       "ListAccounts error suppresses hint (never blocks the real error)",
			err:        &api.HTTPError{StatusCode: 404, Message: "Not Found"},
			usernames:  nil,
			active:     "",
			listErr:    errors.New("cannot read tokens.json"),
			wantSubstr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := authHint(tc.err, stubAccounts(tc.usernames, tc.active, tc.listErr))
			if tc.wantSubstr == "" {
				if got != "" {
					t.Errorf("authHint: got %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tc.wantSubstr) {
				t.Errorf("authHint: got %q, want substring %q", got, tc.wantSubstr)
			}
		})
	}
}
