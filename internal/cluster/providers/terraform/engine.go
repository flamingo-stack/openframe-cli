package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/hashicorp/terraform-exec/tfexec"
)

// Runner is the subset of *tfexec.Terraform the engine uses; an interface so
// engine logic is testable without a terraform binary.
type Runner interface {
	Init(ctx context.Context, opts ...tfexec.InitOption) error
	Apply(ctx context.Context, opts ...tfexec.ApplyOption) error
	Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error
	Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error)
	Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)
}

// Engine drives terraform init/plan/apply/destroy/output in a workspace's
// terraform directory via the terraform-exec library.
type Engine struct {
	verbose bool
	// newRunner is the construction seam for tests; production builds a
	// *tfexec.Terraform on the resolved binary.
	newRunner func(workdir string) (Runner, error)
}

// FindTerraform resolves the terraform binary, preferring the CLI-managed
// pinned install in ~/.openframe/bin over whatever is on PATH.
func FindTerraform() (string, error) {
	if binDir, err := download.UserBinDir(); err == nil {
		download.PrependToPath(binDir)
	}
	path, err := exec.LookPath("terraform")
	if err != nil {
		return "", fmt.Errorf("terraform binary not found (the prerequisite installer provides a verified %s): %w", download.Terraform.Version, err)
	}
	return path, nil
}

// NewEngine builds the production engine. Verbose streams terraform's own
// human output to the terminal; otherwise the engine stays quiet and the
// caller's spinner owns the UX.
func NewEngine(verbose bool) *Engine {
	return &Engine{
		verbose: verbose,
		newRunner: func(workdir string) (Runner, error) {
			bin, err := FindTerraform()
			if err != nil {
				return nil, err
			}
			tf, err := tfexec.NewTerraform(workdir, bin)
			if err != nil {
				return nil, fmt.Errorf("initializing terraform runner: %w", err)
			}
			if verbose {
				tf.SetStdout(os.Stdout)
				tf.SetStderr(os.Stderr)
			}
			return tf, nil
		},
	}
}

// NewEngineWithRunner is the test constructor.
func NewEngineWithRunner(newRunner func(workdir string) (Runner, error)) *Engine {
	return &Engine{newRunner: newRunner}
}

// Init runs terraform init in dir.
func (e *Engine) Init(ctx context.Context, dir string) error {
	tf, err := e.newRunner(dir)
	if err != nil {
		return err
	}
	if err := tf.Init(ctx, tfexec.Upgrade(false)); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	return nil
}

// Apply runs terraform apply in dir. It is idempotent: re-running after a
// partial failure resumes from the recorded state.
func (e *Engine) Apply(ctx context.Context, dir string) error {
	tf, err := e.newRunner(dir)
	if err != nil {
		return err
	}
	if err := tf.Apply(ctx); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}
	return nil
}

// Destroy runs terraform destroy in dir.
func (e *Engine) Destroy(ctx context.Context, dir string) error {
	tf, err := e.newRunner(dir)
	if err != nil {
		return err
	}
	if err := tf.Destroy(ctx); err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}
	return nil
}

// Plan runs terraform plan in dir and reports whether changes are pending.
func (e *Engine) Plan(ctx context.Context, dir string) (bool, error) {
	tf, err := e.newRunner(dir)
	if err != nil {
		return false, err
	}
	changes, err := tf.Plan(ctx)
	if err != nil {
		return false, fmt.Errorf("terraform plan failed: %w", err)
	}
	return changes, nil
}

// Outputs returns the root-module outputs of dir as raw JSON values.
func (e *Engine) Outputs(ctx context.Context, dir string) (map[string]json.RawMessage, error) {
	tf, err := e.newRunner(dir)
	if err != nil {
		return nil, err
	}
	metas, err := tf.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("terraform output failed: %w", err)
	}
	out := make(map[string]json.RawMessage, len(metas))
	for k, v := range metas {
		out[k] = v.Value
	}
	return out, nil
}

// StringOutput decodes a string-typed output value.
func StringOutput(outputs map[string]json.RawMessage, key string) (string, error) {
	raw, ok := outputs[key]
	if !ok {
		return "", fmt.Errorf("terraform output %q missing", key)
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", fmt.Errorf("terraform output %q is not a string: %w", key, err)
	}
	return s, nil
}
