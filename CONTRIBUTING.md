# Contributing to loko

Thank you for your interest in contributing to loko! We're building this tool in public and welcome contributions from developers of all experience levels.

## 🎯 Ways to Contribute

- 🐛 **Report bugs** - Help us find and fix issues
- 💡 **Suggest features** - Share ideas for improvements
- 📖 **Improve documentation** - Clarify, expand, or fix docs
- 🔧 **Submit code** - Bug fixes, features, tests
- 🎨 **Design templates** - Create C4 templates for common patterns
- 🧪 **Test and validate** - Try loko on real projects and report findings

## 🚀 Getting Started

### Prerequisites

- **Go 1.23+** ([install](https://go.dev/doc/install))
- **d2** ([install](https://d2lang.com))
- **git**
- Optional: **veve-cli** (for PDF tests)

### Development Setup

```bash
# 1. Fork and clone
git clone https://github.com/madstone-tech/loko
cd loko

# 2. Install dependencies
go mod download

# 3. Install d2
brew install d2  # macOS
# or download from https://github.com/terrastruct/d2/releases

# 4. Run tests
go test ./...

# 5. Build
go build -o loko .

# 6. Try it out
./loko --help
```

## 🏗️ Architecture Guide

loko uses **Clean Architecture**. Understanding this will help you contribute effectively.

### The Dependency Rule

Dependencies point **inward**. Inner layers never know about outer layers.

```
┌─────────────────────────────────────────────────┐
│           Interfaces (CLI, MCP, API)            │  ← Thin wrappers
├─────────────────────────────────────────────────┤
│         Adapters (d2, filesystem, toon)         │  ← Implements Ports
├─────────────────────────────────────────────────┤
│        Use Cases (CreateSystem, Build)          │  ← Defines Ports
├─────────────────────────────────────────────────┤
│      Entities (Project, System, Container)      │  ← Pure Go
└─────────────────────────────────────────────────┘
```

### Project Structure

```
loko/
├── cmd/                      # CLI commands (thin wrappers)
├── internal/
│   ├── core/                 # THE HEART - zero external deps
│   │   ├── entities/         # Domain objects
│   │   ├── usecases/         # Application logic + ports
│   │   └── errors/           # Domain errors
│   ├── adapters/             # Infrastructure
│   │   ├── config/           # TOML loader
│   │   ├── filesystem/       # File operations
│   │   ├── d2/               # Diagram renderer
│   │   ├── encoding/         # JSON + TOON
│   │   └── html/             # Site builder
│   ├── mcp/                  # MCP server
│   ├── api/                  # HTTP API
│   └── ui/                   # Terminal UI
├── templates/                # Starter templates
└── docs/                     # Documentation
```

### Where to Add Code

| I want to...                | Where to add it                           |
| --------------------------- | ----------------------------------------- |
| Add a new entity field      | `internal/core/entities/`                 |
| Add validation logic        | `internal/core/entities/` (on the entity) |
| Add a new operation         | `internal/core/usecases/` (new use case)  |
| Add a CLI command           | `cmd/` (thin wrapper calling use case)    |
| Add an MCP tool             | `internal/mcp/tools/` (thin wrapper)      |
| Add an API endpoint         | `internal/api/handlers/` (thin wrapper)   |
| Change how files are stored | `internal/adapters/filesystem/`           |
| Change diagram rendering    | `internal/adapters/d2/`                   |
| Add output format           | `internal/adapters/encoding/`             |

### Adding a New Use Case

1. Define input/output structs in `internal/core/usecases/your_usecase.go`
2. If you need new infrastructure, add interface to `internal/core/usecases/ports.go`
3. Implement the use case
4. Add adapter implementation if needed in `internal/adapters/`
5. Wire it up in `main.go`
6. Add thin wrappers in `cmd/`, `internal/mcp/tools/`, `internal/api/handlers/`

### Example: Adding "Archive System" Feature

```go
// 1. internal/core/usecases/archive_system.go
type ArchiveSystemInput struct {
    SystemName string
}

type ArchiveSystemOutput struct {
    ArchivedAt time.Time
    BackupPath string
}

type ArchiveSystemUseCase struct {
    projects ProjectRepository
    archiver Archiver           // New port
}

func (uc *ArchiveSystemUseCase) Execute(ctx context.Context, input ArchiveSystemInput) (*ArchiveSystemOutput, error) {
    // Business logic here
}
```

```go
// 2. internal/core/usecases/ports.go (add new port)
type Archiver interface {
    Archive(ctx context.Context, path string) (string, error)
}
```

```go
// 3. internal/adapters/filesystem/archiver.go
type ZipArchiver struct{}

func (a *ZipArchiver) Archive(ctx context.Context, path string) (string, error) {
    // Implementation
}
```

```go
// 4. cmd/archive.go (thin CLI wrapper - under 50 lines!)
func archiveCmd(uc *usecases.ArchiveSystemUseCase) *cobra.Command {
    return &cobra.Command{
        Use: "archive [system]",
        RunE: func(cmd *cobra.Command, args []string) error {
            output, err := uc.Execute(ctx, usecases.ArchiveSystemInput{
                SystemName: args[0],
            })
            // Format and display output
        },
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests (Use Cases)

Mock the ports to test business logic in isolation:

```go
func TestArchiveSystemUseCase(t *testing.T) {
    mockRepo := &MockProjectRepo{...}
    mockArchiver := &MockArchiver{...}

    uc := usecases.NewArchiveSystemUseCase(mockRepo, mockArchiver)

    output, err := uc.Execute(ctx, usecases.ArchiveSystemInput{
        SystemName: "PaymentService",
    })

    assert.NoError(t, err)
    assert.True(t, mockArchiver.ArchiveCalled)
}
```

### Integration Tests

Use real adapters with temp directories:

```go
func TestArchiveSystemIntegration(t *testing.T) {
    tmpDir := t.TempDir()
    // Set up real file system
    // Use real adapters
    // Verify actual files created
}
```

### Golden Tests

For output formatting:

```go
func TestBuildHTMLGolden(t *testing.T) {
    got := builder.Build(project)
    golden.Assert(t, got, "testdata/expected.html")
}
```

## 🔧 Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/bug-description
```

### 2. Make Changes

- Follow Go best practices ([Effective Go](https://go.dev/doc/effective_go))
- Write tests for new functionality
- Update documentation as needed
- Run `go fmt` before committing

### 3. Test Your Changes

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/core/usecases/...

# Run integration tests
go test -tags=integration ./tests/integration/...
```

### 4. Commit

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
feat: add MCP tool for component creation
fix: resolve d2 caching issue on Windows
docs: update installation guide for Homebrew
test: add integration tests for watch mode
chore: update dependencies
```

### 5. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then open a Pull Request with:

- Clear title and description
- Reference any related issues (#123)
- Screenshots/demos if applicable

## 📝 Code Style

### General Principles

- **Simplicity** - Prefer clear code over clever code
- **Interfaces** - Use interfaces for dependencies
- **Error handling** - Always handle errors with context
- **Documentation** - Public APIs must have godoc comments

### Example: Good Error Handling

```go
func (e *Engine) RenderDiagram(d2File string) error {
    if !strings.HasSuffix(d2File, ".d2") {
        return &errors.ValidationError{
            Path:    d2File,
            Message: "file must have .d2 extension",
        }
    }

    if err := e.renderer.Render(d2File); err != nil {
        return fmt.Errorf("render diagram %s: %w", d2File, err)
    }

    return nil
}
```

### Interface Design

```go
// Good - testable and swappable
type DiagramRenderer interface {
    Render(ctx context.Context, opts RenderOptions) (*RenderResult, error)
    Available() bool
}

type Engine struct {
    renderer DiagramRenderer  // Can mock in tests
}
```

## 🐛 Reporting Bugs

Include:

- loko version (`loko --version`)
- Operating system and version
- Go version (`go version`)
- d2 version (`d2 --version`)
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (run with `--debug`)

## 💡 Suggesting Features

Consider:

- Does it align with loko's core mission?
- Is it simple to use?
- Can it be composed with existing features?
- Would it benefit most users or is it niche?

## 📦 Adding Dependencies

We minimize dependencies. Before adding a new one:

1. **Check if stdlib can do it** - Go's standard library is excellent
2. **Evaluate maintenance** - Is it actively maintained?
3. **Check size** - Will it bloat the binary?
4. **Discuss first** - Open an issue to discuss necessity

## 🏗️ Architecture Decisions

Major decisions are documented in [ADRs](docs/adr/):

- [ADR 0001: Clean Architecture](docs/adr/0001-clean-architecture.md)
- [ADR 0002: Token-Efficient MCP](docs/adr/0002-token-efficient-mcp.md)
- [ADR 0003: TOON Format Support](docs/adr/0003-toon-format.md)

Discuss in issues before implementing major changes.

## 🎯 Pull Request Checklist

Before submitting:

- [ ] Tests pass (`go test ./...`)
- [ ] Code is formatted (`go fmt ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Documentation updated (if needed)
- [ ] Changelog updated (CHANGELOG.md)
- [ ] Commit messages follow convention
- [ ] PR description is clear and complete
- [ ] Interface code is under 50 lines (for new commands/tools)

## 🌟 Recognition

Contributors are recognized in:

- CHANGELOG.md (for each release)
- README.md (top contributors)
- GitHub release notes

## ❓ Questions?

- **General questions** → [GitHub Discussions](https://github.com/madstone-tech/loko/discussions)
- **Bug reports** → [GitHub Issues](https://github.com/madstone-tech/loko/issues)
- **Security issues** → Email <security@madstone.tech>

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

---

**Thank you for contributing to loko!** 🪇

Every contribution, no matter how small, helps make architecture documentation better for everyone.
