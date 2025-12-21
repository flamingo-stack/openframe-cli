# Development Documentation

Welcome to the OpenFrame CLI development documentation. This section provides comprehensive guides for developers who want to contribute to OpenFrame, extend its functionality, or understand its architecture.

## 📚 Documentation Overview

This development documentation is organized into focused sections to help you find exactly what you need:

### 🛠️ Setup & Environment

| Guide | Purpose | Audience |
|-------|---------|----------|
| **[Environment Setup](./setup/environment.md)** | IDE configuration, tools, and extensions | All developers |
| **[Local Development](./setup/local-development.md)** | Running OpenFrame locally, debugging, hot reload | Contributors |

### 🏗️ Architecture & Design

| Guide | Purpose | Audience |
|-------|---------|----------|
| **[Architecture Overview](./architecture/overview.md)** | System design, components, data flow | All developers |

### 🧪 Testing & Quality

| Guide | Purpose | Audience |
|-------|---------|----------|
| **[Testing Overview](./testing/overview.md)** | Test structure, running tests, writing tests | Contributors |

### 🤝 Contributing

| Guide | Purpose | Audience |
|-------|---------|----------|
| **[Contributing Guidelines](./contributing/guidelines.md)** | Code standards, PR process, review checklist | Contributors |

## 🚀 Quick Navigation

### For New Contributors

1. **Start Here**: [Environment Setup](./setup/environment.md) - Set up your development environment
2. **Get Running**: [Local Development](./setup/local-development.md) - Build and run OpenFrame locally
3. **Learn the Code**: [Architecture Overview](./architecture/overview.md) - Understand the codebase
4. **Follow the Rules**: [Contributing Guidelines](./contributing/guidelines.md) - Coding standards and process

### For Platform Engineers

1. **Understanding**: [Architecture Overview](./architecture/overview.md) - High-level system design
2. **Extension**: [Environment Setup](./setup/environment.md) - Tools for extending OpenFrame
3. **Testing**: [Testing Overview](./testing/overview.md) - Quality assurance practices

### For Users Who Want to Understand

1. **How it Works**: [Architecture Overview](./architecture/overview.md) - System internals
2. **Development Flow**: [Local Development](./setup/local-development.md) - See the development process

## 🛠️ Development Stack

OpenFrame CLI is built with:

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Core Language** | Go 1.19+ | CLI application and business logic |
| **CLI Framework** | Cobra | Command structure and flag parsing |
| **Kubernetes Client** | client-go | Kubernetes API interactions |
| **Container Orchestration** | K3d | Local Kubernetes clusters |
| **Package Management** | Helm | Chart installation and management |
| **GitOps** | ArgoCD | Application deployment and sync |
| **Testing** | Go testing + Testify | Unit and integration tests |
| **Documentation** | Markdown + Mermaid | Technical documentation |

## 🎯 Common Development Tasks

### Setting Up Development Environment

```bash
# 1. Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# 2. Set up development tools
make dev-setup  # See Environment Setup guide

# 3. Build and test
make build
make test

# 4. Run locally
./openframe bootstrap --help
```

### Making Changes

```bash
# 1. Create feature branch
git checkout -b feature/your-feature-name

# 2. Make changes and test
make test
make lint

# 3. Build and verify
make build
./openframe cluster create test-cluster

# 4. Submit PR (see Contributing Guidelines)
```

### Running Tests

```bash
# Run all tests
make test

# Run specific test package
go test ./internal/cluster/...

# Run with coverage
make test-coverage

# Run integration tests
make test-integration
```

## 🗂️ Project Structure Overview

