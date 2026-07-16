package uninstall

import (
	"context"
	"errors"
	"testing"
)

type fakeApps struct {
	n          int
	err        error
	finalizers int
	finErr     error
	finCalled  bool
}

func (f *fakeApps) DeleteApplications(context.Context) (int, error) { return f.n, f.err }

func (f *fakeApps) RemoveApplicationFinalizers(context.Context) (int, error) {
	f.finCalled = true
	return f.finalizers, f.finErr
}

type fakeHelm struct {
	calls   []string // "release@context"
	failOn  string   // release name to fail on
	callErr error
}

func (f *fakeHelm) UninstallRelease(_ context.Context, releaseName, _, kubeContext string) error {
	f.calls = append(f.calls, releaseName+"@"+kubeContext)
	if releaseName == f.failOn {
		return f.callErr
	}
	return nil
}

type fakeNS struct {
	deleted string
	err     error
}

func (f *fakeNS) DeleteNamespace(_ context.Context, name string) error {
	if f.err != nil {
		return f.err
	}
	f.deleted = name
	return nil
}

func TestUninstall_HappyPath(t *testing.T) {
	apps := &fakeApps{n: 5, finalizers: 3}
	helm := &fakeHelm{}
	svc := NewService(apps, helm, &fakeNS{}, "k3d-demo")

	res, err := svc.Uninstall(context.Background(), Options{})
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if res.AppsDeleted != 5 {
		t.Fatalf("AppsDeleted = %d, want 5", res.AppsDeleted)
	}
	// Releases removed in order, app-of-apps before argo-cd, with the context.
	want := []string{"app-of-apps@k3d-demo", "argo-cd@k3d-demo"}
	if len(helm.calls) != 2 || helm.calls[0] != want[0] || helm.calls[1] != want[1] {
		t.Fatalf("helm calls = %v, want %v", helm.calls, want)
	}
	// Leftover finalizers are stripped after the releases are removed.
	if !apps.finCalled {
		t.Fatal("RemoveApplicationFinalizers must be called")
	}
	if res.FinalizersCleared != 3 {
		t.Fatalf("FinalizersCleared = %d, want 3", res.FinalizersCleared)
	}
	if res.NamespaceDeleted {
		t.Fatal("namespace should not be deleted without the option")
	}
}

// A finalizer-removal failure surfaces, and it must run only after the Helm
// releases are gone (ArgoCD must be removed before we strip its finalizers).
func TestUninstall_FinalizerErrorSurfacesAfterHelm(t *testing.T) {
	helm := &fakeHelm{}
	apps := &fakeApps{n: 1, finErr: errors.New("patch failed")}
	svc := NewService(apps, helm, &fakeNS{}, "k3d-demo")

	_, err := svc.Uninstall(context.Background(), Options{})
	if err == nil {
		t.Fatal("expected the finalizer error to surface")
	}
	if len(helm.calls) != 2 {
		t.Fatalf("both releases must be uninstalled before finalizer removal, got %v", helm.calls)
	}
	if !apps.finCalled {
		t.Fatal("finalizer removal must have been attempted")
	}
}

func TestUninstall_DeletesNamespaceWhenRequested(t *testing.T) {
	ns := &fakeNS{}
	svc := NewService(&fakeApps{}, &fakeHelm{}, ns, "")

	res, err := svc.Uninstall(context.Background(), Options{DeleteNamespace: true})
	if err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if !res.NamespaceDeleted || ns.deleted != "argocd" {
		t.Fatalf("expected argocd namespace deleted, got %q (%v)", ns.deleted, res.NamespaceDeleted)
	}
}

func TestUninstall_AppDeleteErrorStopsBeforeHelm(t *testing.T) {
	helm := &fakeHelm{}
	svc := NewService(&fakeApps{n: 2, err: errors.New("boom")}, helm, &fakeNS{}, "")

	res, err := svc.Uninstall(context.Background(), Options{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if len(helm.calls) != 0 {
		t.Fatalf("helm should not run after an app-delete failure, got %v", helm.calls)
	}
	if res.AppsDeleted != 2 {
		t.Fatalf("AppsDeleted = %d, want 2 (count is reported even on error)", res.AppsDeleted)
	}
}

func TestUninstall_ReleaseErrorStops(t *testing.T) {
	helm := &fakeHelm{failOn: "argo-cd", callErr: errors.New("helm failed")}
	svc := NewService(&fakeApps{}, helm, &fakeNS{}, "")

	res, err := svc.Uninstall(context.Background(), Options{DeleteNamespace: true})
	if err == nil {
		t.Fatal("expected an error")
	}
	// app-of-apps removed, argo-cd attempted and failed → namespace not touched.
	if len(res.ReleasesRemoved) != 1 || res.ReleasesRemoved[0] != "app-of-apps" {
		t.Fatalf("ReleasesRemoved = %v, want [app-of-apps]", res.ReleasesRemoved)
	}
	if res.NamespaceDeleted {
		t.Fatal("namespace must not be deleted after a release failure")
	}
}

func TestUninstall_NamespaceRequestedWithoutDeleter(t *testing.T) {
	svc := NewService(&fakeApps{}, &fakeHelm{}, nil, "")
	if _, err := svc.Uninstall(context.Background(), Options{DeleteNamespace: true}); err == nil {
		t.Fatal("expected an error when no namespace deleter is configured")
	}
}
