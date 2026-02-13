# Implementation Plan: TOON Alignment & Handler Refactoring

**Branch**: `005-toon-alignment` | **Date**: 2026-02-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/005-toon-alignment/spec.md`

## Summary

This feature pays down handler debt (10 files violating Thin Handler principle) and aligns the TOON encoder with the official TOON v3.0 specification. The approach is: extract business logic from CLI/MCP handlers into use cases first, then replace the custom TOON encoder with a spec-compliant adapter using `toon-format/toon-go`.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: toon-format/toon-go (TOON v3.0), cobra/viper (CLI), lipgloss (UI)
**Storage**: Filesystem (ProjectRepository adapter)
**Testing**: `go test` with concrete mock structs (no mocking libraries)
**Target Platform**: Linux, macOS, Windows (single binary)
**Project Type**: Single Go binary with Clean Architecture
**Performance Goals**: Encoding < 10ms, > 30% token reduction vs JSON
**Constraints**: `internal/core/` has zero external dependencies; handlers < 50 lines (CLI) / < 30 lines (MCP)
**Scale/Scope**: Architecture data: 5-50 systems, 15-200 containers

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate | Status |
|-----------|------|--------|
| I. Clean Architecture | core/ has zero external deps; dependency direction enforced | PASS — toon-go goes in adapter only |
| II. Interface-First | All deps through ports.go | PASS — OutputEncoder port exists; new DiagramGenerator port needed |
| III. Thin Handlers | CLI < 50 lines, MCP < 30 lines | FAIL (current) → PASS (after refactoring) |
| IV. Entity Validation | Validation in entities only | PASS — no changes to entities |
| V. Test-First | Red-Green-Refactor | PASS — tests before implementation |
| VI. Token Efficiency | TOON support for 30-60% reduction | PASS (after TOON alignment) |
| VII. Simplicity & YAGNI | Use toon-go library, don't build from scratch | PASS |

**Gate result**: Principle III is currently violated. This plan resolves it.

## Project Structure

### Documentation (this feature)

```text
specs/005-toon-alignment/
├── plan.md              # This file
├── research.md          # Phase 0: TOON spec research, handler analysis
├── data-model.md        # Phase 1: Use case I/O models, TOON entity mapping
├── quickstart.md        # Phase 1: Acceptance test scenarios
├── contracts/           # Phase 1: Use case contracts, MCP tool schemas
│   └── use-cases.md     # Use case input/output contracts
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/
├── core/
│   ├── entities/                    # No changes (already clean)
│   └── usecases/
│       ├── ports.go                 # Add: DiagramGenerator, UserPrompter, ReportFormatter
│       ├── create_system.go         # Exists ✓
│       ├── create_container.go      # NEW: Extract from cmd/new.go + mcp/tools.go
│       ├── create_component.go      # NEW: Extract from cmd/new.go + mcp/tools.go
│       ├── update_diagram.go        # NEW: Extract from mcp/tools.go
│       ├── scaffold_entity.go       # NEW: Orchestrates create + D2 generation
│       ├── build_docs.go            # Exists ✓ — Enhance with format handling from cmd/build.go
│       ├── validate_architecture.go # Exists ✓
│       ├── query_architecture.go    # Exists ✓
│       └── build_architecture_graph.go # Exists ✓
├── adapters/
│   ├── d2/
│   │   ├── renderer.go              # Exists (DiagramRenderer)
│   │   └── generator.go             # NEW: Move from cmd/d2_generator.go
│   ├── encoding/
│   │   ├── encoder.go               # Exists — Update with toon-go
│   │   ├── toon.go                  # REPLACE: Custom encoder → toon-go wrapper
│   │   └── toon_test.go             # REPLACE: Update tests for TOON v3.0
│   ├── cli/
│   │   ├── prompts.go               # NEW: Extract from cmd/new.go
│   │   ├── progress_reporter.go     # NEW: Extract from cmd/build.go
│   │   └── report_formatter.go      # NEW: Extract from cmd/validate.go
│   └── ...                          # Other adapters unchanged
├── mcp/
│   └── tools/
│       ├── registry.go              # NEW: Tool registration
│       ├── create_system.go         # NEW: Split from tools.go (< 30 lines)
│       ├── create_container.go      # NEW: Split from tools.go (< 30 lines)
│       ├── create_component.go      # NEW: Split from tools.go (< 30 lines)
│       ├── update_diagram.go        # NEW: Split from tools.go (< 30 lines)
│       ├── build_docs.go            # NEW: Split from tools.go (< 30 lines)
│       ├── validate.go              # NEW: Split from tools.go (< 30 lines)
│       ├── validate_diagram.go      # NEW: Split from tools.go (< 30 lines)
│       ├── query_architecture.go    # Exists ✓ (103 lines — already separate)
│       ├── query_project.go         # Exists ✓ (74 lines — already separate)
│       ├── graph_tools.go           # Exists ✓ (348 lines — clean, properly delegating)
│       └── schemas.go               # Exists ✓ (172 lines)
│
cmd/
├── root.go                          # ASSESS: 162 lines, likely acceptable (Cobra wiring)
├── new.go                           # REFACTOR: 504 → < 50 lines
├── new_cobra.go                     # REFACTOR: 199 → < 50 lines
├── build.go                         # REFACTOR: 251 → < 50 lines
├── build_cobra.go                   # REFACTOR: 107 → < 50 lines
├── watch.go                         # ASSESS: 146 lines — mostly event loop (may be acceptable)
├── validate.go                      # REFACTOR: 142 → < 50 lines
├── d2_generator.go                  # DELETE: Moved to internal/adapters/d2/generator.go
└── ...                              # Other commands unchanged
```

**Structure Decision**: Existing Clean Architecture structure is preserved. New use cases go in `internal/core/usecases/`, new adapters in `internal/adapters/`, CLI adapters in `internal/adapters/cli/`. MCP tools split into individual files.

## Part A: Handler Refactoring

### Refactoring Map

| Source | Lines | Target Use Case | Target Adapter | Shared CLI+MCP |
|--------|-------|----------------|----------------|----------------|
| `cmd/new.go` createSystem() | ~150 | `scaffold_entity.go` | — | Yes |
| `cmd/new.go` createContainer() | ~100 | `create_container.go` + `scaffold_entity.go` | — | Yes |
| `cmd/new.go` createComponent() | ~100 | `create_component.go` + `scaffold_entity.go` | — | Yes |
| `cmd/new.go` prompts | ~30 | — | `cli/prompts.go` | No (CLI only) |
| `cmd/new.go` template discovery | ~60 | Part of `TemplateEngine` | — | Yes |
| `cmd/d2_generator.go` (entire file) | 282 | — | `d2/generator.go` | Yes |
| `cmd/build.go` orchestration | ~80 | Already in `build_docs.go` | — | Yes |
| `cmd/build.go` progress reporter | ~30 | — | `cli/progress_reporter.go` | No |
| `cmd/build.go` format parsing | ~30 | — | Input validation in handler | No |
| `cmd/validate.go` report formatting | ~80 | — | `cli/report_formatter.go` | No |
| `mcp/tools/tools.go` CreateSystem | ~150 | `scaffold_entity.go` | — | Yes |
| `mcp/tools/tools.go` CreateContainer | ~120 | `create_container.go` + `scaffold_entity.go` | — | Yes |
| `mcp/tools/tools.go` CreateComponent | ~120 | `create_component.go` + `scaffold_entity.go` | — | Yes |
| `mcp/tools/tools.go` UpdateDiagram | ~120 | `update_diagram.go` | — | Yes |
| `mcp/tools/tools.go` D2 helpers | ~360 | — | `d2/generator.go` | Yes |

### New Port Interfaces (add to ports.go)

```
DiagramGenerator
  - GenerateSystemContextDiagram(system) → string
  - GenerateContainerDiagram(system) → string
  - GenerateComponentDiagram(container) → string

