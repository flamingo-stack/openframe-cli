package helm

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// createTestHelmManager creates a HelmManager for testing with a fake clientset
// so the native (client-go) connectivity checks and deployment waits work
// without a real cluster.
func createTestHelmManager(exec executor.CommandExecutor) *HelmManager {
	return &HelmManager{
		executor:   exec,
		kubeClient: k8sfake.NewSimpleClientset(),
		verbose:    false,
	}
}

// testRestConfig returns a dummy rest.Config for use in tests
// This is not used in actual tests since createTestHelmManager creates the struct directly
var _ = &rest.Config{} // Used to ensure the import is not removed

// MockExecutor implements CommandExecutor for testing
type MockExecutor struct {
	commands [][]string
	results  map[string]*executor.CommandResult
	errors   map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		commands: make([][]string, 0),
		results:  make(map[string]*executor.CommandResult),
		errors:   make(map[string]error),
	}
}

func (m *MockExecutor) Execute(ctx context.Context, name string, args ...string) (*executor.CommandResult, error) {
	command := append([]string{name}, args...)
	m.commands = append(m.commands, command)

	commandStr := name
	for _, arg := range args {
		commandStr += " " + arg
	}

	// Check for partial match for error handling (for complex commands)
	for errKey, err := range m.errors {
		if strings.Contains(commandStr, errKey) {
			return nil, err
		}
	}

	if result, exists := m.results[commandStr]; exists {
		return result, nil
	}

	// Default success result
	return &executor.CommandResult{
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
	}, nil
}

func (m *MockExecutor) ExecuteWithOptions(ctx context.Context, options executor.ExecuteOptions) (*executor.CommandResult, error) {
	return m.Execute(ctx, options.Command, options.Args...)
}

func (m *MockExecutor) SetResult(command string, result *executor.CommandResult) {
	m.results[command] = result
}

func (m *MockExecutor) SetError(command string, err error) {
	m.errors[command] = err
}

func (m *MockExecutor) GetCommands() [][]string {
	return m.commands
}

func TestHelmManager_IsHelmInstalled(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockExecutor)
		expectError bool
	}{
		{
			name: "helm is installed",
			setupMock: func(m *MockExecutor) {
				m.SetResult("helm version --short", &executor.CommandResult{
					ExitCode: 0,
					Stdout:   "v3.12.0+g4f11b4a",
				})
			},
			expectError: false,
		},
		{
			name: "helm is not installed",
			setupMock: func(m *MockExecutor) {
				m.SetError("helm version --short", assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			tt.setupMock(mockExec)

			manager := createTestHelmManager(mockExec)
			err := manager.IsHelmInstalled(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errors.ErrHelmNotAvailable)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHelmManager_IsChartInstalled(t *testing.T) {
	tests := []struct {
		name         string
		releaseName  string
		namespace    string
		setupMock    func(*MockExecutor)
		expectResult bool
		expectError  bool
	}{
		{
			name:        "chart is installed",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetResult("helm list -q -n argocd -f argocd", &executor.CommandResult{
					ExitCode: 0,
					Stdout:   "argocd\n",
				})
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:        "chart is not installed",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetResult("helm list -q -n argocd -f argocd", &executor.CommandResult{
					ExitCode: 0,
					Stdout:   "",
				})
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:        "helm command fails",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetError("helm list -q -n argocd -f argocd", assert.AnError)
			},
			expectResult: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			tt.setupMock(mockExec)

			manager := createTestHelmManager(mockExec)
			result, err := manager.IsChartInstalled(context.Background(), tt.releaseName, tt.namespace)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, result)
			}
		})
	}
}

// metadataCmd is the argv GetChartStatus issues. `helm get metadata` is used
// instead of `helm status` because status's JSON carries no chart version, and
// its top-level "version" field is the release REVISION — verified against helm
// v4.2.2 on a live release.
const metadataCmd = "helm get metadata argocd -n argocd --output json"

func TestHelmManager_GetChartStatus(t *testing.T) {
	tests := []struct {
		name        string
		releaseName string
		namespace   string
		setupMock   func(*MockExecutor)
		expectError bool
		wantStatus  string
		wantVersion string
		wantApp     string
	}{
		{
			name:        "successful status retrieval",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetResult(metadataCmd, &executor.CommandResult{
					ExitCode: 0,
					Stdout:   `{"name":"argocd","namespace":"argocd","status":"deployed","version":"7.7.5","appVersion":"v2.13.0","revision":3}`,
				})
			},
			wantStatus:  "deployed",
			wantVersion: "7.7.5",
			wantApp:     "v2.13.0",
		},
		{
			// The point of M2.4: the method used to return a literal
			// "deployed"/"1.0.0" regardless of what helm reported, so a broken
			// release looked healthy and every chart claimed version 1.0.0.
			name:        "a failed release is reported as failed",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetResult(metadataCmd, &executor.CommandResult{
					ExitCode: 0,
					Stdout:   `{"name":"argocd","namespace":"argocd","status":"failed","version":"7.7.5","appVersion":"v2.13.0"}`,
				})
			},
			wantStatus:  "failed",
			wantVersion: "7.7.5",
			wantApp:     "v2.13.0",
		},
		{
			name:        "status command fails",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetError(metadataCmd, assert.AnError)
			},
			expectError: true,
		},
		{
			name:        "unparseable output is an error, not a fabricated status",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetResult(metadataCmd, &executor.CommandResult{ExitCode: 0, Stdout: `not json`})
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			tt.setupMock(mockExec)

			manager := createTestHelmManager(mockExec)
			info, err := manager.GetChartStatus(context.Background(), tt.releaseName, tt.namespace)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.releaseName, info.Name)
				assert.Equal(t, tt.namespace, info.Namespace)
				assert.Equal(t, tt.wantStatus, info.Status)
				assert.Equal(t, tt.wantVersion, info.Version, "the chart version must come from helm, not a constant")
				assert.Equal(t, tt.wantApp, info.AppVersion)
			}
		})
	}
}
