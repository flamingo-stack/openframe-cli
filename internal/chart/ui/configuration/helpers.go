package configuration

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
)

// loadBaseValues loads base values from current directory or creates default
func (w *ConfigurationWizard) loadBaseValues() (*types.ChartConfiguration, error) {
	values, err := w.modifier.LoadOrCreateBaseValues()
	if err != nil {
		return nil, err
	}

	baseFilePath := config.DefaultHelmValuesFile

	return &types.ChartConfiguration{
		BaseHelmValuesPath: baseFilePath,
		TempHelmValuesPath: "", // Will be set when temporary file is created
		ExistingValues:     values,
		ModifiedSections:   make([]string, 0),
	}, nil
}

// createTemporaryValuesFile creates the temporary values file for installation
func (w *ConfigurationWizard) createTemporaryValuesFile(config *types.ChartConfiguration) error {
	// Apply configuration changes to values
	if err := w.modifier.ApplyConfiguration(config.ExistingValues, config); err != nil {
		return fmt.Errorf("failed to apply configuration changes: %w", err)
	}

	// Create temporary file in current directory
	tempFilePath, err := w.modifier.CreateTemporaryValuesFile(config.ExistingValues)
	if err != nil {
		return err
	}

	// Update config with temporary file path
	config.TempHelmValuesPath = tempFilePath
	return nil
}
