package app

import (
	"testing"
)

// These tests lock the install command's flag-validation contract. They are
// fully hermetic (no cluster, Docker, or network) and therefore run in every
// unit pass — unlike the live bootstrap e2e which is gated behind real infra.

// The CLI supports only the OSS (oss-tenant) deployment, so there is no
// --deployment-mode flag: --non-interactive alone simply reuses the existing
// openframe-helm-values.yaml.
func TestExtractInstallFlags_NonInteractiveIsValid(t *testing.T) {
	cmd := getInstallCmd()
	if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}

	flags, err := extractInstallFlags(cmd)
	if err != nil {
		t.Fatalf("non-interactive install should be valid: %v", err)
	}
	if !flags.NonInteractive {
		t.Fatalf("expected NonInteractive=true, got %+v", flags)
	}
}

func TestExtractInstallFlags_InteractiveDefaults(t *testing.T) {
	cmd := getInstallCmd()

	flags, err := extractInstallFlags(cmd)
	if err != nil {
		t.Fatalf("interactive install should be valid: %v", err)
	}
	if flags.NonInteractive {
		t.Fatalf("expected interactive defaults, got %+v", flags)
	}
}

// The removed --deployment-mode flag must no longer be registered.
func TestExtractInstallFlags_NoDeploymentModeFlag(t *testing.T) {
	cmd := getInstallCmd()
	if cmd.Flags().Lookup("deployment-mode") != nil {
		t.Fatal("--deployment-mode flag should have been removed")
	}
}
