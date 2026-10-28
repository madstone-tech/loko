# Implementation Plan: Dynamic Views

**Branch**: `019-dynamic-views` | **Date**: 2026-10-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/019-dynamic-views/spec.md`

## Summary

A `view` with labelled `step` blocks is a dynamic view: an ordered flow through declared elements,
each step checked against the declared relationships (R1–R4). The compiler resolves `from`, `to`
and `via` like any reference, matches each step to a relationship through containment and, for a
`reply`, in reverse, and reports a mismatch as an error and an unmatched step as a warning. The
projection builds a dedicated view model: participants inside their system boundary, one numbered
edge per step, and an ordered step list (R6). The D2 adapter draws it with numbered edges or, with
`style = "sequence"`, as a D2 sequence diagram (R7). Pages gain a Steps table; the export, MCP
`describe`, a new `query` kind `flows` and `apply_edit` (a `step` target with `before`) cover the
rest (R8–R10).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, toolchain pinned)

**Primary Dependencies**: existing only: hcl/v2 and hclwrite (decoder, edits), in-process d2
v0.7.1 (`shape: sequence_diagram`, verified with a probe under dagre and ELK), `html/template`.
No new module.

**Storage**: files only, as today.

**Testing**: `go test ./...` with `-race` and the goroutine-leak gate. Decoder goldens for each
declaration rule; matching table tests in use cases; D2 goldens for both styles; HTML and markdown
goldens for the Steps table; an MCP end-to-end case (add, insert with `before`, remove, build);
`TestLanguageParity`; `TestValidationRuleCoverage` for the three new codes; a generated 40-step
flow for SC-006.

**Target Platform**: macOS, Linux and Windows CLI; MCP over stdio.

**Project Type**: CLI compiler with an MCP server.

**Performance Goals**: matching is O(steps × relationships) with containment checks on the parent
map; a 40-step flow over the largest example builds within the existing build budget (SC-006).

**Constraints**: byte-stable output (FR-013): models without steps produce identical exports, D2,
SVG, markdown and HTML. View model invariant (edges sorted) is kept; step order travels in
`ViewModel.Steps`. Exit codes unchanged. `internal/mcp` imports no entity or render adapter.

**Scale/Scope**: 1 nested block, 5 step attributes, 1 view attribute, 3 diagnostic codes, 1 edit
target, 1 query kind, 1 emitter, 1 page table.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle / rule | Status | How |
|---|---|---|
| **I. Clean Architecture** | ✅ | Step declarations and IR are entities; resolution, matching and projection are use cases; HCL decoding and edits stay in `hclsource`, drawing in `d2`, pages in `html` and `markdown`. |
| **II. Interface-First** | ✅ | No new port: steps travel in the source model, IR and view model already crossing the existing ports. |
| **III. Thin Handlers** | ✅ | MCP and CLI `query` gain one kind value; `apply_edit` gains a target handled in the use case and adapter. Handlers stay within budget. |
| **IV. Entity Validation** | ✅ | `before` and the `step` target are validated in `authoring.NewEdit`; value sets come from the language table. |
| **V. Test-First** | ✅ | Each phase opens with failing tests: goldens, matching tables, parity, rule coverage. |
| **VI. Token Efficiency** | ✅ | New fields are `omitempty`; `describe` adds steps only to views that have them; `flows` answers return addresses, not full elements. |
| **VII. Simplicity** | ✅ | Reuses `view`, the resolver, the symbol table, the canvas title, the boundary style and D2's own sequence layout. No new module or port. |
| Dependency Direction | ✅ | `authoring` keeps its own copy of the step attributes, held by `TestLanguageParity`. |
| File-size budgets | ✅ | New files keep budgets: `arch/source_model_view.go` and `arch/ir_view.go` (entities ≤ 200), `usecases/validate_steps.go`, `usecases/project_dynamic.go`, `usecases/query_flows.go` (≤ 200), `d2/emit_sequence.go`, `hclsource/edit_steps.go` (≤ 400). `source_model.go` (184) and `ir_element.go` move view types out rather than grow. |
| Quality gates | ✅ | lint, test (`go test -cover ./internal/core/...`), audit, race and leak gate. ADR-0018 records the decisions. |

**Language change governance**: one permanent v1.x nested block and attribute, in
[contracts/language.md](contracts/language.md) and `docs/language.md`.

**Output change**: none for models without steps. MCP `describe` and `query` add fields only for
views with steps and for the new kind.

**Post-design re-check**: ✅ Passes. The design adds use-case and adapter files and fields on
existing types; no layer gains an import.

## Project Structure

### Documentation (this feature)

```text
specs/019-dynamic-views/
├── spec.md
├── plan.md               # this file
├── research.md           # R1–R12
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── language.md
│   └── cli-mcp.md
├── checklists/
│   └── requirements.md
└── tasks.md              # /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/entities/arch/
├── schemas.go                 # step block; view style (R1)
├── source_model_view.go       # NEW: ViewDecl (moved from source_model.go), StepDecl
├── ir_view.go                 # NEW: View (moved from ir_element.go), Step
├── address.go                 # NewStepAddress
└── diagnostic.go, diagnostic_ignore.go   # 3 codes; step_unmatched ignorable (R4)

