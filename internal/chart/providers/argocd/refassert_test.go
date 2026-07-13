package argocd

import (
	"strings"
	"testing"
)

const ossRepo = "https://github.com/flamingo-stack/openframe-oss-tenant"

// TestVerifyRefPinning_LegacyBranchSilentlyDeploysMain is the V3 case: the user
// asked for a legacy branch, its chart ignored repository.branch, and ArgoCD is
// tracking main. That must be reported — the CLI otherwise prints SUCCESS for a
// deployment of the wrong ref.
func TestVerifyRefPinning_LegacyBranchSilentlyDeploysMain(t *testing.T) {
	apps := []Application{
		{Name: "openframe-api", RepoURL: ossRepo, TargetRevision: "main"},
		{Name: "openframe-ui", RepoURL: ossRepo, TargetRevision: "main"},
	}
	mm := verifyRefPinning(apps, ossRepo, "feature/configuration-updates")
	if len(mm) != 2 {
		t.Fatalf("both children on main must be flagged, got %d: %+v", len(mm), mm)
	}
	err := refMismatchError("feature/configuration-updates", mm)
	for _, want := range []string{"predates ref pinning", "openframe-api", `on "main"`, "feature/configuration-updates"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error must contain %q; got:\n%s", want, err.Error())
		}
	}
}

// TestVerifyRefPinning_ModernBranchMatches: a chart that honours the pin sets
// children's targetRevision to the requested ref — no mismatch.
func TestVerifyRefPinning_ModernBranchMatches(t *testing.T) {
	apps := []Application{
		{Name: "openframe-api", RepoURL: ossRepo, TargetRevision: "feature/x"},
		{Name: "openframe-ui", RepoURL: ossRepo + ".git", TargetRevision: "feature/x"},
	}
	if mm := verifyRefPinning(apps, ossRepo, "feature/x"); len(mm) != 0 {
		t.Errorf("matching refs must not be flagged, got %+v", mm)
	}
}

// TestVerifyRefPinning_DefaultRefSkipped: requesting main (or empty) is
// indistinguishable from no pinning; do not flag.
func TestVerifyRefPinning_DefaultRefSkipped(t *testing.T) {
	apps := []Application{{Name: "a", RepoURL: ossRepo, TargetRevision: "main"}}
	for _, ref := range []string{"", "main", "master", "HEAD"} {
		if mm := verifyRefPinning(apps, ossRepo, ref); len(mm) != 0 {
			t.Errorf("default ref %q must be skipped, got %+v", ref, mm)
		}
	}
}

// TestVerifyRefPinning_ForeignRepoIgnored: a child sourcing a different repo has
// its own revision and must not be judged against the OSS ref.
func TestVerifyRefPinning_ForeignRepoIgnored(t *testing.T) {
	apps := []Application{
		{Name: "third-party", RepoURL: "https://charts.example.com/foo", TargetRevision: "1.2.3"},
		{Name: "openframe-api", RepoURL: ossRepo, TargetRevision: "feature/x"},
	}
	if mm := verifyRefPinning(apps, ossRepo, "feature/x"); len(mm) != 0 {
		t.Errorf("foreign-repo child must be ignored, got %+v", mm)
	}
}

// TestVerifyRefPinning_EmptyRevisionSkipped: a child with no declared revision
// is unknowable, not a mismatch.
func TestVerifyRefPinning_EmptyRevisionSkipped(t *testing.T) {
	apps := []Application{{Name: "a", RepoURL: ossRepo, TargetRevision: ""}}
	if mm := verifyRefPinning(apps, ossRepo, "feature/x"); len(mm) != 0 {
		t.Errorf("empty targetRevision must be skipped, got %+v", mm)
	}
}

// TestVerifyRefPinning_RepoURLNormalization: credentials, scheme, and .git must
// not defeat the repo match (else every child looks "foreign" and nothing is
// checked).
func TestVerifyRefPinning_RepoURLNormalization(t *testing.T) {
	apps := []Application{
		{Name: "openframe-api", RepoURL: "https://x-access-token:ghp_secret@github.com/flamingo-stack/openframe-oss-tenant.git", TargetRevision: "main"},
	}
	mm := verifyRefPinning(apps, ossRepo, "feature/x")
	if len(mm) != 1 {
		t.Fatalf("credentialed/.git URL must still match the repo and be flagged, got %+v", mm)
	}
}
