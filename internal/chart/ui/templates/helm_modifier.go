package templates

import (
	"fmt"
	"os"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"gopkg.in/yaml.v3"
)

// HelmValuesModifier handles reading, modifying, and writing Helm values files
type HelmValuesModifier struct{}

// NewHelmValuesModifier creates a new Helm values modifier
func NewHelmValuesModifier() *HelmValuesModifier {
	return &HelmValuesModifier{}
}

// LoadExistingValues loads existing Helm values from file
func (h *HelmValuesModifier) LoadExistingValues(helmValuesPath string) (map[string]interface{}, error) {
	// Check if file exists
	if _, err := os.Stat(helmValuesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("helm values file not found at %s", helmValuesPath)
	}

	// Read file
	data, err := os.ReadFile(helmValuesPath) // #nosec G304 -- helm values path resolved from config/CLI, read as invoking user
	if err != nil {
		return nil, fmt.Errorf("failed to read helm values file: %w", err)
	}

	// Parse YAML
	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse helm values YAML: %w", err)
	}

	// Handle empty file case - yaml.Unmarshal returns nil for empty content
	if values == nil {
		values = make(map[string]interface{})
	}

	return values, nil
}

// LoadOrCreateBaseValues loads helm values from current directory or creates default if missing
func (h *HelmValuesModifier) LoadOrCreateBaseValues() (map[string]interface{}, error) {
	baseHelmValuesPath := config.DefaultHelmValuesFile

	// Try to load existing file from current directory
	if _, err := os.Stat(baseHelmValuesPath); err == nil {
		return h.LoadExistingValues(baseHelmValuesPath)
	}

	// File doesn't exist, create empty values (only configured sections will be added)
	emptyValues := make(map[string]interface{})

	return emptyValues, nil
}

// CreateTemporaryValuesFile creates a temporary helm values file in the OS
// temp directory (never the user's working directory, which it must not
// pollute). It uses a unique name via os.CreateTemp (O_EXCL, 0600) rather than
// a fixed filename: this avoids clobbering between concurrent runs and prevents
// a pre-created file / symlink from redirecting the write (the file can hold
// registry and repository secrets). The caller registers the returned absolute
// path for cleanup so it does not persist past the install; on Windows the helm
// manager converts the path for WSL before use.
func (h *HelmValuesModifier) CreateTemporaryValuesFile(values map[string]interface{}) (string, error) {
	f, err := os.CreateTemp("", "helm-values-tmp-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary values file: %w", err)
	}
	tempFile := f.Name()
	_ = f.Close()

	if err := h.WriteValues(values, tempFile); err != nil {
		_ = os.Remove(tempFile)
		return "", fmt.Errorf("failed to write temporary values file: %w", err)
	}

	return tempFile, nil
}

// ApplyConfiguration applies configuration changes to Helm values
func (h *HelmValuesModifier) ApplyConfiguration(values map[string]interface{}, config *types.ChartConfiguration) error {
	// Update branch if it was modified — the flattened schema uses a single
	// top-level repository.branch.
	if config.Branch != nil {
		h.setRepositoryBranch(values, *config.Branch)
	}

	// Update Docker registry if it was modified
	if config.DockerRegistry != nil {
		registry, ok := values["registry"].(map[string]interface{})
		if !ok {
			registry = make(map[string]interface{})
			values["registry"] = registry
		}

		docker, ok := registry["docker"].(map[string]interface{})
		if !ok {
			docker = make(map[string]interface{})
			registry["docker"] = docker
		}

		docker["username"] = config.DockerRegistry.Username
		docker["password"] = config.DockerRegistry.Password
		docker["email"] = config.DockerRegistry.Email
	}

	return nil
}

// SetRepositoryBranch pins the app-of-apps repository branch/ref. The flattened
// chart schema uses a single top-level repository.branch (openframe-oss-tenant
// flattened deployment.oss.repository.* to repository.*).
func (h *HelmValuesModifier) SetRepositoryBranch(values map[string]interface{}, branch string) {
	h.setRepositoryBranch(values, branch)
}

// setRepositoryBranch pins the app-of-apps source at the top-level
// repository.branch, creating the map as needed.
func (h *HelmValuesModifier) setRepositoryBranch(values map[string]interface{}, branch string) {
	repository, ok := values["repository"].(map[string]interface{})
	if !ok {
		repository = make(map[string]interface{})
		values["repository"] = repository
	}
	repository["branch"] = branch
}

// WriteValues writes updated values back to the Helm values file
func (h *HelmValuesModifier) WriteValues(values map[string]interface{}, helmValuesPath string) error {
	// Marshal back to YAML
	updatedData, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal updated helm values: %w", err)
	}

	// Write updated values back to file with owner-only permissions (0600):
	// the values may contain secrets (docker registry password), so the file
	// must not be world-readable (audit I2).
	if err := os.WriteFile(helmValuesPath, updatedData, 0o600); err != nil {
		return fmt.Errorf("failed to write updated helm values file: %w", err)
	}

	return nil
}

// GetCurrentOSSBranch extracts the current repository branch from the top-level
// repository.branch (the flattened chart schema).
func (h *HelmValuesModifier) GetCurrentOSSBranch(values map[string]interface{}) string {
	if repo, ok := values["repository"].(map[string]interface{}); ok {
		if branch, ok := repo["branch"].(string); ok && branch != "" {
			return branch
		}
	}
	return "main" // default fallback
}

// GetCurrentDockerSettings extracts current Docker settings from Helm values
func (h *HelmValuesModifier) GetCurrentDockerSettings(values map[string]interface{}) *types.DockerRegistryConfig {
	config := &types.DockerRegistryConfig{
		Username: "default",
		Password: "****",
		Email:    "default@example.com",
	}

	if registry, ok := values["registry"].(map[string]interface{}); ok {
		if docker, ok := registry["docker"].(map[string]interface{}); ok {
			if username, ok := docker["username"].(string); ok {
				config.Username = username
			}
			if password, ok := docker["password"].(string); ok {
				config.Password = password
			}
			if email, ok := docker["email"].(string); ok {
				config.Email = email
			}
		}
	}

	return config
}

// GetCurrentIngressSettings extracts current ingress settings from Helm values
func (h *HelmValuesModifier) GetCurrentIngressSettings(values map[string]interface{}) string {
	if deployment, ok := values["deployment"].(map[string]interface{}); ok {
		if ingress, ok := deployment["ingress"].(map[string]interface{}); ok {
			// Check if ngrok is enabled
			if ngrok, ok := ingress["ngrok"].(map[string]interface{}); ok {
				if enabled, ok := ngrok["enabled"].(bool); ok && enabled {
					return "ngrok"
				}
			}

			// Check if localhost is enabled
			if localhost, ok := ingress["localhost"].(map[string]interface{}); ok {
				if enabled, ok := localhost["enabled"].(bool); ok && enabled {
					return "localhost"
				}
			}
		}
	}

	return "localhost" // default fallback
}
