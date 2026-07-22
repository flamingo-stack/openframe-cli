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

// TestDockerInstaller_Install only exercises the fail-fast error paths. It
// must NEVER call Install() where the real install could proceed: on CI
// runners with Homebrew this test used to run an actual
// `brew install --cask docker-desktop` (~100s, mutating the runner and
// failing on brew's own errors).
func TestDockerInstaller_Install(t *testing.T) {
	if runtime.GOOS == "darwin" && commandExists("brew") {
		t.Skip("would run a real 'brew install --cask docker-desktop'")
	}
	if runtime.GOOS == "linux" {
		t.Skip("would run a real package-manager install")
	}
	if runtime.GOOS == "windows" {
		t.Skip("would attempt a real WSL setup")
	}

	// Only the guaranteed-error path remains: darwin without Homebrew.
	err := NewDockerInstaller().Install()
	if err == nil {
		t.Fatal("expected an error when no install tooling is available")
	}
	if !containsSubstring(err.Error(), "Homebrew is required") {
		t.Errorf("expected a Homebrew hint, got: %v", err)
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
