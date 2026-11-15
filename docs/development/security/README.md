# Security

This document covers how the OpenFrame CLI handles authentication to external systems, how it stores and transmits secrets, how it validates input, which classes of vulnerability its architecture defends against, how security is reviewed and tested, and how environment variables and secrets should be managed when developing or operating it.

The OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is a Go command-line tool. It has no server component of its own, no HTTP API, and no database: its entire attack surface is (1) the credentials and tokens it handles to talk to Docker, Kubernetes, cloud provider APIs and git, and (2) the binaries it downloads and executes on the operator's machine. This document is scoped accordingly — the authentication/authorization patterns below are about the CLI *authenticating to* external systems, not about an application login system, because the CLI does not expose one.

> For the organization-wide vulnerability disclosure process (how to report a vulnerability, response SLAs, scope and supported versions), see the `SECURITY.md` file at the root of this repository. That policy applies to every repository in the flamingo-stack organization, including this one.

## 1. Authentication and authorization patterns

The CLI never implements its own authentication scheme. Instead it defers to, and re-uses, the authentication already configured for each external system it drives:

| System | How the CLI authenticates | Where |
|---|---|---|
| Kubernetes | Existing kubeconfig context (`internal/k8s/`), or a `rest.Config` handed back by a cluster provider | `internal/k8s/accessor.go`, `internal/k8s/restconfig.go`, `internal/k8s/contexts.go` |
| Docker (k3d) | Local Docker daemon socket, same as any other Docker CLI user | `internal/cluster/prerequisites/docker/docker.go` |
| AWS (EKS) | The caller's AWS CLI identity/profile; Terraform's AWS provider inherits it | `cmd/cluster/aws_identity.go`, `internal/cluster/prerequisites/aws/aws.go`, `internal/cluster/providers/eks/` |
| GCP (GKE) | `gcloud` login + Application Default Credentials (ADC), with an interactive login flow when missing | `internal/cluster/discovery/auth.go`, `internal/cluster/discovery/gke.go` |
| Git (app-of-apps chart repo) | HTTP Basic Auth built in-memory from a token, never written to disk or argv | `internal/chart/providers/git/auth.go` |
| GitHub Releases (self-update) | Unauthenticated HTTPS, verified by checksum + Sigstore/cosign signature instead of a credential | `internal/shared/selfupdate/release.go`, `internal/shared/selfupdate/cosign.go` |
| ArgoCD (deployed platform) | CLI prints the generated admin password and a `kubectl port-forward` command; the user authenticates directly against ArgoCD's own login | `cmd/app/access.go` |

### The gcloud auth flow

`internal/cluster/discovery/auth.go` implements `AuthFlow`, the single login path every GKE-touching command goes through. It is deliberately conservative:

