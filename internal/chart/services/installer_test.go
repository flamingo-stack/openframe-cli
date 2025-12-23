package services

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/errors"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockArgoCDService is a mock implementation of ArgoCDService
type MockArgoCDService struct {
	mock.Mock
}

func (m *MockArgoCDService) Install(ctx context.Context, config config.ChartInstallConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockArgoCDService) IsInstalled(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockArgoCDService) GetStatus(ctx context.Context) (models.ChartInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.ChartInfo), args.Error(1)
}

func (m *MockArgoCDService) WaitForApplications(ctx context.Context, config config.ChartInstallConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

// MockAppOfAppsService is a mock implementation of AppOfAppsService
type MockAppOfAppsService struct {
	mock.Mock
}

func (m *MockAppOfAppsService) Install(ctx context.Context, config config.ChartInstallConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAppOfAppsService) IsInstalled(ctx context.Context, namespace string) (bool, error) {
	args := m.Called(ctx, namespace)
	return args.Bool(0), args.Error(1)
}

func (m *MockAppOfAppsService) GetStatus(ctx context.Context, namespace string) (models.ChartInfo, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(models.ChartInfo), args.Error(1)
}

func TestInstaller_InstallCharts(t *testing.T) {
	tests := []struct {
		name             string
		config           config.ChartInstallConfig
		setupMocks       func(*MockArgoCDService, *MockAppOfAppsService)
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "successful installation without app-of-apps",
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
				AppOfApps:   nil, // No app-of-apps
			},
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
				// WaitForApplications should not be called when no app-of-apps
			},
			expectedError: false,
		},
		{
			name: "successful installation with app-of-apps",
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
				AppOfApps: &models.AppOfAppsConfig{
					GitHubRepo: "owner/repo",
				},
			},
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
				appOfApps.On("Install", mock.Anything, mock.Anything).Return(nil)
				argoCD.On("WaitForApplications", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "ArgoCD installation failure",
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
			},
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			expectedError:    true,
			expectedErrorMsg: "installation failed for ArgoCD",
		},
		{
			name: "app-of-apps installation failure",
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
				AppOfApps: &models.AppOfAppsConfig{
					GitHubRepo: "owner/repo",
				},
			},
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
				appOfApps.On("Install", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			expectedError:    true,
			expectedErrorMsg: "installation failed for app-of-apps",
		},
		{
			name: "branch not found error passes through",
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
				AppOfApps: &models.AppOfAppsConfig{
					GitHubRepo: "owner/repo",
				},
			},
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
				branchErr := &sharedErrors.BranchNotFoundError{
					Branch: "main",
				}
				appOfApps.On("Install", mock.Anything, mock.Anything).
					Return(branchErr)
			},
			expectedError:    true,
			expectedErrorMsg: "branch 'main'",
		},
		{
			name: "wait for applications failure",
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
				AppOfApps: &models.AppOfAppsConfig{
					GitHubRepo: "owner/repo",
				},
			},
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
				appOfApps.On("Install", mock.Anything, mock.Anything).Return(nil)
				argoCD.On("WaitForApplications", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			expectedError:    true,
			expectedErrorMsg: "waiting failed for ArgoCD applications",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockArgoCD := new(MockArgoCDService)
			mockAppOfApps := new(MockAppOfAppsService)

			// Setup mocks
			tt.setupMocks(mockArgoCD, mockAppOfApps)

			// Create installer with mocks
			installer := &Installer{
				argoCDService:    mockArgoCD,
				appOfAppsService: mockAppOfApps,
			}

			// Execute
			err := installer.InstallCharts(tt.config)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockArgoCD.AssertExpectations(t)
			mockAppOfApps.AssertExpectations(t)
		})
	}
}

func TestInstaller_InstallCharts_RecoverableError(t *testing.T) {
	// Test that wait failures return recoverable errors
	mockArgoCD := new(MockArgoCDService)
	mockAppOfApps := new(MockAppOfAppsService)

	config := config.ChartInstallConfig{
		ClusterName: "test-cluster",
		AppOfApps: &models.AppOfAppsConfig{
			GitHubRepo: "owner/repo",
		},
	}

	mockArgoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
	mockAppOfApps.On("Install", mock.Anything, mock.Anything).Return(nil)
	mockArgoCD.On("WaitForApplications", mock.Anything, mock.Anything).
		Return(assert.AnError)

	installer := &Installer{
		argoCDService:    mockArgoCD,
		appOfAppsService: mockAppOfApps,
	}

	err := installer.InstallCharts(config)
	assert.Error(t, err)

	// Check that error is NOT recoverable (WaitForApplications failures should not trigger reinstallation)
	// ArgoCD and app-of-apps are already installed, so retrying would reinstall them unnecessarily.
	// WaitForApplications has its own internal retry logic.
	chartErr, ok := err.(*errors.ChartError)
	assert.True(t, ok, "Expected ChartError")
	assert.False(t, chartErr.IsRecoverable(), "Expected non-recoverable error - waiting failures should not trigger reinstallation")
}

