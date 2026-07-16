package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Remote state is opt-in: --backend-config s3://bucket/prefix (EKS) or
// gcs://bucket/prefix (GKE). The default stays local state in the workspace —
// remote state protects against losing the machine that created the cluster.

// BackendConfig is a parsed --backend-config value.
type BackendConfig struct {
	Scheme string // "s3" or "gcs"
	Bucket string
	Prefix string
}

// backendPartRE constrains bucket/prefix to characters that are safe to
// interpolate into generated HCL (defense in depth — values also come from a
// validated flag).
var backendPartRE = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

// ParseBackendURL parses scheme://bucket[/prefix].
func ParseBackendURL(raw string) (BackendConfig, error) {
	scheme, rest, found := strings.Cut(raw, "://")
	if !found || (scheme != "s3" && scheme != "gcs") {
		return BackendConfig{}, fmt.Errorf("invalid backend %q: expected s3://bucket/prefix or gcs://bucket/prefix", raw)
	}
	bucket, prefix, _ := strings.Cut(rest, "/")
	if bucket == "" || !backendPartRE.MatchString(bucket) || (prefix != "" && !backendPartRE.MatchString(prefix)) {
		return BackendConfig{}, fmt.Errorf("invalid backend %q: bucket/prefix may contain only letters, digits, '.', '_', '-' and '/'", raw)
	}
	return BackendConfig{Scheme: scheme, Bucket: bucket, Prefix: prefix}, nil
}

// WriteBackend writes backend.tf next to the generated main.tf. Callers pass
// the rendered HCL for their provider's backend type.
func (w *Workspace) WriteBackend(backendTF []byte) error {
	if err := os.MkdirAll(w.TerraformDir(), 0o750); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.TerraformDir(), "backend.tf"), backendTF, 0o600)
}
