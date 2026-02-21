# Implementation Plan: MCP UX Improvements

**Branch**: `008-mcp-ux-improvements` | **Date**: 2026-02-19 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/008-mcp-ux-improvements/spec.md`

## Summary

Add first-class relationship management to the MCP tool surface by introducing a `Relationship` entity persisted in `relationships.toml` per system, three new MCP tools (`create_relationship`, `list_relationships`, `delete_relationship`), and a new `RelationshipRepository` port. Additionally fix five lower-priority UX gaps: auto-initialize container D2 diagrams on creation (wire the existing `DiagramGenerator` into `CreateContainerTool`), add a `create_components` batch tool, suppress `isolated_component` validation findings project-wide when `relationships.toml` is empty, surface slugified IDs in error messages, and eagerly invalidate `GraphCache` after every relationship mutation.

## Technical Context

**Language/Version**: Go 1.25 (go.mod confirmed)  
**Primary Dependencies**: Cobra + Viper (CLI), Lipgloss (TUI), pelletier/go-toml v2 (TOML R/W), toon-go v0 (encoding), fsnotify (file watching), MCP SDK (stdio/JSON-RPC)  
**Storage**: Filesystem — `src/<system>/relationships.toml` (new); existing component `.md` frontmatter and system/container `.toml` files unchanged  
**Testing**: `go test ./...` (unit), `go test -tags=integration` (integration); target >80% coverage on `internal/core/`  
**Target Platform**: Linux/macOS (single binary, no runtime deps beyond d2 and optionally veve)  
**Project Type**: Single project (no frontend, no mobile)  
**Performance Goals**: `create_relationship` round-trip + cache invalidation + graph rebuild in <2 seconds (SC-001)  
**Constraints**: `internal/core/` must have zero external dependencies; MCP tool handlers must stay <100 lines; `veve` binary is used for PDF export only and is NOT invoked during diagram writes  
**Scale/Scope**: Typical project: 1–10 systems, 5–50 containers, 10–200 components, 10–500 relationships

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Architecture | PASS | `Relationship` entity in `core/entities/`; `RelationshipRepository` port in `usecases/ports.go`; filesystem adapter in `adapters/filesystem/`; MCP tools are thin wrappers |
| II. Interface-First | PASS | New `RelationshipRepository` port defined before adapter; all use cases depend on the port, not the concrete adapter |
| III. Thin Handlers | PASS | Each new MCP tool (create/list/delete relationship, create_components) delegates to a use case; handler logic stays <100 lines |
| IV. Entity Validation | PASS | `NewRelationship()` constructor validates source, target, label; use cases trust validated entities |
| V. Test-First | PASS | Tests written before implementation; mock `RelationshipRepository` required |
| VI. Token Efficiency | PASS | `list_relationships` response uses compact array format; TOON opt-in available via existing encoder |
| VII. Simplicity & YAGNI | PASS | One new entity, one new port, one new adapter, five new use cases/tools — no over-engineering; `create_components` reuses existing `ScaffoldEntity` use case in a loop |

**Constitution violations**: None. No Complexity Tracking table required.

## Project Structure

### Documentation (this feature)

```text
specs/008-mcp-ux-improvements/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── relationship-tools.md
│   └── batch-component-tool.md
└── tasks.md             # Phase 2 output (/speckit.tasks - NOT created here)
```

### Source Code Layout (this feature's changes)

```text
internal/
├── core/
│   ├── entities/
│   │   └── relationship.go          # NEW: Relationship entity + NewRelationship()
│   └── usecases/
│       ├── ports.go                 # MODIFY: add RelationshipRepository port
│       ├── create_relationship.go   # NEW: CreateRelationship use case
│       ├── list_relationships.go    # NEW: ListRelationships use case
│       ├── delete_relationship.go   # NEW: DeleteRelationship use case
│       └── validate_architecture.go # MODIFY: FR-012 suppression logic
├── adapters/
│   └── filesystem/
│       └── relationship_repo.go     # NEW: RelationshipRepository filesystem adapter
└── mcp/
    └── tools/
        ├── create_relationship.go   # NEW: MCP tool handler
        ├── list_relationships.go    # NEW: MCP tool handler
        ├── delete_relationship.go   # NEW: MCP tool handler
        ├── create_components.go     # NEW: batch create MCP tool handler
        ├── create_container.go      # MODIFY: inject DiagramGenerator (FR-007/008)
        ├── helpers.go               # MODIFY: add suggestSlugID() helper (FR-014)
        └── registry.go              # MODIFY: register new tools

src/
└── <system>/
    └── relationships.toml           # NEW file format (per-system, created by adapter)
```

**Structure Decision**: Single project layout. All new files follow the existing pattern exactly — one entity per file, use cases alongside their tests, adapters in `adapters/filesystem/`, MCP tool handlers in `mcp/tools/`.
