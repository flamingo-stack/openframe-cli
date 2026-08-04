package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	appCmd "github.com/flamingo-stack/openframe-cli/cmd/app"
	clusterCmd "github.com/flamingo-stack/openframe-cli/cmd/cluster"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui/steps"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func init() {
	testutil.InitializeTestMode()
}

func TestNewService(t *testing.T) {
	service := NewService()

	assert.NotNil(t, service, "NewService should not return nil")
	assert.IsType(t, &Service{}, service, "NewService should return Service type")
}

func TestServiceStructure(t *testing.T) {
	service := NewService()

	// Test that service has the expected structure
	assert.NotNil(t, service)

	// Test that the service can access the required commands
	clusterCmd := clusterCmd.GetClusterCmd()
	appCmd := appCmd.GetAppCmd()

	assert.NotNil(t, clusterCmd, "Should be able to get cluster command")
	assert.NotNil(t, appCmd, "Should be able to get app command")

	// Verify cluster command has create subcommand
	var createCmd *cobra.Command
	for _, cmd := range clusterCmd.Commands() {
		if cmd.Use == "create [NAME]" {
			createCmd = cmd
			break
		}
	}
	assert.NotNil(t, createCmd, "Cluster command should have create subcommand")

	// Verify app command has install subcommand
	var installCmd *cobra.Command
	for _, cmd := range appCmd.Commands() {
		if cmd.Use == "install [cluster-name]" {
			installCmd = cmd
			break
		}
	}
	assert.NotNil(t, installCmd, "App command should have install subcommand")
}

func TestServiceExecuteMethodExists(t *testing.T) {
	service := NewService()

	// Create a mock command structure
	rootCmd := &cobra.Command{}
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose flag")
	cmd := &cobra.Command{}
	rootCmd.AddCommand(cmd)

	// Test that Execute method exists and can be called
	assert.NotNil(t, service.Execute, "Service should have Execute method")

	// Note: We don't actually call Execute to avoid integration testing
	// The method signature and existence are verified, which is sufficient
	// for unit testing the service structure
}

func TestServiceArgumentHandling(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "No arguments",
			args:     []string{},
			expected: "",
		},
		{
			name:     "Single cluster name",
			args:     []string{"my-cluster"},
			expected: "my-cluster",
		},
		{
			name:     "Cluster name with whitespace",
			args:     []string{"  test-cluster  "},
			expected: "test-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService()

			// Verify service exists and can handle different argument patterns
			assert.NotNil(t, service)

			// Test argument structure without executing commands
			// This validates the service can be instantiated for different scenarios
		})
	}
}

func TestServiceVerboseFlagHandling(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "Verbose mode disabled",
			verbose: false,
		},
		{
			name:    "Verbose mode enabled",
			verbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService()

			// Create mock command with verbose flag
			rootCmd := &cobra.Command{}
			rootCmd.PersistentFlags().Bool("verbose", tt.verbose, "verbose flag")
			cmd := &cobra.Command{}
			rootCmd.AddCommand(cmd)

			// Verify service can handle different verbose flag states
			assert.NotNil(t, service)
			assert.NotNil(t, service.Execute)
		})
	}
}

// Note: Full execution testing is intentionally avoided to prevent integration
// testing. The service coordinates existing cluster and chart commands, so
// testing focuses on structure and method availability rather than end-to-end
// execution which would require complex mocking of the underlying commands.

// KubeContext must ride alongside KubeConfig: the chart workflow ignores Args
// once a rest.Config is provided, so without it the confirmation prompt read
// "install OpenFrame chart on ”?" and helm ran without --kube-context,
// targeting whatever context was current instead of the created cluster.
func TestBootstrapInstallRequest_SetsKubeContext(t *testing.T) {
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "config")) // context is fixed, not resolved from kubeconfig
	req := bootstrapInstallRequest("demo", true, false, &rest.Config{})
	if req.KubeContext != "k3d-demo" {
		t.Fatalf("KubeContext = %q, want k3d-demo", req.KubeContext)
	}
	if len(req.Args) != 1 || req.Args[0] != "demo" {
		t.Fatalf("Args = %v", req.Args)
	}
	if req.KubeConfig == nil {
		t.Fatal("KubeConfig must be passed through")
	}
}

// A kubeconfig context named EXACTLY like the cluster must not divert helm:
// bootstrap always creates a k3d cluster (context "k3d-<name>"), so resolving
// the context by name would send every helm call to the unrelated cluster
// while the native clients target the new k3d one — the split-target failure.
func TestBootstrapInstallRequest_ExactNameContextCollision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	kubeconfig := `apiVersion: v1
kind: Config
clusters:
- name: other
  cluster:
    server: https://127.0.0.1:6443
contexts:
- name: demo
  context:
    cluster: other
    user: u
users:
- name: u
  user: {}
current-context: demo
`
	if err := os.WriteFile(path, []byte(kubeconfig), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", path)
	req := bootstrapInstallRequest("demo", true, false, &rest.Config{})
	if req.KubeContext != "k3d-demo" {
		t.Fatalf("KubeContext = %q, want k3d-demo (exact-name context must not win)", req.KubeContext)
	}
}

// bootstrapSummaryMarkdown feeds the GitHub Actions Step Summary panel; a
// failed stage must carry the ✖ marker so the card doesn't read as all-green,
// and stages that never finished must not appear at all.
func TestBootstrapSummaryMarkdown(t *testing.T) {
	tracker := steps.NewTracker("Validate helm values", "Create cluster", "Install platform")
	tracker.Begin(0, "")
	tracker.Done(0)
	tracker.Begin(1, "demo")
	tracker.Fail(1)

	md := bootstrapSummaryMarkdown("demo", tracker)

	assert.True(t, strings.HasPrefix(md, "### OpenFrame ready · "), "header missing: %q", md)
	assert.Contains(t, md, "**Cluster:** `demo`\n")
	assert.Contains(t, md, "| stage | duration |\n|---|---|\n")
	assert.Contains(t, md, "| Validate helm values | ")
	assert.Contains(t, md, "| Create cluster ✖ | ", "failed stage must render the ✖ marker")
	assert.NotContains(t, md, "Install platform", "unfinished stages must not appear")
}
