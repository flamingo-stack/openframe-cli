package prerequisites

import (
	"testing"

	"github.com/shipyard-run/shipyard/internal/shared/testutil"
)

func TestNewInstaller(t *testing.T) {
	installer := NewInstaller()

	if installer == nil {
		t.Fatal("Expected installer to be created")
		return
	}

	if installer.checker == nil {
		t.Error("Expected installer to have a checker")
	}
}

func TestInstallTool(t *testing.T) {
	installer := NewInstaller()

	// Test that install tool delegates to appropriate installers
	validTools := []string{"docker", "k3d"}

	for _, tool := range validTools {
		err := installer.installTool(tool)
		// We expect errors in test environment, but they should be reasonable
		if err != nil {
			// Should be installation-related errors, not logic errors
			errorStr := err.Error()
			invalidErrors := []string{
				"unknown tool",
				"panic",
			}

			for _, invalidError := range invalidErrors {
				if testutil.ContainsSubstring(errorStr, invalidError) {
					t.Errorf("Tool %s returned unexpected error: %v", tool, err)
				}
			}
		}
	}

	// Test unknown tool
	err := installer.installTool("unknown-tool")
	if err == nil {
		t.Error("Expected error for unknown tool")
	}

	expectedError := "unknown tool: unknown-tool"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestContainsTool covers the case-insensitive membership check used to detect
// a freshly installed Docker that still needs the start/wait phase (B3).
func TestContainsTool(t *testing.T) {
	tools := []string{"Docker", "k3d"}
	if !containsTool(tools, "docker") || !containsTool(tools, "Docker") {
		t.Error("containsTool must match case-insensitively")
	}
	if containsTool(tools, "helm") || containsTool(nil, "docker") {
		t.Error("containsTool must not match absent tools")
	}
}
