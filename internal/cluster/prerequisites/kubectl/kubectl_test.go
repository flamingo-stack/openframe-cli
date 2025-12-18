package kubectl

import (
	"runtime"
	"testing"
)

func TestNewKubectlInstaller(t *testing.T) {
	installer := NewKubectlInstaller()
	
	if installer == nil {
		t.Error("Expected kubectl installer to be created")
	}
}

func TestKubectlInstaller_GetInstallHelp(t *testing.T) {
	installer := NewKubectlInstaller()
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

func TestKubectlInstaller_Install(t *testing.T) {
	installer := NewKubectlInstaller()

	// We can't actually test installation in CI, but we can test error handling
	err := installer.Install()

	// On unsupported platforms, should return specific error
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		expectedPrefix := "automatic kubectl installation not supported on"
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

	// On Linux and Windows, the installation will likely fail in test environments
	// We just verify the function runs without panicking
	_ = err
}

func TestCommandExists(t *testing.T) {
	if !commandExists("echo") {
		t.Error("Expected 'echo' command to exist")
	}
	
	if commandExists("nonexistentcommand12345") {
		t.Error("Expected 'nonexistentcommand12345' to not exist")
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