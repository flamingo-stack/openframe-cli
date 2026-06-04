# Security Best Practices

This document covers security patterns, secrets management, input validation, and common vulnerabilities relevant to developing and deploying OpenFrame CLI.

---

## Authentication and Authorization

### Kubeconfig and Cluster Credentials

OpenFrame CLI interacts with Kubernetes clusters using the standard kubeconfig mechanism. The CLI **never stores or transmits kubeconfig credentials itself** — it reads from the location specified by `$KUBECONFIG` (defaulting to `~/.kube/config`).

**Best practices:**

- Use separate kubeconfig contexts per environment (dev, staging, prod)
- Restrict kubeconfig file permissions to the owning user:

```bash
chmod 600 ~/.kube/config
```

- When running in CI, inject kubeconfig via environment variable rather than writing to disk:

```bash
export KUBECONFIG=/tmp/ci-kubeconfig
echo "$KUBECONFIG_CONTENT" > /tmp/ci-kubeconfig
chmod 600 /tmp/ci-kubeconfig
openframe bootstrap --deployment-mode=oss-tenant --non-interactive
```

### SaaS Credentials

When using `saas-tenant` or `saas-shared` deployment modes, the configuration wizard prompts for SaaS API credentials. These credentials are written into `helm-values-tmp.yaml` — a **temporary file** that the CLI creates, uses, and removes after successful installation.

**Important:** The CLI creates a backup of `helm-values.yaml` before modification and restores it on failure, ensuring that secrets are not left in intermediate states.

---

## Secrets Management

### Never Commit Credentials

OpenFrame CLI generates temporary Helm values files (`helm-values-tmp.yaml`) that may contain credentials. Ensure these are gitignored:

```bash
# Verify .gitignore includes temporary files
echo "helm-values-tmp.yaml" >> .gitignore
echo "*.tmp.yaml" >> .gitignore
```

### Environment Variables Over Flags

Prefer environment variables over command-line flags for secrets — flags may be visible in process listings:

```bash
# Avoid: credentials visible in process list
openframe chart install --saas-token=mysecret123

# Prefer: pass via environment when the CLI supports it
export OPENFRAME_SAAS_TOKEN=mysecret123
openframe chart install --deployment-mode=saas-tenant
```

### CI/CD Secret Injection

In CI/CD pipelines, use your platform's secret store and inject at runtime:

```bash
# GitHub Actions example pattern
# Store KUBECONFIG_BASE64 as a repository secret, then:
echo "$KUBECONFIG_BASE64" | base64 -d > /tmp/kubeconfig
export KUBECONFIG=/tmp/kubeconfig
openframe bootstrap --deployment-mode=oss-tenant --non-interactive
```

---

## TLS Certificate Security

### Local Development TLS (mkcert)

The CLI uses `mkcert` to generate locally-trusted TLS certificates for K3D cluster ingress. These certificates are:

- Generated fresh per environment
- Stored in the chart installation directory
- Trusted only by your local machine's certificate store

```bash
# The CLI handles mkcert automatically — but verify your local CA is installed
mkcert -install
```

### TLS Bypass for Local K3D

The CLI applies an **insecure TLS configuration** (`ApplyInsecureTLSConfig()`) specifically for local K3D clusters to avoid TLS verification errors when communicating with the Kubernetes API. This configuration is:

- Applied **only** to local K3D clusters
- **Not** used for any external API communication
- Appropriate for local development purposes

> **Warning:** Never use insecure TLS configurations in production or staging environments. The TLS bypass is explicitly scoped to K3D cluster contexts.

---

## Input Validation

### Cluster Name Validation

All cluster name inputs are validated before use. The `ValidateClusterName()` function enforces:

- Lowercase alphanumeric characters and hyphens only
- No leading or trailing hyphens
- Maximum length limits

Example of validation usage in code:

```go
if err := models.ValidateClusterName(name); err != nil {
    return errors.CreateValidationError("cluster-name", name, err.Error())
}
```

### Flag Validation

The `ValidateCreateFlags()` and related functions in `internal/cluster/models/flags.go` validate all flag combinations before operations begin, preventing invalid configurations from reaching external tool invocations.

### Shell Injection Prevention

The CLI avoids shell injection by using the `os/exec` package directly (never `sh -c "..."`) through the `CommandExecutor` abstraction:

```go
// Safe: arguments are passed as a slice, never interpolated into a shell string
result, err := executor.Execute(ctx, "k3d", "cluster", "create", clusterName)

// Never do this:
// exec.Command("sh", "-c", "k3d cluster create " + clusterName)
```

All external commands use separate argument slices — user-provided input is never concatenated into shell strings.

---

## Common Security Vulnerabilities and Mitigations

| Vulnerability | Mitigation |
|---------------|-----------|
| **Shell injection** | All external commands use `os/exec` with argument slices, never string interpolation |
| **Path traversal** | Temp directories are created via `os.MkdirTemp()` and cleaned up after use |
| **Credential leakage in logs** | Verbose mode logs command invocations but strips sensitive flag values |
| **Insecure TLS** | TLS bypass is scoped only to local K3D contexts; production HTTPS is fully verified |
| **Exposed kubeconfig** | CLI reads kubeconfig from standard locations; never stores credentials internally |
| **Temporary file exposure** | `helm-values-tmp.yaml` is deleted on success; restored from backup on failure |

---

## Secure File Handling

The `internal/shared/files/` package manages the `helm-values-tmp.yaml` lifecycle:

1. **Backup** — Original `helm-values.yaml` is backed up before modification
2. **Modify** — The temporary file is written with user-provided configuration
3. **Install** — Helm uses the temporary file during chart installation
4. **Cleanup** — On success, the temporary file is deleted and the backup is restored
5. **Recovery** — On failure, the original file is restored from backup

This ensures that a failed installation never leaves secrets on disk.

---

## Environment Variables and Secrets Checklist

Before committing code or creating a PR, verify:

```bash
# Check for accidentally committed secrets
git log --all --full-history -- "*.yaml" "*.env" "*.json"

# Check that temp files are gitignored
git check-ignore helm-values-tmp.yaml

# Verify no hardcoded credentials in source
grep -rn "password\|secret\|token\|api_key" --include="*.go" . | grep -v "_test.go" | grep -v "//.*" | grep -v "flag\|env\|config\|model\|type\|field"
```

---

## Security Testing Guidelines

### Unit Tests

Use the `MockCommandExecutor` to test security-sensitive code paths without needing real credentials:

```go
func TestClusterNameValidation(t *testing.T) {
    testutil.InitializeTestMode()

    // Test rejection of invalid names
    invalidNames := []string{"My Cluster", "cluster!", "../escape", ""}
    for _, name := range invalidNames {
        err := models.ValidateClusterName(name)
        assert.Error(t, err, "expected validation error for: %s", name)
    }
}
```

### Code Review Security Checklist

When reviewing PRs, check for:

- [ ] No credentials or secrets hardcoded in source files
- [ ] All user input validated before use as a command argument
- [ ] External commands use argument slices (not string concatenation)
- [ ] Temporary files are cleaned up in both success and failure paths
- [ ] No new uses of insecure TLS outside of K3D-scoped contexts
- [ ] Error messages do not leak sensitive information (tokens, paths, credentials)

---

## Reporting Security Issues

For security vulnerabilities, please report directly via the **OpenMSP Slack community** (do not open a public GitHub issue for security disclosures):

- [Join OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- [OpenMSP Website](https://www.openmsp.ai/)
