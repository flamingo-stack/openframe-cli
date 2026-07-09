package app

import (
	"strings"
	"testing"
)

// These tests lock the install command's flag-validation contract. They are
// fully hermetic (no cluster, Docker, or network) and therefore run in every
// unit pass — unlike the live bootstrap e2e which is gated behind real infra.

func TestExtractInstallFlags_NonInteractiveRequiresDeploymentMode(t *testing.T) {
	cmd := getInstallCmd()
	if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}
	// deployment-mode deliberately left empty.

	_, err := extractInstallFlags(cmd)
	if err == nil || !strings.Contains(err.Error(), "--deployment-mode is required") {
		t.Fatalf("expected a 'deployment-mode required' error, got %v", err)
	}
}

func TestExtractInstallFlags_RejectsInvalidDeploymentMode(t *testing.T) {
	cmd := getInstallCmd()
	if err := cmd.Flags().Set("deployment-mode", "not-a-real-mode"); err != nil {
		t.Fatal(err)
	}

	if _, err := extractInstallFlags(cmd); err == nil {
		t.Fatal("expected an invalid-deployment-mode error, got nil")
	}
}

func TestExtractInstallFlags_NonInteractiveOSSTenantIsValid(t *testing.T) {
	cmd := getInstallCmd()
	if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("deployment-mode", "oss-tenant"); err != nil {
		t.Fatal(err)
	}

	flags, err := extractInstallFlags(cmd)
	if err != nil {
		t.Fatalf("oss-tenant non-interactive should be valid: %v", err)
	}
	if !flags.NonInteractive || flags.DeploymentMode != "oss-tenant" {
		t.Fatalf("flags not parsed as expected: %+v", flags)
	}
}

func TestExtractInstallFlags_InteractiveAllowsEmptyDeploymentMode(t *testing.T) {
	// Without --non-interactive, deployment mode is optional (chosen via wizard).
	cmd := getInstallCmd()

	flags, err := extractInstallFlags(cmd)
	if err != nil {
		t.Fatalf("interactive install without a mode should be valid: %v", err)
	}
	if flags.NonInteractive || flags.DeploymentMode != "" {
		t.Fatalf("expected interactive defaults, got %+v", flags)
	}
}
