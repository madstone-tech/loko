# CLAUDE.md - AI Assistant Guide for loko

## Project Overview

**loko** is a C4 model architecture documentation tool with LLM integration via MCP.

## Key Files

| File | Purpose |
|------|---------|
| `.specify/memory/constitution.md` | **READ FIRST** - Project principles and coding standards |
| `.specify/memory/spec.md` | Full specification with user stories and requirements |
| `.specify/memory/tasks.md` | Task tracking - check current progress here |
| `docs/adr/` | Architecture Decision Records |
| `ISSUES.md` | GitHub issue templates (reference) |

## Architecture

```
internal/
├── core/                     # ZERO external dependencies
│   ├── entities/             # Domain objects (✅ Complete)
│   ├── usecases/             # Application logic + ports.go interfaces
│   └── errors/               # (moved to entities/errors.go)
├── adapters/                 # Infrastructure implementations
│   ├── filesystem/           # ProjectRepository
│   ├── d2/                   # DiagramRenderer
│   ├── ason/                 # TemplateEngine
│   ├── html/                 # SiteBuilder
│   ├── encoding/             # JSON/TOON encoders
│   └── config/               # TOML config loader
├── mcp/                      # MCP server (thin layer)
├── api/                      # HTTP API (thin layer)
└── ui/                       # Lipgloss styles
cmd/                          # CLI commands (thin layer)
```

## Critical Rules

### 1. Clean Architecture

```
core/ imports nothing from adapters/, mcp/, api/, cmd/
adapters/ imports from core/
mcp/, api/, cmd/ import from core/ and adapters/
```

### 2. Interface-First

All external dependencies go through interfaces in `internal/core/usecases/ports.go`:

```go
// Good - depends on interface
type CreateSystemUseCase struct {
    projectRepo ProjectRepository
    templateEngine TemplateEngine
}

// Bad - depends on concrete implementation
type CreateSystemUseCase struct {
    repo *filesystem.ProjectRepo  // NO!
}
```

### 3. Thin Handlers

CLI commands, MCP tools, and API handlers should be <50 lines:

```go
// Good - thin handler
func (c *NewCommand) Run(args []string) error {
    input := usecases.CreateSystemInput{Name: args[0]}
    output, err := c.useCase.Execute(input)
    if err != nil {
        return c.ui.Error(err)
    }
    return c.ui.Success(output)
}
```

### 4. Entity Validation

All validation in entities, not use cases:

```go
// Good - validation in entity
sys, err := entities.NewSystem(name)  // validates internally

// Bad - validation in use case
if name == "" {  // NO! This belongs in NewSystem
    return ErrEmptyName
}
```

## Current Progress

- ✅ Issue #1: Project structure, CI, configs
- ✅ Issue #2: Domain entities with tests
- 🔲 Issue #3: Use case ports (interfaces) - **NEXT**

## Common Tasks

### Implementing a Port (Interface)

1. Add interface to `internal/core/usecases/ports.go`
2. Create adapter in `internal/adapters/<name>/`
3. Wire in `main.go`
4. Write tests with mock implementation

### Adding a CLI Command

1. Create use case in `internal/core/usecases/`
2. Create command in `cmd/<name>.go` (thin wrapper)
3. Register in `cmd/root.go`
4. Keep handler <50 lines

### Adding an MCP Tool

1. Ensure use case exists
2. Create tool in `internal/mcp/tools/<name>.go`
3. Register in `internal/mcp/tools/registry.go`
4. Keep handler <30 lines

## Testing Strategy

| Layer | Test Type | Mock Strategy |
|-------|-----------|---------------|
| entities/ | Unit | None needed (pure) |
| usecases/ | Unit | Mock all ports |
| adapters/ | Integration | Real external deps |
| cmd/, mcp/ | E2E | Full stack |

## External Dependencies

| Dependency | Interface | Adapter Location |
|------------|-----------|------------------|
| d2 CLI | `DiagramRenderer` | `adapters/d2/` |
| veve-cli | `PDFRenderer` | `adapters/pdf/` |
| ason library | `TemplateEngine` | `adapters/ason/` |
| file system | `ProjectRepository` | `adapters/filesystem/` |

## Commands

```bash
# Development
task test           # Run tests
task lint           # Run linter
task build          # Build binary
task dev            # Build and run --help

# When implementing
go test -v ./internal/core/...     # Test core only
go test -cover ./...               # Coverage report
```

## Context for LLMs

When querying this project via MCP, use progressive loading:

- `detail: "summary"` - ~200 tokens, counts only
- `detail: "structure"` - ~500 tokens, systems + containers
- `detail: "full", target: "system-name"` - full details for one system

## References

- [C4 Model](https://c4model.com/)
- [D2 Language](https://d2lang.com/)
- [ason Docs](https://context7.com/madstone-tech/ason/llms.txt)
- [MCP Protocol](https://modelcontextprotocol.io/)

## Active Technologies
- Go 1.25 + ason template engine (internal), fsnotify, d2 CLI (external) (003-serverless-template)
- Filesystem (template files, project files) (003-serverless-template)
- Go 1.25 + spf13/cobra v1.10.2, spf13/viper v1.21.0 (new); charmbracelet/lipgloss, fsnotify (existing) (002-cobra-viper)
- Filesystem — XDG dirs (`~/.config/loko/`, `~/.local/share/loko/`, `~/.cache/loko/`) + project `loko.toml` (002-cobra-viper)
- Go 1.25+ + oon-format/toon-go (TOON v3.0), cobra/viper (CLI), lipgloss (UI) (005-toon-alignment)
- Filesystem (ProjectRepository adapter) (005-toon-alignment)

## Recent Changes
- 003-serverless-template: Added Go 1.25 + ason template engine (internal), fsnotify, d2 CLI (external)
