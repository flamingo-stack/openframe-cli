# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! We welcome contributions from the community and are grateful for your support.

## 📋 Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Community](#community)

## 🚀 Getting Started

Before you begin:

1. **Join our Slack**: Connect with the community on [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
2. **Read the docs**: Familiarize yourself with the [documentation](./docs/README.md)
3. **Check existing issues**: Browse [open issues](https://github.com/flamingo-stack/openframe-cli/issues) to see what needs help

## 🛠 Development Environment

### Prerequisites

**Hardware Requirements:**
- Minimum: 24GB RAM, 6 CPU cores, 50GB disk space
- Recommended: 32GB RAM, 12 CPU cores, 100GB disk space

**Software Requirements:**
- **Go**: 1.21 or later
- **Docker**: Docker Desktop or Docker Engine
- **Git**: Latest version
- **Make**: For build automation

### Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR-USERNAME/openframe-cli.git
   cd openframe-cli
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the CLI:**
   ```bash
   make build
   # Or manually:
   go build -o openframe .
   ```

4. **Run tests:**
   ```bash
   make test
   # Or manually:
   go test ./...
   ```

5. **Test the CLI locally:**
   ```bash
   ./openframe --help
   ```

### Development Workflow

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**

3. **Run tests and linting:**
   ```bash
   make test
   make lint
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

5. **Push and create a pull request**

## 📝 Code Standards

### Go Code Style

We follow standard Go conventions with some additional guidelines:

- **Formatting**: Use `gofmt` and `goimports`
- **Linting**: Use `golangci-lint` with our configuration
- **Naming**: Follow Go naming conventions (CamelCase for exported, camelCase for unexported)
- **Comments**: Document all exported functions and types
- **Error handling**: Always handle errors explicitly, never ignore them

### Architecture Patterns

OpenFrame CLI follows a hexagonal architecture:

```
cmd/                 # Command definitions (Cobra)
├── cluster/         # Cluster command group
├── chart/          # Chart command group
└── dev/            # Development command group

internal/
├── cluster/        # Cluster business logic
├── chart/          # Chart business logic
├── bootstrap/      # Bootstrap orchestration
├── dev/           # Development tools
└── shared/        # Shared utilities
    ├── executor/   # Command execution
    ├── ui/        # Terminal UI components
    ├── config/    # Configuration management
    └── errors/    # Error handling
```

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

**Examples:**
```
feat(cluster): add support for custom node labels
fix(chart): resolve ArgoCD installation timeout issue
docs: update contributing guidelines
test(cluster): add unit tests for k3d provider
```

## 🧪 Testing

### Test Types

1. **Unit Tests**: Test individual components in isolation
   ```bash
   go test ./internal/cluster/...
   go test ./internal/chart/...
   ```

2. **Integration Tests**: Test component interactions
   ```bash
   go test -tags=integration ./...
   ```

3. **End-to-End Tests**: Test complete workflows
   ```bash
   go test -tags=e2e ./tests/e2e/...
   ```

### Test Requirements

- **Coverage**: Maintain at least 80% code coverage for new code
- **Table-driven tests**: Use table-driven tests for multiple scenarios
- **Mocking**: Use interfaces and mocks for external dependencies
- **Clean up**: Always clean up resources in tests (clusters, files, etc.)

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests (requires Docker)
make test-integration

# With coverage
make test-coverage

# Specific package
go test ./internal/cluster/providers/k3d/
```

## 🔄 Pull Request Process

### Before Submitting

1. **Update documentation** if you're adding/changing functionality
2. **Add tests** for new features or bug fixes
3. **Run the full test suite** and ensure all tests pass
4. **Update CHANGELOG.md** for significant changes
5. **Ensure commits follow** our commit message format

### PR Description Template

```markdown
## What this PR does

Brief description of the change and which issue it fixes.

## Why this change is needed

Explanation of the problem this PR solves.

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed
- [ ] Documentation updated

## Checklist

- [ ] Code follows project conventions
- [ ] Self-review completed
- [ ] Tests pass locally
- [ ] Documentation updated
- [ ] Changelog updated (if needed)
```

### Review Process

1. **Automated checks**: CI/CD runs tests and linting
2. **Code review**: At least one maintainer reviews the code
3. **Testing**: Reviewer tests the changes locally if needed
4. **Approval**: PR is approved and merged

## 🐛 Issue Reporting

### Bug Reports

When reporting bugs, please include:

1. **Environment information:**
   - Operating system and version
   - Go version
   - Docker version
   - OpenFrame CLI version

2. **Steps to reproduce**

3. **Expected vs actual behavior**

4. **Logs and error messages**

5. **Additional context**

Use our [bug report template](https://github.com/flamingo-stack/openframe-cli/issues/new?template=bug_report.md).

### Feature Requests

For feature requests, please include:

1. **Problem description**: What problem does this solve?
2. **Proposed solution**: How should we solve it?
3. **Alternatives considered**: What other approaches did you consider?
4. **Additional context**: Any other relevant information

Use our [feature request template](https://github.com/flamingo-stack/openframe-cli/issues/new?template=feature_request.md).

## 💬 Community

### Communication Channels

- **Primary**: [OpenMSP Slack Community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **Issues**: GitHub Issues for bug reports and feature requests
- **Discussions**: Use Slack for general questions and discussions

### Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

### Getting Help

1. **Documentation**: Check the [docs](./docs/README.md) first
2. **Slack**: Ask questions in our Slack community
3. **GitHub Issues**: Open an issue for bugs or feature requests

## 📚 Additional Resources

- [Development Documentation](./docs/development/README.md)
- [Architecture Overview](./docs/reference/architecture/overview.md)
- [Getting Started Guide](./docs/getting-started/quick-start.md)
- [OpenFrame Website](https://openframe.ai)

## 🙏 Recognition

Contributors will be recognized in:
- Release notes for significant contributions
- GitHub contributors list
- Our community Slack

Thank you for contributing to OpenFrame CLI! 🎉

---

For questions or support, reach out on [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) or open an issue.