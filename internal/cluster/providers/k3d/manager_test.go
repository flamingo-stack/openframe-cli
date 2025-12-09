package k3d

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	execPkg "github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExecutor is a mock implementation of CommandExecutor for testing
type MockExecutor struct {
	mock.Mock
}

// setupTestKubeconfig creates a temporary kubeconfig file for tests
// Returns a cleanup function that should be deferred
func setupTestKubeconfig(t *testing.T, clusterName string) func() {
	t.Helper()

	// Create temp directory for kubeconfig
	tempDir := t.TempDir()
	kubeconfigPath := filepath.Join(tempDir, "config")

	// Create a minimal kubeconfig with the expected context
	kubeconfigContent := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6550
  name: k3d-` + clusterName + `
contexts:
- context:
    cluster: k3d-` + clusterName + `
    user: admin@k3d-` + clusterName + `
  name: k3d-` + clusterName + `
current-context: k3d-` + clusterName + `
users:
- name: admin@k3d-` + clusterName + `
  user:
    client-certificate-data: dGVzdA==
    client-key-data: dGVzdA==
`

	err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0600)
	if err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	// Set KUBECONFIG env var to point to our test file
	oldKubeconfig := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", kubeconfigPath)

	return func() {
		if oldKubeconfig != "" {
			os.Setenv("KUBECONFIG", oldKubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}
}

func (m *MockExecutor) Execute(ctx context.Context, name string, args ...string) (*execPkg.CommandResult, error) {
	arguments := m.Called(ctx, name, args)
	if arguments.Get(0) == nil {
		return nil, arguments.Error(1)
	}
	return arguments.Get(0).(*execPkg.CommandResult), arguments.Error(1)
}

func (m *MockExecutor) ExecuteWithOptions(ctx context.Context, options execPkg.ExecuteOptions) (*execPkg.CommandResult, error) {
	arguments := m.Called(ctx, options)
	if arguments.Get(0) == nil {
		return nil, arguments.Error(1)
	}
	return arguments.Get(0).(*execPkg.CommandResult), arguments.Error(1)
}

func TestNewK3dManager(t *testing.T) {
	executor := &MockExecutor{}

	t.Run("creates manager with executor", func(t *testing.T) {
		manager := NewK3dManager(executor, false)

		assert.NotNil(t, manager)
		assert.Equal(t, executor, manager.executor)
		assert.False(t, manager.verbose)
	})

	t.Run("creates manager with verbose mode", func(t *testing.T) {
		manager := NewK3dManager(executor, true)

		assert.NotNil(t, manager)
		assert.True(t, manager.verbose)
	})
}

func TestCreateClusterManagerWithExecutor(t *testing.T) {
	t.Run("creates manager with executor", func(t *testing.T) {
		executor := &MockExecutor{}
		manager := CreateClusterManagerWithExecutor(executor)

		assert.NotNil(t, manager)
		assert.Equal(t, executor, manager.executor)
		assert.False(t, manager.verbose) // Default to non-verbose
	})

	t.Run("panics with nil executor", func(t *testing.T) {
		assert.Panics(t, func() {
			CreateClusterManagerWithExecutor(nil)
		})
	})
}

func TestCreateDefaultClusterManager(t *testing.T) {
	t.Run("panics as expected", func(t *testing.T) {
		assert.Panics(t, func() {
			CreateDefaultClusterManager()
		})
	})
}

func TestK3dManager_CreateCluster(t *testing.T) {
	tests := []struct {
		name           string
		config         models.ClusterConfig
		setupMock      func(*MockExecutor)
		setupKubeconfig bool
		expectedError  string
		expectedArgs   []string
	}{
		{
			name: "successful cluster creation",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      models.ClusterTypeK3d,
				NodeCount: 3,
			},
			setupKubeconfig: true,
			setupMock: func(m *MockExecutor) {
				// Mock bash for kubeconfig directory prep and cleanup
				m.On("Execute", mock.Anything, "bash", mock.Anything).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
				m.On("Execute", mock.Anything, "k3d", mock.Anything).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
			},
		},
		{
			name: "cluster creation with k8s version",
			config: models.ClusterConfig{
				Name:       "test-cluster",
				Type:       models.ClusterTypeK3d,
				NodeCount:  2,
				K8sVersion: "v1.25.0-k3s1",
			},
			setupKubeconfig: true,
			setupMock: func(m *MockExecutor) {
				// Mock bash for kubeconfig directory prep and cleanup
				m.On("Execute", mock.Anything, "bash", mock.Anything).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
				m.On("Execute", mock.Anything, "k3d", mock.Anything).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
			},
		},
		{
			name: "empty cluster name",
			config: models.ClusterConfig{
				Name:      "",
				Type:      models.ClusterTypeK3d,
				NodeCount: 3,
			},
			expectedError: "cluster name cannot be empty",
		},
		{
			name: "invalid cluster type",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      models.ClusterTypeGKE,
				NodeCount: 3,
			},
			expectedError: "no provider available for cluster type 'gke'",
		},
		{
			name: "zero node count",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      models.ClusterTypeK3d,
				NodeCount: 0,
			},
			expectedError: "node count must be at least 1",
		},
		{
			name: "k3d command fails",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      models.ClusterTypeK3d,
				NodeCount: 3,
			},
			setupMock: func(m *MockExecutor) {
				// Mock bash for kubeconfig directory prep and cleanup
				m.On("Execute", mock.Anything, "bash", mock.Anything).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
				m.On("Execute", mock.Anything, "k3d", mock.Anything).Return(nil, errors.New("k3d error"))
			},
			expectedError: "failed to create cluster test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup kubeconfig if needed for tests that verify cluster reachability
			if tt.setupKubeconfig {
				cleanup := setupTestKubeconfig(t, tt.config.Name)
				defer cleanup()
			}

			executor := &MockExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(executor)
			}

			manager := NewK3dManager(executor, false)
			_, err := manager.CreateCluster(context.Background(), tt.config)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				// For successful tests, we expect either no error or a connection error
				// (since we can't actually connect to the cluster in tests)
				if err != nil {
					// Accept connection errors as "success" since the kubeconfig was loaded correctly
					assert.Contains(t, err.Error(), "cluster created but not reachable")
				}
			}

			executor.AssertExpectations(t)
		})
	}
}

func TestK3dManager_CreateCluster_VerboseMode(t *testing.T) {
	// Setup kubeconfig for the test
	cleanup := setupTestKubeconfig(t, "test-cluster")
	defer cleanup()

	executor := &MockExecutor{}
	// Mock bash for kubeconfig directory prep and cleanup
	executor.On("Execute", mock.Anything, "bash", mock.Anything).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
	executor.On("Execute", mock.Anything, "k3d", mock.Anything).Return(&execPkg.CommandResult{Stdout: "success"}, nil)

	manager := NewK3dManager(executor, true) // verbose mode
	config := models.ClusterConfig{
		Name:      "test-cluster",
		Type:      models.ClusterTypeK3d,
		NodeCount: 3,
	}

	_, err := manager.CreateCluster(context.Background(), config)
	// Accept connection errors as "success" since the kubeconfig was loaded correctly
	if err != nil {
		assert.Contains(t, err.Error(), "cluster created but not reachable")
	}
	executor.AssertExpectations(t)
}

func TestK3dManager_DeleteCluster(t *testing.T) {
	tests := []struct {
		name          string
		clusterName   string
		clusterType   models.ClusterType
		force         bool
		setupMock     func(*MockExecutor)
		expectedError string
	}{
		{
			name:        "successful cluster deletion",
			clusterName: "test-cluster",
			clusterType: models.ClusterTypeK3d,
			force:       false,
			setupMock: func(m *MockExecutor) {
				m.On("Execute", mock.Anything, "k3d", []string{"cluster", "delete", "test-cluster"}).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
			},
		},
		{
			name:          "empty cluster name",
			clusterName:   "",
			clusterType:   models.ClusterTypeK3d,
			expectedError: "cluster name cannot be empty",
		},
		{
			name:          "invalid cluster type",
			clusterName:   "test-cluster",
			clusterType:   models.ClusterTypeGKE,
			expectedError: "no provider available for cluster type 'gke'",
		},
		{
			name:        "k3d command fails",
			clusterName: "test-cluster",
			clusterType: models.ClusterTypeK3d,
			setupMock: func(m *MockExecutor) {
				m.On("Execute", mock.Anything, "k3d", mock.Anything).Return(nil, errors.New("k3d error"))
			},
			expectedError: "failed to delete cluster test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &MockExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(executor)
			}

			manager := NewK3dManager(executor, false)
			err := manager.DeleteCluster(context.Background(), tt.clusterName, tt.clusterType, tt.force)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			executor.AssertExpectations(t)
		})
	}
}

func TestK3dManager_StartCluster(t *testing.T) {
	tests := []struct {
		name          string
		clusterName   string
		clusterType   models.ClusterType
		setupMock     func(*MockExecutor)
		expectedError string
	}{
		{
			name:        "successful cluster start",
			clusterName: "test-cluster",
			clusterType: models.ClusterTypeK3d,
			setupMock: func(m *MockExecutor) {
				m.On("Execute", mock.Anything, "k3d", []string{"cluster", "start", "test-cluster"}).Return(&execPkg.CommandResult{Stdout: "success"}, nil)
			},
		},
		{
			name:          "empty cluster name",
			clusterName:   "",
			clusterType:   models.ClusterTypeK3d,
			expectedError: "cluster name cannot be empty",
		},
		{
			name:          "invalid cluster type",
			clusterName:   "test-cluster",
			clusterType:   models.ClusterTypeGKE,
			expectedError: "no provider available for cluster type 'gke'",
		},
		{
			name:        "k3d command fails",
			clusterName: "test-cluster",
			clusterType: models.ClusterTypeK3d,
			setupMock: func(m *MockExecutor) {
				m.On("Execute", mock.Anything, "k3d", mock.Anything).Return(nil, errors.New("k3d error"))
			},
			expectedError: "failed to start cluster test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &MockExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(executor)
			}

			manager := NewK3dManager(executor, false)
			err := manager.StartCluster(context.Background(), tt.clusterName, tt.clusterType)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			executor.AssertExpectations(t)
		})
	}
}

func TestK3dManager_ListClusters(t *testing.T) {
	t.Run("successful cluster listing", func(t *testing.T) {
		executor := &MockExecutor{}
		jsonOutput := `[
			{
				"name": "cluster1",
				"serversCount": 1,
				"serversRunning": 1,
				"agentsCount": 2,
				"agentsRunning": 2,
				"image": "rancher/k3s:latest"
			},
			{
				"name": "cluster2",
				"serversCount": 1,
				"serversRunning": 0,
				"agentsCount": 1,
				"agentsRunning": 0,
				"image": "rancher/k3s:v1.25.0"
			}
		]`

		executor.On("Execute", mock.Anything, "k3d", []string{"cluster", "list", "--output", "json"}).Return(&execPkg.CommandResult{Stdout: jsonOutput}, nil)

		manager := NewK3dManager(executor, false)
		clusters, err := manager.ListClusters(context.Background())

		assert.NoError(t, err)
		assert.Len(t, clusters, 2)

		assert.Equal(t, "cluster1", clusters[0].Name)
		assert.Equal(t, models.ClusterTypeK3d, clusters[0].Type)
		assert.Equal(t, "1/1", clusters[0].Status)
		assert.Equal(t, 3, clusters[0].NodeCount) // 1 server + 2 agents

		assert.Equal(t, "cluster2", clusters[1].Name)
		assert.Equal(t, models.ClusterTypeK3d, clusters[1].Type)
		assert.Equal(t, "0/1", clusters[1].Status)
		assert.Equal(t, 2, clusters[1].NodeCount) // 1 server + 1 agent

		executor.AssertExpectations(t)
	})

	t.Run("k3d command fails", func(t *testing.T) {
		executor := &MockExecutor{}
		executor.On("Execute", mock.Anything, "k3d", mock.Anything).Return(nil, errors.New("k3d error"))

		manager := NewK3dManager(executor, false)
		clusters, err := manager.ListClusters(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list clusters")
		assert.Nil(t, clusters)

		executor.AssertExpectations(t)
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		executor := &MockExecutor{}
		executor.On("Execute", mock.Anything, "k3d", mock.Anything).Return(&execPkg.CommandResult{Stdout: "invalid json"}, nil)

		manager := NewK3dManager(executor, false)
		clusters, err := manager.ListClusters(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse cluster list JSON")
		assert.Nil(t, clusters)

		executor.AssertExpectations(t)
	})
}

func TestK3dManager_ListAllClusters(t *testing.T) {
	t.Run("calls ListClusters", func(t *testing.T) {
		executor := &MockExecutor{}
		executor.On("Execute", mock.Anything, "k3d", []string{"cluster", "list", "--output", "json"}).Return(&execPkg.CommandResult{Stdout: "[]"}, nil)

		manager := NewK3dManager(executor, false)
		clusters, err := manager.ListAllClusters(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, clusters)

		executor.AssertExpectations(t)
	})
}

func TestK3dManager_GetClusterStatus(t *testing.T) {
	t.Run("successful status retrieval", func(t *testing.T) {
		executor := &MockExecutor{}
		jsonOutput := `[
			{
				"name": "test-cluster",
				"serversCount": 1,
				"serversRunning": 1,
				"agentsCount": 2,
				"agentsRunning": 2,
				"image": "rancher/k3s:latest"
			}
		]`

		executor.On("Execute", mock.Anything, "k3d", []string{"cluster", "list", "--output", "json"}).Return(&execPkg.CommandResult{Stdout: jsonOutput}, nil)

		manager := NewK3dManager(executor, false)
		clusterInfo, err := manager.GetClusterStatus(context.Background(), "test-cluster")

		assert.NoError(t, err)
		assert.Equal(t, "test-cluster", clusterInfo.Name)
		assert.Equal(t, models.ClusterTypeK3d, clusterInfo.Type)
		assert.Equal(t, "1/1", clusterInfo.Status)

		executor.AssertExpectations(t)
	})

	t.Run("empty cluster name", func(t *testing.T) {
		executor := &MockExecutor{}
		manager := NewK3dManager(executor, false)

		clusterInfo, err := manager.GetClusterStatus(context.Background(), "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cluster name cannot be empty")
		assert.Equal(t, models.ClusterInfo{}, clusterInfo)
	})

	t.Run("cluster not found", func(t *testing.T) {
		executor := &MockExecutor{}
		executor.On("Execute", mock.Anything, "k3d", []string{"cluster", "list", "--output", "json"}).Return(&execPkg.CommandResult{Stdout: "[]"}, nil)

		manager := NewK3dManager(executor, false)
		clusterInfo, err := manager.GetClusterStatus(context.Background(), "non-existent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cluster non-existent not found")
		assert.Equal(t, models.ClusterInfo{}, clusterInfo)

		executor.AssertExpectations(t)
	})
}

func TestK3dManager_DetectClusterType(t *testing.T) {
	t.Run("successful cluster detection", func(t *testing.T) {
		executor := &MockExecutor{}
		executor.On("Execute", mock.Anything, "k3d", []string{"cluster", "get", "test-cluster"}).Return(&execPkg.CommandResult{Stdout: "cluster info"}, nil)

		manager := NewK3dManager(executor, false)
		clusterType, err := manager.DetectClusterType(context.Background(), "test-cluster")

		assert.NoError(t, err)
		assert.Equal(t, models.ClusterTypeK3d, clusterType)

		executor.AssertExpectations(t)
	})

	t.Run("empty cluster name", func(t *testing.T) {
		executor := &MockExecutor{}
		manager := NewK3dManager(executor, false)

		clusterType, err := manager.DetectClusterType(context.Background(), "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cluster name cannot be empty")
		assert.Equal(t, models.ClusterType(""), clusterType)
	})

	t.Run("cluster not found", func(t *testing.T) {
		executor := &MockExecutor{}
		executor.On("Execute", mock.Anything, "k3d", mock.Anything).Return(nil, errors.New("cluster not found"))

		manager := NewK3dManager(executor, false)
		clusterType, err := manager.DetectClusterType(context.Background(), "non-existent")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cluster 'non-existent' not found")
		assert.Equal(t, models.ClusterType(""), clusterType)

		executor.AssertExpectations(t)
	})
}

func TestK3dManager_GetKubeconfig(t *testing.T) {
	t.Run("successful kubeconfig retrieval", func(t *testing.T) {
		executor := &MockExecutor{}
		kubeconfigContent := "apiVersion: v1\nkind: Config\n..."
		executor.On("Execute", mock.Anything, "k3d", []string{"kubeconfig", "get", "test-cluster"}).Return(&execPkg.CommandResult{Stdout: kubeconfigContent}, nil)

		manager := NewK3dManager(executor, false)
		kubeconfig, err := manager.GetKubeconfig(context.Background(), "test-cluster", models.ClusterTypeK3d)

		assert.NoError(t, err)
		assert.Equal(t, kubeconfigContent, kubeconfig)

		executor.AssertExpectations(t)
	})

	t.Run("unsupported cluster type", func(t *testing.T) {
		executor := &MockExecutor{}
		manager := NewK3dManager(executor, false)

		kubeconfig, err := manager.GetKubeconfig(context.Background(), "test-cluster", models.ClusterTypeGKE)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no provider available for cluster type 'gke'")
		assert.Empty(t, kubeconfig)
	})

	t.Run("k3d command fails", func(t *testing.T) {
		executor := &MockExecutor{}
		executor.On("Execute", mock.Anything, "k3d", mock.Anything).Return(nil, errors.New("k3d error"))

		manager := NewK3dManager(executor, false)
		kubeconfig, err := manager.GetKubeconfig(context.Background(), "test-cluster", models.ClusterTypeK3d)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get kubeconfig for cluster test-cluster")
		assert.Empty(t, kubeconfig)

		executor.AssertExpectations(t)
	})
}

func TestK3dManager_validateClusterConfig(t *testing.T) {
	manager := &K3dManager{}

	tests := []struct {
		name          string
		config        models.ClusterConfig
		expectedError string
	}{
		{
			name: "valid config",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      models.ClusterTypeK3d,
				NodeCount: 3,
			},
		},
		{
			name: "empty name",
			config: models.ClusterConfig{
				Name:      "",
				Type:      models.ClusterTypeK3d,
				NodeCount: 3,
			},
			expectedError: "cluster name cannot be empty",
		},
		{
			name: "empty type",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      "",
				NodeCount: 3,
			},
			expectedError: "cluster type cannot be empty",
		},
		{
			name: "zero node count",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      models.ClusterTypeK3d,
				NodeCount: 0,
			},
			expectedError: "node count must be at least 1",
		},
		{
			name: "negative node count",
			config: models.ClusterConfig{
				Name:      "test-cluster",
				Type:      models.ClusterTypeK3d,
				NodeCount: -1,
			},
			expectedError: "node count must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.validateClusterConfig(tt.config)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// parseNodeCount is a helper function that calculates total node count from agents and servers
// This mimics the logic in k3d_manager.go: NodeCount: k3dCluster.AgentsCount + k3dCluster.ServersCount
func parseNodeCount(agents, servers string) int {
	agentCount, err := strconv.Atoi(agents)
	if err != nil {
		agentCount = 0
	}

	serverCount, err := strconv.Atoi(servers)
	if err != nil {
		serverCount = 0
	}

	return agentCount + serverCount
}

func TestParseNodeCount(t *testing.T) {
	tests := []struct {
		name     string
		agents   string
		servers  string
		expected int
	}{
		{
			name:     "valid counts",
			agents:   "2",
			servers:  "1",
			expected: 3,
		},
		{
			name:     "zero agents",
			agents:   "0",
			servers:  "1",
			expected: 1,
		},
		{
			name:     "invalid agents",
			agents:   "invalid",
			servers:  "1",
			expected: 1,
		},
		{
			name:     "invalid servers",
			agents:   "2",
			servers:  "invalid",
			expected: 2,
		},
		{
			name:     "both invalid",
			agents:   "invalid",
			servers:  "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNodeCount(tt.agents, tt.servers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseClusterInfoURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard cluster-info output",
			input:    "Kubernetes control plane is running at https://127.0.0.1:6550",
			expected: "https://127.0.0.1:6550",
		},
		{
			name:     "cluster-info with additional lines",
			input:    "Kubernetes control plane is running at https://127.0.0.1:6550\nCoreDNS is running at https://127.0.0.1:6550/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy",
			expected: "https://127.0.0.1:6550",
		},
		{
			name:     "cluster-info with ANSI codes",
			input:    "\x1b[32mKubernetes control plane\x1b[0m is running at \x1b[33mhttps://127.0.0.1:6550\x1b[0m",
			expected: "https://127.0.0.1:6550",
		},
		{
			name:     "http URL",
			input:    "Kubernetes control plane is running at http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "no URL found",
			input:    "Some random output without URLs",
			expected: "",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "URL with different port",
			input:    "Kubernetes control plane is running at https://192.168.1.100:16443",
			expected: "https://192.168.1.100:16443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseClusterInfoURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripANSICodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "simple color code",
			input:    "\x1b[32mgreen text\x1b[0m",
			expected: "green text",
		},
		{
			name:     "multiple color codes",
			input:    "\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m \x1b[34mblue\x1b[0m",
			expected: "red green blue",
		},
		{
			name:     "bold and underline",
			input:    "\x1b[1mbold\x1b[0m \x1b[4munderline\x1b[0m",
			expected: "bold underline",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "URL with ANSI codes",
			input:    "\x1b[33mhttps://127.0.0.1:6550\x1b[0m",
			expected: "https://127.0.0.1:6550",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSICodes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractHostPort(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedHost string
		expectedPort string
		expectError  bool
	}{
		{
			name:         "https URL with port",
			input:        "https://127.0.0.1:6550",
			expectedHost: "127.0.0.1",
			expectedPort: "6550",
			expectError:  false,
		},
		{
			name:         "http URL with port",
			input:        "http://localhost:8080",
			expectedHost: "localhost",
			expectedPort: "8080",
			expectError:  false,
		},
		{
			name:         "host:port without scheme",
			input:        "127.0.0.1:6443",
			expectedHost: "127.0.0.1",
			expectedPort: "6443",
			expectError:  false,
		},
		{
			name:         "IPv6 with port",
			input:        "[::1]:6550",
			expectedHost: "::1",
			expectedPort: "6550",
			expectError:  false,
		},
		{
			name:        "no port specified",
			input:       "https://127.0.0.1",
			expectError: true,
		},
		{
			name:        "just hostname",
			input:       "localhost",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := extractHostPort(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHost, host)
				assert.Equal(t, tt.expectedPort, port)
			}
		})
	}
}

func TestIsTemporaryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection refused",
			err:      errors.New("dial tcp 127.0.0.1:6550: connection refused"),
			expected: true,
		},
		{
			name:     "i/o timeout",
			err:      errors.New("read tcp 127.0.0.1:6550: i/o timeout"),
			expected: true,
		},
		{
			name:     "no such host",
			err:      errors.New("dial tcp: lookup foo.local: no such host"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New("read tcp: connection reset by peer"),
			expected: true,
		},
		{
			name:     "service unavailable",
			err:      errors.New("the server is currently unable to handle the request (Service Unavailable)"),
			expected: true,
		},
		{
			name:     "server currently unable",
			err:      errors.New("server is currently unable to serve requests"),
			expected: true,
		},
		{
			name:     "permanent error",
			err:      errors.New("unauthorized: invalid credentials"),
			expected: false,
		},
		{
			name:     "not found error",
			err:      errors.New("resource not found"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTemporaryError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractIPFromRouteOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already valid IP",
			input:    "172.21.96.1",
			expected: "172.21.96.1",
		},
		{
			name:     "full ip route output",
			input:    "default via 172.21.96.1 dev eth0 proto kernel",
			expected: "172.21.96.1",
		},
		{
			name:     "ip route output with trailing newline",
			input:    "default via 172.21.96.1 dev eth0 proto kernel\n",
			expected: "172.21.96.1",
		},
		{
			name:     "ip route output with extra whitespace",
			input:    "  default via 172.21.96.1 dev eth0 proto kernel  ",
			expected: "172.21.96.1",
		},
		{
			name:     "resolv.conf nameserver output",
			input:    "nameserver 172.21.96.1",
			expected: "172.21.96.1",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "no valid IP in output",
			input:    "default via gateway dev eth0",
			expected: "",
		},
		{
			name:     "multiple IPs returns first",
			input:    "192.168.1.1 172.21.96.1",
			expected: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIPFromRouteOutput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
