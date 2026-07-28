package gke

import (
	"errors"
	"strings"
	"testing"
)

func TestOrphanFromInterruptedCreate(t *testing.T) {
	t.Run("names the orphaned resource on a 409", func(t *testing.T) {
		// Shape of a real terraform google-provider 409 during a resume.
		err := errors.New("Error: Error creating Network: googleapi: Error 409: The resource " +
			"'projects/tenant-y0/global/networks/test-resume-vpc' already exists, alreadyExists")
		hint, ok := orphanFromInterruptedCreate(err, "/ws/test/terraform")
		if !ok {
			t.Fatal("expected a 409 AlreadyExists to be detected as an interrupted-create orphan")
		}
		if !strings.Contains(hint, "projects/tenant-y0/global/networks/test-resume-vpc") {
			t.Fatalf("hint must name the exact orphaned resource, got:\n%s", hint)
		}
		if !strings.Contains(hint, "/ws/test/terraform") {
			t.Fatalf("hint must reference the workspace terraform dir for import, got:\n%s", hint)
		}
	})

	t.Run("ignores unrelated apply errors", func(t *testing.T) {
		err := errors.New("Error: quota 'CPUS' exceeded, limit 24.0 in region us-central1")
		if _, ok := orphanFromInterruptedCreate(err, "/ws/test/terraform"); ok {
			t.Fatal("a non-409 error must not be reported as an orphan")
		}
	})
}
