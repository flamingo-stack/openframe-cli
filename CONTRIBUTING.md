# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and instructions for contributing to the project.

## 🚀 Quick Start for Contributors

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/openframe-cli.git
   cd openframe-cli
   ```

2. **Set Up Development Environment**
   ```bash
   # Install Go 1.21+
   go version
   
   # Install dependencies
   go mod download
   
   # Run tests
   go test ./...
   ```

3. **Make Your Changes**
   ```bash
   git checkout -b feature/your-feature-name
   # Make your changes
   git commit -m "feat: add your feature description"
   git push origin feature/your-feature-name
   ```

4. **Submit Pull Request**

## 📋 Prerequisites

### System Requirements
- **Go**: Version 1.21 or higher
- **Docker**: For running K3D clusters and testing
- **Git**: For version control
- **Make**: For running build tasks (optional)

### Development Tools (Optional but Recommended)
- **kubectl**: For testing Kubernetes interactions
- **k3d**: For testing cluster operations
- **helm**: For testing chart installations

## 🏗 Project Structure

```
openframe-cli/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command and CLI setup
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart installation commands
│   └── dev/               # Development workflow commands
├── internal/              # Internal packages
│   ├── bootstrap/         # Bootstrap service logic
│   ├── cluster/           # Cluster management services
│   ├── chart/             # Chart installation services
│   ├── dev/               # Development tool services
│   └── shared/            # Shared utilities and UI
├── pkg/                   # Public API packages
├── docs/                  # Documentation
├── scripts/               # Build and utility scripts
└── tests/                 # Integration tests
```

## 🧪 Development Workflow

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests (requires Docker)
go test -tags=integration ./...

# Run specific test package
go test ./internal/cluster/...
```

### Building the CLI

```bash
# Build for current platform
go build -o openframe .

# Build for all platforms
make build-all

# Install locally
go install .
```

### Testing Your Changes

```bash
# Test cluster creation
./openframe cluster create test-cluster

# Test bootstrap flow
./openframe bootstrap --deployment-mode=oss-tenant

# Test with verbose output
./openframe --verbose cluster status test-cluster
```

## 📝 Code Style and Standards

### Go Conventions
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Write comprehensive tests for new functionality
- Document exported functions and types
- Handle errors appropriately with proper context

### CLI Design Principles
- **Interactive by default**: Provide guided wizards for complex operations
- **Rich terminal UI**: Use pterm for consistent, beautiful output
- **Fail fast**: Validate prerequisites and inputs early
- **Clear feedback**: Provide detailed progress and error messages
- **Testable**: Design with dependency injection for easy testing

### Commit Message Format

We follow conventional commits:

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (no logic changes)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

**Examples:**
```bash
feat(cluster): add cluster restart command
fix(chart): resolve ArgoCD installation timeout
docs(readme): update installation instructions
test(dev): add intercept service tests
```

## 🐛 Reporting Issues

### Before Submitting an Issue

1. **Search existing issues** to avoid duplicates
2. **Update to the latest version** to see if the issue persists
3. **Gather information** about your environment:
   - Operating system and version
   - Go version (`go version`)
   - Docker version (`docker version`)
   - OpenFrame CLI version (`openframe --version`)

### Issue Template

```markdown
**Bug Description**
A clear description of the bug.

**Steps to Reproduce**
1. Run command: `openframe cluster create`
2. Select options: X, Y, Z
3. See error

**Expected Behavior**
What you expected to happen.

**Environment**
- OS: macOS 14.0
- Go: 1.21.3
- Docker: 24.0.6
- OpenFrame CLI: v1.0.0

**Additional Context**
Any additional information, logs, or screenshots.
```

## ✨ Submitting Pull Requests

### Pull Request Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/descriptive-name
   ```

2. **Make Changes**
   - Write code following project conventions
   - Add tests for new functionality
   - Update documentation if needed

3. **Test Thoroughly**
   ```bash
   go test ./...
   go test -tags=integration ./...
   ```

4. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat(scope): descriptive commit message"
   ```

5. **Push and Create PR**
   ```bash
   git push origin feature/descriptive-name
   # Create pull request on GitHub
   ```

### Pull Request Checklist

- [ ] Code follows project style and conventions
- [ ] Tests pass locally (`go test ./...`)
- [ ] New functionality includes tests
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow conventional format
- [ ] No breaking changes (or clearly documented)
- [ ] PR description explains the changes and motivation

### Pull Request Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that causes existing functionality to change)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Screenshots (if applicable)
Add screenshots or terminal output for UI changes.

## Additional Notes
Any additional context or considerations.
```

## 🏷 Release Process

Releases are handled by maintainers:

1. **Version Bump**: Update version in `main.go`
2. **Changelog**: Update `CHANGELOG.md` with new features and fixes
3. **Tag Release**: Create and push git tag (`v1.0.0`)
4. **GitHub Release**: Automated builds create release artifacts
5. **Documentation**: Update installation instructions if needed

## 🤝 Community Guidelines

### Be Respectful
- Use inclusive language
- Provide constructive feedback
- Help newcomers learn and contribute

### Be Collaborative
- Respond to questions and reviews promptly
- Share knowledge and best practices
- Coordinate with maintainers on major changes

### Be Patient
- Code review takes time for quality assurance
- Maintainers balance multiple priorities
- Complex features may need iteration

## 📞 Getting Help

### Discussion Channels
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and community support
- **Documentation**: Check the [docs/](./docs/) folder first

### Maintainer Contact
For sensitive issues or questions:
- Open a private issue with the maintainer team
- Email: [support@flamingo.run](mailto:support@flamingo.run)

## 📄 License

By contributing to OpenFrame CLI, you agree that your contributions will be licensed under the same license as the project (Flamingo AI Unified License v1.0).

---

Thank you for contributing to OpenFrame CLI! Your contributions help make this tool better for the entire community. 🙏