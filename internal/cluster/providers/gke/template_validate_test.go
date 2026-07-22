//go:build tfvalidate

package gke

// Opt-in template validation: `make test-tfvalidate`. See the EKS twin for
// the rationale — this statically verifies the generated root module against
// the real downloaded module schemas, without cloud credentials.

import (
	"context"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"github.com/hashicorp/terraform-exec/tfexec"
)

func TestTerraformValidate(t *testing.T) {
	bin, err := tfengine.FindTerraform()
	if err != nil {
		t.Fatalf("the tfvalidate tag requires a terraform binary: %v", err)
	}

	vars, err := tfvarsFor(models.ClusterConfig{
		Name:      "validate-only",
		Type:      models.ClusterTypeGKE,
		NodeCount: 3,
		Cloud:     &models.CloudConfig{Region: "us-central1", Project: "validate-only"},
	})
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	if err := tfengine.WriteModule(dir, mainTF, vars); err != nil {
		t.Fatal(err)
	}

	tf, err := tfexec.NewTerraform(dir, bin)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
		t.Fatalf("terraform init: %v", err)
	}

	out, err := tf.Validate(ctx)
	if err != nil {
		t.Fatalf("terraform validate: %v", err)
	}
	for _, d := range out.Diagnostics {
		t.Logf("%s: %s — %s", d.Severity, d.Summary, d.Detail)
	}
	if !out.Valid || out.ErrorCount > 0 {
		t.Fatalf("GKE template is invalid: %d error(s)", out.ErrorCount)
	}
}
