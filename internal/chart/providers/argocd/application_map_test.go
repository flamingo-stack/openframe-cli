package argocd

import "testing"

func TestApplicationFromArgoApp_FullMapping(t *testing.T) {
	var item argoApp
	item.Metadata.Name = "core-api"
	item.Status.Health.Status = "Healthy"
	item.Status.Health.Message = "all good"
	item.Status.Sync.Status = "Synced"
	item.Status.Sync.Revision = "abc123"
	item.Status.OperationState.Phase = "Succeeded"
	item.Status.OperationState.Message = "done"
	item.Status.ReconciledAt = "2026-01-01T00:00:00Z"
	item.Status.Conditions = []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}{
		{Type: "", Message: ""},                      // skipped (empty message)
		{Type: "ComparisonError", Message: "drift!"}, // first non-empty → taken
		{Type: "OtherError", Message: "ignored"},
	}
	item.Spec.Source.RepoURL = "https://github.com/org/repo"
	item.Spec.Source.Path = "manifests/app"
	item.Spec.Source.TargetRevision = "main"

	got := applicationFromArgoApp(item)

	checks := map[string]struct{ got, want string }{
		"Name":           {got.Name, "core-api"},
		"Health":         {got.Health, "Healthy"},
		"HealthMessage":  {got.HealthMessage, "all good"},
		"Sync":           {got.Sync, "Synced"},
		"SyncRevision":   {got.SyncRevision, "abc123"},
		"Condition":      {got.Condition, "drift!"},
		"ConditionType":  {got.ConditionType, "ComparisonError"},
		"OperationPhase": {got.OperationPhase, "Succeeded"},
		"RepoURL":        {got.RepoURL, "https://github.com/org/repo"},
		"Path":           {got.Path, "manifests/app"},
		"TargetRevision": {got.TargetRevision, "main"},
		"ReconciledAt":   {got.ReconciledAt, "2026-01-01T00:00:00Z"},
	}
	for field, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", field, c.got, c.want)
		}
	}
}

// TestApplicationFromArgoApp_DefaultsUnknown guards the display defaulting: an
// application with no health/sync yet must read "Unknown", not "".
func TestApplicationFromArgoApp_DefaultsUnknown(t *testing.T) {
	got := applicationFromArgoApp(argoApp{})
	if got.Health != "Unknown" {
		t.Errorf("empty Health = %q, want Unknown", got.Health)
	}
	if got.Sync != "Unknown" {
		t.Errorf("empty Sync = %q, want Unknown", got.Sync)
	}
}