func TestInstaller_InstallCharts_NoWaitWithoutAppOfApps(t *testing.T) {
	// Verify WaitForApplications is not called when no app-of-apps is configured
	mockArgoCD := new(MockArgoCDService)
	mockAppOfApps := new(MockAppOfAppsService)

	config := config.ChartInstallConfig{
		ClusterName: "test-cluster",
		// No GitHubRepo means no app-of-apps
	}

	mockArgoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
	// Explicitly NOT setting up WaitForApplications expectation

	installer := &Installer{
		argoCDService:    mockArgoCD,
		appOfAppsService: mockAppOfApps,
	}

	err := installer.InstallCharts(config)
	assert.NoError(t, err)

	// Verify Install was called but WaitForApplications was not
	mockArgoCD.AssertExpectations(t)
	mockArgoCD.AssertNotCalled(t, "WaitForApplications", mock.Anything, mock.Anything)
}

func TestInstaller_InstallCharts_ErrorTypes(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockArgoCDService, *MockAppOfAppsService)
		config     config.ChartInstallConfig
		checkError func(t *testing.T, err error)
	}{
		{
			name: "ArgoCD error wrapped as ChartError",
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
			},
			checkError: func(t *testing.T, err error) {
				chartErr, ok := err.(*errors.ChartError)
				assert.True(t, ok)
				assert.Equal(t, "ArgoCD", chartErr.Component)
				assert.Equal(t, "installation", chartErr.Operation)
				assert.Equal(t, "test-cluster", chartErr.ClusterName)
			},
		},
		{
			name: "App-of-apps error wrapped as ChartError",
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
				appOfApps.On("Install", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
				AppOfApps: &models.AppOfAppsConfig{
					GitHubRepo: "owner/repo",
				},
			},
			checkError: func(t *testing.T, err error) {
				chartErr, ok := err.(*errors.ChartError)
				assert.True(t, ok)
				assert.Equal(t, "app-of-apps", chartErr.Component)
				assert.Equal(t, "installation", chartErr.Operation)
				assert.Equal(t, "test-cluster", chartErr.ClusterName)
			},
		},
		{
			name: "Wait error is NOT recoverable",
			setupMocks: func(argoCD *MockArgoCDService, appOfApps *MockAppOfAppsService) {
				argoCD.On("Install", mock.Anything, mock.Anything).Return(nil)
				appOfApps.On("Install", mock.Anything, mock.Anything).Return(nil)
				argoCD.On("WaitForApplications", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			config: config.ChartInstallConfig{
				ClusterName: "test-cluster",
				AppOfApps: &models.AppOfAppsConfig{
					GitHubRepo: "owner/repo",
				},
			},
			checkError: func(t *testing.T, err error) {
				// WaitForApplications failures should NOT trigger reinstallation
				// ArgoCD and app-of-apps are already installed, and WaitForApplications
				// has its own internal retry logic.
				chartErr, ok := err.(*errors.ChartError)
				assert.True(t, ok)
				assert.Equal(t, "ArgoCD applications", chartErr.Component)
				assert.Equal(t, "waiting", chartErr.Operation)
				assert.False(t, chartErr.IsRecoverable(), "Wait errors should not trigger reinstallation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockArgoCD := new(MockArgoCDService)
			mockAppOfApps := new(MockAppOfAppsService)

			tt.setupMocks(mockArgoCD, mockAppOfApps)

			installer := &Installer{
				argoCDService:    mockArgoCD,
				appOfAppsService: mockAppOfApps,
			}

			err := installer.InstallCharts(tt.config)
			assert.Error(t, err)
			tt.checkError(t, err)

			mockArgoCD.AssertExpectations(t)
			mockAppOfApps.AssertExpectations(t)
		})
	}
}
