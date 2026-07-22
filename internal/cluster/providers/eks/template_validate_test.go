//go:build tfvalidate

package eks

// Opt-in template validation: `make test-tfvalidate`. Runs a real
// `terraform init` (downloads the pinned modules + providers, needs network)
// followed by `terraform validate`, which statically checks the generated
// root module against the downloaded module schemas — wrong input names or
// types fail here without any cloud credentials. This is the strongest check
// available before a real (billed) e2e apply.

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
		Type:      models.ClusterTypeEKS,
		NodeCount: 3,
		Cloud:     &models.CloudConfig{Region: "us-east-1"},
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
		t.Fatalf("EKS template is invalid: %d error(s)", out.ErrorCount)
	}
}
