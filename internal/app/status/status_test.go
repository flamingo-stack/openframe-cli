package status

import (
	"context"
	"errors"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
)

type fakeLister struct {
	apps []argocd.Application
	err  error
}

func (f fakeLister) ListApplications(context.Context, bool) ([]argocd.Application, error) {
	return f.apps, f.err
}

type fakeHealth struct {
	h   k8s.Health
	err error
}

func (f fakeHealth) CheckHealth(context.Context) (k8s.Health, error) { return f.h, f.err }

type fakePassword struct {
	pw  string
	err error
}

func (f fakePassword) AdminPassword(context.Context) (string, error) { return f.pw, f.err }

func app(name, health, sync string) argocd.Application {
	return argocd.Application{Name: name, Health: health, Sync: sync}
}

func TestSummarize(t *testing.T) {
	total, synced, healthy := summarize([]argocd.Application{
		app("a", "Healthy", "Synced"),
		app("b", "Healthy", "OutOfSync"),
		app("c", "Degraded", "Synced"),
	})
	if total != 3 || synced != 2 || healthy != 2 {
		t.Fatalf("summarize = (%d,%d,%d), want (3,2,2)", total, synced, healthy)
	}
}

func TestReport_ReadyWhenAllSyncedHealthy(t *testing.T) {
	svc := NewService(
		fakeLister{apps: []argocd.Application{app("a", "Healthy", "Synced"), app("b", "Healthy", "Synced")}},
		fakeHealth{h: k8s.Health{Reachable: true, NodesReady: 1}},
		fakePassword{pw: "s3cret"},
	)
	rep, err := svc.Report(context.Background(), false)
	if err != nil {
		t.Fatalf("Report: %v", err)
	}
	if !rep.Ready() {
		t.Fatalf("expected Ready, got summary %q", rep.Summary())
	}
	if rep.Total != 2 || rep.Synced != 2 || rep.Healthy != 2 {
		t.Fatalf("counts = (%d,%d,%d)", rep.Total, rep.Synced, rep.Healthy)
	}
	if rep.AdminPassword != "s3cret" {
		t.Fatalf("AdminPassword = %q", rep.AdminPassword)
	}
	if rep.Summary() != "2/2 synced, 2/2 healthy — READY" {
		t.Fatalf("Summary = %q", rep.Summary())
	}
}

func TestReport_NotReadyWhenPartial(t *testing.T) {
	svc := NewService(
		fakeLister{apps: []argocd.Application{app("a", "Healthy", "Synced"), app("b", "Progressing", "OutOfSync")}},
		fakeHealth{h: k8s.Health{Reachable: true, NodesReady: 1}},
		nil,
	)
	rep, _ := svc.Report(context.Background(), false)
	if rep.Ready() {
		t.Fatal("expected NOT ready")
	}
	if rep.Summary() != "1/2 synced, 1/2 healthy — NOT READY" {
		t.Fatalf("Summary = %q", rep.Summary())
	}
}

func TestReport_ListErrorIsFatal(t *testing.T) {
	svc := NewService(fakeLister{err: errors.New("boom")}, nil, nil)
	if _, err := svc.Report(context.Background(), false); err == nil {
		t.Fatal("expected the list error to propagate")
	}
}

func TestReport_PasswordAndHealthAreBestEffort(t *testing.T) {
	svc := NewService(
		fakeLister{apps: []argocd.Application{app("a", "Healthy", "Synced")}},
		fakeHealth{h: k8s.Health{Reachable: true, NodesReady: 1}, err: errors.New("partial health")},
		fakePassword{err: errors.New("no secret")},
	)
	rep, err := svc.Report(context.Background(), false)
	if err != nil {
		t.Fatalf("password/health errors must not fail the report: %v", err)
	}
	if rep.HealthErr == nil {
		t.Fatal("expected HealthErr to be surfaced")
	}
	if rep.AdminPassword != "" {
		t.Fatalf("AdminPassword should be empty on error, got %q", rep.AdminPassword)
	}
}

func TestReport_UnreachableSummary(t *testing.T) {
	svc := NewService(fakeLister{}, fakeHealth{h: k8s.Health{Reachable: false}}, nil)
	rep, _ := svc.Report(context.Background(), false)
	if rep.Summary() != "cluster unreachable" {
		t.Fatalf("Summary = %q", rep.Summary())
	}
}

func TestReport_NoAppsSummary(t *testing.T) {
	svc := NewService(fakeLister{apps: nil}, fakeHealth{h: k8s.Health{Reachable: true, NodesReady: 1}}, nil)
	rep, _ := svc.Report(context.Background(), false)
	if rep.Ready() {
		t.Fatal("no apps must not be Ready")
	}
	if got := rep.Summary(); got != "cluster reachable, no OpenFrame applications found (not installed yet?)" {
		t.Fatalf("Summary = %q", got)
	}
}
