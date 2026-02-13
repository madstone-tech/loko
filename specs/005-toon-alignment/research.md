# Research: TOON Alignment & Handler Refactoring

**Feature**: 005-toon-alignment | **Date**: 2026-02-06 | **Phase**: 0 (Research)

---

## 1. TOON v3.0 Specification Research

### 1.1 Official Specification

- **Spec URL**: https://toonformat.dev
- **Version**: TOON v3.0 (current)
- **Purpose**: Token-Optimized Object Notation — structured data format designed for LLM context efficiency

### 1.2 Key TOON v3.0 Features

| Feature | Description | Relevance to loko |
|---------|-------------|-------------------|
| Tabular arrays | Arrays of uniform objects rendered as header + rows | Systems, Containers, Components are uniform arrays |
| Length markers | `[N]` suffix on array headers for parser hints | Enables efficient parsing of architecture data |
| Indentation hierarchy | Nested data via 2-space indentation | Maps naturally to C4 hierarchy (System → Container → Component) |
| Struct tags | `toon:"fieldname"` on Go structs | Same pattern as `json:"fieldname"` — familiar to Go developers |
| Omitempty support | `toon:"name,omitempty"` | Skip empty optional fields (tags, metadata) |

### 1.3 Go Library: toon-format/toon-go

| Attribute | Value |
|-----------|-------|
| Repository | github.com/toon-format/toon-go |
| License | MIT |
| Go version | 1.23+ (loko uses 1.25) |
| External deps | Zero |
| API surface | `toon.Marshal()`, `toon.Unmarshal()`, options via `toon.With*()` |
| Test coverage | 349 test fixtures from official spec |
| Last activity | Dec 2025 |
| Stars | ~102 |

**API Example**:
```
toon.Marshal(value)                         // basic encoding
toon.Marshal(value, toon.WithLengthMarkers(true))  // with [N] markers
toon.Unmarshal(data, &target)               // decoding
```

**Struct Tag Pattern**:
```
type System struct {
    Name string `toon:"name"`
    Desc string `toon:"description,omitempty"`
}
```

### 1.4 Alternatives Considered

| Library | Decision | Reason |
|---------|----------|--------|
| toon-format/toon-go | **Selected** | Official, zero deps, familiar API, full spec compliance |
| alpkeskin/gotoon | Rejected | Slower benchmarks, less adoption, incomplete spec coverage |
| Custom implementation | Rejected | 2-4 weeks effort, maintenance burden, bug risk |
| Keep current custom format | Rejected | Non-standard, confuses LLMs, no ecosystem tooling |

### 1.5 Risk Assessment

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| toon-go has no semver tags yet | Known | Pin to specific commit hash in go.mod |
| Breaking API changes | Low | Library follows Go stdlib patterns; adapter isolates impact |
| Missing features | Medium-Low | Can extend adapter with custom logic if needed |

---

## 2. Handler Audit Results

### 2.1 Methodology

Audited all files in `cmd/` and `internal/mcp/tools/` against constitution Principle III:
- CLI handlers: < 50 lines
- MCP tool handlers: < 30 lines

### 2.2 Violations Found: 10 Files

#### CLI Handlers (8 files)

| File | Lines | Business Logic Found |
|------|-------|---------------------|
| `cmd/new.go` | 504 | Entity creation, D2 generation, template rendering, interactive prompts, validation |
| `cmd/d2_generator.go` | 282 | Full D2 diagram generation (system context, container, component) — not a handler but misplaced domain service |
| `cmd/build.go` | 251 | Build orchestration, progress reporting, format parsing, D2 rendering pipeline |
| `cmd/new_cobra.go` | 199 | Cobra wiring with embedded entity creation and D2 generation logic |
| `cmd/root.go` | 162 | Cobra root setup — **likely acceptable** (pure wiring, no business logic) |
| `cmd/watch.go` | 146 | File watching event loop, rebuild triggering — assess if event loop is inherently handler-level |
| `cmd/validate.go` | 142 | Validation orchestration, report formatting with lipgloss styling |
| `cmd/build_cobra.go` | 107 | Cobra wiring with embedded build orchestration |

#### MCP Tool Handlers (2 files)

