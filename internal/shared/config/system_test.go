package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSystemService(t *testing.T) {
	service := NewSystemService()

	if service == nil {
		t.Fatal("NewSystemService should not return nil")
		return
	}

	// Should have default log directory set
	expectedLogDir := filepath.Join(os.TempDir(), "openframe-deployment-logs")
	if service.logDir != expectedLogDir {
		t.Errorf("expected log directory %q, got %q", expectedLogDir, service.logDir)
	}
}

func TestSystemService_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		logDir  string
		wantErr bool
	}{
		{
			name:    "default temp directory",
			logDir:  "", // Use default
			wantErr: false,
		},
		{
			name:    "custom valid directory",
			logDir:  filepath.Join(os.TempDir(), "test-openframe-logs"),
			wantErr: false,
		},
		{
			name:    "nested directory creation",
			logDir:  filepath.Join(os.TempDir(), "test", "nested", "openframe-logs"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var service *SystemService
			if tt.logDir == "" {
				service = NewSystemService()
			} else {
				service = &SystemService{logDir: tt.logDir}
			}

			err := service.Initialize()
			if (err != nil) != tt.wantErr {
				t.Errorf("SystemService.Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify directory was created
				if _, err := os.Stat(service.logDir); os.IsNotExist(err) {
					t.Errorf("expected log directory %q to be created", service.logDir)
				}

				// Clean up test directory
				if tt.logDir != "" {
					os.RemoveAll(service.logDir)
				}
			}
		})
	}
}

func TestSystemService_InitializeErrorHandling(t *testing.T) {
	// Test with invalid directory path (should fail gracefully)
	// Use a path that is guaranteed to fail - a file as a directory component
	tmpFile := filepath.Join(os.TempDir(), "test-file-not-dir.txt")

	// Create a file
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Try to create a directory inside the file (should fail)
	invalidPath := filepath.Join(tmpFile, "cannot", "create", "here")

	service := &SystemService{logDir: invalidPath}
	err = service.Initialize()

	if err == nil {
		t.Error("expected error when creating directory in invalid path")
	}

	// Error should contain meaningful message
	if err != nil && err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestSystemService_MultipleInitialize(t *testing.T) {
	// Test that multiple Initialize calls are safe
	logDir := filepath.Join(os.TempDir(), "test-multiple-init")
	service := &SystemService{logDir: logDir}

	// First initialize
	err1 := service.Initialize()
	if err1 != nil {
		t.Errorf("first Initialize() failed: %v", err1)
	}

	// Second initialize should also succeed (directory already exists)
	err2 := service.Initialize()
	if err2 != nil {
		t.Errorf("second Initialize() failed: %v", err2)
	}

	// Verify directory exists
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("expected log directory %q to exist", logDir)
	}

	// Clean up
	os.RemoveAll(logDir)
}
