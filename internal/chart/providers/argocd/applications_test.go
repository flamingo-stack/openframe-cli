package argocd

import (
	"context"
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
	tests := []struct {
		name          string
		setupMock     func(*executor.MockCommandExecutor)
		expectedCount int
		verbose       bool
	}{
		{
			name: "successfully counts all applications",
			setupMock: func(m *executor.MockCommandExecutor) {
				// app-of-apps specific call returns empty
				m.SetResponse("kubectl -n argocd get applications.argoproj.io app-of-apps", &executor.CommandResult{
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
			name: "falls back to helm values counting",
			setupMock: func(m *executor.MockCommandExecutor) {
				// App-of-apps specific calls return empty
				m.SetResponse("kubectl -n argocd get applications.argoproj.io app-of-apps", &executor.CommandResult{
					Stdout: "",
				})

				// ArgoCD server pod call returns empty (no server pod found)
				m.SetResponse("kubectl -n argocd get pod -l app.kubernetes.io/name=argocd-server", &executor.CommandResult{
					Stdout: "",
				})

				// General kubectl call returns empty JSON (use -o json pattern)
				m.SetResponse("-o json", &executor.CommandResult{
					Stdout: `{"items":[]}`,
				})

				// Helm values call returns applications
				m.SetResponse("helm get values app-of-apps", &executor.CommandResult{
					Stdout: `applications:
  - name: app1
    repoURL: https://github.com/example/repo1
    targetRevision: main
  - name: app2
    repoURL: https://github.com/example/repo2
    targetRevision: main
  - name: app3
    repoURL: https://github.com/example/repo3
    targetRevision: main`,
				})
			},
			expectedCount: 3,
		},
		{
			name: "estimates from ApplicationSets",
			setupMock: func(m *executor.MockCommandExecutor) {
				// App-of-apps specific calls return empty
				m.SetResponse("kubectl -n argocd get applications.argoproj.io app-of-apps", &executor.CommandResult{
					Stdout: "",
				})

				// ArgoCD server pod call returns empty
				m.SetResponse("kubectl -n argocd get pod", &executor.CommandResult{
					Stdout: "",
				})

				// General kubectl call returns empty JSON (use -o json pattern)
				m.SetResponse("-o json", &executor.CommandResult{
					Stdout: `{"items":[]}`,
				})

				// Helm values call returns empty
				m.SetResponse("helm get values", &executor.CommandResult{
					Stdout: "",
				})

				// ApplicationSets call
				m.SetResponse("applicationsets.argoproj.io", &executor.CommandResult{
					Stdout: "appset1\nappset2\n",
				})
			},
			expectedCount: 14, // 2 appsets * 7 estimated apps each
		},
		{
			name: "returns 0 when no method succeeds",
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
			name: "returns empty list on kubectl error",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetShouldFail(true, "kubectl error")
			},
			expectedApps: []Application{},
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
