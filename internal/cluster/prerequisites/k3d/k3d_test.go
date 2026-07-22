package k3d

import (
	"runtime"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
)

func TestNewK3dInstaller(t *testing.T) {
	installer := NewK3dInstaller()

	if installer == nil {
		t.Error("Expected k3d installer to be created")
	}
}

func TestK3dInstaller_GetInstallHelp(t *testing.T) {
	installer := NewK3dInstaller()
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
		if !containsSubstring(help, "curl") && !containsSubstring(help, "https://") {
			t.Errorf("Linux help should contain curl or https reference: %s", help)
		}
	case "windows":
		if !containsSubstring(help, "https://") && !containsSubstring(help, "chocolatey") {
			t.Errorf("Windows help should contain https or chocolatey reference: %s", help)
		}
	}
}

// TestK3dInstaller_Install only exercises the fail-fast error paths. It must
// NEVER call Install() where the real install could proceed: on CI runners
// this test used to run a real `brew install k3d` (darwin) or download the
// pinned binary into the runner's ~/.openframe/bin (linux).
func TestK3dInstaller_Install(t *testing.T) {
	if runtime.GOOS == "darwin" && commandExists("brew") {
		t.Skip("would run a real 'brew install k3d'")
	}
	if runtime.GOOS == "linux" {
		t.Skip("would run a real package-manager install or verified download")
	}
	if runtime.GOOS == "windows" {
		t.Skip("k3d installs via WSL on windows")
	}

	// Only the guaranteed-error path remains: darwin without Homebrew.
	err := NewK3dInstaller().Install()
	if err == nil {
		t.Fatal("expected an error when no install tooling is available")
	}
	if !containsSubstring(err.Error(), "Homebrew") {
		t.Errorf("expected a Homebrew hint, got: %v", err)
	}
}

func TestCommandExists(t *testing.T) {
	if !commandExists("echo") {
		t.Error("Expected 'echo' command to exist")
	}

	if commandExists("nonexistentcommand12345") {
		t.Error("Expected 'nonexistentcommand12345' to not exist")
	}
}

func TestVerifiedInstallHasPinnedAsset(t *testing.T) {
	// installVerified downloads the pinned k3d binary via
	// download.InstallPinnedTool; the actual download + checksum verification is
	// covered by the download package tests. Here we only assert that a verified
	// asset is pinned for the current platform, so the installer never fails with
	// a "no verified asset" error on a supported OS. Windows installs via WSL and
	// does not use the verified path.
	if runtime.GOOS == "windows" {
		t.Skip("Windows installs k3d via WSL, not the verified download path")
	}

	if _, ok := download.K3d.Asset(runtime.GOOS, runtime.GOARCH); !ok {
		t.Errorf("no pinned k3d asset for %s/%s", runtime.GOOS, runtime.GOARCH)
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
