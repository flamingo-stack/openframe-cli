# Security Best Practices

This document outlines security best practices, authentication patterns, and vulnerability management for OpenFrame CLI development.

## Security Overview

OpenFrame CLI operates in a privileged environment, managing Docker containers, Kubernetes clusters, and external tool integrations. Security is paramount to prevent privilege escalation, data exposure, and malicious code execution.

## Authentication and Authorization Patterns

### Kubernetes Authentication

OpenFrame CLI relies on existing Kubernetes authentication mechanisms:

```go
// Secure kubeconfig handling
type KubeConfigManager struct {
    configPath string
    context    string
}

func (k *KubeConfigManager) GetSecureConfig() (*rest.Config, error) {
    // Validate kubeconfig path
    if !isSecurePath(k.configPath) {
        return nil, fmt.Errorf("insecure kubeconfig path: %s", k.configPath)
    }
    
    // Load with restricted permissions
    config, err := clientcmd.BuildConfigFromFlags("", k.configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
    }
    
    // Apply security constraints
    config.Timeout = 30 * time.Second
    config.QPS = 10
    config.Burst = 15
    
    return config, nil
}
```

### Docker Authentication

Secure Docker daemon communication:

```go
// Secure Docker client configuration
func NewSecureDockerClient() (*client.Client, error) {
    // Use system Docker socket with validation
    host := os.Getenv("DOCKER_HOST")
    if host == "" {
        host = "unix:///var/run/docker.sock"
    }
    
    // Validate Docker socket permissions
    if err := validateDockerSocket(host); err != nil {
        return nil, fmt.Errorf("docker socket validation failed: %w", err)
    }
    
    // Create client with security options
    opts := []client.Opt{
        client.WithAPIVersionNegotiation(),
        client.WithTimeout(30 * time.Second),
    }
    
    return client.NewClientWithOpts(opts...)
}

func validateDockerSocket(host string) error {
    if strings.HasPrefix(host, "unix://") {
        socketPath := strings.TrimPrefix(host, "unix://")
        info, err := os.Stat(socketPath)
        if err != nil {
            return err
        }
        
        // Check socket permissions (should be writable by docker group)
        if info.Mode()&0o066 != 0 {
            return fmt.Errorf("docker socket has overly permissive permissions: %v", info.Mode())
        }
    }
    return nil
}
```

### ArgoCD Integration Security

```go
// Secure ArgoCD client configuration
type ArgoCDClient struct {
    serverAddr string
    token      string
    insecure   bool
}

func (a *ArgoCDClient) NewSecureConnection() (*grpc.ClientConn, error) {
    var opts []grpc.DialOption
    
    if a.insecure {
        // Only allow insecure connections for localhost development
        if !isLocalhost(a.serverAddr) {
            return nil, fmt.Errorf("insecure connections only allowed for localhost")
        }
        opts = append(opts, grpc.WithInsecure())
    } else {
        // Use TLS with certificate verification
        config := &tls.Config{
            ServerName: extractHostname(a.serverAddr),
            MinVersion: tls.VersionTLS12,
        }
        opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
    }
    
    return grpc.Dial(a.serverAddr, opts...)
}
```

## Data Encryption and Secure Storage

### Configuration Data

```go
// Secure configuration handling
type SecureConfig struct {
    data map[string]interface{}
    encrypted bool
}

func (c *SecureConfig) SetSecretValue(key, value string) error {
    // Encrypt sensitive values before storage
    encrypted, err := encrypt(value, getConfigEncryptionKey())
    if err != nil {
        return fmt.Errorf("failed to encrypt value: %w", err)
    }
    
    c.data[key] = encrypted
    c.encrypted = true
    return nil
}

func (c *SecureConfig) GetSecretValue(key string) (string, error) {
    value, exists := c.data[key]
    if !exists {
        return "", fmt.Errorf("key not found: %s", key)
    }
    
    if !c.encrypted {
        return value.(string), nil
    }
    
    decrypted, err := decrypt(value.(string), getConfigEncryptionKey())
    if err != nil {
        return "", fmt.Errorf("failed to decrypt value: %w", err)
    }
    
    return decrypted, nil
}

func getConfigEncryptionKey() []byte {
    // Use system keyring or environment-based key derivation
    key := os.Getenv("OPENFRAME_CONFIG_KEY")
    if key == "" {
        // Derive key from system information (non-portable but secure)
        return deriveSystemKey()
    }
    return []byte(key)
}
```

### Temporary Files

