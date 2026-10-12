
# Security

OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is a Go command-line tool, not a hosted service: it has no HTTP listener, no database, and no end-user login. Its security surface is different from a typical web application — it is about how the CLI handles credentials for the systems it drives (Kubernetes, Git, cloud provider APIs, container registries), how it shells out to external tools safely, and how its own binary distribution is verified. This document explains the patterns actually implemented in this repository, where they live, and what to watch for when extending them.

> Per the ecosystem graph (`get_repo_ecosystem`), OpenFrame CLI publishes a single Go artifact (`github.com/flamingo-stack/openframe-cli`), has no recorded upstream dependencies on other internal repositories, and no recorded downstream consumers inside this organization's graph — coverage is **partial**, so this does not prove nothing imports it, only that no consumer is indexed yet. Treat its `internal/` packages as private implementation detail regardless.

For the Flamingo-wide data handling and privacy commitments (what data is or is not collected across Flamingo products, including OpenFrame), see `SECURITY.md` / the Flamingo AI Privacy Policy at the root of this repository. This document focuses on the CLI's own code-level security posture.

## 1. Authentication and authorization patterns

OpenFrame CLI does not implement its own authentication system. It is a client that authenticates **to other systems** on the operator's behalf, and otherwise defers authorization decisions to those systems:

| Target system | Auth mechanism | Where it lives |
|---|---|---|
| Kubernetes clusters | Resolves the active kubeconfig context; builds a `rest.Config` via `client-go` | `internal/k8s/restconfig.go`, `internal/k8s/contexts.go` |
| Google Cloud (GKE) | Interactive `gcloud auth login` / `gcloud auth application-default login` flow | `internal/cluster/discovery/auth.go` |
| AWS (EKS) | Delegates to the AWS CLI / SDK's own credential chain | `internal/cluster/prerequisites/aws/aws.go`, `internal/cluster/discovery/eks.go` |
| Git repositories (app-of-apps chart repo) | HTTP Basic auth built in-memory from a token embedded in the clone URL | `internal/chart/providers/git/auth.go` |
| ArgoCD | Reads the generated `argocd-initial-admin-secret` from the cluster | `internal/chart/providers/argocd` |
| CLI self-update | Sigstore/cosign keyless signature verification against GitHub Releases | `internal/shared/selfupdate/cosign.go` |

### The GKE interactive auth flow

`internal/cluster/discovery/auth.go` defines `AuthFlow`, the single `gcloud` login flow every GKE-touching command goes through. Its entry point, `Ensure(ctx, requireADC)`, does the following, in order:

