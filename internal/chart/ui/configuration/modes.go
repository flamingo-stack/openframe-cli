package configuration

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
)

// showConfigurationModeSelection shows the initial configuration mode selection
func (w *ConfigurationWizard) showConfigurationModeSelection() (string, error) {
	fmt.Println()
	pterm.Info.Printf("How would you like to configure your chart installation?\n")
	fmt.Println()

	prompt := promptui.Select{
		Label: "Configuration Mode",
		Items: []string{
			"Default configuration",
			"Interactive configuration",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "→ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "{{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if idx == 0 {
		return "default", nil
	}
	return "interactive", nil
}

// configureWithDefaults creates a default configuration without user interaction.
// The default is to use the existing openframe-helm-values.yaml as-is.
func (w *ConfigurationWizard) configureWithDefaults() (*types.ChartConfiguration, error) {
	pterm.Info.Println("Using default configuration for OSS deployment")

	// Load base values from current directory or create default
	config, err := w.loadBaseValues()
	if err != nil {
		return nil, fmt.Errorf("failed to load base values: %w", err)
	}

	// Create temporary file with default configuration
	if err := w.createTemporaryValuesFile(config); err != nil {
		return nil, fmt.Errorf("failed to create temporary values file: %w", err)
	}

	return config, nil
}

// configureInteractive runs the interactive configuration wizard
func (w *ConfigurationWizard) configureInteractive() (*types.ChartConfiguration, error) {
	pterm.Info.Println("Configuring Helm values for OSS deployment")

	// Load base values from current directory or create default
	config, err := w.loadBaseValues()
	if err != nil {
		return nil, fmt.Errorf("failed to load base values: %w", err)
	}

	// Configure each section in the correct order
	if err := w.branchConfig.Configure(config); err != nil {
		return nil, fmt.Errorf("branch configuration failed: %w", err)
	}

	if err := w.dockerConfig.Configure(config); err != nil {
		return nil, fmt.Errorf("docker registry configuration failed: %w", err)
	}

	if err := w.ingressConfig.Configure(config); err != nil {
		return nil, fmt.Errorf("ingress configuration failed: %w", err)
	}

	// Create temporary file with final configuration
	if err := w.createTemporaryValuesFile(config); err != nil {
		return nil, fmt.Errorf("failed to create temporary values file: %w", err)
	}

	return config, nil
}
