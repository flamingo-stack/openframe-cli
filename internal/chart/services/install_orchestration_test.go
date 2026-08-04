package services

import (
	"context"
	stderrors "errors"
	"os"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/config"
	"github.com/flamingo-stack/openframe-cli/internal/chart/utils/types"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/rest"
)

// These tests exercise the top-level install orchestration end to end —
// InstallWithContext → ExecuteWithContext → performInstallationWithRetry →
// performInstallation → Installer — through the collaborator seams on
// ChartService (installServices, newFileCleanup, installRetryPolicy), with the
// REAL Installer, retry executor, and FileCleanup in the loop. Only the
// cluster-touching leaves (ArgoCD install/wait, app-of-apps install) are faked.

// spyFileCleanup wraps the real FileCleanup so tests can assert WHICH cleanup
// path ran (forced RestoreFiles on error vs RestoreFilesOnSuccess) while
// keeping the real file-system semantics — the registered temp values file
// must actually disappear.
type spyFileCleanup struct {
	real            *files.FileCleanup
	calls           []string
	registeredPaths []string
	successOnly     bool
}

func (s *spyFileCleanup) RegisterTempFile(filePath string) error {
	s.calls = append(s.calls, "RegisterTempFile")
	s.registeredPaths = append(s.registeredPaths, filePath)
	return s.real.RegisterTempFile(filePath)
}

func (s *spyFileCleanup) SetCleanupOnSuccessOnly(enabled bool) {
	s.calls = append(s.calls, "SetCleanupOnSuccessOnly")
	s.successOnly = enabled
	s.real.SetCleanupOnSuccessOnly(enabled)
}

func (s *spyFileCleanup) RestoreFiles(verbose bool) error {
	s.calls = append(s.calls, "RestoreFiles")
	return s.real.RestoreFiles(verbose)
}

func (s *spyFileCleanup) RestoreFilesOnSuccess(verbose bool) error {
	s.calls = append(s.calls, "RestoreFilesOnSuccess")
	return s.real.RestoreFilesOnSuccess(verbose)
}

func (s *spyFileCleanup) called(name string) bool {
	for _, c := range s.calls {
		if c == name {
			return true
		}
	}
	return false
}

// orchestrationHarness wires a real ChartService (real HelmManager, real
// Installer, real retry executor) with faked install collaborators.
type orchestrationHarness struct {
	svc        *ChartService
	argoCD     *MockArgoCDService
	appOfApps  *MockAppOfAppsService
	cleanup    *spyFileCleanup
	order      []string
	installCfg *config.ChartInstallConfig // config seen by the collaborator factory
}

// step returns a mock Run-callback recording invocation order across fakes.
func (h *orchestrationHarness) step(name string) func(mock.Arguments) {
	return func(mock.Arguments) { h.order = append(h.order, name) }
}

func newOrchestrationHarness(t *testing.T) *orchestrationHarness {
	t.Helper()
	t.Chdir(t.TempDir()) // no stray openframe-helm-values.yaml

	svc, err := NewChartService(NewMockClusterLister(), &rest.Config{Host: "https://127.0.0.1:1"}, false, false)
	if err != nil {
		t.Fatalf("NewChartService: %v", err)
	}

	h := &orchestrationHarness{
		svc:       svc,
		argoCD:    new(MockArgoCDService),
		appOfApps: new(MockAppOfAppsService),
		cleanup:   &spyFileCleanup{real: files.NewFileCleanup()},
	}
	svc.newFileCleanup = func() installFileCleanup { return h.cleanup }
	svc.installServices = func(_ *ChartService, cfg config.ChartInstallConfig) (types.ArgoCDService, types.AppOfAppsService, error) {
		h.installCfg = &cfg
		return h.argoCD, h.appOfApps, nil
	}
	// Production classification rules (same ExponentialBackoffPolicy type and
	// transient-substring fallback as InstallationRetryPolicy), but with
	// millisecond delays so a retried attempt doesn't stall the test 10s.
	svc.installRetryPolicy = sharedErrors.NewExponentialBackoffPolicy(3, time.Millisecond)
	return h
}

