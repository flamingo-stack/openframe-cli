package prerequisites

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/certificates"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/git"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/memory"
)

func TestNewPrerequisiteChecker(t *testing.T) {
	checker := NewPrerequisiteChecker()

	if len(checker.requirements) != 4 {
		t.Errorf("Expected 4 requirements, got %d", len(checker.requirements))
	}

	expectedNames := []string{"Git", "Helm", "Memory", "Certificates"}
	for i, req := range checker.requirements {
		if req.Name != expectedNames[i] {
			t.Errorf("Expected requirement %d to be %s, got %s", i, expectedNames[i], req.Name)
		}
	}
}

func TestInstallHelp(t *testing.T) {
	tests := []struct {
		name     string
		helpFunc func() string
	}{
		{"git", git.NewGitChecker().GetInstallInstructions},
		{"helm", helm.NewHelmInstaller().GetInstallHelp},
		{"memory", memory.NewMemoryChecker().GetInstallHelp},
		{"certificates", certificates.NewCertificateInstaller().GetInstallHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.helpFunc()
			if help == "" {
				t.Errorf("Install help for %s should not be empty", tt.name)
			}
		})
	}
}

func TestCheckAllWithMissingTools(t *testing.T) {
	checker := NewPrerequisiteChecker()

	// Mock some requirements as missing
	checker.requirements[0].IsInstalled = func() bool { return false } // Git
	checker.requirements[1].IsInstalled = func() bool { return true }  // Helm
	checker.requirements[2].IsInstalled = func() bool { return false } // Memory
	checker.requirements[3].IsInstalled = func() bool { return true }  // Certificates

	allPresent, missing := checker.CheckAll()

	if allPresent {
		t.Error("Expected allPresent to be false when tools are missing")
	}

	if len(missing) != 2 {
		t.Errorf("Expected 2 missing tools, got %d: %v", len(missing), missing)
	}

	// Check that the missing tools are Git and Memory
	expectedMissing := map[string]bool{
		"Git":    true,
		"Memory": true,
	}

	for _, tool := range missing {
		if !expectedMissing[tool] {
			t.Errorf("Unexpected missing tool: %s", tool)
		}
	}

	// Verify Git and Memory are in the list
	hasGit := false
	hasMemory := false
	for _, tool := range missing {
		if tool == "Git" {
			hasGit = true
		}
		if tool == "Memory" {
			hasMemory = true
		}
	}
	if !hasGit || !hasMemory {
		t.Errorf("Expected Git and Memory to be missing, got: %v", missing)
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
	missing := []string{"Git", "Helm"}

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
