# Feature Specification: TOON Alignment & Handler Refactoring

**Feature Branch**: `005-toon-alignment`
**Created**: 2026-02-06
**Status**: Draft
**Spec Version**: 0.3.0
**Input**: User description: "005-toon-alignment should be implemented using clean architecture, meaning if in the future a spec update or a better implementation is available not much rewrite should occur"

---

## Overview

This feature combines two related goals:

1. **TOON Format Alignment** — Replace loko's custom semicolon-delimited format with the official TOON v3.0 specification for interoperability and token efficiency.
2. **Thin Handler Refactoring** — Pay down existing technical debt where CLI commands and MCP tool handlers violate the constitution's thin handler principle (CLI < 50 lines, MCP < 30 lines).

These are bundled because both require extracting business logic into use cases and both enforce the same clean architecture principle: handlers delegate, they don't implement.

**Why now**: A constitution audit revealed 10 files violating the thin handler principle, with business logic embedded in `cmd/` and `internal/mcp/tools/`. Fixing this before adding new features (TOON alignment, MCP tool description updates) prevents compounding the debt.

**Clean architecture priority**: All changes must isolate logic behind well-defined interfaces so that future TOON spec updates, library swaps, or handler changes require modifications only in the adapter layer.

---

## Problem Statement

### Problem 1: Non-Standard TOON Format

loko's TOON encoder outputs a custom semicolon-delimited format:

```
{n:OrderService;d:Handles orders;c:[{n:API;t:Go}]}
```

