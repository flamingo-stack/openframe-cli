package k8s

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// RestConfigForContext builds a *rest.Config for a specific kubeconfig context.
// An empty contextName uses the kubeconfig's current-context. This is how a
// context chosen via SelectContext is turned into a working client (for the
// Accessor health/resource checks and, ultimately, the install).
func RestConfigForContext(kubeconfigPath, contextName string) (*rest.Config, error) {
	// $KUBECONFIG may be a list of files (kubectl convention): merge those via
	// Precedence. A single path stays ExplicitPath, whose stricter
	// missing-file error callers rely on.
	rules := &clientcmd.ClientConfigLoadingRules{}
	if files := filepath.SplitList(kubeconfigPath); len(files) > 1 {
		rules.Precedence = files
	} else {
		rules.ExplicitPath = kubeconfigPath
	}
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
