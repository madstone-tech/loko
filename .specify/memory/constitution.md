<!--
SYNC IMPACT REPORT
==================
Version change: 1.0.0 → 1.1.0 (MINOR — materially expanded guidance, no principle removed or redefined incompatibly)

Modified principles:
  - III. Thin Handlers
      - Granularity changed from per-file to per-function for CLI and MCP handlers.
      - CLI handler functions: < 150 lines (per-file)  →  ≤ 50 effective lines (per-function).
      - MCP tool handler functions: < 100 lines (per-file)  →  ≤ 30 effective lines (per-function).
      - API handlers: < 150 lines (per-file)  →  thin-wrapper expectation retained;
        specific budget deferred to a future amendment.

Added principle-level rules (under "Architecture Rules"):
  - Use-case file size: ≤ 200 effective lines (new).
  - Entity file size: ≤ 300 effective lines (new).

Modified sections:
  - Quality Gates → Before Every PR: replaced the per-file handler-count gate and the
    KNOWN_VIOLATIONS allowlist clause with a structural-compliance-check gate that
    enforces layer-import rules plus the new file-size and function-size budgets.

Removed sections:
  - The KNOWN_VIOLATIONS allowlist mechanism (scripts/audit-constitution.sh) is removed.
    Categorical exemptions only (data-files, *_cobra.go, *_test.go, generated files).

Templates requiring updates:
  - ✅ .specify/templates/plan-template.md       (generic constitution reference; no edit needed)
  - ✅ .specify/templates/spec-template.md       (no constitution-specific text; no edit needed)
  - ✅ .specify/templates/tasks-template.md      (no constitution-specific text; no edit needed)
  - ✅ .specify/templates/checklist-template.md  (no constitution-specific text; no edit needed)
  - ⚠ scripts/audit-constitution.sh             (replaced by tools/archcheck; tracked under feature 009)
  - ⚠ Makefile                                   (add `audit-constitution` target; tracked under feature 009)
  - ⚠ .golangci.yml                              (add `depguard` layer config; tracked under feature 009)
  - ⚠ .github/workflows/ci.yml                   (wire the gating step; tracked under feature 009)

Follow-up TODOs:
  - TODO(IMPLEMENTATION): The audit tool, depguard configuration, allowlist removal,
    and CI wiring are deliverables of feature 009-constitution-compliance — see
    specs/009-constitution-compliance/ (plan.md, contracts/, tasks.md once generated).
  - TODO(ADR): docs/adr/0009-constitution-compliance-tooling.md is to be authored as
    part of feature 009; this constitution references it but does not yet link.
  - TODO(API_BUDGET): Specific per-function limit for HTTP API handlers in
    internal/api/ is intentionally deferred. A future amendment will set it
    (likely ≤ 50 effective lines, mirroring CLI) once feature 009 ships.
-->

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

CLI commands, MCP tools, and API handlers are thin wrappers that delegate to use cases. Limits are measured in **effective lines** (source lines after dropping blank lines, single- and multi-line comments, the `package` declaration, and `import (...)` blocks) and apply at **per-function granularity** for CLI and MCP handlers:

- **CLI handler functions**: ≤ **50** effective lines per function (`cmd/**/*.go`)
- **MCP tool handler functions**: ≤ **30** effective lines per function (`internal/mcp/tools/**/*.go`)
- **API handlers**: thin wrappers that delegate to use cases; specific per-function budget deferred to a future amendment, but the same thin-wrapper pattern MUST be followed

Handlers do three things only: parse input, call use case, format output. No business logic, no validation, no data transformation beyond what the interface protocol requires.

Pure-data files (`schemas.go`, `registry.go`, `helpers.go`, `constants.go`), Cobra flag-wiring files (`*_cobra.go`), test files (`*_test.go`), and generated files (those with `// Code generated ... DO NOT EDIT.` headers) are exempt from file-size and function-size budgets but remain subject to layer-import rules.

