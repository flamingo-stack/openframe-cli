package terraform

import (
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
	cfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil || cfg == nil {
		return false
	}
	_, ok := cfg.Contexts[name]
	return ok
}
