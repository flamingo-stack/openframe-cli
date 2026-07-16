package argocd

import (
	"context"
	goruntime "runtime"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

// capturePatches records the body of every Application patch against the fake
// dynamic client, so a test can assert WHAT was patched, not just that a call
// happened.
func capturePatches(m *Manager) *[]string {
	var bodies []string
	dc := m.dynamicClient.(interface {
		PrependReactor(verb, resource string, fn k8stesting.ReactionFunc)
	})
	dc.PrependReactor("patch", "applications", func(action k8stesting.Action) (bool, runtime.Object, error) {
		bodies = append(bodies, string(action.(k8stesting.PatchAction).GetPatch()))
		return false, nil, nil // fall through to the default reactor
	})
	return &bodies
}

// TestHardRefreshApplications_PatchesHardNotNormal is the core of the fix: the
// annotation must be "hard" (re-fetch from git, bypass the manifest cache), not
// "normal" (re-compare against the cache that a just-restarted repo-server has
// lost). A normal refresh left Unknown apps stuck until the wait timed out.
func TestHardRefreshApplications_PatchesHardNotNormal(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	m := fakeManager(
		appObj("ingress-nginx", ArgoCDStatusUnknown, ArgoCDStatusUnknown),
		appObj("openframe-config", ArgoCDStatusUnknown, ArgoCDStatusUnknown),
	)
	bodies := capturePatches(m)

	n := m.hardRefreshApplications(context.Background(), []string{"ingress-nginx", "openframe-config"})
	if n != 2 {
		t.Fatalf("both apps must be refreshed, got %d", n)
	}
	if len(*bodies) != 2 {
		t.Fatalf("expected 2 patches, got %d", len(*bodies))
	}
	for _, b := range *bodies {
		if !strings.Contains(b, `"argocd.argoproj.io/refresh":"hard"`) {
			t.Errorf("patch must set a HARD refresh, got: %s", b)
		}
		if strings.Contains(b, `"normal"`) {
			t.Errorf("patch must NOT be a normal refresh, got: %s", b)
		}
	}
}

// TestHardRefreshApplications_SkipsEmptyNames: empty names are ignored (no
// patch against a nameless resource) and not counted.
func TestHardRefreshApplications_SkipsEmptyNames(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	m := fakeManager(appObj("real", ArgoCDStatusUnknown, ArgoCDStatusUnknown))
	bodies := capturePatches(m)

	n := m.hardRefreshApplications(context.Background(), []string{"", "real", ""})
	if n != 1 {
		t.Fatalf("only the named app may be refreshed, got %d", n)
	}
	if len(*bodies) != 1 {
		t.Errorf("empty names must not produce patches, got %d", len(*bodies))
	}
}

// TestHardRefreshApplications_NilDynamicClientIsNoOp: without a dynamic client
// (native init unavailable) the call is a safe no-op rather than a panic.
func TestHardRefreshApplications_NilDynamicClient(t *testing.T) {
	m := &Manager{}
	if n := m.hardRefreshApplications(context.Background(), []string{"a"}); n != 0 {
		t.Errorf("nil dynamic client must refresh nothing, got %d", n)
	}
}

// TestTriggerRepoServerRecovery_HardRefreshesTriggerApp guards that the recovery
// path itself uses a hard refresh (it previously hard-coded "normal"). The pod
// restart needs a live kube client, so this drives hardRefreshApplications —
// the exact helper the recovery now delegates to — and asserts the annotation.
func TestTriggerRepoServerRecovery_UsesHardRefreshHelper(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("native cluster ops are refused on Windows (must run inside WSL)")
	}
	m := fakeManager(appObj("argocd-apps", ArgoCDStatusUnknown, ArgoCDStatusUnknown))
	bodies := capturePatches(m)

	m.hardRefreshApplications(context.Background(), []string{"argocd-apps"})
	if len(*bodies) != 1 || !strings.Contains((*bodies)[0], `"hard"`) {
		t.Errorf("recovery refresh must be hard, got: %v", *bodies)
	}
}