**Rationale**: Prevents business logic from leaking into interface-specific code. Per-function granularity (rather than per-file) is enforced because file-level totals can mask one large function buried among small ones. The 50/30 budgets are deliberately tight enough to force the conversation about whether a function is doing too much.

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
| `internal/mcp/` | core/usecases, adapters | `internal/core/entities/` directly (entity types MUST be obtained via use-case return values or adapter outputs); api; cmd |
| `internal/api/` | core/usecases, adapters | `internal/core/entities/` directly (entity types MUST be obtained via use-case return values or adapter outputs); mcp; cmd |
| `cmd/` | core, adapters, mcp, api | `internal/core/entities/` directly (entity types MUST be obtained via use-case return values or adapter outputs) |

### File-Size Budgets

Whole-file budgets apply to the inner core layers and are measured in effective lines (same counting convention as Principle III):

| Path | Budget | Rationale |
|------|--------|-----------|
| `internal/core/usecases/**/*.go` | ≤ **200** effective lines | Use-case files must remain narrative-scale; split by sub-step (e.g., `build_docs.go` → `build_docs.go` + `build_docs_d2.go` + `build_docs_markdown.go`) when needed. |
| `internal/core/entities/**/*.go` | ≤ **300** effective lines | Entities may be longer because they declare types and pure-data validation, but still capped to remain reviewable. |

Outer-layer files (`cmd/`, `internal/mcp/`, `internal/api/`) have no whole-file budget; they have per-function budgets per Principle III.

### File Organization

```
internal/
├── core/                     # ZERO external dependencies
│   ├── entities/             # Domain objects with validation (≤ 300 effective lines/file)
│   └── usecases/             # Application logic + ports.go (≤ 200 effective lines/file)
├── adapters/                 # Infrastructure implementations
│   ├── filesystem/           # ProjectRepository
│   ├── d2/                   # DiagramRenderer
│   ├── ason/                 # TemplateEngine
│   ├── html/                 # SiteBuilder
│   ├── encoding/             # OutputEncoder (JSON, TOON)
│   └── config/               # ConfigLoader (TOML)
├── mcp/                      # MCP server (thin layer; tool handler funcs ≤ 30 effective lines)
├── api/                      # HTTP API (thin layer; per-function budget deferred)
└── ui/                       # Lipgloss styles
cmd/                          # CLI commands (thin layer; handler funcs ≤ 50 effective lines)
tools/
└── archcheck/                # Build-time audit binary (outside enforcement scope)
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
- **Structural-compliance audit**: `tools/archcheck` (custom Go AST binary) + `golangci-lint depguard` (redundant fast-path)

## Quality Gates

### Before Every Commit

- `task test` (or `make test`) passes — all tests green
- `task lint` (or `make lint`) passes — no linter warnings; includes the redundant `depguard` layer-import check
- No new external dependencies in `internal/core/`

### Before Every PR

- Test coverage > **80%** on `internal/core/`
- **Structural compliance check passes** (`make audit-constitution`):
  - Layer-import rules per the Dependency Direction table
  - CLI handler functions ≤ 50 effective lines (per Principle III)
  - MCP tool handler functions ≤ 30 effective lines (per Principle III)
  - Use-case files ≤ 200 effective lines (per Architecture Rules)
  - Entity files ≤ 300 effective lines (per Architecture Rules)
- No port interface used outside of designated layers
- ADR written for any new architectural decision

The structural-compliance check has **no per-file allowlist**. Categorical exemptions (data-files `schemas.go`/`registry.go`/`helpers.go`/`constants.go`, `*_cobra.go`, `*_test.go`, and files marked `// Code generated ... DO NOT EDIT.`) are encoded in the rule set itself; new exemptions require a rule-file change plus an ADR.

## Governance

- This constitution supersedes all other development practices for the loko project
- Amendments require: documented rationale, review of impact on existing code, and a migration plan if breaking
- All PRs and code reviews must verify compliance with these principles
- When in doubt, refer to the ADRs in `docs/adr/` for decision context
- The machine-consumable mirror of the file-size, function-size, layer-import, and exemption rules lives at `specs/009-constitution-compliance/contracts/structural-rules.yaml`. The markdown text in this file remains canonical; the YAML is regenerated/synced by review and a CI cross-check ensures the two never diverge.

**Version**: 1.2.0 | **Ratified**: 2026-02-06 | **Last Amended**: 2026-05-21
