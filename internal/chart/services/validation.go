package services

import (
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
)

// ConfigurationValidator validates helm-values.yaml for non-interactive mode
type ConfigurationValidator struct{}

// NewConfigurationValidator creates a new configuration validator
func NewConfigurationValidator() *ConfigurationValidator {
	return &ConfigurationValidator{}
}

// ValidateConfiguration validates configuration for non-interactive mode. The
// CLI supports only the OSS (oss-tenant) deployment, which deploys from a public
// repository and requires no credentials, so there is nothing to validate.
func (v *ConfigurationValidator) ValidateConfiguration(_ *types.ChartConfiguration) error {
	return nil
}
