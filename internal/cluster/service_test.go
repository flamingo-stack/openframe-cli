package cluster

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

// createTestExecutor creates a mock executor for testing
func createTestExecutor() executor.CommandExecutor {
	mock := executor.NewMockCommandExecutor()

	// Set up mock response for k3d cluster list command
	mockJSON := `[{"name":"test-cluster","serversCount":1,"serversRunning":1,"agentsCount":0,"agentsRunning":0,"nodes":[{"name":"k3d-test-cluster-server-0","role":"server","created":"2024-01-01T00:00:00Z"}]}]`
	mock.SetResponse("k3d cluster list", &executor.CommandResult{
		ExitCode: 0,
		Stdout:   mockJSON,
		Duration: 100,
	})

	return mock
}

func TestNewClusterService(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	if service == nil {
		t.Fatal("NewClusterService should not return nil")
		return
	}

	if service.executor != exec {
		t.Error("service should store the provided executor")
	}

	if service.manager == nil {
		t.Error("service should have a manager initialized")
	}
}

func TestClusterService_CreateCluster(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	// Use a unique cluster name to avoid conflicts with existing clusters
	config := models.ClusterConfig{
		Name:       "test-cluster-unit-test",
		Type:       models.ClusterTypeK3d,
		NodeCount:  1,
		K8sVersion: "v1.25.0",
	}

	_, err := service.CreateCluster(context.Background(), config)
	// With mock executor, error can occur if cluster already exists or kubeconfig issues
	// We just verify it doesn't panic
	_ = err
}

func TestClusterService_CreateCluster_CloudWithoutRegionFailsBeforeAnyCommand(t *testing.T) {
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	for _, clusterType := range []models.ClusterType{models.ClusterTypeEKS, models.ClusterTypeGKE} {
		mock := executor.NewMockCommandExecutor()
		service := NewClusterService(mock)

		_, err := service.CreateCluster(context.Background(), models.ClusterConfig{
			Name:      "cloud-cluster",
			Type:      clusterType,
			NodeCount: 1,
		})

		var invalid models.ErrInvalidClusterConfig
		if !errors.As(err, &invalid) {
			t.Fatalf("expected ErrInvalidClusterConfig for %s without region, got %v", clusterType, err)
		}
		if mock.GetCommandCount() != 0 {
			t.Errorf("no commands should run before validation passes, got: %v", mock.GetExecutedCommands())
		}
	}
}

// ListClusters must not hard-fail when local k3d enumeration fails (e.g. the
// Docker daemon is stopped): a cloud-only user must still be able to list and
// see status of their GKE/EKS clusters. The k3d error degrades to best-effort.
func TestClusterService_ListClusters_K3dFailureIsBestEffort(t *testing.T) {
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	mock := executor.NewMockCommandExecutor()
	// Simulate Docker down: every k3d shell-out fails.
	mock.SetShouldFail(true, "Cannot connect to the Docker daemon")
	service := NewClusterServiceSuppressed(mock)

	clusters, err := service.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters must degrade gracefully when k3d fails, got error: %v", err)
	}
	if len(clusters) != 0 {
		t.Fatalf("expected no clusters (k3d down, empty cloud registry), got %d", len(clusters))
	}
}

func TestClusterService_DeleteCluster_UnknownCloudClusterIsNotFound(t *testing.T) {
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	for _, clusterType := range []models.ClusterType{models.ClusterTypeEKS, models.ClusterTypeGKE} {
		mock := executor.NewMockCommandExecutor()
		service := NewClusterService(mock)

		err := service.DeleteCluster(context.Background(), "cloud-cluster", clusterType, false)

		var notFound models.ErrClusterNotFound
		if !errors.As(err, &notFound) {
			t.Fatalf("expected ErrClusterNotFound for %s, got %v", clusterType, err)
		}
		if mock.GetCommandCount() != 0 {
			t.Errorf("no commands should run for a missing cluster, got: %v", mock.GetExecutedCommands())
		}
	}
}

func TestClusterService_ListClusters_MergesCloudRegistry(t *testing.T) {
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	service := NewClusterService(createTestExecutor())

	clusters, err := service.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters: %v", err)
	}
	baseline := len(clusters)

	// Drop a cloud record into the registry and expect it to appear.
	reg := tfengine.NewRegistry(os.Getenv("OPENFRAME_CLUSTERS_DIR"))
	record := tfengine.Record{
		Name:      "cloudy",
		Type:      models.ClusterTypeEKS,
		Status:    tfengine.StatusReady,
		Region:    "us-east-1",
		NodeCount: 3,
	}
	if err := reg.Workspace("cloudy").Scaffold(record, nil, nil); err != nil {
		t.Fatal(err)
	}

	clusters, err = service.ListClusters()
	if err != nil {
		t.Fatalf("ListClusters: %v", err)
	}
	if len(clusters) != baseline+1 {
		t.Fatalf("expected %d clusters after adding a cloud record, got %d", baseline+1, len(clusters))
	}

	clusterType, err := service.DetectClusterType("cloudy")
	if err != nil || clusterType != models.ClusterTypeEKS {
		t.Fatalf("expected eks for cloudy, got %s / %v", clusterType, err)
	}
}

func TestClusterService_DeleteCluster(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	err := service.DeleteCluster(context.Background(), "test-cluster", models.ClusterTypeK3d, false)
	// With mock executor, this should not fail
	if err != nil {
		t.Errorf("DeleteCluster should not error with mock executor: %v", err)
	}
}

