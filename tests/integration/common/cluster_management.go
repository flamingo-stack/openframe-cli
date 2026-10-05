package common

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// testResourcePatterns is the shared list of substrings used to identify
// Docker resources (networks/containers) created by integration tests, so
// cleanup logic does not drift between the different cleanup helpers.
var testResourcePatterns = []string{
	"test",
	"collision",
	"interrupt",
	"stress",
	"multi",
	"integration",
}

// matchesTestResourcePattern reports whether name contains one of the known
// test-name substrings.
func matchesTestResourcePattern(name string) bool {
	for _, pattern := range testResourcePatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}
	return false
}

// removeDockerResourcesByFilter lists docker resources of the given kind
// (e.g. "network" or "container") using the provided list/format args and a
// docker filter, then removes each listed resource whose name matches the
// known test-name substrings (unless requireMatch is false, in which case
// all listed resources are removed).
func removeDockerResourcesByFilter(listArgs []string, removeArgs func(name string) []string, requireMatch bool) {
	cmd := exec.Command("docker", listArgs...) // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args
	output, err := cmd.Output()
	if err != nil {
		return
	}
	names := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, name := range names {
		if name == "" {
			continue
		}
		if requireMatch && !matchesTestResourcePattern(name) {
			continue
		}
		args := removeArgs(name)
		_ = exec.Command("docker", args...).Run() // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args
	}
}

// GenerateTestClusterName creates a unique cluster name for testing
func GenerateTestClusterName() string {
	return fmt.Sprintf("integration-test-%d", time.Now().Unix())
}

// CreateTestCluster creates a k3d cluster for testing with shorter timeout
func CreateTestCluster(name string) error {
	cmd := exec.Command("k3d", "cluster", "create", name, "--agents", "1", "--timeout", "60s") // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args
	return cmd.Run()
}

// DeleteTestCluster removes a test cluster
func DeleteTestCluster(name string) error {
	cmd := exec.Command("k3d", "cluster", "delete", name) // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args
	return cmd.Run()
}

// ClusterExists checks if a cluster exists (with caching for better performance)
func ClusterExists(name string) (bool, error) {
	// Use k3d directly to check if cluster exists - more reliable and faster than CLI
	cmd := exec.Command("k3d", "cluster", "list", name, "--no-headers") // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args
	cmd.Env = append(cmd.Env, "DOCKER_CLI_EXPERIMENTAL=enabled")        // Speed up docker operations
	err := cmd.Run()
	// If k3d command succeeds, cluster exists
	return err == nil, nil
}

// StopTestCluster stops a test cluster
func StopTestCluster(name string) error {
	cmd := exec.Command("k3d", "cluster", "stop", name) // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args
	return cmd.Run()
}

// CleanupTestCluster ensures a test cluster is cleaned up (optimized)
func CleanupTestCluster(name string) {
	// Use k3d directly for faster cleanup - skip CLI overhead
	_ = DeleteTestCluster(name)

	// Wait briefly for Docker resources to be released
	time.Sleep(200 * time.Millisecond)

	// Ensure any remaining Docker resources for this cluster are cleaned up
	cleanupClusterSpecificResources(name)
}

// CleanupAllTestClusters removes all test clusters to prevent resource conflicts
func CleanupAllTestClusters() {
	// Get list of clusters using k3d directly
	cmd := exec.Command("k3d", "cluster", "list", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	// Parse cluster names and delete test clusters
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	deletedAny := false
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		clusterName := fields[0] // First field is cluster name
		if clusterName != "" && (strings.Contains(clusterName, "test") ||
			strings.Contains(clusterName, "cleanup") ||
			strings.Contains(clusterName, "integration") ||
			strings.Contains(clusterName, "collision") ||
			strings.Contains(clusterName, "interrupt") ||
			strings.Contains(clusterName, "stress") ||
			strings.Contains(clusterName, "list-") ||
			strings.Contains(clusterName, "status-") ||
			strings.Contains(clusterName, "create-") ||
			strings.Contains(clusterName, "delete-") ||
			strings.Contains(clusterName, "multi-") ||
			strings.Contains(clusterName, "debug")) {
			CleanupTestCluster(clusterName)
			deletedAny = true
		}
	}

	// If we deleted any clusters, wait for Docker to settle
	if deletedAny {
		time.Sleep(500 * time.Millisecond)
	}

	// Also clean up any leftover Docker networks and containers
	cleanupDockerResources()
}

// cleanupDockerResources removes leftover k3d Docker networks, containers,
// and registries that match known test-name substrings.
func cleanupDockerResources() {
	// Clean up leftover k3d networks
	removeDockerResourcesByFilter(
		[]string{"network", "ls", "--filter", "name=k3d-", "--format", "{{.Name}}"},
		func(name string) []string { return []string{"network", "rm", name} },
		true,
	)

	// Clean up leftover k3d containers
	removeDockerResourcesByFilter(
		[]string{"ps", "-a", "--filter", "name=k3d-", "--format", "{{.Names}}"},
		func(name string) []string { return []string{"rm", "-f", name} },
		true,
	)

	// Clean up any leftover k3d registries that might conflict, restricted
	// to registries matching known test-name substrings to avoid deleting
	// unrelated k3d registries.
	removeDockerResourcesByFilter(
		[]string{"ps", "-a", "--filter", "name=k3d-.*-registry", "--format", "{{.Names}}"},
		func(name string) []string { return []string{"rm", "-f", name} },
		true,
	)
}

// cleanupClusterSpecificResources removes Docker resources for a specific cluster
func cleanupClusterSpecificResources(clusterName string) {
	// Remove specific cluster network
	networkName := fmt.Sprintf("k3d-%s", clusterName)
	_ = exec.Command("docker", "network", "rm", networkName).Run() // #nosec G204 -- integration test harness runs the built CLI/tools with controlled args

	// Remove specific cluster containers
	removeDockerResourcesByFilter(
		[]string{"ps", "-a", "--filter", fmt.Sprintf("name=k3d-%s", clusterName), "--format", "{{.Names}}"},
		func(name string) []string { return []string{"rm", "-f", name} },
		false,
	)
}
