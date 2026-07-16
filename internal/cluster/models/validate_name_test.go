package models

import "testing"

// TestValidateClusterName_RejectsShellMetacharacters is a security regression
// guard: cluster names flow into shell-outs (WSL `bash -c`, docker filters), so
// the DNS-1123 validation must reject anything that could break out of a shell
// token. If this ever loosens, command injection via a cluster name reopens.
func TestValidateClusterName_RejectsShellMetacharacters(t *testing.T) {
	payloads := []string{
		"evil; rm -rf /",
		"a$(whoami)",
		"a`whoami`",
		"a|b",
		"a&b",
		"a b",
		"a/b",
		"a\nb",
		"a'b",
		"a\"b",
		"a&&b",
		"../../etc",
		"$IFS",
	}
	for _, p := range payloads {
		if err := ValidateClusterName(p); err == nil {
			t.Errorf("ValidateClusterName(%q) = nil, expected rejection", p)
		}
	}
}

func TestValidateClusterName_AcceptsValid(t *testing.T) {
	for _, name := range []string{"openframe-dev", "test1", "a", "my-cluster-2"} {
		if err := ValidateClusterName(name); err != nil {
			t.Errorf("ValidateClusterName(%q) = %v, expected nil", name, err)
		}
	}
}
