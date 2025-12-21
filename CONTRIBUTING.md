# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Code Guidelines](#code-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Code of Conduct](#code-of-conduct)

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Docker Desktop** - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** (optional) - For build scripts

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/openframe-cli.git
   cd openframe-cli
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
   ```

## Development Environment

### Setup

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Build the CLI**:
   ```bash
   go build -o openframe .
   ```

3. **Run tests**:
   ```bash
   go test ./...
   ```

4. **Run with verbose output**:
   ```bash
   ./openframe --verbose cluster create
   ```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our [code guidelines](#code-guidelines)

3. **Test your changes**:
   ```bash
   # Run unit tests
   go test ./...
   
   # Test CLI functionality
   ./openframe cluster create test-cluster
   ./openframe cluster delete test-cluster
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

## Code Guidelines

### Go Standards

- Follow standard Go conventions and formatting
- Use `gofmt` and `goimports` for code formatting
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write idiomatic Go code

### Project Structure

```
├── cmd/                    # Command implementations
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/          # Cluster management commands
│   ├── chart/            # Chart management commands
│   └── dev/              # Development commands
├── internal/              # Internal packages
│   ├── cluster/          # Cluster operations
│   ├── chart/            # Chart operations
│   ├── dev/              # Development tools
│   └── shared/           # Shared utilities
├── docs/                  # Documentation
└── tests/                # Integration tests
```

### Code Style

- **Package naming**: Use lowercase, single words when possible
- **Function naming**: Use camelCase, start with uppercase for exported functions
- **Error handling**: Always handle errors appropriately
- **Comments**: Document exported functions and complex logic
- **Interface design**: Keep interfaces small and focused

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

Examples:
```
feat(cluster): add cluster cleanup command
fix(bootstrap): resolve ArgoCD installation timeout
docs: update quick start guide
test(cluster): add unit tests for create command
```

## Testing

### Unit Tests

Write unit tests for new functionality:

```go
func TestClusterCreate(t *testing.T) {
    // Test cluster creation logic
    service := NewClusterService()
    err := service.CreateCluster("test-cluster", config)
    assert.NoError(t, err)
}
```

### Integration Tests

Test CLI commands end-to-end:

```bash
# Test cluster creation
./openframe cluster create test-cluster --dry-run

# Test bootstrap process
./openframe bootstrap --dry-run --deployment-mode=oss-tenant
```

### Manual Testing

Before submitting PRs:

1. Test on your local system
2. Verify all commands work as expected
3. Test error scenarios
4. Check that help messages are clear

## Pull Request Process

### Before Submitting

- [ ] Code follows project guidelines
- [ ] Tests pass locally
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] Branch is up to date with main

### Submission

1. **Open a Pull Request** with:
   - Clear title describing the change
   - Detailed description of what was changed and why
   - Link to any related issues
   - Screenshots/examples if applicable

2. **PR Template**:
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Manual testing completed
   - [ ] Integration tests pass
   
   ## Checklist
   - [ ] Code follows style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   ```

### Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by maintainers
3. **Manual testing** for significant changes
4. **Approval** from at least one maintainer

### After Approval

- Maintainers will merge using "Squash and merge"
- Delete your feature branch after merge
- Update your local repository:
  ```bash
  git checkout main
  git pull upstream main
  ```

## Issue Reporting

### Bug Reports

When reporting bugs, include:

- **Environment details**: OS, Go version, Docker version
- **CLI version**: `openframe version`
- **Steps to reproduce**: Exact commands and configuration
- **Expected vs actual behavior**
- **Error messages**: Full error output with `--verbose`
- **Logs**: Relevant log files

### Feature Requests

When requesting features:

- **Use case**: Describe the problem you're solving
- **Proposed solution**: How you envision the feature working
- **Alternatives**: Other solutions you've considered
- **Additional context**: Screenshots, examples, etc.

### Security Issues

For security vulnerabilities:

- **Do not** open public issues
- Email security@flamingo.run with details
- Include steps to reproduce if possible

## Code of Conduct

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what's best for the community
- Showing empathy towards other community members

**Unacceptable behavior includes:**
- Harassment, trolling, or discriminatory comments
- Personal or political attacks
- Public or private harassment
- Publishing others' private information without permission
- Other conduct inappropriate for a professional setting

### Enforcement

Report any violations to conduct@flamingo.run. All reports will be reviewed and investigated promptly and fairly.

## Development Resources

### Useful Links

- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://cobra.dev/)
- [K3d Documentation](https://k3d.io/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)

### Getting Help

- **Documentation**: Check [docs/](./docs/) for detailed guides
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions
- **Community**: Join our community channels (links in README)

## Recognition

Contributors will be recognized in:
- GitHub contributor list
- Release notes for significant contributions
- Project documentation

Thank you for contributing to OpenFrame CLI! 🎉

---

For questions about contributing, please open an issue or reach out to the maintainers.