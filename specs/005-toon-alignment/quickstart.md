# Acceptance Test Scenarios: TOON Alignment & Handler Refactoring

**Feature**: 005-toon-alignment | **Date**: 2026-02-06 | **Phase**: 1 (Design)

---

## Part A: Handler Refactoring Scenarios

### A1: CLI handler line counts

**Given** all handler refactoring tasks are complete
**When** line counts are measured for all CLI command files
**Then**:
- `cmd/new.go` < 50 lines
- `cmd/new_cobra.go` < 50 lines
- `cmd/build.go` < 50 lines
- `cmd/build_cobra.go` < 50 lines
- `cmd/validate.go` < 50 lines
- `cmd/watch.go` < 50 lines (or documented exception)
- `cmd/root.go` documented as acceptable (Cobra wiring)
- `cmd/d2_generator.go` deleted (moved to adapter)

### A2: MCP handler line counts

**Given** all handler refactoring tasks are complete
**When** line counts are measured for all MCP tool handlers
**Then**:
- `internal/mcp/tools/tools.go` is split into individual files
- Each tool's `Call()` method < 30 lines
- `internal/mcp/tools/graph_tools.go` handler methods < 30 lines

### A3: CLI and MCP share use cases

**Given** a `ScaffoldEntityUseCase` exists
**When** `loko new system TestSystem` is run via CLI
**And** `create_system` is called via MCP with the same parameters
**Then** both create identical output files (same entity, same D2 diagram, same template files)

### A4: Existing behavior preserved

**Given** the codebase before handler refactoring
**When** the full test suite is run after refactoring
**Then** all existing tests pass without modification (behavior preservation)

### A5: D2 generator moved to adapter

**Given** `cmd/d2_generator.go` has been moved to `internal/adapters/d2/generator.go`
**When** the `DiagramGenerator` interface is used by `ScaffoldEntityUseCase`
**Then**:
- System context diagrams are generated correctly
- Container diagrams are generated correctly
- Component diagrams are generated correctly
- No import of `cmd/` package from `internal/`

### A6: New ports in ports.go

**Given** handler refactoring is complete
**When** `internal/core/usecases/ports.go` is inspected
**Then** it contains:
- `DiagramGenerator` interface
- `UserPrompter` interface
- `ReportFormatter` interface
- All existing 18 interfaces unchanged

---

## Part B: TOON v3.0 Alignment Scenarios

### B1: TOON encoding produces valid TOON v3.0

**Given** a Project entity with 3 systems, each with 2 containers
**When** `EncodeTOON()` is called on the project
**Then**:
- Output is valid TOON v3.0 (parseable by any TOON v3.0 parser)
- Tabular arrays are used for uniform collections
- Length markers `[N]` are present on arrays
- Indentation uses 2-space hierarchy

### B2: TOON round-trip encoding

**Given** an architecture data structure
**When** `EncodeTOON()` produces TOON output
**And** `DecodeTOON()` parses it back
**Then** the decoded data matches the original input (field-by-field comparison)

### B3: TOON encoding performance

**Given** a project with 20 systems, 60 containers, 200 components
**When** `EncodeTOON()` is called
**Then** encoding completes in < 10ms

### B4: Token efficiency improvement

**Given** identical architecture data
**When** encoded as JSON via `EncodeJSON()`
**And** encoded as TOON via `EncodeTOON()`
**Then** TOON output uses > 30% fewer tokens than JSON

### B5: Token efficiency for tabular arrays

**Given** a uniform array of 10 containers with name and technology fields
**When** encoded as TOON with tabular array notation
**Then** TOON output uses > 50% fewer tokens than JSON equivalent

### B6: TOON struct tags on entities

**Given** entity structs in `internal/core/entities/`
**When** inspected for struct tags
**Then**:
- `Project` has `toon:"..."` tags on serializable fields
- `System` has `toon:"..."` tags on serializable fields
- `Container` has `toon:"..."` tags on serializable fields
- `Component` has `toon:"..."` tags on serializable fields
- Existing `json:"..."` tags are preserved

### B7: TOON error handling

**Given** invalid TOON input (malformed syntax)
**When** `DecodeTOON()` is called
**Then** a clear error message is returned with location information

### B8: Deprecated custom format

**Given** the old custom TOON format
**When** a user requests `--format compact`
**Then**:
- A deprecation warning is displayed
- Output is generated (backward compatibility)
- Warning suggests using `--format toon` instead

### B9: MCP query with TOON format

**Given** an architecture with 5 systems
**When** `query_architecture` MCP tool is called with `format: "toon"` and `detail: "structure"`
**Then**:
- Response is valid TOON v3.0
- Response uses tabular arrays for system/container lists
- Response is ~500 tokens or fewer

### B10: Clean architecture isolation

**Given** `internal/core/` directory
**When** checked for imports
**Then**:
- No imports of `toon-format/toon-go` from `internal/core/`
- `toon-format/toon-go` is only imported in `internal/adapters/encoding/`
- All TOON encoding goes through the `OutputEncoder` port interface

---

## Verification Checklist

| # | Scenario | Automated | Manual |
|---|----------|-----------|--------|
| A1 | CLI line counts | `wc -l cmd/*.go` | Review |
| A2 | MCP line counts | `wc -l internal/mcp/tools/*.go` | Review |
| A3 | CLI/MCP parity | Integration test | - |
| A4 | Behavior preserved | `go test ./...` | - |
| A5 | D2 adapter moved | `go build ./...` | Import check |
| A6 | New ports exist | `go vet ./...` | Review |
| B1 | Valid TOON output | Unit test | - |
| B2 | Round-trip | Unit test | - |
| B3 | Performance | Benchmark test | - |
| B4 | Token reduction > 30% | Benchmark test | - |
| B5 | Tabular reduction > 50% | Benchmark test | - |
| B6 | Struct tags | Unit test (reflect) | Review |
| B7 | Error handling | Unit test | - |
| B8 | Deprecated format | Unit test | - |
| B9 | MCP TOON query | Integration test | - |
| B10 | Clean architecture | Import analysis | Review |
