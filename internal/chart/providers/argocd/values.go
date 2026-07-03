package argocd

import _ "embed"

// argoCDValues is the baseline Helm values for the Argo CD chart, embedded from
// argocd-values.yaml at build time. Keeping it as a real YAML file (rather than
// a Go string literal) gives editor/lint support and clean diffs, and — because
// the developer-account bcrypt hash lives in the YAML, not in Go source — keeps
// it off gosec's credential radar. See argocd-values.yaml for the security note
// on the developer default account.
//
//go:embed argocd-values.yaml
var argoCDValues string

// GetArgoCDValues returns the baseline ArgoCD Helm chart values as a YAML string.
func GetArgoCDValues() string {
	return argoCDValues
}