```go
// Secure temporary file handling
func CreateSecureTempFile(prefix string) (*os.File, error) {
    // Create temporary file with restricted permissions
    tmpDir := os.TempDir()
    
    // Ensure temp directory has proper permissions
    if err := os.Chmod(tmpDir, 0o700); err != nil {
        return nil, fmt.Errorf("failed to secure temp directory: %w", err)
    }
    
    // Create file with owner-only permissions
    file, err := os.CreateTemp(tmpDir, prefix)
    if err != nil {
        return nil, fmt.Errorf("failed to create temp file: %w", err)
    }
    
    // Set restrictive permissions
    if err := file.Chmod(0o600); err != nil {
        file.Close()
        os.Remove(file.Name())
        return nil, fmt.Errorf("failed to set file permissions: %w", err)
    }
    
    return file, nil
}

// Automatic cleanup with secure deletion
func SecureCleanup(filepath string) error {
    // Overwrite file contents before deletion
    file, err := os.OpenFile(filepath, os.O_WRONLY, 0)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Get file size
    info, err := file.Stat()
    if err != nil {
        return err
    }
    
    // Overwrite with random data
    randomData := make([]byte, info.Size())
    rand.Read(randomData)
    
    if _, err := file.WriteAt(randomData, 0); err != nil {
        return err
    }
    
    if err := file.Sync(); err != nil {
        return err
    }
    
    // Remove file
    return os.Remove(filepath)
}
```

## Input Validation and Sanitization

### Command Input Validation

```go
// Secure command input validation
type InputValidator struct {
    allowedChars *regexp.Regexp
    maxLength    int
}

func NewInputValidator() *InputValidator {
    return &InputValidator{
        // Allow alphanumeric, hyphens, and underscores only
        allowedChars: regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
        maxLength:    63, // Kubernetes name length limit
    }
}

func (v *InputValidator) ValidateClusterName(name string) error {
    if len(name) == 0 {
        return fmt.Errorf("cluster name cannot be empty")
    }
    
    if len(name) > v.maxLength {
        return fmt.Errorf("cluster name too long: %d > %d", len(name), v.maxLength)
    }
    
    if !v.allowedChars.MatchString(name) {
        return fmt.Errorf("cluster name contains invalid characters: %s", name)
    }
    
    // Prevent reserved names
    reserved := []string{"kubernetes", "default", "kube-system", "kube-public", "kube-node-lease"}
    for _, r := range reserved {
        if strings.EqualFold(name, r) {
            return fmt.Errorf("cluster name conflicts with reserved name: %s", r)
        }
    }
    
    return nil
}

func (v *InputValidator) ValidatePath(path string) error {
    // Prevent path traversal attacks
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "..") {
        return fmt.Errorf("path contains directory traversal: %s", path)
    }
    
    // Ensure path is within allowed directories
    allowedPrefixes := []string{
        os.TempDir(),
        os.Getenv("HOME"),
        "/opt/openframe",
    }
    
    isAllowed := false
    for _, prefix := range allowedPrefixes {
        if strings.HasPrefix(cleanPath, prefix) {
            isAllowed = true
            break
        }
    }
    
    if !isAllowed {
        return fmt.Errorf("path not in allowed directories: %s", path)
    }
    
    return nil
}
```

### YAML/JSON Configuration Validation

```go
// Secure configuration parsing
func ParseSecureConfig(data []byte) (*Config, error) {
    // Limit parsing size to prevent DoS
    if len(data) > 10*1024*1024 { // 10MB limit
        return nil, fmt.Errorf("configuration file too large: %d bytes", len(data))
    }
    
    // Use secure YAML parser with restrictions
    decoder := yaml.NewDecoder(bytes.NewReader(data))
    decoder.KnownFields(true) // Reject unknown fields
    
    var config Config
    if err := decoder.Decode(&config); err != nil {
        return nil, fmt.Errorf("failed to parse configuration: %w", err)
    }
    
    // Validate configuration structure
    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return &config, nil
}

func validateConfig(config *Config) error {
    // Validate all string fields
    validator := NewInputValidator()
    
    if err := validator.ValidateClusterName(config.ClusterName); err != nil {
        return fmt.Errorf("invalid cluster name: %w", err)
    }
    
    // Validate resource limits
    if config.Resources.Memory < 0 || config.Resources.Memory > 64*1024*1024*1024 { // 64GB max
        return fmt.Errorf("invalid memory limit: %d", config.Resources.Memory)
    }
    
    if config.Resources.CPU < 0 || config.Resources.CPU > 32 { // 32 CPU max
        return fmt.Errorf("invalid CPU limit: %f", config.Resources.CPU)
    }
    
    return nil
}
```

