package prerequisites

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/certificates"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/chart/prerequisites/memory"
)

func TestNewPrerequisiteChecker(t *testing.T) {
	checker := NewPrerequisiteChecker()

	if len(checker.requirements) != 3 {
		t.Errorf("Expected 3 requirements, got %d", len(checker.requirements))
	}

	// Git is intentionally absent — cloning uses go-git, not the git binary.
	expectedNames := []string{"Helm", "Memory", "Certificates"}
	for i, req := range checker.requirements {
		if req.Name != expectedNames[i] {
			t.Errorf("Expected requirement %d to be %s, got %s", i, expectedNames[i], req.Name)
		}
	}
	for _, req := range checker.requirements {
		if req.Name == "Git" {
			t.Error("Git must not be a prerequisite anymore (go-git replaced the binary)")
		}
	}
}

func TestInstallHelp(t *testing.T) {
	tests := []struct {
		name     string
		helpFunc func() string
	}{
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

	// Requirements order: Helm(0), Memory(1), Certificates(2).
	checker.requirements[0].IsInstalled = func() bool { return false } // Helm
	checker.requirements[1].IsInstalled = func() bool { return false } // Memory
	checker.requirements[2].IsInstalled = func() bool { return true }  // Certificates

	allPresent, missing := checker.CheckAll()

	if allPresent {
		t.Error("Expected allPresent to be false when tools are missing")
	}
	if len(missing) != 2 {
		t.Errorf("Expected 2 missing tools, got %d: %v", len(missing), missing)
	}

	expectedMissing := map[string]bool{"Helm": true, "Memory": true}
	for _, tool := range missing {
		if !expectedMissing[tool] {
			t.Errorf("Unexpected missing tool: %s", tool)
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
