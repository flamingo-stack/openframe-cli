# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Code Guidelines](#code-guidelines)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Issue Reporting](#issue-reporting)
- [Community Guidelines](#community-guidelines)

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** - Usually pre-installed on Unix systems

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/openframe-cli.git
   cd openframe-cli
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
   ```

## Development Environment

### Initial Setup

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Build the project:**
   ```bash
   make build
   # Or directly with Go
   go build -o openframe .
   ```

3. **Verify your build:**
   ```bash
   ./openframe --version
   ./openframe --help
   ```

### Development Workflow

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our [code guidelines](#code-guidelines)

3. **Test your changes:**
   ```bash
   make test
   make test-integration  # If applicable
   ```

4. **Build and test manually:**
   ```bash
   make build
   ./openframe cluster create test-cluster
   ./openframe cluster delete test-cluster
   ```

### Project Structure

```
openframe-cli/
├── cmd/                    # Command definitions (Cobra)
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart management commands
│   └── dev/               # Development workflow commands
├── internal/              # Private application code
│   ├── cluster/           # Cluster management logic
│   │   ├── models/        # Data structures
│   │   ├── services/      # Business logic
│   │   └── ui/            # UI components
│   ├── chart/             # Chart management logic
│   ├── dev/               # Development tools integration
│   └── shared/            # Shared utilities and components
├── pkg/                   # Public library code
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
└── tests/                 # Test files
```

## Code Guidelines

### Go Style

- Follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Use [gofmt](https://pkg.go.dev/cmd/gofmt) for formatting
- Use [golint](https://github.com/golang/lint) for linting
- Write clear, self-documenting code with meaningful variable names

### Architecture Patterns

1. **Command Structure:**
   ```go
   // Each command should follow this pattern
   func NewCommand() *cobra.Command {
       cmd := &cobra.Command{
           Use:   "command-name",
           Short: "Brief description",
           RunE:  runCommand,
       }
       // Add flags
       return cmd
   }
   ```

2. **Service Layer:**
   ```go
   // Business logic should be in service structs
   type Service struct {
       // dependencies
   }
   
   func (s *Service) Execute(ctx context.Context) error {
       // implementation
   }
   ```

3. **Error Handling:**
   ```go
   // Use wrapped errors for context
   if err := doSomething(); err != nil {
       return fmt.Errorf("failed to do something: %w", err)
   }
   ```

### Documentation Standards

- **Public functions** must have godoc comments
- **Complex logic** should have inline comments explaining the "why"
- **Examples** should be included in godoc when helpful

```go
// CreateCluster creates a new K3d cluster with the specified configuration.
// It validates the configuration, checks for conflicts, and sets up the cluster
// with appropriate networking and storage configurations.
//
// Example:
//   config := ClusterConfig{Name: "my-cluster", Nodes: 3}
//   err := service.CreateCluster(ctx, config)
func (s *ClusterService) CreateCluster(ctx context.Context, config ClusterConfig) error {
    // implementation
}
```

### UI/UX Guidelines

- **Interactive prompts** should be clear and provide helpful context
- **Progress indicators** for long-running operations
- **Error messages** should be actionable and user-friendly
- **Success messages** should confirm what was accomplished

```go
// Good: Clear, actionable error message
return fmt.Errorf("cluster 'my-cluster' already exists. Use 'openframe cluster delete my-cluster' to remove it first")

// Bad: Generic error message  
return fmt.Errorf("cluster creation failed")
```

## Testing

### Test Types

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test component interactions
3. **End-to-End Tests** - Test complete user workflows

### Running Tests

```bash
# Run unit tests
make test
go test ./...

# Run tests with coverage
make test-coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests (requires Docker)
make test-integration

# Run specific test
go test -v ./internal/cluster/services -run TestClusterCreation
```

### Writing Tests

1. **Use table-driven tests** for multiple scenarios:
   ```go
   func TestClusterValidation(t *testing.T) {
       tests := []struct {
           name    string
           config  ClusterConfig
           wantErr bool
       }{
           {"valid config", ClusterConfig{Name: "test", Nodes: 1}, false},
           {"empty name", ClusterConfig{Name: "", Nodes: 1}, true},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := ValidateConfig(tt.config)
               if (err != nil) != tt.wantErr {
                   t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
               }
           })
       }
   }
   ```

2. **Mock external dependencies:**
   ```go
   // Use interfaces for testability
   type DockerClient interface {
       CreateContainer(config ContainerConfig) error
   }
   
   // Mock in tests
   type MockDockerClient struct {
       createContainerFunc func(ContainerConfig) error
   }
   ```

3. **Test error conditions** as well as happy paths

### Test Requirements

- All new features must include tests
- Bug fixes should include regression tests
- Tests should be deterministic and not rely on external services
- Integration tests should clean up resources

## Submitting Changes

### Pull Request Process

1. **Update your branch:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Ensure all tests pass:**
   ```bash
   make test
   make test-integration
   make build
   ```

3. **Commit your changes** with clear messages:
   ```bash
   git commit -m "feat: add cluster node scaling support
   
   - Add scale-up and scale-down commands
   - Validate node count ranges
   - Update cluster status display
   
   Closes #123"
   ```

4. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a pull request** on GitHub

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(cluster): add multi-node cluster support
fix(bootstrap): handle missing kubectl gracefully
docs: update installation instructions
test(cluster): add integration tests for cluster deletion
```