## Common Security Vulnerabilities and Mitigations

### 1. Command Injection

```go
// Secure command execution
func ExecuteSecureCommand(command string, args ...string) error {
    // Whitelist allowed commands
    allowedCommands := map[string]bool{
        "k3d":     true,
        "kubectl": true,
        "helm":    true,
        "docker":  true,
    }
    
    if !allowedCommands[command] {
        return fmt.Errorf("command not allowed: %s", command)
    }
    
    // Validate all arguments
    for _, arg := range args {
        if err := validateCommandArgument(arg); err != nil {
            return fmt.Errorf("invalid argument '%s': %w", arg, err)
        }
    }
    
    // Use exec.Command instead of shell execution
    cmd := exec.Command(command, args...)
    
    // Set secure environment
    cmd.Env = getSecureEnvironment()
    
    // Set working directory to safe location
    cmd.Dir = "/tmp"
    
    return cmd.Run()
}

func validateCommandArgument(arg string) error {
    // Prevent shell metacharacters
    dangerous := []string{";", "&", "|", "$", "`", "$(", "${", ">", "<", "&&", "||"}
    for _, d := range dangerous {
        if strings.Contains(arg, d) {
            return fmt.Errorf("argument contains dangerous characters: %s", d)
        }
    }
    return nil
}
```

### 2. Path Traversal

```go
// Secure file operations
func SecureFileRead(basePath, userPath string) ([]byte, error) {
    // Clean and validate the path
    cleanBase := filepath.Clean(basePath)
    cleanUser := filepath.Clean(userPath)
    fullPath := filepath.Join(cleanBase, cleanUser)
    
    // Ensure the resulting path is within the base directory
    if !strings.HasPrefix(fullPath, cleanBase) {
        return nil, fmt.Errorf("path traversal detected: %s", userPath)
    }
    
    // Check if file exists and is readable
    info, err := os.Stat(fullPath)
    if err != nil {
        return nil, fmt.Errorf("file access error: %w", err)
    }
    
    // Prevent reading of special files
    if !info.Mode().IsRegular() {
        return nil, fmt.Errorf("not a regular file: %s", fullPath)
    }
    
    // Limit file size
    if info.Size() > 10*1024*1024 { // 10MB limit
        return nil, fmt.Errorf("file too large: %d bytes", info.Size())
    }
    
    return os.ReadFile(fullPath)
}
```

### 3. Privilege Escalation

```go
// Run commands with minimal privileges
func ExecuteWithMinimalPrivileges(command string, args []string) error {
    cmd := exec.Command(command, args...)
    
    // Drop privileges if running as root
    if os.Getuid() == 0 {
        // Find the nobody user
        nobody, err := user.Lookup("nobody")
        if err != nil {
            return fmt.Errorf("failed to lookup nobody user: %w", err)
        }
        
        uid, _ := strconv.Atoi(nobody.Uid)
        gid, _ := strconv.Atoi(nobody.Gid)
        
        cmd.SysProcAttr = &syscall.SysProcAttr{
            Credential: &syscall.Credential{
                Uid: uint32(uid),
                Gid: uint32(gid),
            },
        }
    }
    
    return cmd.Run()
}
```

## Security Testing and Code Review Guidelines

### Security Testing Checklist

- [ ] **Input Validation**: All user inputs are validated and sanitized
- [ ] **Command Injection**: No user input is directly passed to shell commands
- [ ] **Path Traversal**: File paths are validated and contained within safe directories
- [ ] **Privilege Escalation**: Commands run with minimal required privileges
- [ ] **Secrets Management**: No secrets are logged or stored in plain text
- [ ] **Network Security**: All network connections use appropriate encryption
- [ ] **Error Handling**: Error messages don't leak sensitive information

### Code Review Security Focus

```go
// Example of security-focused code review
func ReviewSecurityChecklist(code string) []string {
    issues := []string{}
    
    // Check for common security anti-patterns
    if strings.Contains(code, "exec.Command(") {
        issues = append(issues, "Review exec.Command usage for command injection")
    }
    
    if strings.Contains(code, "os.Open(") {
        issues = append(issues, "Review file operations for path traversal")
    }
    
    if strings.Contains(code, "http.Get(") {
        issues = append(issues, "Review HTTP requests for proper TLS configuration")
    }
    
    if strings.Contains(code, "fmt.Printf") && strings.Contains(code, "password") {
        issues = append(issues, "Potential secret logging detected")
    }
    
    return issues
}
```

## Environment Variables and Secrets Management

### Secure Environment Variable Handling

```go
// Secure environment variable management
type SecureEnv struct {
    sensitiveKeys map[string]bool
}

func NewSecureEnv() *SecureEnv {
    return &SecureEnv{
        sensitiveKeys: map[string]bool{
            "OPENFRAME_CONFIG_KEY": true,
            "DOCKER_PASSWORD":      true,
            "KUBECONFIG":          true,
            "ARGOCD_AUTH_TOKEN":   true,
        },
    }
}

func (s *SecureEnv) Get(key string) (string, error) {
    value := os.Getenv(key)
    if value == "" {
        return "", fmt.Errorf("environment variable not set: %s", key)
    }
    
    // Log access to sensitive variables
    if s.sensitiveKeys[key] {
        log.Info("accessing sensitive environment variable", "key", key)
    }
    
    return value, nil
}

func (s *SecureEnv) Set(key, value string) error {
    if s.sensitiveKeys[key] {
        // Validate sensitive values
        if len(value) < 8 {
            return fmt.Errorf("sensitive value too short for %s", key)
        }
    }
    
    return os.Setenv(key, value)
}

// Sanitize environment for subprocesses
func GetSecureEnvironment() []string {
    allowedVars := []string{
        "PATH", "HOME", "USER", "TMPDIR",
        "DOCKER_HOST", "KUBECONFIG",
        "OPENFRAME_CONFIG_DIR",
    }
    
    var secureEnv []string
    for _, key := range allowedVars {
        if value := os.Getenv(key); value != "" {
            secureEnv = append(secureEnv, key+"="+value)
        }
    }
    
    return secureEnv
}
```

### Kubernetes Secrets Integration

```go
// Secure secrets management with Kubernetes
func StoreSecretInCluster(name, namespace string, data map[string][]byte) error {
    config, err := rest.InClusterConfig()
    if err != nil {
        return fmt.Errorf("failed to get cluster config: %w", err)
    }
    
    client, err := kubernetes.NewForConfig(config)
    if err != nil {
        return fmt.Errorf("failed to create client: %w", err)
    }
    
    secret := &corev1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
            Annotations: map[string]string{
                "openframe.io/managed": "true",
                "openframe.io/created": time.Now().Format(time.RFC3339),
            },
        },
        Type: corev1.SecretTypeOpaque,
        Data: data,
    }
    
    _, err = client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
    if err != nil {
        return fmt.Errorf("failed to create secret: %w", err)
    }
    
    return nil
}
```

## Compliance and Audit

### Audit Logging

```go
// Security audit logging
type SecurityAuditor struct {
    logger *log.Logger
}

