
# Security

OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is a local operator tool: it provisions Kubernetes clusters (k3d, EKS, GKE) and deploys the OpenFrame platform onto them via Helm and ArgoCD's app-of-apps pattern. It is not a hosted service — it runs on the operator's own machine, under the operator's own cloud/Kubernetes credentials. Its security surface is therefore dominated by three concerns: how it handles the credentials it shells out with (gcloud, aws, kubeconfig, git tokens), how it verifies the binaries and charts it installs, and how it avoids leaking secrets into logs, terminal output, or the process table.

This document explains the authentication, data-handling, validation, and secrets-management patterns actually implemented in this codebase, and the practices expected of contributors. For general contribution workflow see [Contributing Guidelines](../contributing/guidelines.md); for the full architecture see [Architecture Overview](../architecture/README.md) and [Reference Architecture](../../reference/architecture/overview.md).

> **Scope note.** OpenFrame CLI has no recorded upstream or downstream dependents in the organization's dependency graph (per `get_repo_ecosystem`) — it is a leaf artifact. It links no application database, issues no user sessions, and stores no customer data of its own; the platform it deploys (OpenFrame itself) lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository, whose own security documentation should be consulted for the deployed platform's authentication/authorization model. This document covers only the CLI's own code.

## 1. Authentication and authorization patterns

The CLI does not implement its own identity system. It **delegates all authentication to the backing systems it drives** — cloud provider CLIs, Kubernetes, Git hosts — and layers a consistent, auditable flow on top of each.

### 1.1 Cloud provider authentication (GKE / EKS)

`internal/cluster/discovery/auth.go` implements `AuthFlow`, the single gcloud login flow every GKE-touching command goes through. Its `Ensure(ctx, requireADC)` method:

1. Checks whether `gcloud` is installed at all (`CLIMissing`) and fails fast with an install link if not.
2. Checks whether the primary account is authenticated, using `gcloud auth print-access-token --quiet` rather than `gcloud auth list` — the list command can report an **expired** account as still `ACTIVE`, so the CLI forces a real token refresh to catch this before it surfaces mid-Terraform-apply as a raw provider error.
3. When `requireADC` is true (Terraform's `google` provider needs Application Default Credentials), separately verifies `gcloud auth application-default print-access-token`.
4. In an interactive terminal, offers to run the relevant `gcloud auth login` / `gcloud auth application-default login` command right there, attached to the real stdin/stdout/stderr (browser-based flows need this). In a non-interactive session (CI), it **never prompts** — it returns an actionable error telling the operator which command to run manually.

AWS/EKS follows the equivalent pattern via `internal/cluster/prerequisites/aws/aws.go` and `internal/cluster/discovery/eks.go`, and `cmd/cluster/aws_identity.go` surfaces the resolved AWS identity so operators can confirm which account/role they are about to provision against before anything is created.

```go
// internal/cluster/discovery/auth.go — never prompts outside an interactive session
func (f *AuthFlow) login(ctx context.Context, prompt string, args []string, verified func() bool, manualHint string) error {
	if !f.interactive() {
		return fmt.Errorf("%s", manualHint)
	}
	...
}
```

**Pattern for contributors:** any new cloud-facing command must go through the existing `AuthFlow`/prerequisite-checker pattern rather than shelling out to a provider CLI directly. Never assume credentials are present; always probe first and fail with an actionable message.

### 1.2 Kubernetes cluster access

`internal/k8s` isolates *read/inspect* access to an already-existing cluster from `internal/cluster`, which handles cluster *creation*. `RestConfigForContext` (`internal/k8s/restconfig.go`) builds a `rest.Config` from a named kubeconfig context — the CLI relies entirely on the operator's existing kubeconfig and RBAC permissions; it does not mint, store, or escalate credentials of its own. This means authorization for any cluster operation (install, status, uninstall) is enforced by the Kubernetes API server against whatever identity the operator's kubeconfig context already carries, not by the CLI.

This design has a direct authorization-safety consequence that `cmd/app/app.go` encodes explicitly: the `app` command group rejects stray positional arguments (`noPositionalArgs`), because Cobra's default `ArbitraryArgs` behavior would otherwise let `openframe app status my-eks` silently run against whichever kube-context happened to be active — destructively so for `uninstall`. Targeting a cluster must go through the explicit `--context` flag.

```go
// cmd/app/app.go
func noPositionalArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("%q takes no positional arguments (got %q) — use --context to target a cluster",
			cmd.CommandPath(), strings.Join(args, " "))
	}
	return nil
}
```

`ui.ShowContextHeader()` prints the active kube-context as a dim line before any destructive subcommand runs — the cheapest guard against operating on the wrong cluster.

### 1.3 Git authentication (app-of-apps chart repository)

`internal/chart/providers/git/auth.go` implements credential handling for cloning the chart repository. Since the CLI supports only the OSS (`oss-tenant`) deployment, which is pulled from a **public** repository, `internal/chart/services/validation.go`'s `ConfigurationValidator.ValidateConfiguration` is intentionally a no-op — there are no credentials to validate for this deployment type. The git auth plumbing exists as the extension point for any future deployment type that does require a private repository and a token.

> If you are extending the CLI to support a private chart source, follow the pattern in `auth.go` precisely (see §2.1) rather than passing a token in a URL string or environment variable read ad hoc.

### 1.4 Self-update authenticity (supply-chain, not user auth)

`internal/shared/selfupdate/` is the closest thing this CLI has to an authorization boundary against a hostile actor: it decides whether a downloaded binary is allowed to **replace the one currently running**. `internal/shared/selfupdate/cosign.go` performs cosign keyless signature verification (via `sigstore/sigstore-go`) before `update.go` applies a new binary, and `internal/shared/selfupdate/rollback.go` exists specifically so that a bad or unverifiable update can be reverted. Treat this verification path as security-critical: any change that weakens or bypasses it (e.g., skipping verification on a particular OS/arch, or falling back silently on verification failure) must be treated as a regression, not a convenience fix.

## 2. Data encryption and secure storage

OpenFrame CLI does not operate its own datastore, so "encryption at rest" in the traditional sense does not apply. The relevant concerns are: never persisting secrets to disk unnecessarily, keeping credentials out of process argument lists and logs, and verifying integrity of anything downloaded or cloned.

### 2.1 Credentials never touch argv, files, or logs

`internal/chart/providers/git/auth.go`'s `gitAuth` struct is the canonical pattern in this codebase for handling any credential:

- `extractGitAuth` parses a clone URL and separates an embedded token from the URL itself, returning a `cleanURL` with no userinfo.
- The token is handed to go-git **only as an in-memory `transport.AuthMethod`** (`buildAuth`, returning `githttp.BasicAuth`) — never written back into a URL string, never passed as a CLI argument (which would be visible via `ps` or `/proc/<pid>/cmdline`), and never written to a credentials file on disk.
- `maskToken` redacts every occurrence of the token in a string before that string is used in logging or error output.
- The code explicitly handles the common single-field GitHub PAT shorthand (`https://<token>@host/...`), promoting it to the token role rather than silently dropping it — so it is both used correctly for auth and reliably masked.

```go
// internal/chart/providers/git/auth.go
type gitAuth struct {
	cleanURL string // URL with any userinfo stripped
	username string // username from the URL, if any
	token    string // password / PAT; empty for public repos
}
```

**This is the pattern to replicate** for any new credential-bearing input: parse and strip the secret from anything that might be logged or exec'd, carry it only in memory, and provide a masking helper alongside it.

### 2.2 Centralized redaction

`internal/shared/redact/redact.go` exposes a package-level `Redact(s string) string` helper used wherever command output might contain sensitive values before it reaches the terminal or a log line. It is called from `internal/chart/providers/helm/manager.go` (`InstallArgoCDWithProgress`) and from `internal/shared/executor/executor.go` (`RealCommandExecutor.ExecuteWithOptions`) — i.e., redaction is applied at the generic command-execution layer, not left to each individual caller to remember. Any new code path that streams subprocess output to the user should route through the executor's existing redaction rather than printing raw stdout/stderr directly.

### 2.3 Integrity verification for downloaded tools

`internal/shared/download/` is the CLI's substrate for fetching pinned third-party tool binaries (k3d, Helm, mkcert, Terraform, infracost — see `internal/shared/download/pins.go`). `VerifyChecksum` (`internal/shared/download/verify.go`) performs a SHA-256 comparison against an expected hex digest, and is called from `Downloader.FetchVerified` — meaning every pinned-tool download in this CLI is checksum-verified before use, not merely fetched and trusted. Do not add a new downloaded dependency without wiring it through `FetchVerified`/`VerifyChecksum`.

### 2.4 Infrastructure-as-code secrets

The Terraform root modules the CLI generates (`internal/cluster/providers/eks/templates/main.tf`, `internal/cluster/providers/gke/templates/main.tf`) mark cluster endpoint and CA data as `sensitive = true` outputs (see the GKE template's `cluster_endpoint` / `cluster_certificate_authority_data` outputs), so Terraform itself will not print these values in its own plan/apply output. Module versions in both templates are **pinned to exact versions**, not ranges — the EKS template's comment explains that a floating `~> 21.0` constraint once resolved to a new minor version whose changed defaults broke cluster creation in three independent ways. Pin exactly; bump deliberately.

Both cloud templates also follow least-exposure patterns for the clusters they create:
- The GKE template enables private nodes (`enable_private_nodes = true`, no external IPs) with egress routed through a dedicated Cloud NAT, keeping only the API endpoint public so the operator's machine can reach it without a VPN.
- The EKS template sets `http_tokens = "required"` (IMDSv2 enforced) on node metadata options, narrowing the credential-exposure surface of the node's instance role.

## 3. Input validation and sanitization

Because this CLI's primary "input" is operator-supplied flags and interactive wizard answers that are turned into shell/Terraform/Kubernetes operations, validation focuses on: rejecting ambiguous targeting, fixed-argv command construction (no shell injection via string concatenation), and typed-name confirmation before destructive actions.

### 3.1 No positional-argument ambiguity

As shown in §1.2, `cmd/app/app.go`'s `noPositionalArgs` rejects any stray positional argument for the `app` command group, forcing explicit `--context` targeting. This is a direct input-validation control against an entire class of "ran against the wrong cluster" incidents.

### 3.2 Typed-name confirmation for destructive operations

Per `docs/reference/architecture/overview.md`, `openframe cluster delete [NAME]` requires **typed-name confirmation** for cloud clusters before deletion proceeds — an explicit control against fat-fingered or scripted accidental deletion of billed cloud infrastructure.

### 3.3 Fixed-argv command construction

`internal/cluster/discovery/auth.go`'s `runLogin` builds the `gcloud` invocation as a fixed argument slice passed to `exec.CommandContext`, not a shell string:

```go
// internal/cluster/discovery/auth.go
cmd := osexec.CommandContext(ctx, "gcloud", args...) // #nosec G204 -- fixed argv assembled by this package, no user input
```

The `#nosec G204` annotation is paired with an explicit justification comment (fixed argv, no user input interpolated) rather than a blanket suppression — follow this convention if you need to suppress a linter security finding anywhere else in the codebase: justify it inline, scoped to the exact line.

### 3.4 Executor abstraction as a validation chokepoint

All shelled-out commands (Docker, k3d, Helm, Terraform, gcloud, aws) go through `internal/shared/executor`'s `CommandExecutor` abstraction rather than being invoked ad hoc throughout the codebase (per the architecture overview's dependency notes). This gives the codebase one place to enforce consistent argument handling, output capture, redaction (§2.2), and exit-code fidelity (§4.3) instead of many bespoke `exec.Command` call sites with inconsistent handling.

## 4. Common security vulnerabilities and mitigations

| Risk | Where it could arise | Mitigation implemented in this codebase |
|---|---|---|
| Token/secret leakage via process arguments | Git clone URLs with embedded PATs | `extractGitAuth` strips credentials from the URL before use; token passed only via in-memory `transport.AuthMethod` (`internal/chart/providers/git/auth.go`) |
| Token/secret leakage via logs/output | Any command output containing a credential | `maskToken` (git auth) and the general-purpose `redact.Redact` called from the executor and Helm manager (`internal/shared/redact/redact.go`) |
| Shell/argument injection | Shelling out to `gcloud`, `helm`, `terraform`, etc. | Fixed argv construction via `exec.CommandContext(ctx, name, args...)`, never shell string interpolation (`internal/cluster/discovery/auth.go`) |
| Operating on the wrong cluster | Ambiguous CLI arguments defaulting to the current kube-context | `noPositionalArgs` validation plus `ui.ShowContextHeader()` before destructive subcommands (`cmd/app/app.go`) |
| Accidental destruction of billed cloud resources | `cluster delete` on a cloud (EKS/GKE) cluster | Typed-name confirmation requirement (per `docs/reference/architecture/overview.md`) |
| Tampered or corrupted downloaded binaries | k3d/Helm/mkcert/Terraform/infracost downloads | SHA-256 `VerifyChecksum` gate inside `Downloader.FetchVerified` (`internal/shared/download/verify.go`) |
| Malicious or corrupted self-update binary | `openframe update` replacing the running binary | Cosign keyless signature verification (`internal/shared/selfupdate/cosign.go`) plus a `rollback` path (`internal/shared/selfupdate/rollback.go`) |
| Expired-but-reported-active cloud credentials | `gcloud auth list` can report an expired account as `ACTIVE` | `AuthFlow.hasFreshLogin` forces a real token mint via `print-access-token` instead of trusting `auth list` (`internal/cluster/discovery/auth.go`) |
| Retrying a destructive or already-failed operation blindly | Generic retry logic around install/upgrade | `NonRetryableError` interface lets an error veto retry outright, and `InstallationRetryPolicy` only matches specific transient substrings — pod-log/event text embedded in wait-failures is not blindly treated as retryable (`internal/shared/errors/retry_policy.go`) |
| Unverified IaC module drift breaking cluster security posture | Terraform module version ranges | Exact version pins (not `~>` ranges) in both `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf`, with the reasoning documented inline |
| Exit-code laundering hiding a failed external command's real status | Subprocess failures collapsed to a generic exit code | `executor.CommandError` preserves the original exit code through to `main.exitCode()` (`main.go`), enforced by rule `OPENFRAM-005` |

### 4.1 Error wrapping discipline

The organization-wide rule `FLEETMDM-002-2` ("Go errors must be wrapped with `fmt.Errorf(\"...: %w\", err)` at every call boundary") applies here and is a security-relevant practice, not just a style rule: unwrapped bare `return err` loses the context needed to tell whether a failure was a permission error, a network error, or something attacker-influenced, which in turn makes it harder to apply `classifyTransient` correctly in `internal/shared/errors/retry_policy.go` and harder for an operator to diagnose an auth failure versus a transient one.

### 4.2 Already-handled error sentinel (no duplicate/confusing output)

Rule `OPENFRAM-004` requires that any error already shown to the user via `sharedErrors.HandleGlobalError` be wrapped in `*sharederrors.AlreadyHandledError` before propagating to `main`. `main.go` checks for this sentinel with `errors.As` specifically to avoid re-printing an error the user already saw — relevant to security review because a double-printed or reformatted error message is a place where a redacted value could inadvertently be re-emitted in an unredacted form by a second print path.

### 4.3 Exit-code fidelity

`main.go`'s `exitCode` function unwraps `*executor.CommandError` to preserve a failed external command's real exit code (when it is a valid Unix code), rather than collapsing every failure to a generic `1`. This matters for CI/automation security gates that key off specific exit codes from Helm/Terraform/etc.

## 5. Security testing and code review guidelines

This repository enforces a set of automated, deterministic and judged review rules (see `get_code_rules`) on every change. The ones most relevant to a security-focused reviewer:

- **DUP-001** — a new or changed exported definition must not duplicate an existing one in the org. For security code specifically, this means: do not write a second ad hoc token-masking function or checksum verifier when `redact.Redact` or `download.VerifyChecksum` already exist — reuse them.
- **TEST-001** — a new or changed exported definition should be referenced by a test. Security-sensitive helpers (auth flows, redaction, checksum verification) are exactly the kind of code this rule exists to protect; an untested change to `gitAuth`, `AuthFlow`, or `VerifyChecksum` should be treated as a review blocker even though the rule's own severity is `info`.
- **OPENFRAM-001** — cross-package imports that would collide on their last path segment (e.g. two different `errors` or `config` packages) must use a disambiguating camelCase alias (`sharedErrors`, `uiCluster`, etc.). This is not cosmetic for security review: an unaliased import collision is exactly the kind of ambiguity that can cause a reviewer (or the author) to wire the wrong package's error-handling or config logic into a security-sensitive path.
- **OPENFRAM-002 / OPENFRAM-002-11** — business logic lives in an unexported `Service` struct behind a `New<Name>Service` constructor, with dependencies (including executors) injected rather than derived internally. Applied to security code, this keeps things like `AuthFlow` constructible with fake seams for tests (see `NewAuthFlowWithSeams`) instead of hardcoding real `gcloud`/filesystem access that would make negative-path testing (expired token, missing CLI, non-interactive session) impractical.
- **OPENFRAM-004** — see §4.2.
- **OPENFRAM-005** — see §4.3.
- **OPENFRAM-007** — all user-facing output must go through `pterm` printers or `internal/shared/ui`, never raw `fmt.Print*` to stdout. This matters for security because raw prints bypass the redaction and `--silent`/`--plain` behavior the shared UI layer provides.

When reviewing a change that touches authentication, credential handling, download verification, or self-update, additionally check:

1. Does any new credential value ever reach a string that is logged, printed, or passed as a CLI argument unmasked?
2. Is any new external-tool invocation built via fixed argv (`exec.CommandContext(ctx, name, args...)`) rather than a shell string, and if a `#nosec` suppression is needed, is it scoped and justified inline as in `internal/cluster/discovery/auth.go`?
3. Does a new retryable-error classification risk retrying a destructive operation that should instead fail closed (see the `NonRetryableError` discussion in `internal/shared/errors/retry_policy.go`)?
4. Does a new Terraform-generated resource expose a sensitive output without `sensitive = true`?

For the general test layout and how to run the test suite, see [Testing](../testing/README.md). For local environment setup used while developing security-sensitive code (credentials needed for GKE/EKS testing, etc.), see [Environment Setup](../setup/environment.md) and [Local Development](../setup/local-development.md).

## 6. Environment variables and secrets management

### 6.1 Reading environment-controlled behavior

`internal/shared/config/envbool.go` exposes `EnvBool(name string) bool`, the shared helper for interpreting boolean-valued environment variables consistently across the CLI (rather than each call site re-implementing its own truthy-string parsing). Any new environment-variable-gated behavior should be read through this helper for consistency.

### 6.2 No credentials are owned by this CLI

OpenFrame CLI does not define, store, or distribute its own application secrets, API keys, or default credentials — it has no user database, no session store, and no bundled default login. All credentials it uses at runtime belong to the systems it drives:

- **Cloud credentials** — `gcloud`'s own credential store and Application Default Credentials file, and the AWS CLI's own credential chain. The CLI probes and, interactively, offers to refresh these (§1.1), but never writes to or manages the underlying credential files itself beyond invoking the providers' own login commands.
- **Kubernetes credentials** — the operator's existing kubeconfig, resolved per-context by `internal/k8s/restconfig.go`. The CLI reads this; it does not generate or store kubeconfig credentials.
- **ArgoCD admin password** — per the architecture overview, `openframe app access` reads and prints the ArgoCD admin credentials (sourced from the `argocd-initial-admin-secret` Kubernetes secret, as referenced in the App Status Aggregation data flow) for the operator's own use; the CLI does not persist a copy of this password anywhere outside that read-and-display action.
- **Git tokens** (for any future private chart source) — handled exclusively via the in-memory pattern in §2.1; never written to a credentials file by this CLI.

> Refer to your environment configuration for any default usernames, passwords, or endpoints — this document intentionally does not list any, since none are owned or hardcoded by the CLI.

### 6.3 Signing material for self-update

The self-update signature verification in `internal/shared/selfupdate/cosign.go` relies on **cosign keyless verification**, meaning there is no long-lived private signing key that this CLI (or its operators) need to protect locally — verification is against the public Sigstore transparency log and OIDC-issued certificate chain, not a secret the CLI must manage. If this mechanism is ever extended to non-keyless signing, any resulting private key material must never be embedded in the repository or shipped inside the binary; follow the pattern already set by `internal/shared/download`'s checksum-pin approach (`internal/shared/download/pins.go`), which stores only public, verifiable digests.

### 6.4 Shell scripts

`scripts/sign-binary.sh` is the one shell script in this repository's setup surface (per the verified repository facts). Signing credentials or tokens it may require at build/release time should be supplied exclusively via the CI environment's own secrets store, never committed to the script or to the repository.

---

Because OpenFrame CLI has no recorded downstream consumers in the organization's dependency graph (coverage: partial — this is not proof that nothing anywhere depends on it, only that the indexed portion of the graph records none), there is no public Go API contract from this repository that other services import. Its `internal/` packages are private implementation detail, and security review of this repository can therefore focus entirely on the CLI's own runtime behavior described above rather than on a public interface contract.
