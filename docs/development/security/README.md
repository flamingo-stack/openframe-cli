# Security Guidelines

This document outlines the security practices, patterns, and considerations for developing and operating the OpenFrame CLI.

---

## Authentication and Authorization

### Kubernetes Access Control

OpenFrame CLI uses `kubeconfig` for all Kubernetes API authentication. The CLI never stores Kubernetes credentials; it reads them from the standard kubeconfig file at the path specified by the `$KUBECONFIG` environment variable or the default `~/.kube/config`.

```bash
# Use a specific kubeconfig
export KUBECONFIG="$HOME/.kube/my-cluster-config"
openframe cluster status my-cluster
```

**Best practices:**
- Use dedicated kubeconfig contexts per cluster (do not use the `admin` context in production)
- Rotate kubeconfig credentials regularly
- Do not share kubeconfig files across team members

### ArgoCD Authentication

The CLI communicates with ArgoCD using the native Kubernetes API client (`k8s.io/client-go`) — it reads ArgoCD `Application` resources directly from the cluster rather than through ArgoCD's HTTP API. This avoids storing ArgoCD passwords and relies on the RBAC already granted via kubeconfig.

### GitHub Credentials

For `saas-tenant` and `saas-shared` deployment modes that require access to private GHCR (GitHub Container Registry) images, the CLI prompts for GitHub credentials at runtime using the `CredentialsPrompter`:

- Credentials are provided interactively via masked terminal input
- Credentials are **never stored to disk** by the CLI
- They are passed directly to Helm as chart values or to Docker login commands in-process

```bash
# The CLI will prompt you when credentials are needed:
# Enter GitHub username:
# Enter GitHub personal access token (PAT): ****
```

Use a **Personal Access Token (PAT)** with only the `read:packages` scope for GHCR access — never use your GitHub password.

---

## Secrets Management

### Do Not Store Secrets in Code

Never commit credentials, tokens, or certificates to the repository. The project's `internal/shared/config/credentials.go` provides runtime prompting specifically to avoid hardcoded secrets.

```go
// CORRECT — prompt at runtime
creds, err := prompter.PromptForGitHubCredentials(repoURL)

// WRONG — never do this
const githubToken = "ghp_abc123..." // ❌ Never hardcode secrets
```

### Environment Variables

For CI/CD pipelines, pass secrets via environment variables rather than CLI flags (flags may appear in process listings):

```bash
# CI/CD: pass via environment variables to your pipeline secrets manager
# Then reference them in non-interactive mode
openframe bootstrap my-cluster \
  --deployment-mode=oss-tenant \
  --non-interactive
```

### TLS Certificates

The CLI uses `mkcert` to generate locally-trusted TLS certificates for K3D clusters. Key behaviors:

- `internal/shared/config/transport.go` applies `ApplyInsecureTLSConfig()` only for K3D local clusters (never for production clusters)
- Certificate files are written to a temp directory and cleaned up after use
- The `chart/prerequisites/certificates` module manages certificate lifecycle

> **Important:** The `InsecureTLSConfig` is **only** applied for local K3D development environments. Do not use this configuration pattern for connecting to production clusters.

---

## Input Validation and Sanitization

### Cluster Name Validation

All user-provided cluster names go through `ValidateClusterName()` in `internal/cluster/models/flags.go` before being used in any system command. This prevents command injection via specially crafted cluster names.

```go
// Validation enforces safe name format before use
if err := models.ValidateClusterName(clusterName); err != nil {
    return errors.CreateValidationError("cluster-name", clusterName, err.Error())
}
```

### Port Validation

Intercept port values are validated by `validateInputs()` in `internal/dev/services/intercept/service.go` to ensure they are valid TCP port numbers before being passed to Telepresence.

### Command Injection Prevention

The `shared/executor` abstraction executes all external tool commands by passing arguments as a slice (not a shell-interpolated string). This is the idiomatic Go pattern that prevents shell injection:

```go
// CORRECT — args are individual slice elements, no shell interpolation
executor.Execute("k3d", []string{"cluster", "create", clusterName, "--agents", "2"})

// WRONG — never build shell strings with user input
exec.Command("sh", "-c", "k3d cluster create " + clusterName) // ❌ Injection risk
```

---

## Sensitive Data in Logs

### Verbose Mode Caution

The `--verbose` flag enables detailed logging including ArgoCD sync progress. Review verbose output before sharing logs publicly — it may include:

- Kubernetes namespace names and resource names
- Helm chart values (which may reference registry URLs)
- kubeconfig context names

Do not share verbose output in public channels without reviewing it first.

### Error Messages

The `shared/errors` package formats user-facing error messages using `pterm`. Error messages are designed to be informative without leaking internal details. If you add new error types, follow this pattern:

```go
// CORRECT — descriptive but safe
return CreateValidationError("port", portValue, "must be between 1 and 65535")

// WRONG — may expose internal state
return fmt.Errorf("internal state: %+v failed at %s", internalStruct, secretPath)
```

---

## Dependency Security

### Vulnerability Scanning

Run the Go vulnerability scanner against the project dependencies regularly:

```bash
govulncheck ./...
```

Fix any identified vulnerabilities by updating the affected dependency in `go.mod`:

```bash
go get github.com/vulnerable/package@latest
go mod tidy
```

### Dependency Review

When adding new dependencies, evaluate:
1. Is the package actively maintained?
2. Does it have known CVEs?
3. Is the scope minimal (does not pull in unnecessary transitive dependencies)?
4. Is it from a trusted source?

---

## Code Review Security Checklist

When reviewing pull requests, check for:

- [ ] No hardcoded secrets, tokens, passwords, or API keys
- [ ] User input is validated before use in system commands
- [ ] External commands use argument slices (not shell string interpolation)
- [ ] Sensitive values are not logged in verbose output
- [ ] New error types do not expose internal implementation details
- [ ] TLS configuration is not relaxed outside the K3D local context
- [ ] New dependencies have been reviewed for security issues
- [ ] Certificate or credentials handling matches existing patterns in `internal/shared/config/`

---

## Reporting Security Issues

Do not report security vulnerabilities in public GitHub Issues. Instead, reach out directly through the **OpenMSP Slack** community:

👉 [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)

- 🌐 OpenMSP Community: [https://www.openmsp.ai/](https://www.openmsp.ai/)