func (a *SecurityAuditor) LogSecurityEvent(event SecurityEvent) {
    a.logger.Info("security event",
        "type", event.Type,
        "user", event.User,
        "action", event.Action,
        "resource", event.Resource,
        "timestamp", event.Timestamp,
        "success", event.Success,
    )
}

type SecurityEvent struct {
    Type      string    // e.g., "authentication", "authorization", "data_access"
    User      string    // username or system account
    Action    string    // specific action performed
    Resource  string    // resource accessed
    Timestamp time.Time
    Success   bool
    Details   map[string]interface{}
}
```

### Compliance Checks

```go
// Automated compliance checking
func RunComplianceCheck() error {
    checks := []ComplianceCheck{
        CheckFilePermissions,
        CheckEnvironmentSecurity,
        CheckNetworkSecurity,
        CheckSecretHandling,
    }
    
    for _, check := range checks {
        if err := check(); err != nil {
            return fmt.Errorf("compliance check failed: %w", err)
        }
    }
    
    return nil
}

func CheckFilePermissions() error {
    configDir := os.Getenv("OPENFRAME_CONFIG_DIR")
    info, err := os.Stat(configDir)
    if err != nil {
        return err
    }
    
    if info.Mode()&0o077 != 0 {
        return fmt.Errorf("config directory has overly permissive permissions: %v", info.Mode())
    }
    
    return nil
}
```

> 🔒 **Security First**: Security is not an afterthought in OpenFrame CLI. Every feature must be designed with security in mind, following the principle of least privilege and defense in depth. When in doubt, choose the more secure option.