| File | Lines | Business Logic Found |
|------|-------|---------------------|
| `internal/mcp/tools/tools.go` | 1,084 | 7 tool handlers with entity creation, D2 generation, build triggering, validation — all inline |
| `internal/mcp/tools/graph_tools.go` | 348 | Graph query construction, architecture traversal — should be a use case |

### 2.3 Already Compliant Files

| File | Lines | Status |
|------|-------|--------|
| `internal/mcp/tools/query_architecture.go` | 103 | Separate file, delegates to use case |
| `internal/mcp/tools/query_project.go` | 74 | Separate file, delegates to use case |
| `internal/mcp/tools/schemas.go` | 172 | Schema definitions only (data, not logic) |

### 2.4 Root Causes

1. **No use cases existed** when CLI/MCP were first written — logic was placed in handlers by necessity
2. **Code duplication** between CLI and MCP — same entity creation logic exists in both `cmd/new.go` and `mcp/tools/tools.go`
3. **Misplaced domain services** — `cmd/d2_generator.go` contains D2 generation logic that belongs in adapters
4. **UI concerns mixed with logic** — `cmd/validate.go` mixes validation orchestration with lipgloss report formatting

### 2.5 Shared Logic Identified

| Logic | Found In | Target |
|-------|----------|--------|
| Create system + save + generate D2 | `cmd/new.go`, `mcp/tools/tools.go` | `ScaffoldEntityUseCase` |
| Create container + save + generate D2 | `cmd/new.go`, `mcp/tools/tools.go` | `CreateContainerUseCase` + `ScaffoldEntityUseCase` |
| Create component + save + generate D2 | `cmd/new.go`, `mcp/tools/tools.go` | `CreateComponentUseCase` + `ScaffoldEntityUseCase` |
| D2 diagram generation | `cmd/d2_generator.go`, `mcp/tools/tools.go` | `internal/adapters/d2/generator.go` |
| Build orchestration | `cmd/build.go` | Already in `build_docs.go` (enhance) |
| Validation + report | `cmd/validate.go`, `mcp/tools/tools.go` | Already in `validate_architecture.go` + new `ReportFormatter` adapter |

---

## 3. Current Encoder Analysis

### 3.1 Custom Format (Non-Standard)

The current `internal/adapters/encoding/toon.go` implements a custom format:
- Semicolon-delimited key-value pairs: `{n:PaymentService;d:Handles payments}`
- Abbreviated keys: `n=name`, `d=description`, `t=technology`, etc.
- Array notation: `[v1;v2;v3]`
- Booleans: `T`/`F`
- Null: `-`

### 3.2 Problems with Current Format

| Problem | Impact |
|---------|--------|
| Not TOON v3.0 compliant | LLMs trained on TOON can't parse it; tooling incompatible |
| No tabular array support | Uniform arrays (systems, containers) not optimized |
| No decoder | `DecodeTOON()` returns error — round-trip impossible |
| Custom key abbreviations | Non-standard; LLMs must learn custom format per-project |
| Reflection-based | Fragile; no struct tag support for TOON-specific field names |

### 3.3 Migration Path

- Replace `EncodeTOON()` body with `toon.Marshal(value, toon.WithLengthMarkers(true))`
- Replace `DecodeTOON()` body with `toon.Unmarshal(data, value)`
- Add `toon:"fieldname"` struct tags to entities
- Remove custom `keyAbbreviations` map and reflection helpers
- Keep `ArchitectureSummary` and related types but add `toon` struct tags

---

## 4. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | Use `toon-format/toon-go` for TOON encoding | Official, zero deps, spec-compliant, familiar Go API |
| D2 | Pin toon-go to commit hash | No semver tags yet; prevents breaking changes |
| D3 | Refactor handlers BEFORE TOON alignment | Clean use case boundaries needed for proper encoder integration |
| D4 | Create `ScaffoldEntityUseCase` as orchestrator | Single entry point for create + D2 + template prevents logic duplication |
| D5 | `cmd/root.go` (162 lines) is acceptable | Pure Cobra wiring, no business logic — document as exception |
| D6 | Assess `cmd/watch.go` (146 lines) during implementation | Event loop may be inherently handler-level code |