func TestClusterService_ListClusters(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	clusters, err := service.ListClusters()
	// Mock executor might return an error due to parsing mock output, which is acceptable
	// We're mainly testing that the method doesn't panic and returns a valid result
	if err == nil && clusters == nil {
		t.Error("ListClusters should not return nil slice when successful")
	}
}

func TestClusterService_GetClusterStatus(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	_, err := service.GetClusterStatus("test-cluster")
	// Mock executor might return an error for non-existent cluster, which is acceptable
	// We're mainly testing that the method doesn't panic
	_ = err // Ignore error for mock executor
}

func TestClusterService_DetectClusterType(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	_, err := service.DetectClusterType("test-cluster")
	// Mock executor might return an error for non-existent cluster, which is acceptable
	// We're mainly testing that the method doesn't panic
	_ = err // Ignore error for mock executor
}

func TestClusterService_CleanupCluster(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	_, err := service.CleanupCluster(context.Background(), "test-cluster", models.ClusterTypeK3d, false, false)
	if err != nil {
		t.Errorf("CleanupCluster should not error: %v", err)
	}
}

func TestClusterService_ShowClusterStatus(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	// This might fail with mock data, but should not panic
	err := service.ShowClusterStatus("test-cluster", false, false, false)
	// We allow error here since mock data might not be complete
	_ = err
}

func TestClusterService_DisplayClusterList(t *testing.T) {
	exec := createTestExecutor()
	service := NewClusterService(exec)

	// Test with empty cluster list
	clusters := []models.ClusterInfo{}
	err := service.DisplayClusterList(clusters, false, false)
	if err != nil {
		t.Errorf("DisplayClusterList should not error with empty list: %v", err)
	}

	// Test with quiet mode
	err = service.DisplayClusterList(clusters, true, false)
	if err != nil {
		t.Errorf("DisplayClusterList should not error with quiet mode: %v", err)
	}
}

func TestClusterService_WithRealExecutor(t *testing.T) {
	// Test with real executor (dry-run mode)
	exec := executor.NewRealCommandExecutor(true, false) // dry-run mode
	service := NewClusterService(exec)

	if service == nil {
		t.Fatal("service should not be nil")
	}

	// Test that service can be created with real executor
	config := models.ClusterConfig{
		Name:      "test-dry-run",
		Type:      models.ClusterTypeK3d,
		NodeCount: 1,
	}

	// In dry-run mode, this should not actually create anything
	_, err := service.CreateCluster(context.Background(), config)
	// Dry-run might still error if k3d is not available, which is acceptable in tests
	_ = err
}

// TestShouldResumeCloudCreate locks the resume decision (audit П1): a cloud
// registry record with a non-Ready status must RESUME through the provider,
// not short-circuit into the "already exists" reuse path — that made the
// documented "re-run create to resume" unreachable from the CLI.
func TestShouldResumeCloudCreate(t *testing.T) {
	cases := []struct {
		clusterType models.ClusterType
		status      string
		want        bool
	}{
		{models.ClusterTypeGKE, "Failed", true},
		{models.ClusterTypeGKE, "Creating", true},
		{models.ClusterTypeGKE, "Ready", false},
		{models.ClusterTypeEKS, "Failed", true},
		{models.ClusterTypeEKS, "Ready", false},
		{models.ClusterTypeK3d, "1/1", false},
		{models.ClusterTypeK3d, "0/1", false},
	}
	for _, tc := range cases {
		if got := shouldResumeCloudCreate(tc.clusterType, tc.status); got != tc.want {
			t.Errorf("shouldResumeCloudCreate(%s, %q) = %v, want %v", tc.clusterType, tc.status, got, tc.want)
		}
	}
}

// The default (text) status path must resolve through the cloud-aware lookup:
// a registry-recorded GKE/EKS cluster used to be reported "not found" here
// while `--output json` rendered it fine.
func TestClusterService_ShowClusterStatus_FindsCloudCluster(t *testing.T) {
	t.Setenv("OPENFRAME_CLUSTERS_DIR", t.TempDir())
	service := NewClusterService(createTestExecutor())

	reg := tfengine.NewRegistry(os.Getenv("OPENFRAME_CLUSTERS_DIR"))
	record := tfengine.Record{
		Name:      "cloudy",
		Type:      models.ClusterTypeGKE,
		Status:    tfengine.StatusReady,
		Region:    "us-central1",
		Project:   "proj",
		NodeCount: 3,
	}
	if err := reg.Workspace("cloudy").Scaffold(record, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := service.ShowClusterStatus("cloudy", false, false, false); err != nil {
		t.Fatalf("a registry-recorded cloud cluster must not be 'not found': %v", err)
	}

	// An unknown name still errors.
	if err := service.ShowClusterStatus("ghost", false, false, false); err == nil {
		t.Fatal("an unknown cluster must still report not found")
	}
}

// The status box must not hardcode "1/1" as the only healthy shape: cloud
// providers report "Ready" and multi-server k3d clusters report "n/n", and
// both used to display as Partial.
func TestReadinessDisplay(t *testing.T) {
	cases := []struct {
		status string
		want   string
	}{
		{"1/1", "Ready (1/1)"},
		{"3/3", "Ready (3/3)"},
		{"1/3", "Partial (1/3)"},
		{"0/1", "Partial (0/1)"},
		{"0/0", "Partial (0/0)"}, // zero servers is never healthy
		{"Ready", "Ready (Ready)"},
		{"Stopped", "Partial (Stopped)"},
		{"", "Partial ()"},
	}
	for _, tc := range cases {
		if got := readinessDisplay(tc.status); got != tc.want {
			t.Errorf("readinessDisplay(%q) = %q, want %q", tc.status, got, tc.want)
		}
	}
}
