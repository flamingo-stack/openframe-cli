// Package status aggregates OpenFrame platform status (cluster health, ArgoCD
// application sync/health, and admin access) for the `openframe app status` and
// `openframe app access` commands.
package status

import (
	"context"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
)

// Lister lists the ArgoCD applications in the target cluster.
type Lister interface {
	ListApplications(ctx context.Context, verbose bool) ([]argocd.Application, error)
}

// PasswordReader reads the initial ArgoCD admin password.
type PasswordReader interface {
	AdminPassword(ctx context.Context) (string, error)
}

// HealthChecker reports cluster reachability/readiness.
type HealthChecker interface {
	CheckHealth(ctx context.Context) (k8s.Health, error)
}

// Report is the aggregated platform status.
type Report struct {
	Health        k8s.Health
	HealthErr     error
	Apps          []argocd.Application
	Total         int
	Synced        int
	Healthy       int
	AdminPassword string // empty when unavailable
}

// Ready reports whether the platform is fully up: the cluster is reachable and
// every application is both Synced and Healthy.
func (r Report) Ready() bool {
	return r.Health.Reachable && r.Total > 0 && r.Synced == r.Total && r.Healthy == r.Total
}

// Summary is a one-line, human-readable status suitable for logs or a header.
func (r Report) Summary() string {
	if !r.Health.Reachable {
		return "cluster unreachable"
	}
	if r.Total == 0 {
		return "cluster reachable, no OpenFrame applications found (not installed yet?)"
	}
	state := "NOT READY"
	if r.Ready() {
		state = "READY"
	}
	return fmt.Sprintf("%d/%d synced, %d/%d healthy — %s", r.Synced, r.Total, r.Healthy, r.Total, state)
}

// Service aggregates platform status from its injected sources.
type Service struct {
	lister   Lister
	health   HealthChecker
	password PasswordReader
}

// NewService wires a status service. health and password may be nil (their
// contributions are then skipped); lister is required.
func NewService(lister Lister, health HealthChecker, password PasswordReader) *Service {
	return &Service{lister: lister, health: health, password: password}
}

// Report gathers cluster health, application status, and admin access. Health
// and password are best-effort — their errors are surfaced (HealthErr) or
// swallowed (password) but do not fail the report. A hard error is returned
// only when the applications cannot be listed.
func (s *Service) Report(ctx context.Context, verbose bool) (Report, error) {
	var rep Report

	if s.health != nil {
		rep.Health, rep.HealthErr = s.health.CheckHealth(ctx)
	}

	apps, err := s.lister.ListApplications(ctx, verbose)
	if err != nil {
		return rep, err
	}
	rep.Apps = apps
	rep.Total, rep.Synced, rep.Healthy = summarize(apps)

	if s.password != nil {
		if pw, perr := s.password.AdminPassword(ctx); perr == nil {
			rep.AdminPassword = pw
		}
	}

	return rep, nil
}

// summarize counts total, synced and healthy applications.
func summarize(apps []argocd.Application) (total, synced, healthy int) {
	total = len(apps)
	for _, a := range apps {
		if a.Sync == argocd.ArgoCDSyncSynced {
			synced++
		}
		if a.Health == argocd.ArgoCDHealthHealthy {
			healthy++
		}
	}
	return total, synced, healthy
}
