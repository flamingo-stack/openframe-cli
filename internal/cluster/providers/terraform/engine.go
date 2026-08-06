package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flamingo-stack/openframe-cli/internal/shared/download"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

// Runner is the subset of *tfexec.Terraform the engine uses; an interface so
// engine logic is testable without a terraform binary.
type Runner interface {
	Init(ctx context.Context, opts ...tfexec.InitOption) error
	Apply(ctx context.Context, opts ...tfexec.ApplyOption) error
	ApplyJSON(ctx context.Context, w io.Writer, opts ...tfexec.ApplyOption) error
	Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error
	DestroyJSON(ctx context.Context, w io.Writer, opts ...tfexec.DestroyOption) error
	Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error)
	ShowPlanFile(ctx context.Context, planPath string, opts ...tfexec.ShowOption) (*tfjson.Plan, error)
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
				tf.SetStderr(os.Stderr)
				return &verboseRunner{Terraform: tf}, nil
			}
			return tf, nil
		},
	}
}

// verboseRunner streams terraform's human-readable output (init, plan) to the
// terminal in verbose mode WITHOUT tee-ing the machine-readable commands.
// Holding tf.SetStdout(os.Stdout) for the runner's whole life did exactly
// that: tfexec merges its JSON parse buffer with the configured stdout, so
// `terraform show -json` dumped the entire plan — one 641 KB line, CA cert
// and user-data included — into the terminal, burying the very errors
// --verbose exists to reveal. Stdout is therefore enabled only around the
// human-output commands; ApplyJSON/DestroyJSON are unaffected either way
// (they pipe stdout to their own progress writer).
type verboseRunner struct {
	*tfexec.Terraform
}

// withStdout runs fn with terraform's human stdout streaming to the terminal,
// then silences it again. tfexec lazily runs `terraform version -json` before
// the first command of an instance — priming it first keeps even that blob
// off the terminal.
func (r *verboseRunner) withStdout(ctx context.Context, fn func() error) error {
	if _, _, err := r.Terraform.Version(ctx, false); err != nil {
		return err
	}
	r.Terraform.SetStdout(os.Stdout)
	defer r.Terraform.SetStdout(io.Discard)
	return fn()
}

func (r *verboseRunner) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	return r.withStdout(ctx, func() error { return r.Terraform.Init(ctx, opts...) })
}

func (r *verboseRunner) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	var changes bool
	err := r.withStdout(ctx, func() error {
		var planErr error
		changes, planErr = r.Terraform.Plan(ctx, opts...)
		return planErr
	})
	return changes, err
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

// Apply runs terraform apply in dir, streaming per-resource progress lines
// (via terraform's machine-readable -json output) so a 15-minute cloud apply
// is never a silent wait. It is idempotent: re-running after a partial
// failure resumes from the recorded state.
func (e *Engine) Apply(ctx context.Context, dir string) error {
	tf, err := e.newRunner(dir)
	if err != nil {
		return err
	}
	if err := tf.ApplyJSON(ctx, newProgressWriter(e.verbose)); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}
	return nil
}

// Destroy runs terraform destroy in dir, streaming progress like Apply.
//
// Guardrail: terraform destroy tears down exactly what the state in `dir`
// tracks, so `dir` must be a materialized cluster workspace — one holding the
// CLI-generated main.tf root module. Refusing an empty, wrong, or parent path
// keeps a destroy strictly inside one cluster's own workspace and never lets
// terraform be pointed at a directory whose state we did not generate.
func (e *Engine) Destroy(ctx context.Context, dir string) error {
	if err := assertClusterWorkspace(dir); err != nil {
		return err
	}
	tf, err := e.newRunner(dir)
	if err != nil {
		return err
	}
	if err := tf.DestroyJSON(ctx, newProgressWriter(e.verbose)); err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}
	return nil
}

