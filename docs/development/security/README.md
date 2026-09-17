# Security Best Practices

OpenFrame CLI shells out to powerful external tools (Docker, Terraform, Helm, cloud CLIs) and self-updates by downloading and executing new binaries. Its security model focuses on **supply-chain integrity of downloaded artifacts**, **avoiding shell injection**, and **not leaking secrets in logs** — rather than traditional web-app authn/authz, since the CLI itself has no server component.

## Binary & Release Integrity

### Signed, Checksum-Verified Downloads

External CLI tools (`k3d`, `helm`, `terraform`, `mkcert`, `infracost`) are **never** installed via unverified `curl | bash`. Instead, `internal/shared/download` pins an exact version and SHA256 checksum per `GOOS`/`GOARCH` platform for each tool (see `internal/shared/download/pins.go`), verifies the checksum after download, and only then installs the binary into `~/.openframe/bin`.

### Self-Update Signature Verification

The `openframe update` command downloads new CLI releases from GitHub and verifies them with [Sigstore/cosign](https://www.sigstore.dev/) (`internal/shared/selfupdate/cosign.go`) before applying:

- The OIDC issuer is pinned to `https://token.actions.githubusercontent.com`.
- The certificate SAN is pinned to this repository's `release.yml` workflow (running on `main` or a tag ref).
- Signatures from any other repository, workflow file, or OIDC issuer are **rejected**.
- The Sigstore trust root is fetched via TUF and cached under `~/.openframe/state/tuf`.

> **Escape hatch:** `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY=1` bypasses signature verification. This exists only for emergency recovery scenarios and should never be used in normal operation or CI.

### Signed Release Binaries

Release binaries are additionally signed at the OS level via `scripts/sign-binary.sh`, invoked by the GoReleaser build post-hook (only when `OPENFRAME_SIGN=1`, set exclusively by the release workflow — local builds and CI compile checks never sign):

| OS | Signing Mechanism |
|---|---|
| macOS | `codesign` (Developer ID, hardened runtime) + `notarytool` notarization |
| Windows | Azure Trusted Signing via `jsign` (Authenticode) |
| Linux | Unsigned; integrity is covered by `checksums.txt` + the cosign bundle |

## Avoiding Shell Injection

All external command execution goes through `internal/shared/executor.CommandExecutor`, which invokes commands with **discrete argv arrays** (`os/exec`) rather than constructing shell strings — this prevents shell metacharacter injection (e.g., `$(...)`, backticks, `;`) from user-supplied input like cluster names or branch refs.

The test suite enforces this: `MockCommandExecutor.Commands()` returns structured `RecordedCommand` values with discrete `Args []string`, specifically so tests can assert that no argument contains shell metacharacters — see `internal/shared/executor/mock.go`.

Input that reaches shell-out boundaries is also validated early:

- Cluster names are validated with `models.ValidateClusterName` (DNS-1123 rules: max 63 chars, alphanumeric/hyphen, must start/end alphanumeric) **before** any provider code runs.
- `openframe bootstrap`'s cluster-name argument is validated at the command boundary specifically so "no unsafe input reaches downstream shell-outs."

## Secret Redaction

`internal/shared/redact` scrubs sensitive values from all log/debug output before it's printed:

- `redact.RegisterSecret(value)` registers a runtime secret (e.g., a fetched ArgoCD admin password) for redaction; values shorter than 4 characters are ignored to avoid over-redacting common substrings.
- `redact.Redact(s)` replaces all registered secrets with `***`, and unconditionally scrubs `user:pass@` patterns from URLs — catching credentials embedded in URLs that were never explicitly registered.
- Longer secrets are redacted before shorter ones, so a secret that happens to be a substring of another isn't partially unmasked.
- This is used whenever the CLI logs the exact external commands it runs (in `--verbose` mode), so credentials passed as CLI flags or embedded in URLs never leak into terminal output or CI logs.

## Credential & Config Handling

- **ArgoCD admin password**: retrieved on demand via `openframe app access` from the live cluster; never written to disk by the CLI.
- **Helm values**: `openframe-helm-values.yaml` (resolved via `internal/chart/utils/config.PathResolver.GetHelmValuesFile()`) lives in your working directory — treat it as sensitive if it contains secrets, and exclude it from version control if so.
- **TLS certificates**: local development certs (via `mkcert`) are stored under `~/.config/openframe/certs` with restrictive directory permissions (`0750`), falling back to a repo-relative path only if the user's home directory can't be resolved.
- **Cloud credentials**: AWS/GCP credentials are never handled directly by the CLI — it delegates to the already-authenticated `aws`/`gcloud` CLIs and Terraform's native provider auth.

## Destructive Operation Confirmation

Commands that delete resources (`app uninstall`, `cluster delete`) require interactive confirmation via `ui.RequireConfirmation` unless `--yes`/`--force` is passed. Confirmation-prompt errors are handled carefully (`internal/shared/errors/handler.go`):

- A Ctrl-C interruption is preserved as-is so the top-level error handler can print a clean cancellation message and exit gracefully — it is never misreported as a generic failure.
- In non-interactive/CI environments without a TTY, commands fail fast rather than hanging on a prompt that can never be answered.

## Common Vulnerabilities & Mitigations

| Risk | Mitigation |
|---|---|
| Shell injection via cluster names, refs, or flags | Argv-based execution (no shell string construction) + early input validation (`ValidateClusterName`) |
| Supply-chain compromise of downloaded tools | Pinned versions + SHA256 checksum verification (`internal/shared/download`) |
| Malicious/forged CLI update | Cosign signature verification pinned to this repo's release workflow OIDC identity |
| Secrets leaking into `--verbose` logs or CI output | Centralized `redact.Redact()` applied to all command logging |
| Destructive commands run accidentally in scripts | Explicit `--yes`/`--force` required to skip confirmation in non-interactive contexts |
| Unsafe cluster names reaching shell-outs | RFC1123-style validation enforced at the command boundary, before provider dispatch |

## Security Testing & Code Review Guidelines

- When adding a new external command invocation, use `internal/shared/executor.CommandExecutor` — never call `os/exec` directly, and never build a shell string with user input.
- When adding a new test around command execution, prefer asserting against `MockCommandExecutor.Commands()` (structured argv) over `GetExecutedCommands()` (flattened strings), since only the structured form can reliably catch injected shell metacharacters.
- When adding a new prerequisite installer, use `internal/shared/download`'s pinned/verified download pattern rather than introducing a new unverified download or `curl | bash` step.
- When logging anything that might contain a secret (passwords, tokens, credential-bearing URLs), route it through `redact.Redact()` first.
- Any change to the self-update signing/verification flow (`internal/shared/selfupdate/cosign.go`) should be reviewed with extra scrutiny, since it is the CLI's primary supply-chain trust boundary.

## Environment Variables & Secrets Management

| Variable | Sensitivity | Notes |
|---|---|---|
| `GITHUB_TOKEN` / `OPENFRAME_GITHUB_TOKEN` | Sensitive | Forwarded into WSL for authenticated GitHub API access; never logged in plain form |
| `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY` | Security-relevant | Disables cosign verification — restrict to emergency/manual use only |
| `OPENFRAME_NO_WSL_FORWARD` | Non-sensitive | Behavioral only |
| `OPENFRAME_WSL_DISTRO` | Non-sensitive | Behavioral only |

Cloud provider credentials (AWS/GCP) and Kubernetes credentials (kubeconfig) are managed entirely by the underlying `aws`, `gcloud`, and `kubectl`/client-go tooling already present on your machine — OpenFrame CLI does not introduce its own credential store.