1. Checks `gcloud` is installed (`CLIMissing` → actionable error pointing at the Google Cloud SDK install docs).
2. Checks the primary login is both present and **fresh** — not just "ACTIVE" in `gcloud auth list` (which can lie about an expired token), but actually able to mint a token via `gcloud auth print-access-token --quiet`. This two-step check exists because `gcloud auth list` reports an expired account as still active.
3. In an interactive session, prompts before running `gcloud auth login` or `gcloud auth application-default login`; in a non-interactive session (CI), it **never prompts** and returns an actionable error instead.
4. When `requireADC` is true (required for `cluster create --type gke`, where Terraform's Google provider consumes Application Default Credentials), separately verifies ADC via `gcloud auth application-default print-access-token`.

This flow is the reference pattern for any future "authenticate to an external system" feature in this CLI: check state first, prompt only when interactive, never block CI with a hanging prompt, and give a copy-paste command in the error message when automation must do it manually.

### Git credential handling

`internal/chart/providers/git/auth.go` is the pattern to follow for any credential that must reach a subprocess or library call without leaking. Its `gitAuth` struct deliberately separates a **credential-free URL** from the extracted token:

```go
// gitAuth holds credentials extracted from a clone URL together with the
// cleaned, credential-free URL.
type gitAuth struct {
	cleanURL string // URL with any userinfo stripped
	username string // username from the URL, if any
	token    string // password / PAT; empty for public repos
}
```

`extractGitAuth` parses a raw clone URL and strips any embedded userinfo, including the common single-field GitHub PAT shorthand (`https://<token>@host`), promoting it to the token role rather than silently discarding it. `buildAuth()` then returns an in-memory go-git `transport.AuthMethod` (HTTP Basic, username `x-access-token` by convention — the `GitTokenUser` constant) — the token is **never** written back into the URL, so it never appears in process argv, in `/proc/<pid>/cmdline`, in a credentials file on disk, or in error text built from the URL.

Per `internal/chart/services/validation.go`, the CLI's `ConfigurationValidator` is intentionally a no-op: the only supported deployment type is the public OSS (`oss-tenant`) app-of-apps repository, which requires no credentials at all. If a future deployment type needs authenticated chart repositories, this is the extension point — and it should reuse the `gitAuth` pattern above rather than reinvent credential handling.

### Authorization

The CLI has no internal RBAC: whatever the active kubeconfig context, `gcloud` account, or AWS credentials are authorized to do on the target system is what the CLI can do. `cmd/app/app.go` layers one guardrail on top of this: `noPositionalArgs` rejects stray positional arguments on `app` subcommands so that, for example, `openframe app uninstall my-eks` cannot silently run against the current kube-context while the operator believes they named a different target — it forces the destructive command to use `--context` explicitly. Destructive cloud operations (`cluster delete`) require typed-name confirmation per the architecture overview's command reference.

## 2. Data encryption and secure storage

OpenFrame CLI does not persist secrets to its own data store — it has none. Instead:

- **Kubeconfig and cloud credentials** are read from the standard locations each underlying tool already uses (kubeconfig file, `gcloud`/`aws` credential stores); the CLI does not create a parallel credential store.
- **Terraform state** for cloud clusters (EKS via `internal/cluster/providers/eks/templates/main.tf`, GKE via `internal/cluster/providers/gke/templates/main.tf`) is generated per-cluster under a CLI-owned workspace; these templates are explicitly marked "do not edit by hand — the CLI owns this workspace." Terraform outputs such as the GKE cluster endpoint and CA certificate data are marked `sensitive = true` in `main.tf` so Terraform itself does not print them in plan/apply output.
- **Downloaded third-party binaries** (k3d, Helm, mkcert, Terraform, infracost) are checksum-verified before use. `internal/shared/download/verify.go` exposes `VerifyChecksum(data []byte, wantHex string)`, called from `Downloader.FetchVerified` — every pinned tool binary is hashed and compared against an expected value before it is trusted, rather than executed on trust alone.
- **Self-update binaries** go through signature verification, not just a checksum: `internal/shared/selfupdate/cosign.go` performs Sigstore/cosign keyless verification against the GitHub Release artifact before replacing the running binary, and `internal/shared/selfupdate/rollback.go` provides a rollback path if an update turns out to be bad.
- **Secret values are redacted before they reach logs or terminal output.** `internal/shared/redact/redact.go` exposes `Redact(s string)`, called from `HelmManager.InstallArgoCDWithProgress` (`internal/chart/providers/helm/manager.go`) and from the real command executor's `ExecuteWithOptions` (`internal/shared/executor/executor.go`) — meaning redaction is applied at the executor layer itself, so any shelled-out command's streamed output passes through the same scrubbing regardless of which higher-level caller invoked it.

> There is no application-level "encryption at rest" concern inside this repository: the CLI's job is to hand credentials to already-encrypted transports (HTTPS/TLS to Git, Kubernetes, and cloud APIs) and to avoid writing them to disk or argv in the clear itself.

## 3. Input validation and sanitization

The CLI's main validation surface is **command-line flags and cluster-name arguments**, since it spends most of its time building arguments for other command-line tools (`helm`, `terraform`, `gcloud`, `aws`, `docker`, `k3d`).

- `cmd/app/app.go`'s `noPositionalArgs` rejects any stray positional argument on `app` subcommands outright, with an error naming the exact flag (`--context`) that should have been used — this is validation aimed at preventing a *semantic* mistake (operating on the wrong cluster), not just a syntax error.
- `cmd/bootstrap/bootstrap.go` validates the cluster name before using it (per the architecture overview's description of `bootstrap`'s command definition).
- `internal/cluster/models/flags.go` and `internal/shared/flags/flags.go` centralize flag definitions, which keeps validation rules for cluster names, types, and regions in one place rather than duplicated per-command.
- Subprocess arguments are passed as discrete `argv` elements through the `executor.CommandExecutor` abstraction (`internal/shared/executor/executor.go`), **not** interpolated into a shell string — this is the same class of protection applied explicitly in `internal/cluster/discovery/auth.go`, where the `gcloud` login subprocess is built with `osexec.CommandContext(ctx, "gcloud", args...)` and annotated `#nosec G204 -- fixed argv assembled by this package, no user input`. That comment records the actual security property being relied on: the argv is fixed by the package, not built from unsanitized user input, so there is no shell-injection surface even though the linter would otherwise flag a dynamic `exec.Command`.
- Chart configuration intended for non-interactive mode is validated through `internal/chart/services/validation.go`'s `ConfigurationValidator.ValidateConfiguration` — currently a deliberate no-op because the only supported configuration (OSS/public repo) has nothing sensitive to validate, but it is the seam where future configuration-shape validation belongs.

### Retry classification is not an injection surface, but is a correctness-adjacent trap

`internal/shared/errors/retry_policy.go` classifies whether a failed operation should be retried partly by **scanning the error's string content** for substrings like `"connection refused"` or `"i/o timeout"` (`ExponentialBackoffPolicy.ShouldRetry`, via the `contains` helper). This fallback exists because shelled-out tools (helm, kubectl, docker) only expose a failure as an exit code plus stderr text. The comments in this file document a real incident shape worth knowing before touching this logic: an app-install-wait failure embeds pod logs and Kubernetes events in its error text, which can *contain* transient-looking phrases even when the underlying condition is not transient — blindly retrying would reinstall an already-installed ArgoCD. The fix implemented is the `NonRetryableError` interface, which lets a specific error explicitly veto retries regardless of what substrings appear in its text; structural classification (`classifyTransient`, checking real Go error types like `net.Error`, `syscall.ECONNREFUSED`, Kubernetes API error helpers) is tried before the substring fallback for the same reason. Any new retryable-error pattern should prefer a structural check over adding another substring.

## 4. Common security vulnerabilities and mitigations

| Risk | Where it could occur | Mitigation in this codebase |
|---|---|---|
| Credential leakage via process arguments | Git clone, cloud CLI invocations | `gitAuth.buildAuth()` keeps tokens out of URLs and argv entirely (`internal/chart/providers/git/auth.go`); `gcloud` login uses a fixed, no-user-input argv (`internal/cluster/discovery/auth.go`) |
| Credential leakage via logs/terminal output | Any command whose stdout/stderr is streamed to the user | `redact.Redact` is applied inside the real command executor (`internal/shared/executor/executor.go`) and in ArgoCD install progress output (`internal/chart/providers/helm/manager.go`), so secrets are scrubbed at the lowest common layer |
| Shell injection via dynamically built commands | Any `exec.Command`/`osexec.CommandContext` call | Arguments are passed as discrete slice elements, not concatenated into a shell string; the one place this is explicitly called out (`internal/cluster/discovery/auth.go`) is annotated to record that the argv is fixed, not built from user input (`#nosec G204`) |
| Tampered/corrupted downloaded binaries | Pinned tool downloads (k3d, Helm, mkcert, Terraform, infracost) | `internal/shared/download/verify.go`'s `VerifyChecksum` runs before any downloaded binary is used, inside `Downloader.FetchVerified` |
| Tampered CLI self-update | `openframe update` replacing the running binary | `internal/shared/selfupdate/cosign.go` performs cosign keyless signature verification before applying an update; `internal/shared/selfupdate/rollback.go` exists for recovery if something still goes wrong |
| Operating on the wrong cluster | Any `app` subcommand, especially `uninstall` | `noPositionalArgs` in `cmd/app/app.go` forces explicit `--context` targeting instead of silently falling back to whatever kube-context happens to be active |
| Silently retrying a non-idempotent, already-completed operation | Install/upgrade wait loops | `NonRetryableError` interface in `internal/shared/errors/retry_policy.go` lets an error explicitly veto the substring-based retry heuristic |
| Floating infrastructure module versions causing unreviewed behavior changes | Terraform modules for EKS/GKE | `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf` pin exact module versions (not ranges) — the EKS template's comments document a real incident where a floating `~> 21.0` constraint silently changed defaults (no bootstrapped addons, IMDS hop limit 1, gated Kubernetes-version data source) and broke cluster creation three independent ways |
| Unencrypted or exposed cloud cluster endpoints | GKE node networking | `internal/cluster/providers/gke/templates/main.tf` provisions GKE nodes as **private** (`enable_private_nodes = true`, no external IPs) behind Cloud NAT for egress, specifically to satisfy orgs enforcing `constraints/compute.vmExternalIpAccess`; the control-plane endpoint stays public only so the operator's own machine can reach it without a VPN |
| Dropped cloud resources after `terraform destroy` | EBS volumes provisioned outside Terraform state by the AWS EBS CSI driver | `main.tf` (EKS) tags CSI-provisioned volumes with the cluster name specifically so the CLI's post-destroy orphan sweep (`internal/cluster/providers/eks/orphan.go`, called from `teardown.go`) has a handle to find and remove them |

### Gosec suppressions must document the actual safety property

The one `#nosec` annotation present in this codebase (`internal/cluster/discovery/auth.go`, on the `gcloud` login subprocess) is accompanied by a comment stating exactly why the suppression is safe: "fixed argv assembled by this package, no user input." Any new `#nosec` suppression added to this repository should follow the same pattern — state the specific property that makes the flagged pattern safe here, not just silence the linter.

## 5. Security testing and code review guidelines

This repository has no dedicated `docs/development/testing/security` suite; security-relevant behavior is exercised through the same unit and integration tests as the rest of the CLI, under `tests/testutil/` and `tests/integration/common/`. When changing anything described above, keep these practices:

- **Prefer structural error classification over string matching** when adding retry or error-handling logic (see the `classifyTransient` discussion above) — string-matching heuristics over subprocess stderr are a known source of subtle bugs in this codebase and should be a last resort, clearly commented.
- **Never put secrets in URLs, command-line arguments, or struct fields that get logged.** Follow the `gitAuth` pattern: parse out the credential, keep a clean value for anything that might be printed or compared, and pass the credential only through an in-memory auth object.
- **Route all shelled-out commands through `internal/shared/executor`**, not direct `os/exec` calls, so that redaction (`redact.Redact`) and consistent error wrapping apply uniformly.
- **Any new `#nosec` suppression needs an inline justification** describing why the flagged pattern cannot be exploited here (fixed argv, no user input reaching the command line, etc.).

This repository's active review rules (enforced automatically) include several that bear directly on security-sensitive code:

- **Errors must be wrapped with context at every call boundary** (`fmt.Errorf("...: %w", err)`), so that a failure inside, say, the Git auth flow or the cosign verification path carries enough context to debug without re-exposing internals as a raw error string.
- **External-command exit codes must be preserved through `executor.CommandError`, not collapsed to 1** — this matters for security tooling in CI pipelines that branch on exit codes (e.g. a failed checksum or signature verification must not look like a generic failure).
- **Errors already displayed to the user must be wrapped with `sharedErrors.AlreadyHandledError`** before returning to `main`, so authentication/authorization failures surfaced interactively (e.g. a failed `gcloud` login) are not double-printed.
- **User-facing output must go through `pterm` printers or `internal/shared/ui`**, never raw `fmt.Print*` — this keeps `--silent` and `--plain` modes from accidentally leaking output (including anything not yet redacted) that the operator expected to be suppressed.
- **A new or changed exported definition should be referenced by a test** — applies to any new credential-handling or verification helper added to packages like `internal/shared/redact`, `internal/chart/providers/git`, or `internal/shared/selfupdate`.

Before changing files in `internal/chart/providers/git/`, `internal/shared/redact/`, `internal/shared/errors/`, or similar security-relevant paths, check the active rules for the exact files you intend to touch — the full catalog is available via the organization's code-rules tooling and is kept current with the repository.

## 6. Environment variables and secrets management

OpenFrame CLI reads a small set of environment variables to control behavior, and relies on ambient tool configuration (kubeconfig, `gcloud` config, AWS credential files) for the actual secrets it needs to operate:

- **`internal/shared/config/envbool.go`** exposes `EnvBool(name string)`, a helper for reading boolean feature-flag-style environment variables. This is the pattern to use for any new environment-driven toggle rather than parsing `os.Getenv` ad hoc.
- **`internal/shared/config/transport.go` and `internal/shared/config/system.go`** hold additional environment-derived configuration for HTTP transport and system behavior.
- The CLI does **not** require or read a long-lived API key or secret of its own — it has no backend to authenticate to. Instead, it depends on the credential state already established for the tools it drives:
  - Kubernetes: the active kubeconfig (location resolved via the standard `KUBECONFIG` convention and `internal/k8s/contexts.go`).
  - GKE: `gcloud`'s own credential store, refreshed interactively via `internal/cluster/discovery/auth.go` when stale or absent.
  - EKS: the AWS CLI/SDK credential chain, checked by `internal/cluster/prerequisites/aws/aws.go`.
  - Git (app-of-apps repository): a token embedded in the clone URL when a private repository is used, handled per the `gitAuth` pattern above; the default OSS deployment uses a public repository and needs no token at all.

> Never commit credentials, kubeconfig files, or cloud service-account keys into this repository or into chart value files. If you need to pass a private Git repository URL with a token to the CLI, supply it the same way `extractGitAuth` expects to receive it (embedded in the URL at invocation time), and rely on the CLI's existing stripping/redaction rather than printing or storing it elsewhere.

For environment setup instructions (which tools must be installed and how configuration flows into the CLI), see the development environment guide. For the broader system architecture this security model sits inside, see the architecture overview and the ecosystem diagram.
