package types

import "testing"

// TestGetRepositoryURL pins the platform repository URL. The CLI supports only
// the OSS (oss-tenant) deployment, so this must always be the public OSS repo
// with no embedded credentials.
func TestGetRepositoryURL(t *testing.T) {
	const want = "https://github.com/flamingo-stack/openframe-oss-tenant"
	if got := GetRepositoryURL(); got != want {
		t.Errorf("GetRepositoryURL() = %q, want %q", got, want)
	}
}