- It **never prompts in a non-interactive session** (CI rule) — `interactive()` checks `sharedUI.IsNonInteractive()`, and a non-interactive run gets an actionable error (e.g. `"gcloud is not authenticated — run 'gcloud auth login'"`) instead of hanging on a prompt.
- It distinguishes "not authenticated" from "authenticated but token expired": `gcloud auth list` reports an expired account as still `ACTIVE`, so `AuthFlow.hasFreshLogin` forces a real token refresh via `gcloud auth print-access-token --quiet` rather than trusting the account list.
- Application Default Credentials (needed by Terraform's `google` provider) are checked and requested **separately** from the primary login, because `cluster create --type gke` needs both while `cluster use`/discovery only need the first (`AuthFlow.Ensure(ctx, requireADC bool)`).
- The actual `gcloud auth login` subprocess is run attached to the real terminal (`os.Stdin`/`os.Stdout`/`os.Stderr`), not through the capturing executor, because the browser-based login flow needs to read and print directly.

### Authorization is delegated, not modeled

The CLI has no concept of users, roles or permissions of its own. Whatever the operator's kubeconfig, AWS profile, or gcloud account is authorized to do, the CLI can do — nothing more, nothing less. This is intentional: `internal/cluster/providers/eks/templates/main.tf` sets `enable_cluster_creator_admin_permissions = true` precisely so that the identity running Terraform is the same identity later used for `kubeconfig` auth via `aws eks get-token`, avoiding a second IAM/RBAC wiring step to manage.

When documenting or extending a command, treat "is the current context/profile allowed to do X" as a question for the underlying platform (Kubernetes RBAC, AWS IAM, GCP IAM) to answer, not something to re-implement in the CLI.

### Defensive command construction

`cmd/app/app.go` shows a recurring pattern worth following in new commands: the `app` command group installs its own `noPositionalArgs` validator instead of relying on cobra's default `ArbitraryArgs`. The comment in the source explains why this matters for authorization-adjacent safety: without it, `openframe app uninstall my-eks` would silently run against whatever the *current* kube-context happens to be, not the cluster named on the command line — for `uninstall`, that is a destructive mistake, not just a confusing one. Any new subcommand that targets a specific cluster should validate its arguments as strictly.

## 2. Data encryption and secure storage

The CLI does not run a datastore and does not persist application data. Its "storage" surface is: local config files, the Terraform workspace it manages, and the pinned binaries it downloads.

- **Downloaded binaries are integrity-verified, not encrypted.** `internal/shared/download/verify.go` exposes `VerifyChecksum(data []byte, wantHex string)`, called from `Downloader.FetchVerified`. Every tool the CLI fetches for itself — k3d, Helm, mkcert, Terraform, infracost — is downloaded, hashed, and compared against a pinned SHA-256 digest recorded in `internal/shared/download/pins.go` before it is ever executed. A mismatch aborts the install; nothing is silently trusted because it came from "the internet."
- **Self-update binaries are signature-verified.** `internal/shared/selfupdate/cosign.go` verifies the CLI's own release artifacts with Sigstore/cosign keyless signatures, anchored via TUF to the public-good Sigstore trust root, before an in-place binary replacement (`internal/shared/selfupdate/update.go`) or rollback (`internal/shared/selfupdate/rollback.go`) occurs. No binary is installed on checksum alone during self-update — both the digest and the signature must check out.
- **Terraform state and cloud credentials are the Terraform/cloud SDK's own responsibility.** The CLI shells out to Terraform (`internal/cluster/providers/terraform/`) and to the `aws`/`gcloud` CLIs; it does not read, copy, or re-store their credential files. Treat any `~/.aws`, `~/.kube`, or `~/.config/gcloud` content as out of this CLI's custody — it is passed through, not duplicated.
- **Git credentials are kept in memory only.** See `internal/chart/providers/git/auth.go` in detail below — this is the one place in the codebase where a secret token is actively handled, and it is deliberately never written to disk.

### The git-token handling pattern (read before touching `auth.go`)

`internal/chart/providers/git/auth.go` is the canonical example of "no secret touches disk or argv" in this codebase. Key points from the file, since any future git-auth change should preserve all of them:

```go
// gitAuth holds credentials extracted from a clone URL together with the
// cleaned, credential-free URL.
//
// Keeping the token OUT of the URL is what keeps it off the process command
// line and out of any on-disk file (audit I1): it is handed to go-git only via
// an in-memory HTTP basic-auth method (see buildAuth), so it never appears in
// argv (visible to `ps`/`/proc/<pid>/cmdline`), in a credentials file, nor in
// process/error output.
```

`extractGitAuth(rawURL string)` parses any embedded userinfo out of the clone URL (including the common single-field GitHub PAT shorthand, `https://<token>@host`, which it recognizes as a token-only credential rather than silently dropping it), strips it from the URL, and returns a `gitAuth` struct carrying the token only in memory. `buildAuth()` turns that into a `go-git` `http.BasicAuth` method handed directly to the clone call — the token is never interpolated back into a URL string, never passed as a CLI argument, and never written to a credentials file. `maskToken(s, token string)` is provided specifically so that any log or error message that might echo the clone URL has the token replaced with `***` before it is ever printed.

If you add a new git operation, route its auth through this same `gitAuth`/`buildAuth` pair rather than constructing a `user:token@host` URL by hand — that is exactly the pattern this file exists to avoid.

### Redaction at the executor boundary

Secrets can also leak through command output, not just through storage. `internal/shared/redact/redact.go` exposes `Redact(s string)`, called from two places:

- `internal/chart/providers/helm/manager.go` (`HelmManager.InstallArgoCDWithProgress`) — scrubbing Helm's own streamed output before it reaches the terminal.
- `internal/shared/executor/executor.go` (`RealCommandExecutor.ExecuteWithOptions`) — scrubbing the output of *every* shelled-out command (Docker, k3d, Helm, Terraform, gcloud, AWS CLI) that runs through the executor abstraction.

Because redaction is wired into the executor itself rather than into each individual call site, any new command that goes through `executor.CommandExecutor` gets this scrubbing for free. Prefer the executor abstraction over a raw `os/exec` call for exactly this reason — see `internal/shared/executor/executor.go` and its mock in `internal/shared/executor/mock.go`, used so tests never need network access or real credentials.

## 3. Input validation and sanitization

Because the CLI's primary job is assembling arguments for other programs (`docker`, `k3d`, `helm`, `terraform`, `gcloud`, `aws`) and for Kubernetes/git/HTTP clients, its validation concerns are different from a typical web application: the goal is to stop user-supplied strings from being interpreted as additional shell arguments, additional Terraform resources, or a redirect to an unintended destination.

Patterns to follow, found in the material:

- **Fixed argv, no shell interpolation.** `internal/cluster/discovery/auth.go` builds the `gcloud` login subprocess with `osexec.CommandContext(ctx, "gcloud", args...)`, annotated `#nosec G204 -- fixed argv assembled by this package, no user input`. The argument list is constructed by the package itself from known-safe literals (`"auth", "login"`, `"auth", "application-default", "login"`) — never by splicing user input into a command string. Any new subprocess invocation should follow the same shape: a fixed verb plus a `[]string` built from validated, typed values, never a string passed to a shell.
- **Required positional/flag validation before action.** `cmd/app/app.go`'s `noPositionalArgs` function is a validation gate run in `PersistentPreRunE`, rejecting any unexpected positional argument with an error that names the correct flag (`--context`) to use instead. This both prevents a confusing silent fallback and closes the "ran against the wrong cluster" class of mistake described above.
- **Pinned, not operator-supplied, download sources and digests.** `internal/shared/download/pins.go` hard-codes the versions and SHA-256 digests for every tool the CLI downloads. The URL and digest to fetch are not assembled from user input at runtime, which removes an entire class of "fetch whatever this flag says and trust it" vulnerability.
- **Config validation is explicit, even when currently a no-op.** `internal/chart/services/validation.go`'s `ConfigurationValidator.ValidateConfiguration` is the dedicated seam for validating `openframe-helm-values.yaml` in non-interactive mode. Today it returns `nil` because the CLI only supports the OSS (`oss-tenant`) deployment, which uses a public repository and needs no credentials — but the seam exists precisely so that if a future deployment mode accepts credentials or custom values, validation has a single, already-wired place to live rather than being bolted on ad hoc.
- **Terraform inputs are typed, not interpolated strings.** Both `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf` declare every CLI-supplied value (`cluster_name`, `region`, `instance_type`, node counts, etc.) as a typed Terraform `variable`, passed through `terraform.tfvars.json`. There is no string-built HCL; malformed or unexpected values fail Terraform's own type checking before `apply` runs.

When adding a new flag that flows into a subprocess call, a Terraform variable, or a file path, validate its shape (allowed characters, expected format) before it reaches the executor or template layer, and prefer extending the existing `models.ClusterType`/flag-parsing structures in `internal/cluster/models/flags.go` and `internal/shared/flags/flags.go` over inventing a new ad hoc parser.

## 4. Common security vulnerabilities and mitigations

| Vulnerability class | How it could show up here | Mitigation already in place |
|---|---|---|
| Command injection | A flag value spliced into a shell string before exec | Fixed argv construction (`osexec.CommandContext(ctx, "gcloud", args...)`), `#nosec G204` only where argv is provably fixed; all other subprocesses go through `internal/shared/executor` |
| Secret leakage via logs/output | A token echoed in a cloned-URL error, a verbose Helm/Terraform log line | `redact.Redact` wired into `RealCommandExecutor.ExecuteWithOptions` and `HelmManager.InstallArgoCDWithProgress`; `maskToken` in `internal/chart/providers/git/auth.go` |
| Secret leakage via process list or disk | A `user:token@host` URL passed as a CLI argument or written to a credentials file | `gitAuth`/`buildAuth` keep the token in memory only, handed to `go-git` as an `http.BasicAuth` value, never interpolated into a URL string |
| Supply-chain compromise of a downloaded tool | A compromised or tampered k3d/Helm/Terraform/mkcert/infracost binary | SHA-256 pinning (`internal/shared/download/pins.go`) verified by `VerifyChecksum` before any pinned binary is used; never fetched via an unverified `curl \| bash` |
| Supply-chain compromise of the CLI's own binary | A tampered `openframe` release artifact during self-update | Sigstore/cosign signature verification (`internal/shared/selfupdate/cosign.go`) in addition to checksum verification, anchored to the public Sigstore trust root |
| Wrong-cluster / wrong-context destructive action | A command silently operating on the current kube-context instead of the named target | `noPositionalArgs` validation in `cmd/app/app.go`; `ui.ShowContextHeader()` prints the active kube-context before any destructive subcommand runs |
| Retrying a destructive or already-failed operation inappropriately | A transient-looking error string inside pod logs/events triggering an unwanted retry of an already-applied install | `NonRetryableError` interface in `internal/shared/errors/retry_policy.go`, checked *before* the substring-based transient classification, so an explicit "never retry" marking always wins |
| Overbroad cloud IAM surface from a floating module version | A Terraform module version bump silently changing security-relevant defaults (e.g. IMDS hop limit, default addons) | Exact (non-range) pins on the `terraform-aws-modules/eks/aws` and `terraform-aws-modules/vpc/aws` module versions in `internal/cluster/providers/eks/templates/main.tf`, bumped deliberately with a full create/delete pass, per the comment in that file |
| IMDS credential exposure to workload pods | A node's instance-metadata service reachable with a weak hop limit, letting a pod on an overlay network reach node credentials unexpectedly | `metadata_options.http_tokens = "required"` (IMDSv2 only) with a deliberately chosen `http_put_response_hop_limit = 2` in `internal/cluster/providers/eks/templates/main.tf`, scoped to let the EBS CSI controller authenticate without granting broader access |
| Public exposure of a cluster's nodes | GKE nodes with public IPs reachable directly from the internet | `enable_private_nodes = true`, `enable_private_endpoint = false` plus a dedicated Cloud NAT for egress in `internal/cluster/providers/gke/templates/main.tf` — nodes have no external IP, only the API endpoint remains public so the operator's machine can reach it without a VPN |
| Exit-code laundering masking a failed external command | A failed `helm`/`terraform` invocation reported to automation as a generic failure (exit 1) instead of its real code | `executor.CommandError` preserves the original exit code end-to-end; `main.exitCode()` unwraps it via `errors.As` so CI can distinguish failure types (rule `OPENFRAM-005`) |

### A note on `#nosec` annotations

The codebase does use `#nosec` to suppress a static-analysis finding in exactly one place shown in this material (`internal/cluster/discovery/auth.go`, `G204`), and does so with an inline justification ("fixed argv assembled by this package, no user input"). Any new `#nosec` suppression should carry the same kind of specific, verifiable justification — not a blanket silencing of the linter — so a reviewer can confirm the claim against the surrounding code.

## 5. Security testing and code review guidelines

### Active review rules

This repository's automated reviewer enforces a set of structural rules that double as security guardrails even though most are framed as architecture conventions. The ones most relevant to security-sensitive changes:

- **OPENFRAM-005** — external-command exit codes must be preserved through `executor.CommandError`, never collapsed to a generic `1`. This matters for automation correctness and for not hiding *which* external tool failed and why.
- **OPENFRAM-004** — an error that has already been shown to the user must be wrapped in `sharedErrors.AlreadyHandledError` before it propagates to `main`, so it is not printed twice. Relevant to security review because a double-printed error can obscure the actual failure reason (including a security-relevant one) in long output.
- **OPENFRAM-007** — all user-facing output must go through `pterm` printers or `internal/shared/ui`, never raw `fmt.Print*` to stdout. This is what lets `--silent` and `--plain` modes, and the redaction machinery, apply consistently; a stray `fmt.Println` of command output is exactly the kind of gap that could leak an unredacted secret.
- **FLEETMDM-002-2** — every error returned from a called function must be wrapped with `fmt.Errorf("...: %w", err)` at the call boundary. Preserving context on errors matters for diagnosing security-relevant failures (auth failures, verification failures) without having to reproduce them.
- **FLEETMDM-001** — any edit to a file that originates from upstream `fleetdm/fleet` must be wrapped in `// >>> OPENFRAME(<slug>): <why>` / `// <<< OPENFRAME(<slug>)` sentinel comments, naming a doc under `openframe/docs/`. This applies to the files in this security-sensitive area (`internal/chart/providers/git/auth.go`, `internal/shared/redact/redact.go`, `internal/shared/selfupdate/cosign.go`, `internal/shared/download/verify.go`) and exists so that upstream syncs do not silently clobber fork-specific security logic such as the git-token handling or checksum verification described above.

When changing any file under `internal/chart/providers/git/`, `internal/shared/redact/`, `internal/shared/selfupdate/`, or `internal/shared/download/`, check these rules deliberately before opening a PR — they are the review rules the automated reviewer applies to exactly that area.

### Testing patterns to reuse

- **Mock the executor, never shell out in tests.** `internal/shared/executor/mock.go` provides a mock `CommandExecutor` so unit and integration tests exercise command-construction logic (including the fixed-argv patterns above) without actually invoking `docker`, `helm`, `terraform`, `gcloud` or `aws`, and without needing real credentials in CI.
- **Integration tests drive the real CLI binary.** `tests/integration/common/cli_runner.go` and `tests/integration/common/cluster_management.go` show the pattern for end-to-end coverage of cluster lifecycle commands; `tests/integration/common/dependencies.go` manages prerequisite tool state for those runs.
- **Shared assertion/test helpers.** `tests/testutil/assertions.go`, `tests/testutil/command_assertions.go`, and `tests/testutil/flag_contract.go` centralize common checks (including flag validation contracts) — extend these rather than duplicating assertion logic when adding tests for a new security-relevant flag or command guard.
- **Retry-policy classification is unit-testable in isolation.** Because `classifyTransient` in `internal/shared/errors/retry_policy.go` is a pure function over error types (not a string match), and the `NonRetryableError`/`RecoverableError` interfaces are explicit contracts, tests can assert retry/no-retry behavior for a given error shape directly — valuable for confirming that a failed signature or checksum verification is never silently retried into a bypass.

For broader testing conventions beyond the security-specific points above, see this repository's testing documentation.

## 6. Environment variables and secrets management

The CLI itself does not require application secrets to run its default (OSS/k3d) path — `ConfigurationValidator.ValidateConfiguration` in `internal/chart/services/validation.go` is a no-op precisely because the OSS tenant deployment clones a public repository and needs no credentials. Secrets enter the picture only when:

1. **A private chart/app-of-apps repository is used.** In that case a git token is supplied to `internal/chart/providers/git/auth.go` via the clone URL's userinfo (or the `https://<token>@host` shorthand) and is handled exactly as described in section 2 above — extracted once, held in memory, never written to disk, masked in any output via `maskToken`.
2. **Cloud provider credentials are needed (EKS/GKE).** These are never read or stored by the CLI directly. For AWS, the CLI relies on the ambient AWS CLI/SDK credential chain (profiles, environment variables such as standard AWS credential variables, or instance/role credentials) consumed directly by Terraform's `aws` provider and the `aws` CLI (`internal/cluster/prerequisites/aws/aws.go`, `cmd/cluster/aws_identity.go`). For GCP, `internal/cluster/discovery/auth.go` drives `gcloud auth login` and `gcloud auth application-default login` so credentials live in the standard gcloud credential store, not in a CLI-managed file.
3. **ArgoCD admin access.** `cmd/app/access.go` retrieves and prints the ArgoCD-generated admin password for the deployed cluster along with the `kubectl port-forward` command needed to reach it. This password is generated and stored by ArgoCD itself (inside the cluster, as a Kubernetes Secret) — the CLI only reads and displays it; refer to your environment's ArgoCD configuration rather than a document like this one for the current value.
4. **Signing the CLI's own release artifacts.** `scripts/sign-binary.sh` is part of the release pipeline that produces the signed binaries verified by `internal/shared/selfupdate/cosign.go` at update time. Day-to-day development and use of the CLI does not require running this script or holding its signing material — it is a release-time concern, not a runtime one.

### Practical guidance for contributors

- Never hard-code a token, password, or API key into a `_test.go` file, a fixture, or a committed Terraform `.tfvars`. Use the `internal/shared/executor` mock and `tests/testutil` helpers so tests do not need real credentials at all.
- If a change needs to pass a new secret through to a subprocess or an HTTP client, follow the `gitAuth` pattern: parse/accept it once, hold it only in memory, hand it to the library via an in-memory auth object, and never place it in a URL, a file, or a subprocess argument list.
- If a change adds new output that might echo a URL, header, or environment value that could carry a credential, route it through `internal/shared/redact` before it reaches `internal/shared/executor` or a `pterm` printer, following the existing call sites in `helm/manager.go` and `executor.go`.
- Treat any new pinned download (a new tool the CLI fetches for itself) as requiring an entry in `internal/shared/download/pins.go` with a verified SHA-256 digest — never add an unpinned or unverified download path.
- For cloud infrastructure changes, keep Terraform module versions as exact pins (as both `main.tf` templates already do) rather than version ranges, and re-run a full create/delete pass when bumping one, since a changed default (as documented in the EKS template's comments) can silently alter the security posture of the resulting cluster.

For environment setup unrelated to secrets (installing prerequisite tools, configuring a local development environment), see this repository's environment and local development setup documentation.
