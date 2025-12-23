package docker

import (
	"runtime"
	"testing"
)

func TestNewDockerInstaller(t *testing.T) {
	installer := NewDockerInstaller()
	
	if installer == nil {
		t.Error("Expected Docker installer to be created")
	}
}

func TestDockerInstaller_GetInstallHelp(t *testing.T) {
	installer := NewDockerInstaller()
	help := installer.GetInstallHelp()
	
	if help == "" {
		t.Error("Install help should not be empty")
	}
	
	switch runtime.GOOS {
	case "darwin":
		if !containsSubstring(help, "brew") && !containsSubstring(help, "https://") {
			t.Errorf("macOS help should contain brew or https reference: %s", help)
		}
	case "linux":
		if !containsSubstring(help, "package manager") && !containsSubstring(help, "https://") {
			t.Errorf("Linux help should contain package manager or https reference: %s", help)
		}
	case "windows":
		if !containsSubstring(help, "https://") {
			t.Errorf("Windows help should contain https reference: %s", help)
		}
	}
}

func TestDockerInstaller_Install(t *testing.T) {
	installer := NewDockerInstaller()

	// We can't actually test installation in CI, but we can test error handling
	err := installer.Install()

	// On unsupported platforms, should return specific error
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		expectedPrefix := "automatic Docker installation not supported on"
		if err == nil || !containsSubstring(err.Error(), expectedPrefix) {
			t.Errorf("Expected error containing '%s', got: %v", expectedPrefix, err)
		}
		return
	}

	// On macOS without brew, should suggest installing brew
	if runtime.GOOS == "darwin" && !commandExists("brew") {
		if err == nil {
			t.Error("Expected error when Homebrew is not installed")
		} else {
			expectedSubstring := "Homebrew is required"
			if !containsSubstring(err.Error(), expectedSubstring) {
				t.Errorf("Expected error containing '%s', got: %v", expectedSubstring, err)
			}
		}
		return
	}

	// On Linux without sudo or package managers, should fail
	if runtime.GOOS == "linux" && !commandExists("sudo") {
		if err != nil {
			// This is expected, installation needs sudo
			return
		}
	}

	// On Windows, may attempt WSL setup (will likely fail in test environment)
	// Just verify it doesn't panic and returns some result
	if runtime.GOOS == "windows" {
		// Windows installation will likely fail due to WSL not being set up in tests
		// We just verify the function runs without panicking
		_ = err
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && 
		   func() bool {
			   for i := 0; i <= len(str)-len(substr); i++ {
				   if str[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }()
}