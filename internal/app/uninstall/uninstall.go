// Package uninstall removes the OpenFrame platform (ArgoCD + app-of-apps) from a
// cluster without deleting the cluster itself. It backs `openframe app uninstall`.
package uninstall

import (
	"context"
	"fmt"
)

// namespace is where ArgoCD and the app-of-apps live.
const namespace = "argocd"

// releases lists the Helm releases to remove, ordered so the app-of-apps release
// (which owns the child Applications) goes before ArgoCD itself.
var releases = []string{"app-of-apps", "argo-cd"}

// ApplicationDeleter deletes ArgoCD Application CRs (cascading their workloads).
type ApplicationDeleter interface {
	DeleteApplications(ctx context.Context) (int, error)
}

// ReleaseUninstaller removes a Helm release.
type ReleaseUninstaller interface {
	UninstallRelease(ctx context.Context, releaseName, namespace, kubeContext string) error
}

// NamespaceDeleter deletes a namespace.
type NamespaceDeleter interface {
	DeleteNamespace(ctx context.Context, name string) error
}

// Options controls optional uninstall behavior.
type Options struct {
	DeleteNamespace bool
}

// Result records what was removed.
type Result struct {
	AppsDeleted      int
	ReleasesRemoved  []string
	NamespaceDeleted bool
}

// Service orchestrates a platform uninstall.
type Service struct {
	apps        ApplicationDeleter
	helm        ReleaseUninstaller
	ns          NamespaceDeleter
	kubeContext string
}

// NewService wires an uninstall service bound to a kube-context (empty = current).
func NewService(apps ApplicationDeleter, helm ReleaseUninstaller, ns NamespaceDeleter, kubeContext string) *Service {
	return &Service{apps: apps, helm: helm, ns: ns, kubeContext: kubeContext}
}

// Uninstall removes the platform in a safe order: delete the ArgoCD Applications
// first (so ArgoCD cascades workload cleanup while it is still running), then
// uninstall the Helm releases (app-of-apps before ArgoCD), then optionally delete
// the argocd namespace. It stops at the first hard error.
func (s *Service) Uninstall(ctx context.Context, opts Options) (Result, error) {
	var res Result

	deleted, err := s.apps.DeleteApplications(ctx)
	res.AppsDeleted = deleted
	if err != nil {
		return res, fmt.Errorf("removing ArgoCD applications: %w", err)
	}

	for _, rel := range releases {
		if err := s.helm.UninstallRelease(ctx, rel, namespace, s.kubeContext); err != nil {
			return res, err
		}
		res.ReleasesRemoved = append(res.ReleasesRemoved, rel)
	}

	if opts.DeleteNamespace {
		if s.ns == nil {
			return res, fmt.Errorf("namespace deletion requested but no namespace deleter is configured")
		}
		if err := s.ns.DeleteNamespace(ctx, namespace); err != nil {
			return res, err
		}
		res.NamespaceDeleted = true
	}

	return res, nil
}