```text
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── bootstrap/          # Bootstrap command
│   ├── cluster/            # Cluster management commands
│   ├── chart/              # Chart management commands
│   └── dev/                # Development tool commands
├── internal/               # Internal packages (not exported)
│   ├── bootstrap/          # Bootstrap business logic
│   ├── cluster/            # Cluster management logic
│   │   ├── models/         # Data structures and validation
│   │   ├── services/       # Business logic
│   │   ├── ui/             # Interactive prompts
│   │   └── utils/          # Shared utilities
│   ├── chart/              # Chart management logic
│   ├── dev/                # Development tools logic
│   └── shared/             # Shared components
├── docs/                   # Documentation
│   ├── getting-started/    # User guides
│   ├── development/        # This section
│   └── reference/          # Technical reference
├── scripts/                # Build and development scripts
├── tests/                  # Test files and fixtures
└── Makefile               # Build automation
```

## 🎓 Learning Path

### Week 1: Getting Familiar
- [ ] Read [Architecture Overview](./architecture/overview.md)
- [ ] Set up [Development Environment](./setup/environment.md)
- [ ] Complete [Local Development](./setup/local-development.md) setup
- [ ] Run existing tests and explore codebase

### Week 2: Contributing
- [ ] Read [Contributing Guidelines](./contributing/guidelines.md)
- [ ] Review [Testing Overview](./testing/overview.md)
- [ ] Find a "good first issue" and implement
- [ ] Submit your first PR

### Week 3: Advanced Development
- [ ] Understand internal package organization
- [ ] Write comprehensive tests for new features
- [ ] Review and contribute to documentation
- [ ] Help review other contributors' PRs

## 🔧 Development Tools & IDE Setup

### Recommended IDEs

| IDE | Extensions | Configuration |
|-----|------------|---------------|
| **VS Code** | Go, Kubernetes, YAML | See [Environment Setup](./setup/environment.md) |
| **GoLand** | Built-in Go support | Native Kubernetes integration |
| **Vim/Neovim** | vim-go, coc-go | Lightweight terminal-based |

### Required Tools

- **Go 1.19+**: Core language
- **Docker**: For K3d clusters
- **kubectl**: Kubernetes CLI
- **helm**: Package management
- **make**: Build automation
- **git**: Version control

See [Environment Setup](./setup/environment.md) for detailed installation instructions.

## 📋 Development Workflows

### Feature Development

1. **Planning**: Create GitHub issue with requirements
2. **Design**: Document architecture changes if needed
3. **Implementation**: Follow coding standards and patterns
4. **Testing**: Add comprehensive test coverage
5. **Documentation**: Update relevant docs
6. **Review**: Submit PR following contributing guidelines

### Bug Fixes

1. **Reproduction**: Create test case that reproduces the bug
2. **Investigation**: Understand root cause
3. **Fix**: Implement minimal, targeted fix
4. **Verification**: Ensure fix resolves issue without regressions
5. **Testing**: Add test to prevent future regressions

### Documentation Updates

1. **Identify Gap**: Find missing or outdated documentation
2. **Research**: Understand current behavior and requirements
3. **Write**: Create clear, actionable documentation
4. **Review**: Test instructions with fresh environment
5. **Integrate**: Ensure proper linking and navigation

## 🤝 Getting Help

### For Development Questions

- **GitHub Issues**: Technical questions and bug reports
- **GitHub Discussions**: Feature ideas and general questions
- **Code Comments**: Inline documentation and examples
- **Architecture Docs**: Design decisions and patterns

### For Contributing Questions

- **[Contributing Guidelines](./contributing/guidelines.md)**: Process and standards
- **PR Reviews**: Feedback on specific changes
- **Code Review**: Learning from existing implementations

## 🎯 Next Steps

Choose your path based on your goals:

**🏁 I want to contribute code**
→ Start with [Environment Setup](./setup/environment.md)

**🔍 I want to understand the architecture**
→ Read [Architecture Overview](./architecture/overview.md)

**🧪 I want to improve testing**
→ Check out [Testing Overview](./testing/overview.md)

**📝 I want to improve documentation**
→ Review [Contributing Guidelines](./contributing/guidelines.md)

**🚀 I want to build on OpenFrame**
→ Study [Local Development](./setup/local-development.md)

---

**Happy coding!** 🎉 The OpenFrame CLI development community welcomes your contributions, questions, and ideas.