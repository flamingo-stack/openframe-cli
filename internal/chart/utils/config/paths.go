package config

import (
	"os"
	"path/filepath"
)

// PathResolver handles path resolution for chart-related files and directories
type PathResolver struct{}

// NewPathResolver creates a new path resolver
func NewPathResolver() *PathResolver {
	return &PathResolver{}
}

// GetCertificateDirectory returns the directory where certificates are stored
func (p *PathResolver) GetCertificateDirectory() string {
	// Certificates are stored in ~/.config/openframe/certs as per the certificate installer
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to working directory if home directory can't be determined
		if wd, err := os.Getwd(); err == nil {
			return filepath.Join(wd, "internal", "chart", "prerequisites", "certs")
		}
		return ""
	}

	certDir := filepath.Join(homeDir, ".config", "openframe", "certs")

	// Check if the directory exists
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		_ = os.MkdirAll(certDir, 0750)
	}

	return certDir
}

// GetHelmValuesFile returns the path to the helm values file
func (p *PathResolver) GetHelmValuesFile() string {
	// Return relative path to helm-values.yaml in CLI directory
	// This file should be dynamically read at runtime
	return "./helm-values.yaml"
}

// GetCertificateFiles returns the paths to the certificate files
func (p *PathResolver) GetCertificateFiles() (certFile, keyFile string) {
	certDir := p.GetCertificateDirectory()
	return filepath.Join(certDir, "localhost.pem"), filepath.Join(certDir, "localhost-key.pem")
}