UserPrompter (CLI-only, optional dependency)
  - PromptString(prompt, defaultValue) → string
  - PromptStringMulti(prompt) → []string

ReportFormatter (CLI-only, optional dependency)
  - PrintReport(report)
```

### New Use Cases

1. **CreateContainerUseCase** — Validates input, creates container entity, saves via ProjectRepository
2. **CreateComponentUseCase** — Validates input, creates component entity, saves via ProjectRepository
3. **ScaffoldEntityUseCase** — Orchestrates: create entity + generate D2 diagram + update parent diagram + render template. Called by both CLI `new` and MCP `create_*` tools.
4. **UpdateDiagramUseCase** — Validates D2 source, writes to project via ProjectRepository

### Execution Order

```
Step 1: Add new ports to ports.go (DiagramGenerator, UserPrompter, ReportFormatter)
Step 2: Move cmd/d2_generator.go → internal/adapters/d2/generator.go (implements DiagramGenerator)
Step 3: Create use cases (CreateContainer, CreateComponent, ScaffoldEntity, UpdateDiagram)
Step 4: Extract CLI adapters (prompts, progress_reporter, report_formatter)
Step 5: Refactor cmd/new.go → thin handler calling ScaffoldEntityUseCase
Step 6: Refactor cmd/build.go → thin handler calling BuildDocsUseCase
Step 7: Refactor cmd/validate.go → thin handler calling ValidateArchitectureUseCase + ReportFormatter
Step 8: Split mcp/tools/tools.go → individual files, each calling shared use cases
Step 9: Verify all handlers under line limits, all tests pass
```

## Part B: TOON v3.0 Alignment

### Research Decision: Use toon-format/toon-go

**Decision**: Use `github.com/toon-format/toon-go` official Go library.

**Rationale**:
- Spec-compliant TOON v3.0 implementation (349 test fixtures)
- Zero external dependencies (aligns with clean architecture)
- API matches Go stdlib pattern: `toon.Marshal()` / `toon.Unmarshal()`
- Struct tags: `toon:"fieldname"` (same pattern as `json:"fieldname"`)
- Full tabular array support with `toon.WithLengthMarkers(true)`
- Go 1.23+ compatible (loko uses 1.25)
- Actively maintained (102 stars, last commit Dec 2025)

**Alternatives rejected**:
- `alpkeskin/gotoon`: Slower performance, less adoption
- Build from scratch: 2-4 weeks effort, maintenance burden, bug risk
- Keep custom format: Non-standard, confuses LLMs, no ecosystem tooling

**Risk**: Pre-release version (no semver tags yet).
**Mitigation**: Pin to specific commit hash in go.mod.

### Implementation Approach

The existing `OutputEncoder` interface is already correct:

```
EncodeJSON(value any) ([]byte, error)
EncodeTOON(value any) ([]byte, error)
DecodeJSON(data []byte, value any) error
DecodeTOON(data []byte, value any) error
```

Replace the custom encoder body with `toon.Marshal(value, toon.WithLengthMarkers(true))` for encode and `toon.Unmarshal(data, value)` for decode. Add `toon:"fieldname"` struct tags to entities.

### Entity Tag Requirements

Add `toon` struct tags to entities that need TOON serialization:
- `entities.Project` — project-level fields
- `entities.System` — system fields for tabular arrays
- `entities.Container` — container fields for tabular arrays
- `entities.Component` — component fields for tabular arrays

### Execution Order

```
Step 1: Add toon-go dependency (go get github.com/toon-format/toon-go@latest)
Step 2: Add toon struct tags to entities (non-breaking, stdlib only)
Step 3: Replace EncodeTOON/DecodeTOON in internal/adapters/encoding/toon.go
Step 4: Write unit tests (spec compliance, round-trip, error handling)
Step 5: Run token efficiency benchmarks
Step 6: Deprecate old custom format (rename to --format compact with warning)
Step 7: Update MCP tool descriptions and documentation
```

## Complexity Tracking

| Concern | Why Acceptable | Simpler Alternative Rejected Because |
|---------|---------------|--------------------------------------|
| ScaffoldEntityUseCase orchestrates create + D2 + template | Single entry point for both CLI and MCP prevents logic duplication | Separate use cases per entity type would duplicate orchestration |
| DiagramGenerator port added to ports.go | D2 generation is a domain concern (how to represent C4 as diagrams) that belongs behind an interface | Inline D2 generation in use cases would couple core to D2 syntax |
| 4 new ports (DiagramGenerator, UserPrompter, ReportFormatter + existing) | Each represents a genuine external concern (diagram syntax, user input, terminal output) | Combining them violates single responsibility |
