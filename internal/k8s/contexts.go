// Package k8s provides read/inspect access to an EXISTING Kubernetes cluster:
// kubeconfig contexts, cluster health, and resource sufficiency.
//
// It is deliberately isolated from cluster *creation* (internal/cluster/...):
// the app-install flow uses this to target a cluster the user already has —
// one created by `openframe cluster`, by the user directly, or anywhere else.
package k8s

import (
	"os"
	"path/filepath"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
)

// ContextInfo describes a single kubeconfig context.
type ContextInfo struct {
	Name    string // context name
	Cluster string // cluster the context points at
	Current bool   // whether this is the kubeconfig's current-context
}

// DefaultKubeconfigPath returns the kubeconfig path from $KUBECONFIG, or
// ~/.kube/config, falling back to the client-go recommended location.
func DefaultKubeconfigPath() string {
	if p := os.Getenv("KUBECONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return clientcmd.RecommendedHomeFile
	}
	return filepath.Join(home, ".kube", "config")
}

// LoadContexts reads the kubeconfig at path and returns its contexts (sorted by
// name) together with the current-context name. This is what the interactive
// context-selection menu is built on.
func LoadContexts(path string) (contexts []ContextInfo, current string, err error) {
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, "", err
	}

	current = cfg.CurrentContext
	contexts = make([]ContextInfo, 0, len(cfg.Contexts))
	for name, c := range cfg.Contexts {
		cluster := ""
		if c != nil {
			cluster = c.Cluster
		}
		contexts = append(contexts, ContextInfo{
			Name:    name,
			Cluster: cluster,
			Current: name == current,
		})
	}
	sort.Slice(contexts, func(i, j int) bool { return contexts[i].Name < contexts[j].Name })
	return contexts, current, nil
}
