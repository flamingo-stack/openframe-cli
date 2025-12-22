# Contributing Guidelines

Welcome to OpenFrame CLI! We appreciate your interest in contributing to this project. This guide outlines our development process, coding standards, and best practices for contributors.

## Getting Started

### Before You Contribute

1. **Read the Documentation**: Familiarize yourself with the project structure and architecture
2. **Set Up Development Environment**: Follow the [Environment Setup](../setup/environment.md) guide
3. **Understand the Codebase**: Review the [Architecture Overview](../architecture/overview.md)
4. **Check Existing Issues**: Look for existing issues or create a new one for discussion

### Ways to Contribute

| Contribution Type | Description | Good for |
|-------------------|-------------|----------|
| **Bug Fixes** | Fix existing issues and regressions | First-time contributors |
| **Features** | Add new functionality or commands | Experienced contributors |
| **Documentation** | Improve guides, examples, and API docs | All skill levels |
| **Testing** | Add test coverage and improve test quality | Quality-focused contributors |
| **Performance** | Optimize existing code and operations | Performance experts |
| **Refactoring** | Improve code structure and maintainability | Architecture enthusiasts |

## Development Workflow

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then clone your fork
git clone git@github.com:YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# Add upstream remote
git remote add upstream git@github.com:flamingo-stack/openframe-cli.git

# Verify remotes
git remote -v
```

### 2. Create a Branch

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/my-awesome-feature

# Or for bug fixes
git checkout -b fix/issue-description
```

### 3. Make Changes

Follow our coding standards and ensure your changes:
- ✅ Follow Go conventions and project patterns
- ✅ Include comprehensive tests
- ✅ Add documentation for new features
- ✅ Pass all existing tests
- ✅ Follow commit message conventions

### 4. Test Your Changes

```bash
# Run all tests
make test

# Run integration tests
make test-integration

# Run linting
make lint

# Format code
make fmt

# Run complete check
make dev
```

### 5. Commit and Push

```bash
# Stage your changes
git add .

# Commit with descriptive message
git commit -m "feat: add cluster scaling functionality"

# Push to your fork
git push origin feature/my-awesome-feature
```

### 6. Create Pull Request

1. Go to your fork on GitHub
2. Click "New Pull Request"
3. Choose base: `main` ← compare: `your-branch`
4. Fill out the PR template thoroughly
5. Submit the pull request

## Coding Standards

### Go Code Style

We follow standard Go conventions with some project-specific additions:

**Formatting:**
```bash
# Use goimports (includes gofmt)
goimports -w .

# Our linter configuration enforces:
# - Line length: 120 characters
# - No unused variables or imports
# - Proper error handling
# - Consistent naming conventions
```

**Naming Conventions:**
```go
// Packages: lowercase, single word when possible
package cluster

// Interfaces: noun or adjective + "er" suffix
type ClusterProvider interface{}
type ExecutorService interface{}

// Structs: CamelCase nouns
type ClusterService struct{}
type ClusterConfig struct{}

// Functions/Methods: CamelCase verbs
func CreateCluster() error {}
func (s *ClusterService) GetStatus() error {}

// Constants: CamelCase or UPPER_CASE for exported
const DefaultTimeout = 30 * time.Second
const MAX_RETRIES = 3

// Variables: camelCase
var defaultConfig ClusterConfig
```

**Error Handling:**
```go
// Always handle errors explicitly
result, err := operation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use custom error types for structured errors
type ClusterError struct {
    Type    ErrorType
    Message string
    Cause   error
}

// Wrap errors to preserve context
return fmt.Errorf("failed to create cluster %s: %w", name, err)
```

**Interface Design:**
```go
// Keep interfaces small and focused
type ClusterCreator interface {
    CreateCluster(config ClusterConfig) error
}

// Use interfaces for testability
type Service struct {
    provider ClusterProvider  // Interface, not concrete type
    executor ExecutorService
}
```

### Project Structure Conventions

**Package Organization:**
```text
cmd/                    # CLI commands (thin layer)
├── bootstrap/         # Bootstrap command
├── cluster/           # Cluster commands  
├── chart/             # Chart commands
└── dev/               # Development commands

internal/               # Internal packages
├── bootstrap/         # Bootstrap service logic
├── cluster/           # Cluster management
├── chart/             # Chart installation
├── dev/               # Development workflows
└── shared/            # Shared utilities
```

