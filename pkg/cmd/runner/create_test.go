package runner

import (
	"reflect"
	"testing"
)

func TestParseLabelFlags(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"single repeated flags", []string{"self.hosted", "linux"}, []string{"self.hosted", "linux"}},
		{"comma-separated in one flag", []string{"self.hosted,linux,amd64"}, []string{"self.hosted", "linux", "amd64"}},
		{"mixed key=value + bare tokens", []string{"os=linux,arch=amd64", "self.hosted"}, []string{"os=linux", "arch=amd64", "self.hosted"}},
		{"dedupe preserves order", []string{"linux", "linux,arm64"}, []string{"linux", "arm64"}},
		{"strips whitespace + empties", []string{" linux , ", ",arm64,"}, []string{"linux", "arm64"}},
		{"nil input", nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseLabelFlags(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseLabelFlags(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