// installRequest is the canonical explicit-target request: KubeConfig from
// --context IS the install target (no cluster selection), non-interactive
// (no prompts), with app-of-apps configured so the wait step runs.
func installRequest() types.InstallationRequest {
	return types.InstallationRequest{
		NonInteractive: true,
		GitHubRepo:     "https://github.com/test/repo",
		GitHubBranch:   "main",
		KubeConfig:     &rest.Config{Host: "https://127.0.0.1:1"},
		KubeContext:    "k3d-target",
	}
}

// requireTempValuesGone asserts the workflow registered a temp values file and
// that it no longer exists on disk.
func requireTempValuesGone(t *testing.T, h *orchestrationHarness) {
	t.Helper()
	if len(h.cleanup.registeredPaths) == 0 {
		t.Fatal("workflow must register the temp values file for cleanup")
	}
	for _, p := range h.cleanup.registeredPaths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("temp values file %s must be removed (stat err: %v)", p, err)
		}
	}
}

func TestInstallWithContext_HappyPathOrderAndCleanup(t *testing.T) {
	h := newOrchestrationHarness(t)
	h.argoCD.On("Install", mock.Anything, mock.Anything).Run(h.step("argocd.Install")).Return(nil)
	h.appOfApps.On("Install", mock.Anything, mock.Anything).Run(h.step("appofapps.Install")).Return(nil)
	h.argoCD.On("WaitForApplications", mock.Anything, mock.Anything).Run(h.step("argocd.Wait")).Return(nil)

	err := h.svc.InstallWithContext(context.Background(), installRequest())
	assert.NoError(t, err)

	// The three-step orchestration runs exactly once each, in order.
	assert.Equal(t, []string{"argocd.Install", "appofapps.Install", "argocd.Wait"}, h.order)
	h.argoCD.AssertExpectations(t)
	h.appOfApps.AssertExpectations(t)

	// F4 one-target rule: the explicit kube-context reaches the collaborators'
	// config, and no cluster name is invented alongside it.
	if assert.NotNil(t, h.installCfg) {
		assert.Equal(t, "k3d-target", h.installCfg.KubeContext)
		assert.Empty(t, h.installCfg.ClusterName)
		assert.True(t, h.installCfg.HasAppOfApps())
	}

	// Cleanup-on-success semantics: temp files are cleaned only after success
	// (SetCleanupOnSuccessOnly(true)), via the success path — never the forced
	// error-path restore.
	assert.True(t, h.cleanup.successOnly, "InstallWithContext must set cleanup-on-success-only")
	assert.True(t, h.cleanup.called("RestoreFilesOnSuccess"), "success must clean temp files via RestoreFilesOnSuccess")
	assert.False(t, h.cleanup.called("RestoreFiles"), "forced restore must not run on success")
	requireTempValuesGone(t, h)
}

func TestInstallWithContext_WaitFailureRestoresFilesAndNeverReinstalls(t *testing.T) {
	h := newOrchestrationHarness(t)
	h.argoCD.On("Install", mock.Anything, mock.Anything).Run(h.step("argocd.Install")).Return(nil)
	h.appOfApps.On("Install", mock.Anything, mock.Anything).Run(h.step("appofapps.Install")).Return(nil)
	// The wait diagnostics embed pod logs whose text would pattern-match the
	// retry policy's transient table; the installer's explicit never-retry
	// marker must win, or a wait failure would reinstall ArgoCD.
	h.argoCD.On("WaitForApplications", mock.Anything, mock.Anything).Run(h.step("argocd.Wait")).
		Return(stderrors.New("applications not ready: dial tcp 10.0.0.1:8080: connection refused"))

	err := h.svc.InstallWithContext(context.Background(), installRequest())

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "waiting failed for ArgoCD applications")
	}
	// Exactly one pass through the pipeline: transient-looking wait text must
	// not re-trigger the (already completed) install steps.
	assert.Equal(t, []string{"argocd.Install", "appofapps.Install", "argocd.Wait"}, h.order)

	// Cleanup-on-error semantics: the forced restore runs and removes the temp
	// values file even though cleanup was configured as success-only.
	assert.True(t, h.cleanup.called("RestoreFiles"), "error path must force-restore files")
	assert.False(t, h.cleanup.called("RestoreFilesOnSuccess"))
	requireTempValuesGone(t, h)
}

