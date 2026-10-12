// Package kubeconfigmerge holds the provider-agnostic logic for merging and
// removing a single cluster context from the user's default kubeconfig. It
// is shared by providers (EKS, GKE, ...) whose kubeconfig entries differ only
// in how the exec-plugin / cluster / auth-info objects are constructed.
package kubeconfigmerge

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Entry carries the pieces a provider needs to contribute to the merged
// kubeconfig for a single named context.
type Entry struct {
	Name     string
	Endpoint string
	Cluster  *clientcmdapi.Cluster
	AuthInfo *clientcmdapi.AuthInfo
	Context  *clientcmdapi.Context
}

// MergeIntoDefault writes the entry's context into the user's kubeconfig
// (honoring $KUBECONFIG) and switches the current context to it. It refuses
// to overwrite a same-named context that points at a DIFFERENT server.
func MergeIntoDefault(entry Entry) error {
	pathOpts := clientcmd.NewDefaultPathOptions()
	existing, err := pathOpts.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}
	if prior, ok := existing.Contexts[entry.Name]; ok {
		if cluster, ok := existing.Clusters[prior.Cluster]; ok && cluster.Server != entry.Endpoint {
			return fmt.Errorf("kubeconfig context '%s' already exists and points at %s — refusing to overwrite it; rename the existing context or pick another cluster name", entry.Name, cluster.Server)
		}
	}
	existing.Clusters[entry.Name] = entry.Cluster
	existing.AuthInfos[entry.Name] = entry.AuthInfo
	existing.Contexts[entry.Name] = entry.Context
	existing.CurrentContext = entry.Name
	if err := clientcmd.ModifyConfig(pathOpts, *existing, true); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	return nil
}

// RemoveFromDefault drops the named context after a destroy — but ONLY when
// the entry still points at the given endpoint. Best-effort: a missing entry
// is not an error.
func RemoveFromDefault(name, endpoint string) error {
	pathOpts := clientcmd.NewDefaultPathOptions()
	existing, err := pathOpts.GetStartingConfig()
	if err != nil {
		return err
	}
	if prior, ok := existing.Contexts[name]; ok {
		if cluster, ok := existing.Clusters[prior.Cluster]; ok && cluster.Server != endpoint {
			return nil // same name, different server — not ours anymore
		}
	}
	delete(existing.Clusters, name)
	delete(existing.AuthInfos, name)
	delete(existing.Contexts, name)
	if existing.CurrentContext == name {
		existing.CurrentContext = ""
	}
	return clientcmd.ModifyConfig(pathOpts, *existing, true)
}
