package eks

import (
	"errors"
	"strings"
	"testing"
)

func TestOrphanFromInterruptedCreate(t *testing.T) {
	t.Run("names the orphaned resource on a ResourceInUseException", func(t *testing.T) {
		// Shape of a real terraform aws-provider collision during a resume.
		err := errors.New("Error: creating EKS Cluster (demo): operation error EKS: CreateCluster, " +
			"ResourceInUseException: Cluster already exists with name: demo, " +
			"arn:aws:eks:us-east-1:123456789012:cluster/demo")
		hint, ok := orphanFromInterruptedCreate(err, "/ws/test/terraform")
		if !ok {
			t.Fatal("expected a ResourceInUseException to be detected as an interrupted-create orphan")
		}
		if !strings.Contains(hint, "arn:aws:eks:us-east-1:123456789012:cluster/demo") {
			t.Fatalf("hint must name the exact orphaned resource, got:\n%s", hint)
		}
		if !strings.Contains(hint, "/ws/test/terraform") {
			t.Fatalf("hint must reference the workspace terraform dir for import, got:\n%s", hint)
		}
	})

	t.Run("detects IAM EntityAlreadyExists without an ARN", func(t *testing.T) {
		err := errors.New("Error: creating IAM Role (demo-node-group): EntityAlreadyExists: " +
			"Role with name demo-node-group already exists.")
		hint, ok := orphanFromInterruptedCreate(err, "/ws/test/terraform")
		if !ok {
			t.Fatal("expected EntityAlreadyExists to be detected as an interrupted-create orphan")
		}
		if !strings.Contains(hint, "the resource named in the error above") {
			t.Fatalf("without an ARN the hint must fall back to the generic pointer, got:\n%s", hint)
		}
	})

	t.Run("ignores unrelated apply errors", func(t *testing.T) {
		err := errors.New("Error: creating EC2 Instance: VcpuLimitExceeded: You have requested more vCPU capacity than your current limit")
		if _, ok := orphanFromInterruptedCreate(err, "/ws/test/terraform"); ok {
			t.Fatal("a non-collision error must not be reported as an orphan")
		}
	})
}