### Pull Request Guidelines

- **Title** should be descriptive and follow commit message format
- **Description** should explain what changed and why
- **Link related issues** using "Closes #123" or "Fixes #456"
- **Include screenshots** for UI changes
- **Update documentation** if needed

### Review Process

1. Automated checks must pass (tests, linting, build)
2. At least one maintainer review required
3. Address review feedback promptly
4. Maintainer will merge when ready

## Issue Reporting

### Bug Reports

Use the bug report template and include:

- **OpenFrame CLI version:** `openframe --version`
- **Operating system:** macOS, Linux, Windows
- **Docker version:** `docker --version`
- **Expected behavior:** What should happen
- **Actual behavior:** What actually happened
- **Steps to reproduce:** Minimal reproduction steps
- **Logs:** Relevant error messages or logs

### Feature Requests

Use the feature request template and include:

- **Use case:** Why is this needed?
- **Proposed solution:** How should it work?
- **Alternatives:** Other approaches considered
- **Additional context:** Screenshots, examples, etc.

### Getting Help

- **Documentation:** Check existing docs first
- **Search issues:** Look for existing discussions
- **Discussion forum:** For general questions
- **Discord/Slack:** Real-time community help

## Community Guidelines

### Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- **Be respectful** and constructive in all interactions
- **Be patient** with newcomers and those learning
- **Be collaborative** and help others succeed
- **Be inclusive** and welcoming to all backgrounds

### Communication

- **GitHub Issues** - Bug reports, feature requests, technical discussions
- **Pull Requests** - Code review discussions
- **Discussion Forum** - General questions and community help
- **Discord/Slack** - Real-time chat and community building

### Recognition

Contributors are recognized through:

- **GitHub contributors page**
- **Release notes acknowledgments**
- **Community highlights**
- **Maintainer nominations** for significant contributors

## Development Tips

### Debugging

```bash
# Enable verbose logging
./openframe bootstrap --verbose

# Debug with Delve
dlv debug . -- cluster create test
```

### Local Testing

```bash
# Test with different cluster configurations
./openframe cluster create single-node --nodes 1
./openframe cluster create multi-node --nodes 3

# Test bootstrap process
./openframe bootstrap test-env --deployment-mode=oss-tenant

# Clean up test resources
./openframe cluster cleanup
```

### IDE Setup

Recommended VS Code extensions:
- Go extension pack
- Docker extension
- Kubernetes extension
- GitLens

Recommended settings:
```json
{
    "go.formatTool": "gofmt",
    "go.lintTool": "golint",
    "go.testFlags": ["-v"]
}
```

## Getting Started Checklist

- [ ] Fork and clone the repository
- [ ] Set up development environment
- [ ] Build the project successfully
- [ ] Run the test suite
- [ ] Create a test cluster with your build
- [ ] Read through the codebase structure
- [ ] Find a "good first issue" to work on
- [ ] Join the community discussion channels

## Questions?

- Check the [documentation](docs/)
- Search [existing issues](https://github.com/flamingo-stack/openframe-cli/issues)
- Ask in our [discussion forum](https://github.com/flamingo-stack/openframe-cli/discussions)
- Join our community chat

Thank you for contributing to OpenFrame CLI! 🚀