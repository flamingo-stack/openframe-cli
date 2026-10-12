# Security Best Practices

OpenFrame CLI is a privileged tool: it shells out to Docker, cloud CLIs, and Terraform, downloads and executes third-party binaries, and handles ArgoCD admin credentials. This page covers the security patterns already established in the codebase and what to follow when extending it.

## Authentication and Authorization Patterns

OpenFrame CLI does not implement its own authentication system — it delegates to the credential/context mechanisms of the tools and clusters it orchestrates:

- **Kubernetes access** is resolved through kubeconfig contexts (`internal/k8s/contexts.go`, `internal/k8s/restconfig.go`), either the current context or an explicit `--context` flag. There are no OpenFrame-specific credentials stored for cluster access.
- **ArgoCD admin credentials** are read directly from the cluster's `argocd-initial-admin-secret` (via `internal/chart/providers/argocd`) — never generated, stored, or transmitted by the CLI itself. `openframe app access` reads and displays them; it does not create them.
- **Cloud provider authentication** (AWS/GCP) reuses the user's existing AWS CLI / gcloud CLI configuration and credential chain; the CLI does not manage its own cloud credentials.
- **CLI self-update authenticity** is verified via Sigstore/cosign keyless signatures (`internal/shared/selfupdate/cosign.go`). The OIDC issuer is pinned to `https://token.actions.githubusercontent.com` and the certificate SAN is pinned to this repository's `release.yml` workflow — signatures from any other repository, workflow, or issuer are rejected.

## Secure Handling of Credentials and Secrets

- **Never print secrets in the clear where avoidable.** `internal/shared/redact` provides `RegisterSecret()`/`Redact()` to scrub known sensitive values (and any `user:pass@` URL-embedded credentials) from log/debug output before it reaches the terminal.
- **ArgoCD passwords are redacted in error/debug paths.** When adding new logging around ArgoCD or Helm operations that might include secret values, register those values with `redact.RegisterSecret()` first.
- **No secrets are persisted to disk by the CLI.** Kubeconfig and cloud CLI credential files are managed by their respective tools (kubectl, aws, gcloud), not by OpenFrame CLI.

> When writing new code that handles a password, token, or API key, register it with `internal/shared/redact.RegisterSecret()` immediately after it's obtained, before it can appear in any log line, error message, or `--verbose` output.

## Verified, Integrity-Checked Downloads

A core security principle in this codebase: **no unsafe `curl | bash` style installs.**

- `internal/shared/download` (`verify.go`) implements `Downloader.FetchVerified`/`InstallVerified*`, which:
  - Pins every prerequisite tool (k3d, Helm, mkcert, Terraform, infracost) to an exact version and per-platform SHA256 checksum (`PinnedTool`/`PinnedAsset`).
  - Verifies `sha256(data)` against the expected digest before any bytes are trusted.
  - Writes files atomically (temp file + rename) so a failed/interrupted download never leaves a partial or unverified binary in place.
  - Caps downloads (512 MiB) and archive member extraction (200 MiB) to guard against decompression bombs.
- This pattern replaced earlier `curl | bash`-style installers as part of a security audit (see notes in the k3d installer). **Any new prerequisite installer should use this same verified-download path**, not a raw shell pipe to an installer script.
- CLI self-updates go through the same discipline, plus cosign signature verification (see above) — a checksum match alone is not sufficient for self-update binaries.

## Input Validation and Sanitization

- **Cluster names are validated** before reaching any shell-out (`internal/cluster/models.ValidateClusterName`), enforcing DNS-1123-like rules (max 63 chars, alphanumeric/hyphen, must start/end alphanumeric). This is enforced consistently at command boundaries (e.g., `bootstrap`, `cluster create`) so unsafe input never reaches downstream tool invocations.
- **Flag combinations are validated up front** (`internal/cluster/models.ValidateCreateFlags` and friends) — e.g., rejecting `--project` with `--type eks`, or requiring `--region`/`--project` for GKE when `--skip-wizard` is set — so ambiguous or cross-provider flag combinations are rejected before any cloud resources are touched.
- **All external commands use structured argv, not shell strings.** The `CommandExecutor` abstraction (`internal/shared/executor`) executes commands with discrete argument slices rather than interpolating user input into a shell string, which prevents shell-injection via crafted flag values.

## Testing for Shell-Injection and Command Safety

The test mock executor (`internal/shared/executor.MockCommandExecutor`) exposes `Commands()`, returning structured `RecordedCommand{Args, Env, Stdin}` records specifically so tests can assert no shell metacharacters (e.g., `$(...)`) were ever passed as a literal argument:

```go
for _, cmd := range exec.Commands() {
    for _, arg := range cmd.Args {
        if strings.Contains(arg, "$(") {
            t.Errorf("shell injection in argv: %q", arg)
        }
    }
}
```

> Prefer `Commands()` (structured argv) over `GetExecutedCommands()` (flattened strings) whenever a test's purpose is a security assertion — a flattened log cannot distinguish a literal `$(x)` argument from a shell-constructed string.

## Common Vulnerabilities and Mitigations

| Risk | Mitigation in this codebase |
|---|---|
| Unsafe binary installs (`curl \| bash`) | Replaced with pinned-version, SHA256-verified downloads (`internal/shared/download`) |
| Tampered/malicious self-update binary | Sigstore/cosign signature verification pinned to this repo's release workflow (`internal/shared/selfupdate/cosign.go`) |
| Shell injection via cluster names/flags | Strict cluster-name validation + structured argv execution (no shell string interpolation) |
| Secrets leaking into logs/`--verbose` output | Centralized `internal/shared/redact` registration/scrubbing |
| Decompression bombs from downloaded archives | Size-capped extraction (200 MiB) in `internal/shared/download` |
| Partial/corrupted files from interrupted downloads | Atomic write (temp file + rename) in `writeFileAtomic` |

## Emergency Escape Hatches (Use With Caution)

`OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY=1` bypasses cosign signature verification during self-update. This exists only as an emergency fallback (e.g., trust-root fetch outage) — it must never be recommended in normal documentation, scripts, or defaults, and should not be set in CI pipelines that fetch untrusted releases.

## Secrets and Environment Variables in CI/Release

The release signing pipeline (`scripts/sign-binary.sh`) is a good reference for how secrets should be handled in automation:

- It is a strict no-op unless `OPENFRAME_SIGN=1` is explicitly set by the release workflow — local builds and CI compile checks never attempt to sign or touch signing secrets.
- macOS signing/notarization and Windows Authenticode signing (Azure Trusted Signing) both require their credentials (Apple ID/team, Azure tenant/client secret) to be provided as environment variables sourced from the CI secret store — never hard-coded, and each `: "${VAR:?}"` guard fails fast if a required secret is missing rather than silently proceeding.
- Tokens (e.g., the Azure AAD token) are fetched fresh per invocation rather than cached long-lived, limiting the blast radius of a leaked token.

## Code Review Guidelines

When reviewing changes touching security-sensitive areas, confirm:

- [ ] New secrets/credentials are registered with `redact.RegisterSecret()` before any logging path can reach them.
- [ ] New external tool installers use `internal/shared/download`'s verified-download path, not a raw shell pipe.
- [ ] New shelled-out commands go through `CommandExecutor` with argv slices, not interpolated shell strings.
- [ ] New user-supplied identifiers (names, refs, paths) are validated before being passed to any external command.
- [ ] Nothing writes cloud or cluster credentials to disk outside of the standard tool-managed locations (kubeconfig, AWS/gcloud config).
