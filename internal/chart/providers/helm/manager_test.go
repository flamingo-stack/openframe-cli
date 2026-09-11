package helm

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	k8sfake "k8s.io/client-go/kubernetes/fake"
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

func TestHelmManager_IsHelmInstalled(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*executor.MockCommandExecutor)
		expectError bool
	}{
		{
			name: "helm is installed",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse("helm version --short", &executor.CommandResult{
					ExitCode: 0,
					Stdout:   "v3.12.0+g4f11b4a",
				}, nil)
			},
			expectError: false,
		},
		{
			name: "helm is not installed",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse("helm version --short", nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := executor.NewMockCommandExecutor()
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
		setupMock    func(*executor.MockCommandExecutor)
		expectResult bool
		expectError  bool
	}{
		{
			name:        "chart is installed",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse("helm list -q -n argocd -f argocd", &executor.CommandResult{
					ExitCode: 0,
					Stdout:   "argocd\n",
				}, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:        "chart is not installed",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse("helm list -q -n argocd -f argocd", &executor.CommandResult{
					ExitCode: 0,
					Stdout:   "",
				}, nil)
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:        "helm command fails",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse("helm list -q -n argocd -f argocd", nil, assert.AnError)
			},
			expectResult: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := executor.NewMockCommandExecutor()
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
		setupMock   func(*executor.MockCommandExecutor)
		expectError bool
		wantStatus  string
		wantVersion string
		wantApp     string
	}{
		{
			name:        "successful status retrieval",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse(metadataCmd, &executor.CommandResult{
					ExitCode: 0,
					Stdout:   `{"name":"argocd","namespace":"argocd","status":"deployed","version":"7.7.5","appVersion":"v2.13.0","revision":3}`,
				}, nil)
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
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse(metadataCmd, &executor.CommandResult{
					ExitCode: 0,
					Stdout:   `{"name":"argocd","namespace":"argocd","status":"failed","version":"7.7.5","appVersion":"v2.13.0"}`,
				}, nil)
			},
			wantStatus:  "failed",
			wantVersion: "7.7.5",
			wantApp:     "v2.13.0",
		},
		{
			name:        "status command fails",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse(metadataCmd, nil, assert.AnError)
			},
			expectError: true,
		},
		{
			name:        "unparseable output is an error, not a fabricated status",
			releaseName: "argocd",
			namespace:   "argocd",
			setupMock: func(m *executor.MockCommandExecutor) {
				m.SetResponse(metadataCmd, &executor.CommandResult{ExitCode: 0, Stdout: `not json`}, nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := executor.NewMockCommandExecutor()
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
