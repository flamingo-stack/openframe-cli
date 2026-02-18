# Quick Start Guide

Get OpenFrame CLI running in under 5 minutes with this streamlined setup process.

## TL;DR Installation

```bash
# 1. Download OpenFrame CLI
# Windows AMD64
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip
unzip openframe-cli_windows_amd64.zip

# Linux/macOS - Replace with your platform
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz
tar -xzf openframe-cli_linux_amd64.tar.gz

# 2. Make executable and add to PATH (Linux/macOS)
chmod +x openframe-cli
sudo mv openframe-cli /usr/local/bin/openframe

# 3. Verify installation
openframe --version

# 4. Bootstrap complete environment
openframe bootstrap

# 5. Access ArgoCD dashboard
kubectl port-forward -n argocd svc/argocd-server 8080:443 &
open https://localhost:8080
```

## Step-by-Step Installation

### Step 1: Download OpenFrame CLI

Choose your platform and download the latest release:

**Windows AMD64:**
```bash
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip
```

**Linux AMD64:**
```bash
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz
```

**macOS:**
```bash
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz
```

### Step 2: Extract and Install

**Windows:**
```bash
unzip openframe-cli_windows_amd64.zip
# Add the directory to your PATH environment variable
```

**Linux/macOS:**
```bash
tar -xzf openframe-cli_linux_amd64.tar.gz
chmod +x openframe-cli
sudo mv openframe-cli /usr/local/bin/openframe
```

### Step 3: Verify Installation

```bash
openframe --version
```

Expected output:
```text
OpenFrame CLI v1.0.0
Commit: abc123def
Built: 2024-01-01T00:00:00Z
```

## One-Command Setup

The fastest way to get a complete OpenFrame environment:

```bash
openframe bootstrap
```

This single command will:

1. ✅ **Check Prerequisites** - Verify Docker, install missing tools
2. ✅ **Create K3D Cluster** - Local Kubernetes cluster with networking
3. ✅ **Install ArgoCD** - GitOps continuous delivery platform  
4. ✅ **Deploy Charts** - App-of-apps pattern for service management
5. ✅ **Configure Certificates** - Local HTTPS with trusted certificates
6. ✅ **Verify Health** - Ensure all services are running correctly

### Bootstrap Process

```text
🚀 OpenFrame Bootstrap
├── Prerequisites Check ✓
├── Creating K3D cluster "openframe-local" ✓
├── Installing ArgoCD ✓
├── Configuring GitOps repository ✓
├── Deploying applications ✓
├── Generating certificates ✓
└── Verification complete ✓

🎉 Your OpenFrame environment is ready!

Access Points:
• ArgoCD: https://argocd.local.openframe.ai
• API Gateway: https://api.local.openframe.ai
• Documentation: https://docs.local.openframe.ai

Next Steps:
• Run 'openframe cluster status' to view cluster details
• Use 'openframe dev intercept' for service development
• Visit the First Steps guide for feature exploration
```

## Verify Your Installation

### Check Cluster Status

```bash
openframe cluster status
```

### List Running Services

```bash
kubectl get pods -A
```

### Access ArgoCD Dashboard

```bash
# Port forward ArgoCD (run in background)
kubectl port-forward -n argocd svc/argocd-server 8080:443 &

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Open browser
open https://localhost:8080
```

**Login Credentials:**
- Username: `admin`
- Password: Use the decoded secret from above

## Your First OpenFrame Workflow

### 1. Explore Available Commands

```bash
openframe --help
```

### 2. Check Cluster Information

```bash
# Detailed cluster status
openframe cluster status --detailed

# List all clusters
openframe cluster list
```

### 3. View Deployed Applications

```bash
# ArgoCD applications
kubectl get applications -n argocd

# All pods across namespaces
kubectl get pods --all-namespaces
```

### 4. Development Workflow

```bash
# Start a development intercept
openframe dev intercept

# Generate a new service scaffold
openframe dev scaffold
```

## Expected Results

After successful bootstrap, you should see:

### ✅ Running Services

```bash
kubectl get pods -A
```
```text
NAMESPACE     NAME                                READY   STATUS
argocd        argocd-server-xxx                   1/1     Running
argocd        argocd-application-controller-xxx   1/1     Running  
argocd        argocd-repo-server-xxx              1/1     Running
kube-system   coredns-xxx                         1/1     Running
```

### ✅ Network Access

- Local services accessible via `.local.openframe.ai` domains
- HTTPS certificates trusted by your browser
- Port forwarding available for external access

### ✅ GitOps Ready

- ArgoCD monitoring your configured Git repositories
- Automatic synchronization of application manifests
- Visual application dependency graph

## Troubleshooting Quick Issues

### Docker Not Running

```bash
# Start Docker service
sudo systemctl start docker  # Linux
# or restart Docker Desktop application
```

### Port Conflicts

```bash
# Check what's using common ports
lsof -i :8080
lsof -i :6443

# Kill conflicting processes or use different ports
```

### Certificate Issues

```bash
# Regenerate certificates
mkcert -install
openframe chart install --regenerate-certs
```

## Next Steps

🎉 **Congratulations!** You now have a fully functional OpenFrame environment.

Continue your journey:

1. **[First Steps Guide](first-steps.md)** - Explore key features and workflows
2. **[Development Setup](../development/setup/local-development.md)** - Configure your IDE and development environment  
3. **[Architecture Overview](../development/architecture/README.md)** - Understand the system design

## Need Help?

- **Community Support**: Join our Slack at https://www.openmsp.ai/
- **Documentation**: Explore additional guides in this documentation
- **Video Tutorials**: Watch our YouTube channel for walkthroughs

[![OpenFrame Product Walkthrough (Beta Access)](https://img.youtube.com/vi/awc-yAnkhIo/maxresdefault.jpg)](https://www.youtube.com/watch?v=awc-yAnkhIo)

You're now ready to build and deploy MSP services with OpenFrame!