# Contributing to OpenFrame CLI

We're excited that you're interested in contributing to OpenFrame CLI! This guide will help you get started with contributing to the project.

## Getting Started

### Prerequisites

Before you begin, ensure you have:

- Go 1.24.6 or higher
- Docker 20.10+ (with daemon running)
- kubectl 1.25+
- Helm 3.10+
- K3D 5.0+
- Git

### Development Environment Setup

1. **Fork and Clone the Repository**
```bash
# Fork the repo on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# Add upstream remote
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
```

2. **Set Up Your Development Environment**

Follow the [Development Environment Setup](./docs/development/setup/environment.md) guide for detailed IDE configuration, tools, and environment variables.

3. **Install Go Tools**
```bash
# Install essential Go development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/rakyll/gotest@latest
```

4. **Build and Test**
```bash
# Build the project
go build -o openframe main.go

# Run tests
go test ./...

# Run linter
golangci-lint run
```

## Development Workflow

### Branch Management

1. **Create a Feature Branch**
```bash
git checkout -b feature/your-feature-name
```

2. **Keep Your Branch Up to Date**
```bash
git fetch upstream
git rebase upstream/main
```

### Code Standards

#### Go Code Style
- Follow standard Go conventions and idioms
- Use `gofmt` and `goimports` for formatting
- Write clear, self-documenting code with meaningful names
- Include comments for exported functions and complex logic

#### Project Structure
```text
openframe-cli/
├── cmd/                    # CLI command definitions
├── internal/
│   ├── cluster/           # Cluster management logic
│   ├── chart/            # Chart and ArgoCD management
│   ├── dev/              # Development tools
│   ├── bootstrap/        # Environment bootstrapping
│   └── shared/           # Common utilities
├── docs/                 # Documentation
├── scripts/              # Build and utility scripts
└── main.go               # Application entry point
```

#### Testing Guidelines
- Write unit tests for all business logic
- Include integration tests for external tool interactions
- Use table-driven tests where appropriate
- Mock external dependencies using interfaces

Example test structure:
```go
func TestClusterCreate(t *testing.T) {
    tests := []struct {
        name     string
        input    ClusterConfig
        expected error
    }{
        {
            name: "valid cluster creation",
            input: ClusterConfig{
                Name:  "test-cluster",
                Nodes: 3,
            },
            expected: nil,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateCluster(tt.input)
            assert.Equal(t, tt.expected, err)
        })
    }
}
```

### Commit Guidelines

Follow conventional commit format:

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Build process or auxiliary tool changes

**Examples:**
```bash
git commit -m "feat(cluster): add support for custom node configurations"
git commit -m "fix(bootstrap): resolve ArgoCD installation timeout"
git commit -m "docs: update prerequisites and installation guide"
```

### Pull Request Process

1. **Prepare Your PR**
```bash
# Ensure your branch is up to date
git fetch upstream
git rebase upstream/main

# Run all checks
go fmt ./...
goimports -w .
golangci-lint run
go test ./...
```

2. **Submit Your Pull Request**
- Use a clear, descriptive title
- Include a detailed description of changes
- Reference any related issues
- Add screenshots/logs for UI or behavioral changes
- Ensure all CI checks pass

3. **PR Template**
```markdown
## Description
Brief description of the changes and their purpose.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
```

## Code Review Process

### For Contributors
- Respond promptly to review feedback
- Address all comments and suggestions
- Ask questions if feedback is unclear
- Update documentation if your changes affect user-facing behavior

### For Reviewers
- Provide constructive, actionable feedback
- Focus on code quality, maintainability, and correctness
- Check that tests adequately cover new functionality
- Verify documentation updates are included

## Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/cluster/...

# Run integration tests (requires Docker)
go test -tags=integration ./...
```

### Test Categories
- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test interactions with external tools (Docker, kubectl, etc.)
- **End-to-End Tests**: Test complete workflows from CLI to cluster

### Writing Good Tests
- Test both happy path and error conditions
- Use descriptive test names that explain what is being tested
- Keep tests focused and atomic
- Use test fixtures and helpers to reduce duplication

## Documentation

### Types of Documentation
- **Code Comments**: Explain complex logic and public APIs
- **README Updates**: Keep installation and usage instructions current
- **Developer Docs**: Architecture, design decisions, and development guides
- **User Guides**: Step-by-step tutorials and reference material

### Documentation Guidelines
- Write clear, concise instructions
- Include code examples where helpful
- Update docs when making user-facing changes
- Use proper Markdown formatting

## Release Process

### Version Management
We use semantic versioning (SemVer):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Creating a Release
1. Update version in `main.go`
2. Update `CHANGELOG.md`
3. Create and push version tag
4. GitHub Actions handles the build and release

## Issue Management

### Reporting Issues
When reporting bugs or requesting features:
- Check existing issues first
- Use issue templates when available
- Provide detailed reproduction steps for bugs
- Include system information and versions

### Working on Issues
- Comment on issues before starting work
- Ask for clarification if requirements are unclear
- Link your PR to the issue when ready

## Community Guidelines

### Communication Channels
- **Primary Support**: [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **Development Discussion**: GitHub PR comments and code reviews
- **Feature Requests**: GitHub Issues

### Code of Conduct
- Be respectful and inclusive in all interactions
- Focus on constructive feedback and solutions
- Help newcomers get started
- Follow the project's technical standards and conventions

## Getting Help

Need assistance? Here's how to get help:

1. **Development Questions**: Ask in OpenMSP Slack #dev channel
2. **Documentation Issues**: Create a GitHub issue with the "documentation" label
3. **Bug Reports**: File a GitHub issue with reproduction steps
4. **Feature Ideas**: Discuss in Slack first, then create GitHub issue

## External Dependencies

### CLI Tools Integration
This repository contains OpenFrame CLI code. The main OpenFrame application code is maintained separately:

- **OpenFrame Main Repository**: [flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **CLI Documentation**: [CLI Documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

When contributing CLI-related changes, coordinate with the main repository team through Slack.

## Acknowledgments

Thank you for contributing to OpenFrame CLI! Your efforts help make IT operations more accessible and cost-effective for MSPs worldwide.

---

**Questions?** Join our [OpenMSP Slack community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) - we're here to help!