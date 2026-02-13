# loko Constitution

## Core Principles

### I. Clean Architecture (NON-NEGOTIABLE)

loko exposes functionality through three interfaces (CLI, MCP, HTTP API) that all share the same business logic. The architecture enforces strict dependency direction:

```
core/ imports NOTHING from adapters/, mcp/, api/, cmd/
adapters/ imports from core/ only
mcp/, api/, cmd/ import from core/ and adapters/
```

- `internal/core/` has **zero external dependencies** — stdlib only
- Core defines interfaces (ports) in `usecases/ports.go`; adapters implement them
- Use cases contain all business logic — no logic in handlers or adapters
- Dependencies are injected at startup in `main.go`
- Swapping any adapter (filesystem, renderer, encoder) requires zero changes to core

**Rationale**: Three consumer interfaces (CLI, MCP, API) must share logic without duplication. External tools (d2, veve-cli) and libraries (ason, toon-go) must be replaceable without cascading rewrites.

### II. Interface-First

All external dependencies are accessed through interfaces defined in `internal/core/usecases/ports.go`. No use case or entity may reference a concrete adapter implementation.

- New external dependency? Define the port interface first, then implement the adapter
- Port interfaces live exclusively in `usecases/ports.go`
- Adapters live in `internal/adapters/<name>/`
- Wiring (interface → implementation) happens only in `main.go`

**Rationale**: Enables testing with mocks, swapping implementations, and enforcing the dependency rule.

### III. Thin Handlers

CLI commands, MCP tools, and API handlers are thin wrappers that delegate to use cases:

- CLI commands: **< 50 lines** of handler code
- MCP tool handlers: **< 30 lines** of handler code
- API handlers: **< 50 lines** of handler code

Handlers do three things only: parse input, call use case, format output. No business logic, no validation, no data transformation beyond what's needed for the interface protocol.

**Rationale**: Prevents business logic from leaking into interface-specific code. If a handler grows beyond the line limit, logic belongs in a use case.

### IV. Entity Validation

All domain validation lives in entities, not in use cases or handlers.

- Constructors (`NewSystem`, `NewContainer`, etc.) validate internally and return errors
- Use cases trust that entities are valid once constructed
- No validation code in CLI commands, MCP tools, or API handlers
- Entities are pure Go structs with methods — no external dependencies

**Rationale**: Validation rules are domain knowledge. Centralizing them in entities prevents inconsistent validation across three consumer interfaces.

### V. Test-First

Tests are written before implementation code. The Red-Green-Refactor cycle is enforced.

- **Entities**: Unit tests, no mocks needed (pure structs)
- **Use cases**: Unit tests with concrete mock implementations of ports (no mocking libraries)
- **Adapters**: Integration tests with real external dependencies
- **CLI/MCP/API**: End-to-end tests with full stack
- Target: **> 80% coverage** across `internal/core/`

**Rationale**: Three consumer interfaces multiply the risk of regression. Tests are the safety net that enables confident refactoring and adapter swaps.

### VI. Token Efficiency

loko is designed for LLM consumption. Every output format and query response must minimize token usage without sacrificing correctness.

- Progressive context loading: summary (~200 tokens) → structure (~500 tokens) → full (targeted)
- TOON format support for 30-60% token reduction over JSON on architecture data
- Tabular arrays for uniform data (systems, containers, components)
- Default to JSON for compatibility; TOON is opt-in

**Rationale**: LLM context windows are finite and tokens cost money. Architecture data has highly uniform structure (arrays of systems, containers, components) that benefits enormously from compact formats.

### VII. Simplicity & YAGNI

Start with the simplest solution that works. Do not build for hypothetical future requirements.

- No feature flags or backward-compatibility shims when you can just change the code
- No abstractions for one-time operations
- No third-party mocking libraries — concrete mock structs are sufficient
- If three similar lines of code work, don't create a premature abstraction
- Single binary with no runtime dependencies except d2 (and optionally veve-cli)

**Rationale**: Complexity is the enemy of maintainability. Every abstraction must justify its existence against the cost of indirection.

## Architecture Rules

### Dependency Direction

| Layer | May Import | Must Not Import |
|-------|-----------|-----------------|
| `internal/core/entities/` | stdlib only | anything else |
| `internal/core/usecases/` | entities, stdlib | adapters, mcp, api, cmd |
| `internal/adapters/` | core (entities + usecases interfaces) | mcp, api, cmd |
| `internal/mcp/` | core, adapters | api, cmd |
| `internal/api/` | core, adapters | mcp, cmd |
| `cmd/` | core, adapters, mcp, api | — |

### File Organization

```
internal/
├── core/                     # ZERO external dependencies
│   ├── entities/             # Domain objects with validation
│   └── usecases/             # Application logic + ports.go
├── adapters/                 # Infrastructure implementations
│   ├── filesystem/           # ProjectRepository
│   ├── d2/                   # DiagramRenderer
│   ├── ason/                 # TemplateEngine
│   ├── html/                 # SiteBuilder
│   ├── encoding/             # OutputEncoder (JSON, TOON)
│   └── config/               # ConfigLoader (TOML)
├── mcp/                      # MCP server (thin layer)
├── api/                      # HTTP API (thin layer)
└── ui/                       # Lipgloss styles
cmd/                          # CLI commands (thin layer)
```

### External Dependencies

| Dependency | Type | Interface | Adapter |
|------------|------|-----------|---------|
| d2 CLI | Shell out | `DiagramRenderer` | `adapters/d2/` |
| veve-cli | Shell out | `PDFRenderer` | `adapters/pdf/` |
| ason | Go library | `TemplateEngine` | `adapters/ason/` |
| toon-go | Go library | `OutputEncoder` | `adapters/encoding/` |
| fsnotify | Go library | `FileWatcher` | `adapters/filesystem/` |
| file system | OS | `ProjectRepository` | `adapters/filesystem/` |

## Technology Stack

- **Language**: Go 1.25+
- **CLI framework**: Cobra + Viper (adapter layer only)
- **TUI/styling**: Lipgloss (UI layer only)
- **Diagram rendering**: d2 CLI (behind interface)
- **Template engine**: ason (behind interface)
- **Encoding**: JSON (stdlib) + TOON v3.0 (behind interface)
- **File watching**: fsnotify (behind interface)
- **MCP transport**: stdio, JSON-RPC
- **Configuration**: TOML (loko.toml)
- **Paths**: XDG Base Directory Specification

## Quality Gates

### Before Every Commit

- `task test` passes (all tests green)
- `task lint` passes (no linter warnings)
- No new external dependencies in `internal/core/`

### Before Every PR

- Test coverage > 80% on `internal/core/`
- Handler line counts within limits (CLI < 50, MCP < 30)
- No port interface used outside of designated layers
- ADR written for any new architectural decision

## Governance

- This constitution supersedes all other development practices for the loko project
- Amendments require: documented rationale, review of impact on existing code, and a migration plan if breaking
- All PRs and code reviews must verify compliance with these principles
- When in doubt, refer to the ADRs in `docs/adr/` for decision context

**Version**: 1.0.0 | **Ratified**: 2026-02-06 | **Last Amended**: 2026-02-06