func TestInstallWithContext_ArgoCDFailureSkipsLaterSteps(t *testing.T) {
	h := newOrchestrationHarness(t)
	// Non-transient failure text: no retry, and the pipeline stops at step one.
	h.argoCD.On("Install", mock.Anything, mock.Anything).Run(h.step("argocd.Install")).
		Return(stderrors.New("values don't meet the specifications of the schema"))

	err := h.svc.InstallWithContext(context.Background(), installRequest())

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "ArgoCD")
	}
	assert.Equal(t, []string{"argocd.Install"}, h.order, "app-of-apps and wait must not be attempted")
	h.appOfApps.AssertNotCalled(t, "Install", mock.Anything, mock.Anything)
	h.argoCD.AssertNotCalled(t, "WaitForApplications", mock.Anything, mock.Anything)

	assert.True(t, h.cleanup.called("RestoreFiles"), "error path must force-restore files")
	requireTempValuesGone(t, h)
}

func TestInstallWithContext_TransientFailureIsRetried(t *testing.T) {
	h := newOrchestrationHarness(t)
	// First attempt fails with a transient failure (classified retryable by the
	// policy's substring table for shelled-out tools); the retry succeeds.
	h.argoCD.On("Install", mock.Anything, mock.Anything).Run(h.step("argocd.Install")).
		Return(stderrors.New("Kubernetes cluster unreachable: dial tcp 127.0.0.1:6443: connection refused")).Once()
	h.argoCD.On("Install", mock.Anything, mock.Anything).Run(h.step("argocd.Install")).Return(nil)
	h.appOfApps.On("Install", mock.Anything, mock.Anything).Run(h.step("appofapps.Install")).Return(nil)
	h.argoCD.On("WaitForApplications", mock.Anything, mock.Anything).Run(h.step("argocd.Wait")).Return(nil)

	err := h.svc.InstallWithContext(context.Background(), installRequest())
	assert.NoError(t, err)

	assert.Equal(t, []string{"argocd.Install", "argocd.Install", "appofapps.Install", "argocd.Wait"}, h.order)
	assert.True(t, h.cleanup.called("RestoreFilesOnSuccess"))
	requireTempValuesGone(t, h)
}

func TestInstallWithContext_NonRetryableFailureIsNotRetried(t *testing.T) {
	h := newOrchestrationHarness(t)
	h.argoCD.On("Install", mock.Anything, mock.Anything).Run(h.step("argocd.Install")).
		Return(stderrors.New("another operation (install/upgrade/rollback) is in progress"))

	err := h.svc.InstallWithContext(context.Background(), installRequest())

	assert.Error(t, err)
	// "another operation is in progress" was deliberately REMOVED from the
	// retryable set: it means a release wedged in pending-* by a killed run,
	// and retrying hits the same wedged release every time.
	assert.Equal(t, []string{"argocd.Install"}, h.order, "non-retryable failures must run exactly once")
	requireTempValuesGone(t, h)
}

