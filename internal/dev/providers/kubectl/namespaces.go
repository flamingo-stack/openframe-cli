package kubectl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetNamespaces returns all namespaces in the current cluster
func (p *Provider) GetNamespaces(ctx context.Context) ([]string, error) {
	result, err := p.executor.Execute(ctx, "kubectl", "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	namespaces := strings.Fields(result.Stdout)
	return namespaces, nil
}

// ValidateNamespace checks if a namespace exists
func (p *Provider) ValidateNamespace(ctx context.Context, namespace string) error {
	_, err := p.executor.Execute(ctx, "kubectl", "get", "namespace", namespace)
	if err != nil {
		return fmt.Errorf("namespace '%s' not found", namespace)
	}
	return nil
}

// FindResourceNamespace searches for K8s resources matching the service name and returns their namespace
func (p *Provider) FindResourceNamespace(ctx context.Context, serviceName string) (string, error) {
	if serviceName == "" {
		return "default", nil
	}
	
	// Search for deployments with matching name
	if namespace := p.searchResourceByType(ctx, "deployment", serviceName); namespace != "" {
		return namespace, nil
	}
	
	// Search for statefulsets with matching name
	if namespace := p.searchResourceByType(ctx, "statefulset", serviceName); namespace != "" {
		return namespace, nil
	}
	
	// Search for pods with matching name prefix
	if namespace := p.searchResourceByType(ctx, "pod", serviceName); namespace != "" {
		return namespace, nil
	}
	
	// Default to "default" namespace if no resources found
	return "default", nil
}

// resourceList represents a generic Kubernetes resource list for JSON parsing
type resourceList struct {
	Items []struct {
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	} `json:"items"`
}

// searchResourceByType searches for a specific resource type with the given name/prefix
func (p *Provider) searchResourceByType(ctx context.Context, resourceType, serviceName string) string {
	// Search across all namespaces for resources matching the service name
	// Use -o json to avoid Windows WSL escaping issues with jsonpath
	result, err := p.executor.Execute(ctx, "kubectl", "get", resourceType, "--all-namespaces", "-o", "json")
	if err != nil {
		if p.verbose {
			fmt.Printf("DEBUG: kubectl search failed for %s: %v\n", resourceType, err)
		}
		return ""
	}

	if p.verbose {
		fmt.Printf("DEBUG: Searching %s for service '%s'\n", resourceType, serviceName)
	}

	var resources resourceList
	if err := json.Unmarshal([]byte(result.Stdout), &resources); err != nil {
		if p.verbose {
			fmt.Printf("DEBUG: Failed to parse JSON for %s: %v\n", resourceType, err)
		}
		return ""
	}

	for _, item := range resources.Items {
		namespace := item.Metadata.Namespace
		resourceName := item.Metadata.Name

		if p.verbose {
			fmt.Printf("DEBUG: Checking resource '%s' in namespace '%s'\n", resourceName, namespace)
		}

		// Check if resource name matches or starts with service name
		if resourceName == serviceName || strings.HasPrefix(resourceName, serviceName+"-") || strings.HasPrefix(resourceName, serviceName) {
			if p.verbose {
				fmt.Printf("DEBUG: Found match! Resource '%s' in namespace '%s'\n", resourceName, namespace)
			}
			return namespace
		}
	}

	if p.verbose {
		fmt.Printf("DEBUG: No matching %s found for service '%s'\n", resourceType, serviceName)
	}
	return ""
}