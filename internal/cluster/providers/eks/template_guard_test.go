package eks

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Since Kubernetes 1.23 EKS ships no in-tree EBS provisioner: without the
// aws-ebs-csi-driver addon every PVC stays Pending forever and the OpenFrame
// platform (Kafka, MongoDB, Cassandra, … in 'datasources') can never come up.
// This guard fails loudly the moment the addon — or the node-role policy that
// lets its controller talk to EC2 — disappears from the template.
func TestTemplate_InstallsEBSCSIDriver(t *testing.T) {
	src := string(mainTF)
	require.Contains(t, src, "aws-ebs-csi-driver",
		"the EBS CSI addon is required or PVCs stay Pending on EKS >= 1.23")
	assert.Contains(t, src, "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy",
		"the CSI controller authenticates through the node role — it needs the EBS policy attached")
}

// The post-destroy orphan sweep (teardown.go) finds CSI-provisioned volumes by
// the tag the template tells the driver to stamp. The two sides live in
// different files; this pins them together so neither can drift alone.
func TestTemplate_VolumeTagMatchesSweepFilter(t *testing.T) {
	src := string(mainTF)
	require.Contains(t, src, "extraVolumeTags",
		"CSI volumes live outside terraform state — without the tag the sweep cannot find them")
	assert.Contains(t, src, `"`+orphanVolumeTagKey+`"`,
		"the template's volume tag key must equal teardown.go's sweep filter (orphanVolumeTagKey)")
}

// The CLI drives the cluster from the operator's machine with the operator's
// identity: the API endpoint must stay publicly reachable and the creating
// identity must be granted cluster admin, or every post-create step (kubeconfig
// merge, chart install) fails against a cluster that was just billed into
// existence.
func TestTemplate_OperatorCanReachTheCluster(t *testing.T) {
	src := string(mainTF)
	assert.Contains(t, src, "endpoint_public_access = true")
	assert.Contains(t, src, "enable_cluster_creator_admin_permissions = true")
}

// EKS module v21 bootstraps NO addons itself (bootstrap_self_managed_addons =
// false): everything a working cluster needs must be declared. Without vpc-cni
// no pod gets a network and nodes never become Ready; without kube-proxy no
// Service routes; without coredns nothing resolves. This guard fails the
// moment any of them — or the CNI's before_compute ordering — leaves the
// template.
func TestTemplate_DeclaresCoreAddons(t *testing.T) {
	src := string(mainTF)
	// Match the addon map ENTRIES, not the raw source: the addon names also
	// appear in comments, which must not be able to satisfy this guard.
	for _, addon := range []string{"vpc-cni", "kube-proxy", "coredns"} {
		assert.Regexpf(t, `(?m)^\s*`+regexp.QuoteMeta(addon)+`\s*=\s*\{`, src,
			"module v21 installs no addons by itself — %s must be declared or the cluster is born broken", addon)
	}
	assert.Regexp(t, `(?ms)^\s*vpc-cni\s*=\s*\{[^}]*^\s*before_compute\s*=\s*true`, src,
		"vpc-cni itself must install before the node group, or nodes wait on a missing CNI")
}

// The EBS CSI controller is a regular pod reaching the node role via IMDS —
// one hop more than host network. Module v21 defaults the hop limit to 1,
// which cuts the controller off ("no EC2 IMDS role found") and no volume ever
// binds. Separately, since EKS 1.30 AWS ships no default StorageClass, so the
// addon must create one or every PVC with no explicit class stays Pending.
func TestTemplate_CSIControllerCanActuallyProvision(t *testing.T) {
	src := string(mainTF)
	assert.Contains(t, src, "http_put_response_hop_limit = 2",
		"IMDS hop limit must be 2 or the (non-hostNetwork) CSI controller cannot reach the node role")
	assert.Contains(t, src, "defaultStorageClass",
		"since EKS 1.30 there is no default StorageClass unless the CSI addon creates one")
}

// Module versions must be exact pins. A floating "~> 21.0" resolved to a new
// minor whose changed defaults broke create in three independent ways; the
// docs also promise pinned modules.
func TestTemplate_ModuleVersionsAreExactPins(t *testing.T) {
	src := string(mainTF)
	moduleVersions := regexp.MustCompile(
		`source\s*=\s*"terraform-aws-modules/[^"]+"\s*\n\s*version\s*=\s*"([^"]+)"`,
	).FindAllStringSubmatch(src, -1)
	require.Len(t, moduleVersions, 2, "expected exactly the eks and vpc module blocks")
	for _, m := range moduleVersions {
		assert.Regexpf(t, `^\d+\.\d+\.\d+$`, m[1],
			"module versions must be exact pins, bumped deliberately — a range (%q) lets upstream default changes arrive unannounced", m[1])
	}
}

// The kubernetes_version default must be a concrete version: module v21 gates
// a data source on `kubernetes_version == null` and fails at plan time when
// the value is unknown ("Invalid count argument") — the documented no-flags
// create must be able to plan.
func TestTemplate_KubernetesVersionHasConcreteDefault(t *testing.T) {
	src := string(mainTF)
	assert.NotContains(t, src, `var.kubernetes_version != ""`,
		"kubernetes_version must never map to null — module v21 cannot plan with an unknown version")
	assert.Regexp(t, `variable "kubernetes_version" \{[^}]*default\s*=\s*"\d+\.\d+"`, src,
		"the kubernetes_version variable must default to a concrete <major>.<minor>")
}

// Subnets must reach the EKS module as real module.vpc references (not string
// literals), so terraform keeps the destroy-ordering edge from the cluster to
// the VPC — the reason the GKE template needs an explicit depends_on is exactly
// that its strings sever this edge.
func TestTemplate_SubnetsAreModuleReferences(t *testing.T) {
	src := string(mainTF)
	require.Contains(t, src, "subnet_ids = module.vpc.private_subnets",
		"subnet_ids must be a module.vpc output reference so destroy tears the cluster down before the VPC")
	assert.False(t, strings.Contains(src, `subnet_ids = "`),
		"subnet_ids must not be a string literal")
}
