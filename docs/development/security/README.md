# Security Best Practices

This guide covers security considerations for developing, deploying, and operating OpenFrame CLI. It addresses credential management, TLS configuration, secrets handling, and secure development practices.

---

## Overview

OpenFrame CLI interacts with several sensitive systems:

- **Kubernetes clusters** — via kubeconfig and REST API
- **Docker registries** — with optional authentication
- **GitHub repositories** — via personal access tokens (PATs) or SSH
- **Helm chart repositories** — with potential authentication
- **ArgoCD** — via Kubernetes RBAC

Security is a shared responsibility between the CLI, the operator's environment, and the underlying platform.

---

## Authentication and Authorization

### Kubernetes Access

The CLI uses standard Kubernetes authentication via kubeconfig. It calls `clientcmd.BuildConfigFromFlags` to load the kubeconfig from the default location or the `KUBECONFIG` environment variable.

**Best practices:**

- Use separate kubeconfig files or contexts for each environment (dev, staging, production)
- Rotate kubeconfig credentials regularly
- Never commit kubeconfig files to version control
- Use RBAC-scoped service accounts for CI/CD pipelines — avoid using admin credentials in automated workflows

```bash
# Use environment variable to specify kubeconfig
export KUBECONFIG="$HOME/.kube/openframe-dev-config"
openframe cluster status my-cluster
```

### GitHub Credentials

When the chart installation requires cloning private repositories, the CLI prompts for GitHub credentials via the `CredentialsPrompter`:

```text
Enter your GitHub credentials for https://github.com/org/private-repo:
Username: your-github-username
Token: ****  (personal access token)
```

**Best practices:**

- Use **GitHub Personal Access Tokens (PATs)** with the minimum required scopes — typically `repo` (read-only) is sufficient for chart cloning
- Use **fine-grained PATs** scoped to specific repositories where possible
- Rotate PATs regularly (every 90 days recommended)
- Never hardcode PATs in scripts or environment files that get committed

### Docker Registry Credentials

For deployments using private Docker registries:

```text
Enter Docker registry credentials for registry.example.com:
Username: your-username
Password: ****
```

**Best practices:**

- Use short-lived registry credentials (e.g., via `docker login` with token-based auth)
- Prefer image pull secrets in Kubernetes over embedding credentials in Helm values
- Use read-only registry credentials for deployments

---

## Secrets Management

### Environment Variables vs. Flags

The CLI never stores credentials on disk between sessions. Credentials are:
1. Prompted interactively at runtime
2. Passed directly to the tool that needs them
3. Never written to log files or stdout (passwords are masked)

**What to avoid:**

```bash
# NEVER DO THIS — credentials visible in shell history and process list
openframe chart install my-cluster --github-token=ghp_my_token
```

**Instead, use interactive prompts** which mask input and avoid shell history exposure.

### Protecting Shell History

If you must pass credentials via environment variables for CI/CD:

```bash
# Set in CI environment secrets, not in scripts
export GITHUB_TOKEN="$(vault kv get -field=token secret/github)"
```

Use your CI/CD platform's native secrets management:
- **GitHub Actions**: Repository Secrets (`${{ secrets.GITHUB_TOKEN }}`)
- **GitLab CI**: CI/CD Variables (masked)
- **Jenkins**: Credentials Store
- **ArgoCD**: Sealed Secrets or External Secrets Operator

---

## TLS and Certificate Security

### Local TLS with mkcert

The CLI uses `mkcert` to generate locally trusted TLS certificates for K3D clusters. This ensures HTTPS works correctly in local development without certificate warnings.

**How it works:**

1. mkcert installs a local Certificate Authority (CA) into your system trust store
2. The CLI generates a certificate signed by this local CA
3. The certificate is used for the Kubernetes API server and ingress

**Security considerations:**

- The local CA is trusted **only on your machine** — certificates are not valid externally
- Do not distribute local CA certificates to other machines
- For production deployments, use certificates from a trusted public CA (Let's Encrypt, your organization's PKI)

### TLS Configuration in Code

The CLI applies an insecure TLS configuration for local K3D cluster connections:

```go
// internal/shared/config/transport.go
// ApplyInsecureTLSConfig is used ONLY for local K3D clusters
// Never use InsecureSkipVerify in production connections
```

> **Warning**: `InsecureSkipVerify` is enabled only for local K3D clusters (127.0.0.1 / localhost). This is acceptable for local development but must never be used for remote or production clusters.

---

## Input Validation and Sanitization

### Cluster Name Validation

The CLI validates cluster names before use via `ValidateClusterName`. Names are checked for:
- Valid character sets (alphanumeric, hyphens)
- Length limits
- Reserved names

### Port Validation

Port numbers are validated before creating Telepresence intercepts:

```go
// Validation occurs in internal/dev/services/intercept/service.go
// validateInputs() checks port ranges (1-65535) and service name format
```

### Flag Sanitization

All user-provided flag values go through structured validation before being passed to external tools. The `ValidationError` type provides field-level error reporting:

```text
✗ Validation failed:
  Field: port
  Value: "invalid"
  Error: must be a valid port number (1-65535)
```

---

## Common Security Vulnerabilities and Mitigations

| Vulnerability | Risk | Mitigation |
|---|---|---|
| Credential exposure in logs | Passwords visible in verbose output | Passwords are always masked; never log credential values |
| Command injection | User input passed to shell commands | All external commands use `exec.Command` with args array — no shell expansion |
| Insecure TLS | MITM attacks on cluster connections | InsecureSkipVerify used only for `localhost`/`127.0.0.1` K3D clusters |
| Kubeconfig exposure | Cluster access via leaked kubeconfig | Never log kubeconfig paths; use scoped contexts; rotate credentials |
| Dependency vulnerabilities | CVEs in transitive Go dependencies | Run `govulncheck ./...` regularly; keep `go.sum` up to date |
| Path traversal in temp files | Reading files outside temp directory | Temp directory cleanup via `defer` in `git.Repository` |

---

## Secure Handling of Temporary Files

The CLI creates temporary files during chart installation (cloned repositories, temporary Helm values files). These are cleaned up via deferred functions:

```go
// internal/chart/providers/git/repository.go
// Shallow clone goes to os.MkdirTemp() — always cleaned up via defer os.RemoveAll(tempDir)
```

**Best practices for operators:**

- Ensure `/tmp` (or OS temp directory) is on an encrypted volume in sensitive environments
- Monitor for unexpected files persisting after CLI operations (may indicate cleanup failures)

---

## Security Testing

### Running Vulnerability Checks

```bash
# Check for known CVEs in Go dependencies
govulncheck ./...

# Check Go module dependencies for known vulnerabilities
go list -m all | nancy sleuth
```

### Code Review Security Checklist

When reviewing CLI changes, check for:

- [ ] No credentials or secrets hardcoded in source or test files
- [ ] All user-provided inputs are validated before use
- [ ] External commands use `exec.Command` with separate args (not shell string interpolation)
- [ ] Temporary files are created in OS temp directory and cleaned up with `defer`
- [ ] No `InsecureSkipVerify` added outside the local K3D TLS helper
- [ ] Verbose mode does not log credential values
- [ ] Error messages do not expose internal paths or system information unnecessarily

---

## Environment Variables and Secrets in CI/CD

For non-interactive CI/CD usage of the CLI:

```bash
# Bootstrap in CI — use platform secret management
openframe bootstrap my-ci-cluster \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --github-branch=main
```

**CI security checklist:**

- [ ] Store all tokens and credentials in your CI platform's secret store
- [ ] Use short-lived credentials where possible
- [ ] Rotate credentials after every major release pipeline
- [ ] Audit CI logs to ensure no secrets are printed (check for masked values)
- [ ] Use a dedicated service account for CI cluster operations, not personal credentials

---

## Reporting Security Issues

Security vulnerabilities should be reported via the **OpenMSP Slack community** — not as public GitHub Issues.

- **Slack**: [Join OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **Community**: [openmsp.ai](https://www.openmsp.ai/)

Please use a private Slack message to the maintainers for sensitive security disclosures.
