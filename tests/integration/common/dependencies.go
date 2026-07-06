package common

import (
	"os/exec"
	"testing"
)

// Dependency represents a dependency required for integration tests
type Dependency struct {
	Name       string
	CheckCmd   []string
	InstallMsg string
}

// IsAvailable checks if a dependency is available
func (d *Dependency) IsAvailable() bool {
	if len(d.CheckCmd) == 0 {
		return false
	}
	cmd := exec.Command(d.CheckCmd[0], d.CheckCmd[1:]...) // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args
	return cmd.Run() == nil
}

// RequireOrSkip skips (does NOT pass) the test when the dependency is missing.
func (d *Dependency) RequireOrSkip(t *testing.T, skipMsg string) {
	if !d.IsAvailable() {
		// SKIPPED, not passed: the code path was never exercised. %s tells the
		// developer how to install it so the test can actually run.
		t.Logf("SKIPPING (not run): %s unavailable — %s", d.Name, d.InstallMsg)
		t.Skip(skipMsg)
	}
}

// Predefined dependencies
var (
	Docker = &Dependency{
		Name:       "Docker not running",
		CheckCmd:   []string{"docker", "info"},
		InstallMsg: "Start Docker daemon to run actual cluster tests.",
	}

	K3d = &Dependency{
		Name:       "k3d not available",
		CheckCmd:   []string{"k3d", "version"},
		InstallMsg: "Install k3d to run actual cluster tests.",
	}

	Kubectl = &Dependency{
		Name:       "kubectl not available",
		CheckCmd:   []string{"kubectl", "version", "--client"},
		InstallMsg: "Install kubectl to run actual cluster tests.",
	}

	Helm = &Dependency{
		Name:       "helm not available",
		CheckCmd:   []string{"helm", "version"},
		InstallMsg: "Install helm to run actual cluster tests.",
	}
)

// RequireClusterDependencies checks all cluster-related dependencies
func RequireClusterDependencies(t *testing.T) {
	Docker.RequireOrSkip(t, "Docker not running, skipping integration test")
	K3d.RequireOrSkip(t, "k3d not available, skipping integration test")
}

// RequireK8sDependencies checks all k8s tool dependencies
func RequireK8sDependencies(t *testing.T) {
	Kubectl.RequireOrSkip(t, "kubectl not available, skipping integration test")
	Helm.RequireOrSkip(t, "helm not available, skipping integration test")
}

// RequireAllDependencies checks all dependencies
func RequireAllDependencies(t *testing.T) {
	RequireClusterDependencies(t)
	RequireK8sDependencies(t)
}
