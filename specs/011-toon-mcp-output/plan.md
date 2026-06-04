# Implementation Plan: TOON Output Format for MCP Query Tools

**Branch**: `011-toon-mcp-output` | **Date**: 2026-06-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/011-toon-mcp-output/spec.md`

## Summary

Make TOON (Token-Oriented Object Notation) the default output format for all seven MCP read tools, with an optional `format` parameter to fall back to JSON for debugging. Mutation and validation tools remain unchanged (JSON). Add a reproducible benchmark suite that gates the ≥30% token-reduction claim.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: `github.com/toon-format/toon-go` (already in go.mod), stdlib `encoding/json`  
**Storage**: N/A (in-memory response formatting)  
**Testing**: Go test (`go test ./...`)  
**Target Platform**: All platforms (loko runs where Go runs)  
**Project Type**: CLI tool + MCP server  
**Performance Goals**: Zero measurable latency overhead for format selection; benchmark runs in <5s  
**Constraints**: MCP tool handler functions ≤ 30 effective lines; use-case files ≤ 200 effective lines  
**Scale/Scope**: 7 read tools affected; ~30 MCP tools total

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Architecture | ✓ PASS | Changes stay in `internal/mcp/tools/` (handlers) and `internal/adapters/encoding/` (adapter). No core logic changes. |
| II. Interface-First | ✓ PASS | `OutputEncoder` interface already exists in `ports.go`; TOON implementation already exists in `adapters/encoding/`. |
| III. Thin Handlers | ⚠ CHECK | Each tool's `Call()` must stay ≤ 30 effective lines. Format dispatch will be a single helper call. |
| IV. Entity Validation | ✓ PASS | No new entities; no validation changes. |
| V. Test-First | ✓ PASS | Existing tests for `query_architecture` cover format paths. New tests needed for the other 6 tools. |
| VI. Token Efficiency | ⚠ CHECK | TOON as default for MCP read tools conflicts with constitution "Default to JSON for compatibility; TOON is opt-in". See ADR 0011 for resolution. |
| VII. Simplicity & YAGNI | ✓ PASS | Reuses existing `Encoder` and `ExecuteWithFormat` patterns. No new abstractions. |
**Post-design re-check**: After Phase 1, verify handler function line counts remain ≤ 30 effective lines. Verify ADR 0011 is ratified before implementation.


## Project Structure

### Documentation (this feature)

```text
specs/011-toon-mcp-output/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── mcp/tools/                    # MCP tool handlers (7 read tools updated)
│   ├── query_project.go
│   ├── query_architecture.go
│   ├── query_dependencies.go
│   ├── query_related_components.go
│   ├── search_elements.go
│   ├── list_relationships.go
│   ├── analyze_coupling.go
│   └── helpers.go                # New: formatResponse(), validateFormat()
├── adapters/encoding/
│   ├── toon.go                   # Already exists
│   └── toon_benchmark_test.go    # Expand for gating benchmark
├── core/usecases/
│   └── ports.go                  # OutputEncoder already defined
└── ...

tests/
├── benchmarks/
│   └── token_efficiency_test.go  # New: gating benchmark suite
└── mcp/
    └── tool_format_test.go       # New: E2E tests for format param on all 7 tools
```

**Structure Decision**: Changes are confined to existing layers. No new directories. The benchmark suite lives in `tests/benchmarks/` (existing). E2E tests for format behavior live in `tests/mcp/` (existing).

## Complexity Tracking

> No constitution violations anticipated. All changes are thin-handler pattern with helper delegation.

| Risk | Mitigation |
|------|------------|
| Handler function line budget exceeded | Extract format dispatch to `helpers.go`; keep `Call()` bodies to 3 lines (parse format → build payload → return `formatResponse()`) |
| Breaking existing MCP clients | Documented behavior change; JSON escape hatch available via `format: "json"` |
| TOON encoding fails on edge-case data | `Encoder.EncodeTOON` already handles `toon` struct tags; fallback to `json.Marshal` if TOON errors |
