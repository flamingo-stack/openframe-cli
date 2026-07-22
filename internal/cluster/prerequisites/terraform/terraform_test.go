package terraform

import "testing"

// TestVersionSatisfies locks the terraform version gate: the generated root
// modules require >= 1.15, so an older system terraform must count as NOT
// installed (found by a real tfvalidate run against terraform 1.13.3, which
// passed the old binary-presence check and then failed on required_version).
func TestVersionSatisfies(t *testing.T) {
	cases := []struct {
		version string
		want    bool
	}{
		{"1.15.0", true},
		{"1.15.8", true},
		{"1.16.0", true},
		{"2.0.0", true},
		{"v1.15.8", true},
		{"1.13.3", false},
		{"1.14.9", false},
		{"0.15.0", false},
		{"", false},
		{"nonsense", false},
		{"1", false},
	}
	for _, tc := range cases {
		if got := versionSatisfies(tc.version, minMajor, minMinor); got != tc.want {
			t.Errorf("versionSatisfies(%q) = %v, want %v", tc.version, got, tc.want)
		}
	}
}
