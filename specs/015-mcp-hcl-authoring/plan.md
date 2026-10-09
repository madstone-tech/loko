# Implementation Plan: MCP HCL Authoring

**Branch**: `015-mcp-hcl-authoring` | **Date**: 2026-10-05 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/015-mcp-hcl-authoring/spec.md`

**Stage**: 3 of 6 in the [v1.0 roadmap](../012-v1-architecture-dsl/roadmap.md). Builds on features
013 (compiler) and 014 (renderers), both merged, and on the Go 1.27 upgrade.

## Summary

The conversational interface currently registers no tools. This feature gives it five, built on the
compiled IR. Three of them read: `describe`, `query` and `validate`. Two of them write HCL and
nothing else: `apply_edit` and `move`. It also mirrors the query engine as `loko query`.

The approach turns on four decisions from [research.md](./research.md):

1. **Surgical edits through `hclwrite`** (R1). Files are parsed, edited through the
   `Body`/`Block`/`Expression` API and serialised. `hclwrite` keeps every token it was not asked to
   change, so comments, ordering and spacing outside the edited declaration survive byte for byte.
   New declarations are formatted on their own before being spliced in. Existing lines are never
   reformatted.
2. **Plan, compile in memory, then commit** (R2, R3). Edits (single or batch) become in-memory file
   contents. The compiler loads them through an overlay, and only an error-free result is written,
   atomically, after a staleness check against the caller's revision. Preview is the same plan
   without the commit.
3. **One new language construct, the `moved` block** (R4). A rename rewrites the declaring labels
   and every reference with `RenameVariablePrefix`, then appends `moved { from to }`. The IR
   carries moves as an `omitempty` field, so no earlier export golden changes.
4. **One query engine, two front ends** (R5, R11). `usecases.Query` implements dependents,
   dependencies, path, orphans and coupling with descendant inclusion. The MCP tool and
   `loko query` both call it, and a test proves byte-identical output.

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, toolchain go1.27.1)

**Primary Dependencies**: `github.com/hashicorp/hcl/v2` (`hclwrite`, already present) and the
existing `toon-go` encoder. **No new module.** The unified diff for previews is a small in-house
Myers implementation (R9).

**Storage**: Files only. Edits write `*.loko.hcl` files atomically (temp + rename, with rollback).
There is no state file: revisions are computed from file hashes on demand.

**Testing**: `go test`, plus the following.
- Seeded round-trip property test (2,000 sequences per run) and a `FuzzApplyEdits` target (R10).
- Byte-preservation tests on hand-written fixtures.
- An MCP end-to-end build test.
- A CLI-vs-MCP equality test.
- Goroutine-leak gate (`leakcheck.Main`) in the new packages that start goroutines.
- Concrete mocks only (Principle VII).

**Target Platform**: The single loko binary; MCP over stdio.

**Project Type**: Compiler + CLI + MCP server (single Go module, clean architecture).

**Performance Goals**: Reads under 1 s and an edit with its compile check under 2 s on 1,000
elements. A summary-level description of at most 300 TOON tokens (SC-007, SC-008).

**Constraints**:
- Bytes outside the edited declarations unchanged (FR-013).
- Compile before every save, and save all-or-nothing (FR-012, FR-012a).
- Nothing but `*.loko.hcl` is ever written (FR-026).
- Deterministic results (FR-009).
- Handler functions ≤ 30 lines (MCP) and ≤ 50 (CLI).
- Use-case files ≤ 200 lines and functions ≤ 60; entity files ≤ 200; adapter files ≤ 400.

**Scale/Scope**:
- 5 MCP tools, 1 CLI command with 5 subcommands, and 1 language block with 3 diagnostic codes.
- About 10 use-case files and 1 new entity package (about 6 files).
- About 6 new adapter files in `hclsource`.

No `NEEDS CLARIFICATION` remains; the five spec clarifications and research R1–R13 resolve
everything.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design (below).*

| Principle / rule | Status | How |
|---|---|---|
| **I. Clean Architecture** | ✅ | `authoring` entities are stdlib-only. Query, describe and the edit workflow live in use cases. `hclwrite`, file I/O and diffs live in `adapters/hclsource`. MCP and CLI only translate. Wiring is in `cmd/` (v1.4.0). |
| **II. Interface-First** | ✅ | `OverlaySource` and `SourceEditor` live in `ports.go` ([contracts/ports.md](contracts/ports.md)). |
| **III. Thin Handlers** | ✅ | Each MCP tool handler (≤ 30 lines) and `loko query` subcommand (≤ 50) calls one use case. Argument decoding is in shared `tools/helpers.go` (an exempt helper file). |
| **IV. Entity Validation** | ✅ | `authoring.NewEdit` validates op/target combinations, address shape and legal attributes. The `moved` rules are compiler validation, as for every other block. |
| **V. Test-First** | ✅ | Each story's phase opens with failing tests. The property test and the byte-preservation tests come before the editor. |
| **VI. Token Efficiency** | ✅ | TOON is the default for reads (ADR-0011). `describe` levels have token budgets that are asserted (SC-008). Diffs rather than whole files in edit results (R9). |
| **VII. Simplicity** | ✅ | No new module. No state file, cache or database. Five tools, not twenty-five. |
| Dependency Direction | ✅ | MCP and `cmd` never import entities: handlers pass `usecases.EditInput` DTOs. Entity packages import nothing internal, so `authoring` imports neither `arch` nor anything else. `hclwrite` stays confined to `internal/adapters/**`. |
| Removed in v1.0 list | ✅ | The v0 scaffolding tools stay removed (FR-027). No fsnotify, HTTP API or config file. |
| Quality gates | ✅ | lint, test, audit, the goroutine-leak gate and coverage > 80% in core. ADR-0014 records the `moved` block and the authoring engine. |

**Language change governance**: the `moved` block is permanent v1.x surface. It is specified as a
normative contract ([contracts/language-moved.md](contracts/language-moved.md)), documented in
`docs/language.md`, and recorded in ADR-0014.

**Post-design re-check**: ✅ Passes. The design review caught two layering problems before they
became code, and both are corrected in [data-model.md](data-model.md):
- `EditResult` became a use-case DTO, because it carries `arch.Diagnostics` and one entity package
  may not import another.
- Handlers pass `EditInput` DTOs instead of constructing entities.

## Project Structure

### Documentation (this feature)

```text
specs/015-mcp-hcl-authoring/
├── spec.md, plan.md, research.md, data-model.md, quickstart.md
├── contracts/
│   ├── mcp-tools.md        # the five tools, arguments, results, refusals
│   ├── cli-query.md        # loko query
│   ├── language-moved.md   # the moved block (normative language addition)
│   └── ports.md            # OverlaySource, SourceEditor, use-case map
├── checklists/requirements.md
└── tasks.md                # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/core/entities/arch/
├── source_model.go          # EDIT: MovedDecl
├── ir.go                    # EDIT: Moves []Move (omitempty)
└── diagnostic.go            # EDIT: moved_from_declared, moved_to_unresolved, moved_duplicate_from

internal/core/entities/authoring/      # NEW, stdlib only
├── edit.go          # Edit, Op, TargetKind, BindingRef, Attr, NewEdit (validation)
├── edit_schema.go   # legal attributes per target kind
├── batch.go         # Batch (1..100 edits)
├── revision.go      # Revision, FileHash, ParseRevision
└── plan.go          # Plan, FileContent, FileChange, EditError

internal/core/usecases/
├── ports.go                 # EDIT: OverlaySource, SourceEditor
├── resolve_moved.go         # NEW: moved validation (R4)
├── build_ir.go              # EDIT: carry Moves
├── query.go, query_graph.go # NEW: Query (R5)
├── describe.go              # NEW: Describe (R6)
├── validate_project.go      # NEW: Validate DTO + revision
├── authoring_service.go     # NEW: AuthoringService (mutex), EditInput, EditResult
├── apply_edits.go, apply_plan.go, apply_dependents.go   # NEW: the write flow (data-model §5)
└── leak_main_test.go        # existing gate covers the new code

internal/adapters/hclsource/
├── decode_moved.go          # NEW: parse moved blocks
├── parse.go                 # EDIT: overlay-aware file reads; LoadOverlay
├── edit_plan.go             # NEW: SourceEditor.Plan, dispatch per op
├── edit_blocks.go           # NEW: locate, add, update, remove blocks; canonical new blocks
├── edit_rename.go           # NEW: label rewrite, RenameVariablePrefix, append moved
├── edit_commit.go           # NEW: revision hashing, stale check, atomic commit + rollback
├── edit_diff.go             # NEW: unified diff (Myers)
├── edit_property_test.go    # NEW: R10 property test + FuzzApplyEdits
└── testdata/handwritten/    # NEW: comment- and spacing-heavy fixtures

internal/mcp/
├── server.go                # EDIT: sorted tools/list; request context to tools
└── tools/
    ├── helpers.go           # NEW: argument decoding, format/encode, refusal shaping
    ├── describe.go, query.go, validate.go, apply_edit.go, move.go   # NEW: ≤30-line handlers
    ├── schemas.go           # REWRITE: the five input schemas
    └── *_test.go            # protocol-level tests, end-to-end build, writes-only-HCL

cmd/
├── mcp.go                   # EDIT: register the five tools
├── wiring.go                # EDIT: newAuthoringDeps
├── query.go, query_cobra.go # NEW: loko query
└── query_test.go            # NEW: CLI == MCP equality; authoring performance test

docs/language.md (moved), docs/cli-reference.md (query), docs/mcp-integration.md (rewrite)
docs/adr/0014-hcl-authoring.md       # NEW
internal/_parked/mcp_tools/          # DELETE (superseded); _parked/README.md updated
```

**Structure Decision**: The existing clean-architecture layout, extended in place. Editing lives in
`hclsource` beside the parser and formatter that already own HCL, so `hclwrite` stays in one
adapter. The parked v0 MCP tools are deleted rather than reworked, because they wrote the old
file-tree model and nothing in them carries over.

## Complexity Tracking

No constitution violations. The table is intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
