package gke

import (
	"regexp"
	"strings"
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

	// Belt and braces: no google_project* resource of ANY kind — including
	// google_project_service. API enablement moved to a create-time gcloud step
	// (ensureProjectServices) precisely so no cluster workspace owns
	// project-level state in a shared project (report M5).
	projectResRE := regexp.MustCompile(`resource\s+"(google_project[a-z_]*)"`)
	for _, m := range projectResRE.FindAllStringSubmatch(src, -1) {
		assert.Failf(t, "project-scoped resource in the GKE template",
			"%q must not be managed per-cluster: project-level state belongs to no single cluster (M5)", m[1])
	}
}

// The project must be consumed strictly as an input variable, proving the
// template reads the project rather than owning it.
func TestTemplate_ProjectIsAnInputVariable(t *testing.T) {
	require.Contains(t, string(mainTF), `variable "project"`,
		"the project must be a declared input variable, not a resource the template creates")
}

// The GKE cluster's subnet/secondary ranges are passed to module.gke as string
// literals, not module.network outputs, so terraform builds no implicit edge
// from the cluster to the subnet. module.gke must therefore depend_on
// module.network explicitly — otherwise destroy can delete the subnet in
// parallel with the still-running node pool and GCP rejects it. This asserts
// the ordering edge stays in place.
func TestTemplate_GKEDependsOnNetworkForDestroyOrdering(t *testing.T) {
	src := string(mainTF)

	// The literal wiring that removes the implicit edge (the reason the explicit
	// depends_on is required). If these ever become module.network references,
	// this guard can be revisited.
	require.Contains(t, src, `subnetwork        = "${var.cluster_name}-subnet"`,
		"subnetwork is expected to be a string literal — the premise of the explicit dependency")

	depRE := regexp.MustCompile(`depends_on\s*=\s*\[([^\]]*)\]`)
	var gkeDep string
	for _, m := range depRE.FindAllStringSubmatch(src, -1) {
		if strings.Contains(m[1], "module.network") {
			gkeDep = m[1]
		}
	}
	require.NotEmpty(t, gkeDep,
		"module.gke must depend_on module.network so destroy tears the cluster down before the subnet")
}
