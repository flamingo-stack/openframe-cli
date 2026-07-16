package config

import (
	"os"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	sharedConfig "github.com/flamingo-stack/openframe-cli/internal/shared/config"
)

// Service provides centralized configuration management for chart operations
type Service struct {
	systemService *sharedConfig.SystemService
	pathResolver  *PathResolver
}

// NewService creates a new configuration service
func NewService() *Service {
	return &Service{
		systemService: sharedConfig.NewSystemService(),
		pathResolver:  NewPathResolver(),
	}
}

// GetCertificateDirectory returns the certificate directory path
func (s *Service) GetCertificateDirectory() string {
	return s.pathResolver.GetCertificateDirectory()
}

// GetPathResolver returns the path resolver instance
func (s *Service) GetPathResolver() *PathResolver {
	return s.pathResolver
}

// Initialize performs any necessary configuration initialization
func (s *Service) Initialize() error {
	// Initialize shared system service
	if err := s.systemService.Initialize(); err != nil {
		return err
	}

	// Ensure certificate directory exists
	certDir := s.GetCertificateDirectory()
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		if err := os.MkdirAll(certDir, 0750); err != nil {
			return err
		}
	}

	return nil
}

// BuildInstallConfig creates a complete installation configuration
func (s *Service) BuildInstallConfig(
	force, dryRun, verbose bool,
	clusterName string,
	appOfAppsConfig *models.AppOfAppsConfig,
) ChartInstallConfig {
	// Set default certificate directory if not provided
	if appOfAppsConfig != nil && appOfAppsConfig.CertDir == "" {
		appOfAppsConfig.CertDir = s.GetCertificateDirectory()
	}

	return ChartInstallConfig{
		ClusterName: clusterName,
		Force:       force,
		DryRun:      dryRun,
		Verbose:     verbose,
		Silent:      false,
		AppOfApps:   appOfAppsConfig,
	}
}
