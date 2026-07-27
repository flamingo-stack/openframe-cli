package gke

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The GKE root module must never MANAGE the GCP project itself: the project is
// an input (var.project) the operator already owns, not a resource the CLI
// creates. If a `resource "google_project"` (or a project-deleting data flow)
// ever appears in the template, `openframe cluster delete` → terraform destroy
// would tear down the whole project — every unrelated workload in it included.
// This test fails loudly the moment the template stops treating the project as
// read-only.
func TestTemplate_NeverManagesTheProject(t *testing.T) {
	src := string(mainTF)

	// No resource that owns the project's lifecycle (create ⇒ destroy).
	forbidden := []string{
		`resource "google_project"`,
		`resource "google_project_iam_policy"`, // authoritative — would overwrite project IAM
	}
	for _, decl := range forbidden {
		assert.NotContainsf(t, src, decl,
			"GKE template must not declare %s — the project must stay a read-only input, never a managed/destroyable resource", decl)
	}

	// Belt and braces: no google_project* resource of ANY kind. The only
	// project-scoped resource we allow is google_project_service (API
	// enablement), and even that must never disable APIs on destroy.
	projectResRE := regexp.MustCompile(`resource\s+"(google_project[a-z_]*)"`)
	for _, m := range projectResRE.FindAllStringSubmatch(src, -1) {
		assert.Equalf(t, "google_project_service", m[1],
			"unexpected project-scoped resource %q in the GKE template; only google_project_service is permitted", m[1])
	}

	// google_project_service must keep disable_on_destroy=false so a cluster
	// teardown never switches off project APIs other workloads depend on.
	if assert.Contains(t, src, `resource "google_project_service"`) {
		assert.Contains(t, src, "disable_on_destroy = false",
			"google_project_service must set disable_on_destroy=false so destroy never disables project APIs")
	}
}

// The project must be consumed strictly as an input variable, proving the
// template reads the project rather than owning it.
func TestTemplate_ProjectIsAnInputVariable(t *testing.T) {
	require.Contains(t, string(mainTF), `variable "project"`,
		"the project must be a declared input variable, not a resource the template creates")
}