**File Naming:**
```text
service.go              # Main service implementation
interfaces.go           # Interface definitions
models.go              # Data models and types
errors.go              # Error types and handling
service_test.go        # Unit tests
integration_test.go    # Integration tests
```

### Documentation Standards

**Code Comments:**
```go
// Package documentation
// Package cluster provides Kubernetes cluster management functionality.
// It supports creating, deleting, and managing K3d clusters with various
// configuration options and deployment modes.
package cluster

// Exported function documentation
// CreateCluster creates a new Kubernetes cluster with the specified configuration.
// It validates the configuration, checks prerequisites, and delegates to the
// appropriate provider for cluster creation.
//
// Returns an error if:
//   - Configuration is invalid
//   - Prerequisites are not met
//   - Cluster creation fails
func (s *ClusterService) CreateCluster(config ClusterConfig) error {}

// Complex logic documentation
func complexOperation() {
    // Step 1: Validate input parameters
    // This ensures we fail fast before expensive operations
    
    // Step 2: Acquire necessary resources
    // Resource cleanup is handled by defer statements
}
```

**README and Markdown:**
```markdown
# Use descriptive headings
## Clear section organization
### Consistent formatting

- Use bullet points for lists
- Include code examples with language hints
- Add tables for structured information
- Include mermaid diagrams for complex relationships
```

## Testing Requirements

### Test Coverage

All contributions must include appropriate tests:

| Component | Required Tests | Coverage Target |
|-----------|----------------|-----------------|
| **New Features** | Unit + Integration | 80%+ |
| **Bug Fixes** | Regression test + existing | Maintain current |
| **Refactoring** | All existing tests pass | No decrease |
| **Utilities** | Comprehensive unit tests | 90%+ |

### Test Writing Guidelines

**Unit Test Structure:**
```go
func TestClusterService_CreateCluster(t *testing.T) {
    tests := []struct {
        name           string           // Test case name
        input          ClusterConfig    // Input data
        mockSetup      func(*MockProvider) // Mock expectations  
        expectedResult ClusterStatus    // Expected result
        expectedError  string          // Expected error (if any)
    }{
        {
            name: "successful cluster creation",
            input: ClusterConfig{
                Name: "test-cluster",
                Type: ClusterTypeK3d,
            },
            mockSetup: func(m *MockProvider) {
                m.EXPECT().CreateCluster(gomock.Any()).Return(nil)
            },
            expectedError: "",
        },
        {
            name: "invalid cluster name",
            input: ClusterConfig{
                Name: "",  // Invalid empty name
                Type: ClusterTypeK3d,
            },
            expectedError: "cluster name cannot be empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockProvider := mocks.NewMockClusterProvider(ctrl)
            if tt.mockSetup != nil {
                tt.mockSetup(mockProvider)
            }
            
            service := NewClusterService(mockProvider)
            
            // Act
            err := service.CreateCluster(tt.input)
            
            // Assert
            if tt.expectedError != "" {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

**Integration Test Guidelines:**
```go
func TestClusterIntegration(t *testing.T) {
    // Skip in short mode
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Use unique names to avoid conflicts
    clusterName := fmt.Sprintf("test-%d", time.Now().Unix())
    
    // Always clean up
    defer func() {
        exec.Command("k3d", "cluster", "delete", clusterName).Run()
    }()
    
    // Test with real dependencies
    service := cluster.NewClusterService(executor.NewRealExecutor())
    // ... test implementation
}
```

## Commit Message Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) for consistent commit messages:

### Format

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat: add cluster scaling command` |
| `fix` | Bug fix | `fix: handle empty cluster names properly` |
| `docs` | Documentation changes | `docs: update installation instructions` |
| `style` | Code style changes (formatting, etc.) | `style: fix gofmt issues` |
| `refactor` | Code refactoring | `refactor: simplify error handling logic` |
| `test` | Adding or fixing tests | `test: add integration tests for bootstrap` |
| `chore` | Build process or auxiliary tool changes | `chore: update dependencies` |
| `perf` | Performance improvements | `perf: optimize cluster status checking` |
| `ci` | CI/CD changes | `ci: add integration test workflow` |

### Examples

