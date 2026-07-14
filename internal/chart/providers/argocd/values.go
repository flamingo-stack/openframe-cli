package argocd

import (
	_ "embed"
	"fmt"
	"os"
	"sort"

	"sigs.k8s.io/yaml"
)

// argoCDValues is the baseline Helm values for the Argo CD chart, embedded from
// argocd-values.yaml at build time. Keeping it as a real YAML file (rather than
// a Go string literal) gives editor/lint support and clean diffs, and — because
// the developer-account bcrypt hash lives in the YAML, not in Go source — keeps
// it off gosec's credential radar. See argocd-values.yaml for the security note
// on the developer default account.
//
//go:embed argocd-values.yaml
var argoCDValues string

// UserArgoCDKey is the top-level key in the user's openframe-helm-values.yaml
// whose subtree overrides this baseline for the Argo CD install.
//
// It is a DEDICATED key, not the whole file, on purpose. The rest of that file
// uses the flattened app-of-apps schema (repository.branch, registry.docker.*)
// and is meant for the app-of-apps chart — merging it into the Argo CD chart
// would both mismatch schemas and leak the docker registry password into the
// argo-cd release. Scoping to `argocd:` keeps the two value streams (and the
// secret) separate.
const UserArgoCDKey = "argocd"

// GetArgoCDValues returns the baseline ArgoCD Helm chart values as a YAML string.
func GetArgoCDValues() string {
	return argoCDValues
}

// MergedArgoCDValues deep-merges the user's `argocd:` overrides over the
// embedded baseline and returns the YAML to feed helm plus the sorted top-level
// override keys (for a visible warning). userValues is the whole parsed user
// file; when it has no non-empty `argocd:` subtree the baseline is returned
// byte-for-byte unchanged with a nil key list, so the common path is untouched.
//
// Merge semantics match helm's: maps merge recursively, scalars and lists
// replace. Overriding is the user's explicit choice, so nothing in the baseline
// is protected from being replaced — the caller warns loudly instead.
func MergedArgoCDValues(userValues map[string]interface{}) (string, []string, error) {
	raw, present := userValues[UserArgoCDKey]
	// Absent, or bare `argocd:` (null), or `argocd: {}` — nothing to override.
	if !present || raw == nil {
		return argoCDValues, nil, nil
	}
	// Present but not a mapping (scalar/list/typo'd indentation) is a mistake,
	// not a no-op: fail loudly so the caller surfaces it, rather than silently
	// dropping the user's intended override (same silent-failure class as V3).
	sub, ok := raw.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("%q in the values file must be a mapping of ArgoCD chart values, got %T", UserArgoCDKey, raw)
	}
	if len(sub) == 0 {
		return argoCDValues, nil, nil
	}

	var base map[string]interface{}
	if err := yaml.Unmarshal([]byte(argoCDValues), &base); err != nil {
		return "", nil, fmt.Errorf("parsing embedded ArgoCD values: %w", err)
	}
	deepMerge(base, sub)

	out, err := yaml.Marshal(base)
	if err != nil {
		return "", nil, fmt.Errorf("marshaling merged ArgoCD values: %w", err)
	}

	keys := make([]string, 0, len(sub))
	for k := range sub {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return string(out), keys, nil
}

// ValidateUserValuesFile is the pre-flight check for the user's values file:
// a missing file is fine (baseline install), but a file that exists must be
// readable, parse as YAML, and its `argocd:` key — when present — must be a
// mapping. Callers run this BEFORE any cluster work; previously a malformed
// override surfaced only mid-install, after a cluster create and behind an
// ArgoCD pod-diagnostics dump it had nothing to do with (0.4.9 verification
// observation).
func ValidateUserValuesFile(path string) error {
	data, err := os.ReadFile(path) // #nosec G304 -- values path resolved from config/CLI, read as the invoking user
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading values file %s: %w", path, err)
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("values file %s is not valid YAML: %w", path, err)
	}
	if _, _, err := MergedArgoCDValues(m); err != nil {
		return fmt.Errorf("values file %s: %w", path, err)
	}
	return nil
}

// deepMerge overlays src onto dst in place: nested maps merge recursively;
// scalars, lists, and new keys replace (helm's value-merge rule).
func deepMerge(dst, src map[string]interface{}) {
	for k, sv := range src {
		if dv, ok := dst[k]; ok {
			if dm, ok1 := dv.(map[string]interface{}); ok1 {
				if sm, ok2 := sv.(map[string]interface{}); ok2 {
					deepMerge(dm, sm)
					continue
				}
			}
		}
		dst[k] = sv
	}
}
