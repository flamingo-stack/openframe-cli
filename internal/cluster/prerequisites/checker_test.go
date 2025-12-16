package prerequisites

import (
	"runtime"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/kubectl"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/wsl"
)

func TestNewPrerequisiteChecker(t *testing.T) {
	checker := NewPrerequisiteChecker()

	// On native Windows, WSL2 is added as the first requirement
	wslInstaller := wsl.NewWSLInstaller()
	var expectedNames []string
	if wslInstaller.IsApplicable() {
		expectedNames = []string{"WSL2", "Docker", "kubectl", "k3d", "helm"}
	} else {
		expectedNames = []string{"Docker", "kubectl", "k3d", "helm"}
	}

	if len(checker.requirements) != len(expectedNames) {
		t.Errorf("Expected %d requirements, got %d", len(expectedNames), len(checker.requirements))
	}

	for i, req := range checker.requirements {
		if i < len(expectedNames) && req.Name != expectedNames[i] {
			t.Errorf("Expected requirement %d to be %s, got %s", i, expectedNames[i], req.Name)
		}
	}
}

func TestCommandExists(t *testing.T) {
	// Test using docker package since it has commandExists function
	dockerInstaller := docker.NewDockerInstaller()

	// We can't directly test commandExists since it's not exported,
	// but we can test IsInstalled which uses it internally
	_ = dockerInstaller.IsInstalled()
}

func TestInstallHelp(t *testing.T) {
	tests := []struct {
		name     string
		helpFunc func() string
	}{
		{"docker", docker.NewDockerInstaller().GetInstallHelp},
		{"kubectl", kubectl.NewKubectlInstaller().GetInstallHelp},
		{"k3d", k3d.NewK3dInstaller().GetInstallHelp},
		{"helm", helm.NewHelmInstaller().GetInstallHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.helpFunc()
			if help == "" {
				t.Errorf("Install help for %s should not be empty", tt.name)
			}

			switch runtime.GOOS {
			case "darwin":
				if !containsAny(help, []string{"brew", "https://"}) {
					t.Errorf("macOS help should contain brew or https reference: %s", help)
				}
			case "linux":
				if !containsAny(help, []string{"package manager", "https://", "curl"}) {
					t.Errorf("Linux help should contain package manager, https, or curl reference: %s", help)
				}
			case "windows":
				if !containsAny(help, []string{"https://", "chocolatey", "choco"}) {
					t.Errorf("Windows help should contain https, chocolatey, or choco reference: %s", help)
				}
			}
		})
	}
}

func containsAny(str string, substrings []string) bool {
	for _, sub := range substrings {
		if len(str) >= len(sub) {
			for i := 0; i <= len(str)-len(sub); i++ {
				if str[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

func TestCheckAllWithMissingTools(t *testing.T) {
	checker := NewPrerequisiteChecker()

	// Find indices of requirements by name to handle platform differences
	// On Windows, WSL2 is prepended as the first requirement
	findIndex := func(name string) int {
		for i, req := range checker.requirements {
			if req.Name == name {
				return i
			}
		}
		return -1
	}

	dockerIdx := findIndex("Docker")
	kubectlIdx := findIndex("kubectl")
	k3dIdx := findIndex("k3d")
	helmIdx := findIndex("helm")

	// Set all as installed first
	for i := range checker.requirements {
		checker.requirements[i].IsInstalled = func() bool { return true }
	}

	// Mock specific requirements: Docker (missing), kubectl (installed), k3d (missing), helm (installed)
	if dockerIdx >= 0 {
		checker.requirements[dockerIdx].IsInstalled = func() bool { return false }
	}
	if kubectlIdx >= 0 {
		checker.requirements[kubectlIdx].IsInstalled = func() bool { return true }
	}
	if k3dIdx >= 0 {
		checker.requirements[k3dIdx].IsInstalled = func() bool { return false }
	}
	if helmIdx >= 0 {
		checker.requirements[helmIdx].IsInstalled = func() bool { return true }
	}

	allPresent, missing := checker.CheckAll()

	if allPresent {
		t.Error("Expected allPresent to be false when tools are missing")
	}

	if len(missing) != 2 {
		t.Errorf("Expected 2 missing tools, got %d", len(missing))
	}

	expectedMissing := []string{"Docker", "k3d"}
	for i, tool := range missing {
		if tool != expectedMissing[i] {
			t.Errorf("Expected missing tool %d to be %s, got %s", i, expectedMissing[i], tool)
		}
	}
}

func TestCheckAllWithAllTools(t *testing.T) {
	checker := NewPrerequisiteChecker()

	for i := range checker.requirements {
		checker.requirements[i].IsInstalled = func() bool { return true }
	}

	allPresent, missing := checker.CheckAll()

	if !allPresent {
		t.Error("Expected allPresent to be true when all tools are present")
	}

	if len(missing) != 0 {
		t.Errorf("Expected no missing tools, got %d: %v", len(missing), missing)
	}
}

func TestGetInstallInstructions(t *testing.T) {
	checker := NewPrerequisiteChecker()
	missing := []string{"Docker", "k3d"}

	instructions := checker.GetInstallInstructions(missing)

	if len(instructions) != 2 {
		t.Errorf("Expected 2 instructions, got %d", len(instructions))
	}

	for _, instruction := range instructions {
		if instruction == "" {
			t.Error("Instruction should not be empty")
		}
	}
}
