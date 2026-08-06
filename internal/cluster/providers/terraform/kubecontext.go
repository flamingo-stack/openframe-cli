package terraform

import (
	"sort"

	"k8s.io/client-go/tools/clientcmd"
)

// KubeconfigHasContext reports whether the user's kubeconfig (default loading
// rules: $KUBECONFIG or ~/.kube/config) actually contains a context with the
// given name.
//
// The cluster list's CONTEXT column must be derived from this, never from the
// cluster name: the context is merged only after a successful create, so a
// workspace whose apply failed at plan time has NO kubeconfig entry — yet the
// list used to print one fabricated from the name. Any read error counts as
// "no context": inventing one is the failure mode this exists to prevent.
func KubeconfigHasContext(name string) bool {
	return KubeconfigContextMatching(func(have string) bool { return have == name }) != ""
}

// KubeconfigContextMatching returns the first kubeconfig context (in sorted
// order, so the result is deterministic) accepted by match, or "". It lets
// providers recognize the conventional context names beyond the plain cluster
// name — e.g. the ARN-named context `aws eks update-kubeconfig` writes, or
// gcloud's gke_<project>_<location>_<name> — mirroring the candidate shapes
// the discovery package matches.
func KubeconfigContextMatching(match func(name string) bool) string {
	cfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil || cfg == nil {
		return ""
	}
	names := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if match(name) {
			return name
		}
	}
	return ""
}
