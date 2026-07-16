package k8s

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// RestConfigForContext builds a *rest.Config for a specific kubeconfig context.
// An empty contextName uses the kubeconfig's current-context. This is how a
// context chosen via SelectContext is turned into a working client (for the
// Accessor health/resource checks and, ultimately, the install).
func RestConfigForContext(kubeconfigPath, contextName string) (*rest.Config, error) {
	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	overrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		overrides.CurrentContext = contextName
	}

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config for context %q: %w", contextName, err)
	}
	return cfg, nil
}
