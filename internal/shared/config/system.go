package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// SystemService provides system-level configuration and initialization
// This handles CLI setup, logging directories, and system-wide settings
type SystemService struct {
	logDir string
}

// NewSystemService creates a new system service with default configuration
func NewSystemService() *SystemService {
	return &SystemService{
		logDir: filepath.Join(os.TempDir(), "openframe-deployment-logs"),
	}
}

// Initialize performs system initialization tasks
func (s *SystemService) Initialize() error {
	if err := s.setupLogDirectory(); err != nil {
		return fmt.Errorf("system initialization failed: %w", err)
	}
	return nil
}

// setupLogDirectory creates the logging directory structure
func (s *SystemService) setupLogDirectory() error {
	if err := os.MkdirAll(s.logDir, 0750); err != nil {
		return fmt.Errorf("failed to setup log directory %s: %w", s.logDir, err)
	}
	return nil
}