```bash
# Feature addition
git commit -m "feat(cluster): add support for custom K3d configuration"

# Bug fix with body
git commit -m "fix(bootstrap): handle timeout errors gracefully

The bootstrap process now properly handles timeout errors from ArgoCD
sync operations and provides helpful error messages to users."

# Breaking change
git commit -m "feat!: change cluster configuration format

BREAKING CHANGE: The cluster configuration now uses YAML format
instead of JSON. Update your configuration files accordingly."

# Multiple types
git commit -m "feat(chart): add Helm repository management

- Add support for custom Helm repositories
- Include repository validation
- Update documentation and examples

Closes #123"
```

## Pull Request Guidelines

### PR Template

When creating a pull request, use this template:

```markdown
## Description
Brief description of what this PR does and why.

## Changes
- [ ] New feature implementation
- [ ] Bug fix
- [ ] Documentation update
- [ ] Test improvements
- [ ] Refactoring

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated  
- [ ] Manual testing performed
- [ ] All tests pass

## Documentation
- [ ] Code comments added/updated
- [ ] README updated (if needed)
- [ ] Documentation updated (if needed)

## Breaking Changes
- [ ] No breaking changes
- [ ] Breaking changes documented

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Tests cover the changes
- [ ] Documentation is up to date
- [ ] Commit messages follow convention

## Related Issues
Closes #123
References #456
```

### Review Process

**Before Review:**
1. ✅ All CI checks pass
2. ✅ Self-review completed
3. ✅ Tests added and passing
4. ✅ Documentation updated
5. ✅ Commit messages follow convention

**During Review:**
- Address all review comments
- Keep discussions constructive and focused
- Update documentation if requested
- Add tests for edge cases if identified

**After Approval:**
- Squash commits if requested
- Ensure CI is still passing
- Merge when approved by maintainers

## Code Review Standards

### As a Reviewer

**Focus Areas:**
- ✅ **Correctness**: Does the code do what it claims?
- ✅ **Testing**: Are there adequate tests?
- ✅ **Security**: Are there security implications?
- ✅ **Performance**: Is it reasonably efficient?
- ✅ **Maintainability**: Is it readable and well-structured?
- ✅ **Documentation**: Is it properly documented?

**Review Guidelines:**
```markdown
# Good review feedback
"Consider using a context.WithTimeout here to prevent indefinite blocking"
"This error message could be more helpful to users - maybe include the cluster name?"
"Great test coverage! Could we add a test for the edge case where...?"

# Avoid
"This is wrong" (not helpful)
"Why did you do it this way?" (without suggesting alternatives)
"Nit: ..." (use suggestions for minor changes)
```

### As a Contributor

**Responding to Reviews:**
- Address all comments, even if just to acknowledge
- Ask for clarification if feedback is unclear
- Push additional commits for changes (don't force-push during review)
- Mark conversations as resolved after addressing
- Be open to learning and different approaches

## Release Process

### Version Management

We use [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Types

| Release Type | Description | Frequency |
|--------------|-------------|-----------|
| **Patch** | Bug fixes, minor improvements | As needed |
| **Minor** | New features, major improvements | Monthly |
| **Major** | Breaking changes, major refactoring | Quarterly |
| **RC** | Release candidates for testing | Before major releases |

## Community Guidelines

### Communication

- **Be respectful**: Treat all community members with respect
- **Be constructive**: Provide helpful feedback and suggestions  
- **Be patient**: Remember that people have different experience levels
- **Be collaborative**: We're all working toward the same goals

### Code of Conduct

We follow the [Contributor Covenant](https://www.contributor-covenant.org/):
- Use welcoming and inclusive language
- Respect differing viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what's best for the community

### Getting Help

- **GitHub Discussions**: For questions and general discussion
- **GitHub Issues**: For bug reports and feature requests
- **Code Comments**: For implementation questions during review
- **Documentation**: Check existing docs before asking questions

## Recognition

### Contributors

We recognize contributors in several ways:
- **Contributor List**: Maintained in CONTRIBUTORS.md
- **Release Notes**: Notable contributions highlighted in releases
- **GitHub Recognition**: Thanks in PR descriptions and issues

### Contribution Types

All contributions are valued:
- Code contributions (features, fixes, improvements)
- Documentation improvements and examples
- Issue reporting and triage
- Testing and quality assurance
- Community support and mentoring

---

**Thank You!** Your contributions make OpenFrame CLI better for everyone. Whether you're fixing a typo, adding a feature, or helping other users, every contribution matters. We look forward to collaborating with you!