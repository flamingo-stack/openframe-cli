package types

import "testing"

// TestGetRepositoryURL_Mapping is a characterization test (testing plan §3): it
// pins the deployment-mode → repository-URL mapping so the audit-noted "silent
// redirect" behavior cannot change unnoticed. If a mapping must change, this
// test must be updated deliberately.
func TestGetRepositoryURL_Mapping(t *testing.T) {
	cases := map[DeploymentMode]string{
		DeploymentModeOSS:        "https://github.com/flamingo-stack/openframe-oss-tenant",
		DeploymentModeSaaS:       "https://github.com/flamingo-stack/openframe-saas-tenant",
		DeploymentModeSaaSShared: "https://github.com/flamingo-stack/openframe-saas-shared",
	}
	for mode, want := range cases {
		if got := GetRepositoryURL(mode); got != want {
			t.Errorf("GetRepositoryURL(%q) = %q, want %q", mode, got, want)
		}
	}

	// Unknown modes must fall back to the OSS repository (never a SaaS repo).
	if got := GetRepositoryURL(DeploymentMode("bogus")); got != "https://github.com/flamingo-stack/openframe-oss-tenant" {
		t.Errorf("unknown mode fallback = %q, want OSS repo", got)
	}
}
