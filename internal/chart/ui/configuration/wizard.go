package configuration

import (
	"github.com/flamingo-stack/openframe-cli/internal/chart/ui/templates"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
)

// ConfigurationWizard handles the chart configuration workflow
type ConfigurationWizard struct {
	modifier      *templates.HelmValuesModifier
	branchConfig  *BranchConfigurator
	dockerConfig  *DockerConfigurator
	ingressConfig *IngressConfigurator
}

// NewConfigurationWizard creates a new configuration wizard
func NewConfigurationWizard() *ConfigurationWizard {
	modifier := templates.NewHelmValuesModifier()
	return &ConfigurationWizard{
		modifier:      modifier,
		branchConfig:  NewBranchConfigurator(modifier),
		dockerConfig:  NewDockerConfigurator(modifier),
		ingressConfig: NewIngressConfigurator(modifier),
	}
}

// ConfigureHelmValues reads existing Helm values and prompts user for configuration changes
func (w *ConfigurationWizard) ConfigureHelmValues() (*types.ChartConfiguration, error) {
	// Show configuration mode selection (deployment is always OSS)
	modeChoice, err := w.showConfigurationModeSelection()
	if err != nil {
		return nil, err
	}

	if modeChoice == "default" {
		return w.configureWithDefaults()
	}

	return w.configureInteractive()
}