// assertClusterWorkspace verifies dir is a generated cluster workspace before
// a destructive terraform run: it must hold a main.tf root module. This is the
// scoping check behind Destroy — terraform acts on the state in whatever
// directory it is handed, so a directory that is not one of our generated
// workspaces must fail loudly instead of running a destroy.
func assertClusterWorkspace(dir string) error {
	info, err := os.Stat(filepath.Join(dir, "main.tf"))
	if err != nil {
		return fmt.Errorf("refusing to run terraform destroy in %q: not a cluster workspace (no generated main.tf): %w", dir, err)
	}
	if info.IsDir() {
		return fmt.Errorf("refusing to run terraform destroy in %q: main.tf is not a regular file", dir)
	}
	return nil
}

// PlanChange is one resource-level action of a terraform plan, in
// terraform's own diff notation: "+" create, "~" update, "-" destroy,
// "-/+" replace.
type PlanChange struct {
	Action  string
	Address string
}

// PlanSummary is the resource-change footprint of a terraform plan.
type PlanSummary struct {
	Add     int
	Change  int
	Destroy int
	// Changes lists every planned resource action, in plan order — the counts
	// alone don't tell the user WHAT would be created.
	Changes []PlanChange
	// PlanJSON is the machine-readable plan (terraform show -json), used by
	// the optional infracost estimate. Empty when the plan has no changes.
	PlanJSON []byte
}

// HasChanges reports whether the plan would modify anything.
func (s PlanSummary) HasChanges() bool { return s.Add+s.Change+s.Destroy > 0 }

// Plan runs terraform plan in dir and summarizes the pending changes by
// resource action (create/update/delete).
func (e *Engine) Plan(ctx context.Context, dir string) (PlanSummary, error) {
	summary, planFile, err := e.PlanForApply(ctx, dir)
	if planFile != "" {
		_ = os.Remove(planFile)
	}
	return summary, err
}

// PlanForApply plans dir and KEEPS the plan file for a subsequent ApplyPlan —
// the interactive `terraform apply` shape: what was shown to the user is
// EXACTLY what gets applied, with no re-plan drift in between. The caller
// owns removing the returned plan file.
func (e *Engine) PlanForApply(ctx context.Context, dir string) (PlanSummary, string, error) {
	tf, err := e.newRunner(dir)
	if err != nil {
		return PlanSummary{}, "", err
	}
	planFile := filepath.Join(dir, "tfplan")

	changes, err := tf.Plan(ctx, tfexec.Out(planFile))
	if err != nil {
		return PlanSummary{}, "", fmt.Errorf("terraform plan failed: %w", err)
	}
	if !changes {
		return PlanSummary{}, planFile, nil
	}
	plan, err := tf.ShowPlanFile(ctx, planFile)
	if err != nil {
		return PlanSummary{}, planFile, fmt.Errorf("terraform show failed: %w", err)
	}
	var summary PlanSummary
	// Best-effort: the JSON only feeds the optional cost estimate.
	if data, err := json.Marshal(plan); err == nil {
		summary.PlanJSON = data
	}
	for _, rc := range plan.ResourceChanges {
		switch {
		case rc.Change.Actions.Create():
			summary.Add++
			summary.Changes = append(summary.Changes, PlanChange{Action: "+", Address: rc.Address})
		case rc.Change.Actions.Update():
			summary.Change++
			summary.Changes = append(summary.Changes, PlanChange{Action: "~", Address: rc.Address})
		case rc.Change.Actions.Delete():
			summary.Destroy++
			summary.Changes = append(summary.Changes, PlanChange{Action: "-", Address: rc.Address})
		case rc.Change.Actions.Replace():
			summary.Add++
			summary.Destroy++
			summary.Changes = append(summary.Changes, PlanChange{Action: "-/+", Address: rc.Address})
		}
	}
	return summary, planFile, nil
}

// ApplyPlan applies a SAVED plan file produced by PlanForApply, streaming
// per-resource progress like Apply.
func (e *Engine) ApplyPlan(ctx context.Context, dir, planFile string) error {
	tf, err := e.newRunner(dir)
	if err != nil {
		return err
	}
	if err := tf.ApplyJSON(ctx, newProgressWriter(e.verbose), tfexec.DirOrPlan(planFile)); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}
	return nil
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
