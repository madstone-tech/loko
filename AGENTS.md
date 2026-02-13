# AGENTS.md - Guidance for Agentic Coding in loko

This file guides AI agents operating in the loko repository. Read `.specify/memory/constitution.md` first for governance principles.

## Build, Test, Lint Commands

**Build**: `make build` or `go build -o loko .` (creates `loko` binary)

**Test all**: `go test ./...` or `make test` (fast unit tests)

**Single test**: `go test -run TestName ./path/to/package` (e.g., `go test -run TestArchiveSystem ./internal/core/usecases`)

**Verbose**: `go test -v ./...` or `make test-v`

**Coverage**: `task coverage` (generates `coverage.html`; aim for >80% in `internal/core/`)

**Integration tests**: `go test -tags=integration -v ./tests/integration/...`

**Lint**: `task lint` (runs golangci-lint; must pass before commit)

**Format**: `task fmt` (gofmt + goimports; auto-formats imports)

**Single file test**: `go test -run TestName -v ./internal/core/usecases -count=1` (disables caching)

## Code Style Guidelines

**Imports**: Use `goimports` (auto-groups: stdlib, external, internal). Order: stdlib → external → github.com/madstone-tech

**Formatting**: Tab indentation (8 spaces = 1 tab). Run `task fmt` before commit. Max line length ~100 chars (readability).

**Types**: Use simple types (avoid generic overkill). Interfaces ONLY in core/usecases/ports.go. Concrete structs in entities/. Export only what external packages need.

**Naming**: PascalCase for exported types/functions. camelCase for unexported. Acronyms uppercase (HTTPServer, not HttpServer). Entity field names descriptive (Name, not N).

**Error Handling**: Wrap errors with context (use `fmt.Errorf("%w", err)` or `errors.Join()`). Define sentinel errors in entities/errors.go (e.g., `var ErrNotFound = errors.New(...)`). Return custom error types for rich context (ValidationError, NotFoundError). NEVER ignore errors (no `_ = someFunc()`).

**Receivers**: Use pointer receivers on methods (consistency). Exception: small immutable types (time.Time).

**Comments**: Exported types/functions MUST have godoc. Format: `// TypeName describes...` (period-terminated). Entity validation logic in comments. Unexported helpers: inline comment if non-obvious.

**Testing**: Unit tests mock all ports (see MockProjectRepo pattern). Use t.TempDir() for file I/O. Prefer table-driven tests. Golden tests for generated output (HTML, diagrams). No third-party mocking libraries.

**Architecture**: Business logic ONLY in core/ (zero external deps). Adapters implement ports. CLI/MCP/API are thin wrappers (<50 lines). Validate in entities, not use cases. Inject dependencies in main.go.

**Constants**: Unexported magic numbers → const. Exported config → ProjectConfig struct with doc comments.

**Files**: One entity per file (project.go, system.go, etc.). Adapters group by infrastructure type (filesystem/, d2/, encoding/). Tests: \*\_test.go alongside source.

## Key Constraints (from Constitution v1.0.0)

- ✅ Go 1.25+ syntax (use generics where they reduce boilerplate)
- ✅ Clean Architecture with strict dependency inversion
- ✅ Interfaces testable; no global state or init() functions
- ✅ Immutable builds (same input → same output; mock non-determinism in tests)
- ❌ NO arbitrary code execution (no eval, dynamic templates, or shell interpolation)
- ❌ NO third-party mocking libraries (use concrete mocks)
- ❌ NO external dependencies in core/ package

## Workflow

1. **Understand the feature**: Read spec in `.specify/memory/` and Constitution v1.0.0
2. **Check architecture**: Which layer? (entities → usecases → adapters → cli/mcp/api)
3. **Write tests first** (TDD): `go test -run TestName -v` (watch for failures)
4. **Implement**: Follow patterns in internal/core/entities/ and internal/core/usecases/
5. **Lint & format**: `task fmt && task lint` (must pass)
6. **Coverage check**: `task coverage` (aim >80% for core/)
7. **Commit**: Clear message referencing Constitution principle if applicable

---

**Last updated**: 2025-12-17 | **Ref**: Constitution v1.0.0, Makefile, .golangci.yml

## Active Technologies
- Go 1.25+ + Cobra (CLI), Viper (config), Lipgloss (formatting), Bubbletea (interactive prompts), ason (templates), MCP SDK (model context protocol) (001-loko-v0.1.0)
- File system (src/ directory structure + loko.toml configuration); no database (001-loko-v0.1.0)

## Recent Changes
- 001-loko-v0.1.0: Added Go 1.25+ + Cobra (CLI), Viper (config), Lipgloss (formatting), Bubbletea (interactive prompts), ason (templates), MCP SDK (model context protocol)
