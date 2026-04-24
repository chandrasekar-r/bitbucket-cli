package webhook

import (
	"reflect"
	"testing"
)

func TestParseEventFlags(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"single flag, one event", []string{"repo:push"}, []string{"repo:push"}},
		{"single flag, comma-separated", []string{"repo:push,pullrequest:created"}, []string{"repo:push", "pullrequest:created"}},
		{"repeated flags", []string{"repo:push", "pullrequest:created"}, []string{"repo:push", "pullrequest:created"}},
		{"mix of repeat + comma", []string{"repo:push,repo:fork", "pullrequest:created"}, []string{"repo:push", "repo:fork", "pullrequest:created"}},
		{"dedupe preserves order", []string{"repo:push", "repo:push,pullrequest:created"}, []string{"repo:push", "pullrequest:created"}},
		{"strips whitespace", []string{" repo:push , pullrequest:created "}, []string{"repo:push", "pullrequest:created"}},
		{"ignores empty", []string{"", "repo:push,,"}, []string{"repo:push"}},
		{"nil input", nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseEventFlags(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseEventFlags(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