internal/core/entities/authoring/
├── edit.go                    # TargetStep; Edit.Before
├── edit_schema.go             # step attributes; view style
└── render_attrs.go            # style enum

internal/core/entities/viewmodel/
├── view.go                    # View.Style
└── model.go                   # ViewModel.Steps, StepRow

internal/core/usecases/
├── symbol_table.go            # declare step addresses
├── resolve_references.go      # resolve from, to, via
├── validate_steps.go          # NEW: selection conflicts, matching (R2, R3)
├── build_ir.go                # IR steps
├── resolve_views.go           # View.Style
├── project_view.go            # dispatch views with steps
├── project_dynamic.go         # NEW: participants, boundaries, edges, rows (R6)
├── wide_view.go               # skip the sequence style
├── describe.go, describe_types.go   # steps in ViewInfo (R10)
└── query_flows.go, query.go   # NEW kind flows (R10)

internal/adapters/hclsource/
├── decode_project.go          # view style, step blocks
├── edit_locate.go             # locate steps
└── edit_steps.go              # NEW: add with before, update, remove (R9)

internal/adapters/d2/
├── emit.go                    # dispatch on style
└── emit_sequence.go           # NEW: sequence diagram (R7)

internal/adapters/html/theme/view.gohtml    # Steps table
internal/adapters/markdown/backend.go       # Steps table
internal/mcp/tools/                         # target and query kind enums, before field
cmd/query_cobra.go                          # flows kind

testdata/projects/dynamic-views/            # NEW (R11)
docs/language.md, docs/cli-reference.md, docs/mcp-integration.md
docs/adr/0018-dynamic-views.md
CHANGELOG.md
```

**Structure Decision**: the existing single-module layout. Each change lands in the layer the
constitution assigns it; no new package.

## Phases (for `/speckit-tasks`)

1. **Foundation**: language table (step, style), entities (StepDecl, IR Step, codes, address),
   authoring copy and parity. No behaviour change on existing fixtures.
2. **Story 1 (P1)**: decode, resolve, IR, projection with boundaries and numbered edges, Steps
   table on HTML and markdown, export; the `dynamic-views` fixture.
3. **Story 2 (P1)**: selection conflicts, matching, `step_mismatch`, `step_unmatched`; `move`
   rewrites steps.
4. **Story 3 (P2)**: `style`; the sequence emitter; wide-view hint skips it.
5. **Story 4 (P3)**: `step` edit target with `before`; `describe` steps; `query` `flows`; MCP
   end-to-end.
6. **Polish**: docs, ADR-0018, changelog, roadmap, 40-step budget check, quickstart walk.

Story 2 depends on Story 1's resolution; Stories 3 and 4 depend on Story 1.

## Complexity Tracking

No constitution violations to justify.