func TestInstallWithContext_CollaboratorFactoryErrorPropagates(t *testing.T) {
	h := newOrchestrationHarness(t)
	factoryErr := stderrors.New("no route to target cluster")
	h.svc.installServices = func(_ *ChartService, _ config.ChartInstallConfig) (types.ArgoCDService, types.AppOfAppsService, error) {
		return nil, nil, factoryErr
	}

	err := h.svc.InstallWithContext(context.Background(), installRequest())

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to create ArgoCD service for the install target")
		assert.True(t, stderrors.Is(err, factoryErr))
	}
	assert.True(t, h.cleanup.called("RestoreFiles"), "error path must force-restore files")
	requireTempValuesGone(t, h)
}

// TestInstallWithContextDeferred_ResolvesHelmManagerThenInstalls covers the
// deferred entry point: the service starts without a HelmManager (standalone
// install) and the workflow initializes it from the request's rest.Config
// before the collaborators are built.
func TestInstallWithContextDeferred_ResolvesHelmManagerThenInstalls(t *testing.T) {
	t.Chdir(t.TempDir())

	svc, err := NewChartServiceDeferred(NewMockClusterLister(), false, false)
	if err != nil {
		t.Fatalf("NewChartServiceDeferred: %v", err)
	}
	h := &orchestrationHarness{
		svc:       svc,
		argoCD:    new(MockArgoCDService),
		appOfApps: new(MockAppOfAppsService),
		cleanup:   &spyFileCleanup{real: files.NewFileCleanup()},
	}
	svc.newFileCleanup = func() installFileCleanup { return h.cleanup }
	svc.installServices = func(cs *ChartService, cfg config.ChartInstallConfig) (types.ArgoCDService, types.AppOfAppsService, error) {
		// By collaborator-construction time the deferred HelmManager must exist.
		assert.NotNil(t, cs.helmManager, "deferred HelmManager must be initialized before collaborators are built")
		return h.argoCD, h.appOfApps, nil
	}
	svc.installRetryPolicy = sharedErrors.NewExponentialBackoffPolicy(3, time.Millisecond)

	h.argoCD.On("Install", mock.Anything, mock.Anything).Run(h.step("argocd.Install")).Return(nil)
	h.appOfApps.On("Install", mock.Anything, mock.Anything).Run(h.step("appofapps.Install")).Return(nil)
	h.argoCD.On("WaitForApplications", mock.Anything, mock.Anything).Run(h.step("argocd.Wait")).Return(nil)

	err = svc.InstallWithContextDeferred(context.Background(), installRequest())
	assert.NoError(t, err)
	assert.Equal(t, []string{"argocd.Install", "appofapps.Install", "argocd.Wait"}, h.order)
	requireTempValuesGone(t, h)
}

// TestChartService_DefaultSeams pins the production fallbacks: with nothing
// injected, the seams yield the real FileCleanup, the real per-target
// collaborators, and the installation retry policy.
func TestChartService_DefaultSeams(t *testing.T) {
	svc, err := NewChartService(NewMockClusterLister(), &rest.Config{Host: "https://127.0.0.1:1"}, false, false)
	if err != nil {
		t.Fatalf("NewChartService: %v", err)
	}
	assert.NotNil(t, svc.helmManager)
	assert.NotNil(t, svc.kubeConfig)

	_, isReal := svc.fileCleanupOrDefault().(*files.FileCleanup)
	assert.True(t, isReal, "default file cleanup must be the real FileCleanup")

	policy := svc.installRetryPolicyOrDefault()
	assert.Equal(t, sharedErrors.InstallationRetryPolicy().GetMaxAttempts(), policy.GetMaxAttempts())

	argoCDService, appOfAppsService, err := svc.installServicesOrDefault(config.ChartInstallConfig{ClusterName: "test-cluster"})
	assert.NoError(t, err)
	_, isRealArgoCD := argoCDService.(*ArgoCD)
	assert.True(t, isRealArgoCD, "default ArgoCD service must be the real implementation")
	_, isRealAppOfApps := appOfAppsService.(*AppOfApps)
	assert.True(t, isRealAppOfApps, "default app-of-apps service must be the real implementation")
}
