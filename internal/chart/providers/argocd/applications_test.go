package argocd

import (
	"context"
	"os"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	mockExec := executor.NewMockCommandExecutor()
	manager := NewManager(mockExec)

	assert.NotNil(t, manager)
	assert.Equal(t, mockExec, manager.executor)
}

func TestGetTotalExpectedApplications(t *testing.T) {
	// Force kubectl fallback by setting KUBECONFIG to non-existent path
	// This prevents native Kubernetes clients from connecting to a real cluster
	originalKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", "/tmp/nonexistent-kubeconfig")
	defer os.Setenv("KUBECONFIG", originalKubeconfig)

	tests := []struct {
		name          string
		setupMock     func(*executor.MockCommandExecutor)
		expectedCount int
		verbose       bool
	}{
		{
			name: "successfully counts applications via kubectl fallback",
			setupMock: func(m *executor.MockCommandExecutor) {
				// app-of-apps specific call returns empty (no resources found)
				m.SetResponse("app-of-apps -o jsonpath", &executor.CommandResult{
					Stdout: "",
				})
				// Return JSON format for -o json for the general applications query
				m.SetResponse("-o json", &executor.CommandResult{
					Stdout: `{"items":[{"metadata":{"name":"app1"},"status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}},{"metadata":{"name":"app2"},"status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}},{"metadata":{"name":"app3"},"status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}},{"metadata":{"name":"app4"},"status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}},{"metadata":{"name":"app5"},"status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}}]}`,
				})
			},
			expectedCount: 5,
		},
		{
			name: "counts resources from app-of-apps status",
			setupMock: func(m *executor.MockCommandExecutor) {
				// app-of-apps returns resource names via jsonpath
				m.SetResponse("app-of-apps -o jsonpath", &executor.CommandResult{
					Stdout: "app1 app2 app3",
				})
			},
			expectedCount: 3,
		},
		{
			name: "returns 0 when no applications found",
			setupMock: func(m *executor.MockCommandExecutor) {
				// app-of-apps returns empty
				m.SetResponse("app-of-apps -o jsonpath", &executor.CommandResult{
					Stdout: "",
				})
				// General query returns empty items
				m.SetResponse("-o json", &executor.CommandResult{
					Stdout: `{"items":[]}`,
				})
			},
			expectedCount: 0,
		},
		{
			name: "returns 0 when kubectl commands fail",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetDefaultResult(&executor.CommandResult{
					Stdout: "",
				})
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := executor.NewMockCommandExecutor()
			tt.setupMock(mockExec)

			manager := NewManager(mockExec)
			config := config.ChartInstallConfig{
				Verbose: tt.verbose,
			}

			count := manager.getTotalExpectedApplications(context.Background(), config)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestParseApplications(t *testing.T) {
	// Force kubectl fallback by setting KUBECONFIG to non-existent path
	// This prevents native Kubernetes clients from connecting to a real cluster
	originalKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", "/tmp/nonexistent-kubeconfig")
	defer os.Setenv("KUBECONFIG", originalKubeconfig)

	tests := []struct {
		name         string
		setupMock    func(*executor.MockCommandExecutor)
		expectedApps []Application
		expectError  bool
	}{
		{
			name: "successfully parses healthy applications",
			setupMock: func(m *executor.MockCommandExecutor) {
				// Return JSON format for -o json
				m.SetResponse("kubectl -n argocd get applications.argoproj.io", &executor.CommandResult{
					Stdout: `{"items":[{"metadata":{"name":"app1"},"status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}},{"metadata":{"name":"app2"},"status":{"health":{"status":"Progressing"},"sync":{"status":"Synced"}}},{"metadata":{"name":"app3"},"status":{"health":{"status":"Healthy"},"sync":{"status":"OutOfSync"}}}]}`,
				})
			},
			expectedApps: []Application{
				{Name: "app1", Health: "Healthy", Sync: "Synced"},
				{Name: "app2", Health: "Progressing", Sync: "Synced"},
				{Name: "app3", Health: "Healthy", Sync: "OutOfSync"},
			},
		},
		{
			name: "handles applications with unknown status",
			setupMock: func(m *executor.MockCommandExecutor) {
				// Return JSON format with empty/unknown status
				m.SetResponse("kubectl -n argocd get applications.argoproj.io", &executor.CommandResult{
					Stdout: `{"items":[{"metadata":{"name":"app1"},"status":{"health":{"status":"Healthy"},"sync":{"status":"Synced"}}},{"metadata":{"name":"app2"},"status":{"health":{},"sync":{}}},{"metadata":{"name":"app3"},"status":{"health":{"status":"Unknown"},"sync":{"status":"Unknown"}}}]}`,
				})
			},
			expectedApps: []Application{
				{Name: "app1", Health: "Healthy", Sync: "Synced"},
				{Name: "app2", Health: "Unknown", Sync: "Unknown"},
				{Name: "app3", Health: "Unknown", Sync: "Unknown"},
			},
		},
		{
			name: "returns empty list and error on kubectl error",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetShouldFail(true, "kubectl error")
			},
			expectedApps: []Application{},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := executor.NewMockCommandExecutor()
			tt.setupMock(mockExec)

			manager := NewManager(mockExec)
			apps, err := manager.parseApplications(context.Background(), false)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedApps, apps)
			}
		})
	}
}
