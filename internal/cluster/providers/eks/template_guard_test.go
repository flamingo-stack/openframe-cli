package eks

import (
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
