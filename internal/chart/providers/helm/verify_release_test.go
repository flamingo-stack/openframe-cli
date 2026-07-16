package helm

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

func newManagerWithMock() (*HelmManager, *executor.MockCommandExecutor) {
	mock := executor.NewMockCommandExecutor()
	// nil rest.Config → minimal manager; verifyHelmRelease only needs the executor.
	m, _ := NewHelmManager(mock, nil, false)
	return m, mock
}

func TestVerifyHelmRelease_Found(t *testing.T) {
	m, mock := newManagerWithMock()
	mock.SetResponse("helm list", &executor.CommandResult{Stdout: `[{"name":"argo-cd","status":"deployed"}]`})
	mock.SetResponse("helm status", &executor.CommandResult{Stdout: "STATUS: deployed"})

	if err := m.verifyHelmRelease(context.Background(), "argo-cd", "argocd", "demo", false); err != nil {
		t.Fatalf("expected success: %v", err)
	}

	// The cluster name must pin the kube-context on both helm calls.
	joined := strings.Join(mock.GetExecutedCommands(), " | ")
	if !strings.Contains(joined, "--kube-context k3d-demo") {
		t.Errorf("expected --kube-context k3d-demo in commands: %s", joined)
	}
}

func TestVerifyHelmRelease_EmptyListMeansNotFound(t *testing.T) {
	m, mock := newManagerWithMock()
	mock.SetResponse("helm list", &executor.CommandResult{Stdout: "[]"})

	err := m.verifyHelmRelease(context.Background(), "argo-cd", "argocd", "", false)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected a not-found error, got %v", err)
	}
}

func TestVerifyHelmRelease_ListFailure(t *testing.T) {
	m, mock := newManagerWithMock()
	mock.SetShouldFail(true, "helm exploded")

	if err := m.verifyHelmRelease(context.Background(), "argo-cd", "argocd", "", false); err == nil {
		t.Fatal("expected an error when helm list fails")
	}
}

func TestVerifyHelmRelease_StatusFailureAfterFound(t *testing.T) {
	m, mock := newManagerWithMock()
	mock.SetResponse("helm list", &executor.CommandResult{Stdout: `[{"name":"argo-cd"}]`})
	mock.SetResponse("helm status", &executor.CommandResult{Stdout: "", ExitCode: 1})

	err := m.verifyHelmRelease(context.Background(), "argo-cd", "argocd", "", false)
	if err == nil || !strings.Contains(err.Error(), "status check failed") {
		t.Fatalf("expected a status-check error, got %v", err)
	}
}
