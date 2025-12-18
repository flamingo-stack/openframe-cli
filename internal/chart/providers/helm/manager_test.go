package helm

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

// createTestHelmManager creates a HelmManager for testing with a mock rest.Config
func createTestHelmManager(exec executor.CommandExecutor) *HelmManager {
	// Create a minimal rest.Config for testing
	// Note: In tests, we use the manager directly without calling New since the
	// kubernetes clients would fail to initialize with this fake config
	return &HelmManager{
		executor: exec,
		verbose:  false,
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

func TestHelmManager_InstallArgoCD(t *testing.T) {
	tests := []struct {
		name          string
		config        config.ChartInstallConfig
		setupMock     func(*MockExecutor)
		expectError   bool
		checkCommands func(t *testing.T, commands [][]string)
	}{
		{
			name: "successful installation",
			config: config.ChartInstallConfig{
				DryRun: false,
			},
			setupMock: func(m *MockExecutor) {
				// All commands should succeed
			},
			expectError: false,
			checkCommands: func(t *testing.T, commands [][]string) {
				// Verify expected commands were called
				require.GreaterOrEqual(t, len(commands), 3)

				// Commands may be wrapped in wsl on Windows, so check the command name flexibly
				// Should have added repo and updated - check for "helm" or "wsl" as first command
				if len(commands[0]) > 0 {
					firstCmd := commands[0][0]
					assert.True(t, firstCmd == "helm" || firstCmd == "wsl", "First command should be helm or wsl, got %s", firstCmd)
				}

				// Should have upgrade/install command
				installCmd := commands[2]
				// On Windows, command might be: wsl -d Ubuntu helm upgrade...
				// On Unix, command might be: helm upgrade...
				cmdStart := 0
				if len(installCmd) > 0 && installCmd[0] == "wsl" {
					// Skip wsl wrapper args to find actual helm command
					for i, arg := range installCmd {
						if arg == "helm" {
							cmdStart = i
							break
						}
					}
				}

				if cmdStart < len(installCmd) {
					assert.Equal(t, "helm", installCmd[cmdStart])
					if cmdStart+1 < len(installCmd) {
						assert.Equal(t, "upgrade", installCmd[cmdStart+1])
					}
					if cmdStart+2 < len(installCmd) {
						assert.Equal(t, "--install", installCmd[cmdStart+2])
					}
				}

				// Check that install command contains expected flags
				installCmdStr := strings.Join(installCmd, " ")
				assert.Contains(t, installCmdStr, "argo-cd")
				assert.Contains(t, installCmdStr, "argo/argo-cd")
				assert.Contains(t, installCmdStr, "--version=8.2.7")
				assert.Contains(t, installCmdStr, "--namespace")
				assert.Contains(t, installCmdStr, "argocd")
				assert.Contains(t, installCmdStr, "--create-namespace")
				assert.Contains(t, installCmdStr, "--wait")
				assert.Contains(t, installCmdStr, "--timeout")
				// Timeout may vary (7m for ArgoCD, 30m for app-of-apps)
				assert.True(t, strings.Contains(installCmdStr, "7m") || strings.Contains(installCmdStr, "30m"), "Should contain timeout value")
				assert.Contains(t, installCmdStr, "argocd-values")
			},
		},
		{
			name: "dry run installation",
			config: config.ChartInstallConfig{
				DryRun: true,
			},
			setupMock: func(m *MockExecutor) {
				// All commands should succeed
			},
			expectError: false,
			checkCommands: func(t *testing.T, commands [][]string) {
				require.GreaterOrEqual(t, len(commands), 3)
				installCmd := commands[2]
				installCmdStr := strings.Join(installCmd, " ")
				assert.Contains(t, installCmdStr, "--dry-run")
			},
		},
		{
			name: "repo add fails",
			config: config.ChartInstallConfig{
				DryRun: false,
			},
			setupMock: func(m *MockExecutor) {
				m.SetError("helm repo add argo https://argoproj.github.io/argo-helm", assert.AnError)
			},
			expectError:   true,
			checkCommands: func(t *testing.T, commands [][]string) {},
		},
		{
			name: "repo update fails",
			config: config.ChartInstallConfig{
				DryRun: false,
			},
			setupMock: func(m *MockExecutor) {
				m.SetError("helm repo update", assert.AnError)
			},
			expectError:   true,
			checkCommands: func(t *testing.T, commands [][]string) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := NewMockExecutor()
			tt.setupMock(mockExec)

			manager := createTestHelmManager(mockExec)
			err := manager.InstallArgoCD(context.Background(), tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.checkCommands(t, mockExec.GetCommands())
			}
		})
	}
}

func TestHelmManager_GetChartStatus(t *testing.T) {
	tests := []struct {
		name        string
		releaseName string
		namespace   string
		setupMock   func(*MockExecutor)
		expectError bool
	}{
		{
			name:        "successful status retrieval",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetResult("helm status argocd -n argocd --output json", &executor.CommandResult{
					ExitCode: 0,
					Stdout:   `{"name":"argocd","namespace":"argocd","info":{"status":"deployed"}}`,
				})
			},
			expectError: false,
		},
		{
			name:        "status command fails",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *MockExecutor) {
				m.SetError("helm status argocd -n argocd --output json", assert.AnError)
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
				assert.Equal(t, "deployed", info.Status)
			}
		})
	}
}
