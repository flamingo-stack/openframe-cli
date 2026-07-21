// Package k8s provides read/inspect access to an EXISTING Kubernetes cluster:
// kubeconfig contexts, cluster health, and resource sufficiency.
//
// It is deliberately isolated from cluster *creation* (internal/cluster/...):
// the app-install flow uses this to target a cluster the user already has —
// one created by `openframe cluster`, by the user directly, or anywhere else.
package k8s

import (
	"fmt"
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

// ResolveContextForCluster returns the kube-context to use for a named cluster.
// It prefers a context whose name matches the cluster exactly, otherwise the
// k3d convention "k3d-<name>" — which is also the fallback when the kubeconfig
// cannot be read, preserving prior behavior. This stops the chart/helm layer
// from hardcoding the k3d naming and so breaking on renamed or non-k3d contexts.
// An empty cluster name yields "".
func ResolveContextForCluster(kubeconfigPath, clusterName string) string {
	if clusterName == "" {
		return ""
	}
	k3d := "k3d-" + clusterName
	contexts, _, err := LoadContexts(kubeconfigPath)
	if err != nil {
		return k3d
	}
	for _, c := range contexts {
		if c.Name == clusterName {
			return clusterName
		}
	}
	return k3d
}

// HasContext reports whether the kubeconfig at path contains a context with
// the given name. An unreadable kubeconfig counts as "no".
func HasContext(path, name string) bool {
	contexts, _, err := LoadContexts(path)
	if err != nil {
		return false
	}
	for _, c := range contexts {
		if c.Name == name {
			return true
		}
	}
	return false
}

// SwitchContext sets current-context in the kubeconfig at path. It only
// changes the pointer — never contexts, clusters, or users — so it cannot
// damage entries owned by other tools.
func SwitchContext(path, name string) error {
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}
	if _, ok := cfg.Contexts[name]; !ok {
		return fmt.Errorf("kubeconfig has no context '%s'", name)
	}
	cfg.CurrentContext = name
	if err := clientcmd.WriteToFile(*cfg, path); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	return nil
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