**Issues:**
- Not compliant with the official TOON specification
- Custom key abbreviations not recognized by TOON parsers
- No tabular array support (TOON's key efficiency feature)
- Decode not implemented
- Claims "TOON" but isn't TOON — confuses users and LLMs

### Problem 2: Bloated Handlers (Constitution Violation)

A constitution audit found **10 files** violating Principle III (Thin Handlers):

**CLI commands (limit: < 50 lines):**

| File | Lines | Business Logic Found |
|------|-------|---------------------|
| `cmd/new.go` | 504 | Entity creation, D2 generation, template scaffolding |
| `cmd/d2_generator.go` | 282 | Domain service (not a handler at all) |
| `cmd/build.go` | 251 | Adapter instantiation, inline progress reporter |
| `cmd/new_cobra.go` | 199 | Cobra wrapper with embedded logic |
| `cmd/root.go` | 162 | Root command setup with configuration logic |
| `cmd/watch.go` | 146 | File watching orchestration |
| `cmd/validate.go` | 142 | Validation orchestration |
| `cmd/build_cobra.go` | 107 | Cobra wrapper with embedded logic |

**MCP tools (limit: < 30 lines per handler):**

| File | Lines | Business Logic Found |
|------|-------|---------------------|
| `internal/mcp/tools/tools.go` | 1,084 | 10+ tools in one file, handlers 30-100+ lines each |
| `internal/mcp/tools/graph_tools.go` | 348 | Graph query logic in handler layer |

### Desired State

- loko outputs official TOON v3.0 format with tabular arrays and indentation-based hierarchy
- All CLI commands are thin wrappers (< 50 lines) delegating to use cases
- All MCP tool handlers are thin wrappers (< 30 lines) delegating to use cases
- Business logic lives exclusively in `internal/core/usecases/`
- Domain services live in `internal/adapters/` or `internal/core/usecases/`, not `cmd/`

---

## User Scenarios & Testing

### User Story 1 - Spec-Compliant TOON Output (Priority: P1)

As an LLM consuming loko's architecture data, I want TOON output that matches the official specification so I can reliably parse the format I've been trained on.

**Why this priority**: LLMs are increasingly trained on TOON examples from the official spec. Non-standard formats reduce comprehension accuracy.

**Independent Test**: Output from `query_architecture --format toon` can be parsed by any official TOON reference implementation.

**Acceptance Scenarios**:

1. **Given** user requests TOON format, **When** loko outputs architecture, **Then** output validates against official TOON v3.0 grammar
2. **Given** architecture has multiple systems, **When** output as TOON, **Then** systems render as tabular array (fields declared once, rows streamed)
3. **Given** TOON output from loko, **When** parsed by any official TOON parser, **Then** parses successfully with correct data

---

### User Story 2 - Token Efficiency (Priority: P1)

As a user working within context limits, I want TOON output that achieves meaningful token savings over JSON so I can work with larger architectures.

**Why this priority**: Token efficiency is TOON's core value proposition. The custom format may not achieve the savings that TOON's tabular notation provides.

**Independent Test**: Benchmark TOON vs JSON output on representative architecture data shows measurable token reduction.

**Acceptance Scenarios**:

1. **Given** architecture with 5+ systems and 15+ containers, **When** output as TOON vs JSON, **Then** TOON uses at least 30% fewer tokens
2. **Given** tabular data (uniform arrays), **When** output as TOON, **Then** achieves at least 50% reduction vs JSON
3. **Given** nested structures, **When** output as TOON, **Then** indentation-based format preserves hierarchy readably

---

### User Story 3 - Backward Compatibility (Priority: P2)

As an existing loko user, I want the format transition to not break my existing workflows while I migrate to the new format.

**Why this priority**: Existing users may have tooling built around the current JSON output format.

**Independent Test**: Existing `--format json` continues to work unchanged.

**Acceptance Scenarios**:

1. **Given** user specifies `--format json`, **When** querying architecture, **Then** JSON output remains identical to pre-change behavior
2. **Given** user specifies `--format toon`, **When** querying architecture, **Then** official TOON v3.0 format is used
3. **Given** existing code uses the custom format, **When** the old format is deprecated, **Then** a clear warning message guides migration

---

### User Story 4 - Round-Trip Support (Priority: P2)

As a developer, I want to both encode and decode TOON so I can use it for data exchange and import.

**Why this priority**: Current implementation only encodes. Full round-trip support enables importing architecture definitions.

**Independent Test**: Encode architecture to TOON, decode back, verify data matches original.

**Acceptance Scenarios**:

1. **Given** architecture data, **When** encoded to TOON then decoded, **Then** decoded data matches the original
2. **Given** a valid TOON input file, **When** loko reads it, **Then** it parses correctly into architecture data
3. **Given** malformed TOON input, **When** decoded, **Then** a clear error message indicates the problem location

---

### User Story 5 - Clean Architecture Isolation (Priority: P1)

As a maintainer, I want the TOON implementation isolated behind interfaces so that future spec updates, library swaps, or alternative implementations require minimal changes outside the encoding adapter.

**Why this priority**: The user explicitly requires clean architecture. TOON is an evolving specification — the implementation must be swappable without cascading changes.

**Independent Test**: Swapping the TOON adapter (e.g., from library-based to hand-written) requires changes only in the adapter package and its wiring — no changes to use cases, CLI commands, or MCP tools.

**Acceptance Scenarios**:

1. **Given** the TOON encoder/decoder is implemented, **When** a new library or implementation is available, **Then** only the adapter layer needs modification
2. **Given** a use case calls the encoding interface, **When** the underlying TOON format changes, **Then** the use case code remains unchanged
3. **Given** the existing `OutputEncoder` interface in ports, **When** TOON alignment is complete, **Then** all TOON encoding/decoding goes through that interface

---

### User Story 6 - Thin CLI Handlers (Priority: P1)

As a maintainer, I want all CLI command handlers to be thin wrappers (< 50 lines) that delegate to use cases, so that business logic is reusable across CLI, MCP, and API interfaces.

**Why this priority**: 8 CLI files violate the constitution. Business logic embedded in `cmd/` cannot be reused by MCP tools or the HTTP API, causing duplication and inconsistency.

**Independent Test**: After refactoring, every `cmd/*.go` file's handler function (Run/Execute) is under 50 lines, and the business logic it previously contained now lives in `internal/core/usecases/` with unit tests.

**Acceptance Scenarios**:

1. **Given** `cmd/new.go` currently has 504 lines, **When** refactored, **Then** handler is < 50 lines and entity creation logic is in a use case
2. **Given** `cmd/d2_generator.go` is a domain service in `cmd/`, **When** refactored, **Then** it is moved to `internal/adapters/d2/` or `internal/core/usecases/`
3. **Given** `cmd/build.go` contains adapter instantiation and an inline progress reporter, **When** refactored, **Then** handler is < 50 lines, reporter is in an adapter, and build orchestration is in a use case
4. **Given** any CLI command, **When** its handler is inspected, **Then** it only parses input, calls a use case, and formats output

---

### User Story 7 - Thin MCP Tool Handlers (Priority: P1)

As a maintainer, I want all MCP tool handlers to be thin wrappers (< 30 lines) that delegate to use cases, so that tool logic is testable and consistent with CLI behavior.

**Why this priority**: `tools.go` is 1,084 lines with 10+ tools whose handlers range from 30-100+ lines. This makes MCP tools hard to test and maintain independently.

**Independent Test**: After refactoring, each MCP tool's `Call()` method is under 30 lines, tools are split into individual files, and business logic is in shared use cases.

**Acceptance Scenarios**:

1. **Given** `tools.go` is 1,084 lines with 10+ tools, **When** refactored, **Then** each tool is in its own file with a handler < 30 lines
2. **Given** `graph_tools.go` has 348 lines of graph query logic, **When** refactored, **Then** query logic is in a use case and the handler is < 30 lines
3. **Given** an MCP tool and a CLI command perform the same operation (e.g., create system), **When** both are inspected, **Then** both delegate to the same use case

---

### Edge Cases

- Architecture names contain special characters (pipes, colons) → TOON quoting rules must be applied
- Arrays contain non-uniform items (mixed field sets) → Fall back to nested object notation
- Existing scripts depend on custom format output → Deprecation warning issued; JSON remains as the stable alternative
- TOON spec evolves beyond v3.0 → Implementation pinned to v3.0; future updates handled through adapter swap
- Empty architecture (no systems/containers) → Valid TOON output with project header only
- Refactored handlers must preserve existing CLI/MCP behavior exactly (no user-visible changes)
- `cmd/root.go` setup logic (162 lines) may be acceptable if it's purely Cobra wiring — assess whether it contains business logic or just command registration

---

## Requirements

### Functional Requirements

#### TOON Alignment

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-001 | TOON encoder MUST comply with official TOON v3.0 specification | P1 |
| FR-002 | TOON encoder MUST use tabular array notation for uniform arrays (key efficiency feature) | P1 |
| FR-003 | TOON encoder MUST use indentation-based hierarchy for nested structures | P1 |
| FR-004 | TOON decoder MUST parse valid TOON v3.0 documents into architecture data | P2 |
| FR-005 | TOON decoder MUST provide clear, location-aware error messages for invalid input | P2 |
| FR-006 | JSON output format MUST remain unchanged for backward compatibility | P1 |
| FR-007 | TOON output MUST achieve at least 30% token reduction vs equivalent JSON on architecture data | P1 |
| FR-008 | TOON output MUST be parseable by official TOON reference implementations | P1 |
| FR-009 | Custom (non-standard) format SHOULD be deprecated with clear migration guidance | P2 |
| FR-010 | TOON encoding/decoding MUST be isolated behind the existing port interface so adapter implementations are swappable without changing use cases or handlers | P1 |

#### Handler Refactoring

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-011 | All CLI command handlers MUST be < 50 lines of handler code | P1 |
| FR-012 | All MCP tool handlers MUST be < 30 lines of handler code | P1 |
| FR-013 | Business logic currently in `cmd/` MUST be extracted to use cases in `internal/core/usecases/` | P1 |
| FR-014 | Domain services in `cmd/` (e.g., `d2_generator.go`) MUST be moved to `internal/adapters/` or `internal/core/usecases/` | P1 |
| FR-015 | MCP tools in `tools.go` MUST be split into individual files (one tool per file) | P1 |
| FR-016 | CLI commands and MCP tools performing the same operation MUST delegate to the same use case | P1 |
| FR-017 | Refactoring MUST NOT change any user-visible behavior (CLI output, MCP responses) | P1 |

### Key Entities

- **TOON Document**: A text representation of architecture data in official TOON v3.0 format; supports scalar values, indentation-based objects, and tabular arrays
- **Output Format**: An enumeration of supported serialization formats (JSON, TOON); user-selectable via `--format` flag
- **Tabular Array**: TOON's compact representation where field names are declared once and rows contain pipe-separated values — the primary token-saving mechanism
- **Use Case**: Application-level business logic callable from any interface (CLI, MCP, API) via dependency injection

---

## Success Criteria

### Measurable Outcomes

- **SC-001**: All TOON output from loko validates against the official TOON v3.0 grammar (100% compliance)
- **SC-002**: TOON output uses at least 30% fewer tokens than equivalent JSON output on a benchmark dataset of 5 systems / 15 containers
- **SC-003**: TOON output can be parsed by any official TOON reference implementation without errors
- **SC-004**: Existing JSON output (`--format json`) produces identical results before and after the change
- **SC-005**: Encoding architecture data then decoding it back produces data matching the original (round-trip fidelity)
- **SC-006**: Swapping the TOON adapter implementation requires zero changes to use case or handler code (clean architecture validation)
- **SC-007**: All CLI command handler functions are under 50 lines (verified by line count)
- **SC-008**: All MCP tool handler `Call()` methods are under 30 lines (verified by line count)
- **SC-009**: Zero business logic remains in `cmd/` — only input parsing, use case invocation, and output formatting
- **SC-010**: All existing CLI and MCP tests continue to pass after refactoring (behavior preservation)

---

## Scope & Exclusions

### In Scope

- Spec-compliant TOON v3.0 encoder
- TOON v3.0 decoder (for round-trip support)
- Migration path from custom format with deprecation warnings
- Token efficiency benchmarks
- Extract business logic from 8 CLI files into use cases
- Split and slim down 2 MCP tool files into individual thin handlers
- Move `cmd/d2_generator.go` to appropriate layer
- Updated documentation

### Out of Scope (Future)

- TOON schema validation
- TOON streaming encoder
- Custom TOON extensions beyond v3.0
- TOON-based configuration files (e.g., loko.toon)
- Support for TOON versions beyond v3.0
- Refactoring files that are borderline (< 100 lines) and contain only Cobra wiring
- HTTP API handler refactoring (already compliant)

---

## Dependencies & Assumptions

### Dependencies

- [TOON Specification v3.0](https://github.com/toon-format/spec) — Format grammar and rules
- Existing `OutputEncoder` interface in loko's core ports — The contract that adapters must satisfy
- Existing test suite — Must pass before and after refactoring

### Assumptions

- The official TOON v3.0 spec is stable (released 2025)
- A Go-compatible TOON library exists or can be implemented within the adapter layer
- LLMs benefit from standard TOON format over custom formats
- Token savings of 30-60% are achievable with C4 architecture data (which has many uniform arrays)
- The existing `OutputEncoder` port interface is sufficient; if not, it can be extended without breaking consumers
- Handler refactoring is behavior-preserving — no functional changes to CLI or MCP outputs
- Some use cases needed for handler extraction already exist (e.g., `BuildDocs`, `QueryArchitecture`); others will need to be created

---

## Constraints & Tradeoffs

- **Interface stability over implementation flexibility**: The port interface must remain stable; implementation details are confined to the adapter layer
- **Spec compliance over custom optimizations**: Prioritize official TOON v3.0 compliance even if a custom variant could save more tokens
- **Backward compatibility over clean break**: Deprecate old format gradually rather than removing it immediately
- **Pinned spec version**: Target TOON v3.0 specifically; future versions handled as separate adapter updates
- **Refactor before feature**: Handler refactoring should be done before TOON alignment to avoid building new features on top of debt
- **Behavior preservation**: Refactoring must be invisible to users — same inputs produce same outputs

---

## External References

- [TOON Format Website](https://toonformat.dev/) — Official documentation
- [TOON Specification](https://github.com/toon-format/spec) — Formal grammar and rules
- [TOON Benchmarks](https://github.com/toon-format/toon) — Token efficiency data
