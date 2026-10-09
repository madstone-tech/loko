# Implementation Plan: Rendering Fidelity

**Branch**: `016-rendering-fidelity` | **Date**: 2026-10-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/016-rendering-fidelity/spec.md`

## Summary

Generated diagrams should be as readable as hand-drawn ones. This feature adds five optional
attributes to the language:

- `title` on any element;
- `shape` on containers and externals;
- `kind` and `tags` on relationships;
- `direction` on views.

It carries them through the IR, the view model, the D2, site and markdown backends, the MCP
editing tools and `describe`. Container, system and declared views switch to a top-down layout by
default. A project that uses none of the attributes exports byte-identically, and renders
byte-identically apart from that direction line (research R8).

The approach turns on three decisions from [research.md](./research.md):
1. **Triggers are swapped during edge lifting**, before edges are keyed. Queries never see the
   swap, so their results are unchanged (R4).
2. **Async edges use a distinct dash pattern** (`stroke-dash: 8`). Crossing edges keep theirs, so
   existing diagrams do not change (R4).
3. **Titled nodes use plain D2 labels** (title, then `[Kind: Technology] · name`). Markdown labels
   were rejected during implementation because they render blank outside browsers. Untitled nodes
   keep their labels byte for byte (R2).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, toolchain go1.27.1)

**Primary Dependencies**: `oss.terrastruct.com/d2` v0.7.1 (in-process: shapes, markdown labels,
`direction`), `hcl/v2`. **No new module.**

**Storage**: Files only.

**Testing**: `go test`, golden files (projection, D2, site, markdown, export), the 015 round-trip
property test (extended), an aspect-ratio test on the rendered SVG `viewBox`, and the
goroutine-leak gate.

**Target Platform**: The single loko binary.

**Project Type**: Compiler + CLI + MCP server (one Go module, clean architecture).

**Performance Goals**: The 014 build budgets hold: the full build of 1,020 elements stays within
the dev-hardware budget, and rendering cost per node does not grow measurably for untitled nodes.

**Constraints**:
- Byte-identical output across runs and machines (014 FR-008).
- Unchanged projects stay byte-identical apart from the direction line (SC-003).
- Entities stay stdlib-only; D2 stays confined to `internal/adapters/d2`.
- The 015 constraints on the editor (FR-013/FR-014) apply to the new attributes.

**Scale/Scope**:
- 5 attributes and 3 diagnostic codes.
- About 12 files edited across arch, viewmodel, usecases, hclsource, d2, html, markdown and
  authoring.
- 1 fixture (`testdata/projects/serverless-reference`, added with the spec) and 1 ADR.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle / rule | Status | How |
|---|---|---|
| **I. Clean Architecture** | ✅ | The attributes are entity fields (stdlib). Validation, defaults and edge lifting live in use cases. D2 shape names, markdown labels and the `direction:` line live in the D2 adapter; titles in pages live in the html and markdown adapters. |
| **II. Interface-First** | ✅ | No new port. `Backend`, `ArchitectureSource` and `SourceEditor` are unchanged. |
| **III. Thin Handlers** | ✅ | No new handler. MCP and CLI pick up the attributes through existing use cases. |
| **IV. Entity Validation** | ✅ | Value sets live with their constants in `arch`; compile diagnostics come from core validation (R7). `authoring.NewEdit` checks them for edits. |
| **V. Test-First** | ✅ | Every phase opens with failing tests: diagnostics fixtures, lift tests for trigger and async, emitter goldens, the aspect-ratio test, and the byte-identity test. |
| **VI. Token Efficiency** | ✅ | New DTO fields are `omitempty`. `describe` structure stays within its 2,000-token budget, re-asserted by the 015 performance test. |
| **VII. Simplicity** | ✅ | No new module, block or port. Five shapes, not an icon set. One attribute each, no style language. |
| Dependency Direction | ✅ | `mcp` and `cmd` do not import entities; D2 is not imported outside adapters. |
| Quality gates | ✅ | lint, test, audit, race and goroutine-leak gate; ADR-0015 records the relationship-kind semantics and the direction default. |

**Language change governance**: five permanent v1.x attributes, specified in
[contracts/language.md](contracts/language.md) and documented in `docs/language.md`.

**Post-design re-check**: ✅ Passes. The design adds fields to existing types only.

## Project Structure

### Documentation (this feature)

```text
specs/016-rendering-fidelity/
├── spec.md, plan.md, research.md, data-model.md, quickstart.md
├── ux-findings-serverless.md     # the UX report that motivated this feature
├── contracts/
│   ├── language.md               # the five attributes and their diagnostics
│   └── rendering.md              # what each attribute does to diagrams, pages and markdown
└── tasks.md                      # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/core/entities/arch/
├── source_model.go     # EDIT: Title, Shape (ElementDecl); Kind, Tags (RelationDecl); Direction (ViewDecl)
├── ir_element.go       # EDIT: Element.Title/Shape, Relationship.Kind/Tags, View.Direction (omitempty)
├── render_attrs.go     # NEW: shape, kind, direction constants and membership helpers
└── diagnostic.go       # EDIT: invalid_attribute_value, shape_not_allowed, empty_title

internal/core/entities/viewmodel/
├── style.go            # EDIT: five shapes; StyleFor takes an element shape
├── model.go            # EDIT: Node.Title, Edge.Tags, EdgeStyle.Async, View.Direction
└── page.go             # EDIT: ElementPage.Title

internal/core/entities/authoring/
└── edit_schema.go      # EDIT: title, shape (container/external), kind, tags, direction

internal/core/usecases/
├── validate_render_attrs.go   # NEW: the three diagnostics (R7), called from CompileArchitecture
├── build_ir.go                # EDIT: carry the new fields
├── project_lift.go            # EDIT: swap triggers before keying; async aggregation; tag union
├── project_view.go            # EDIT: titles and shapes on nodes; direction defaults (R5)
├── project_pages.go           # EDIT: titles on pages
└── describe.go                # EDIT: new DTO fields

internal/adapters/hclsource/
├── schema.go                  # EDIT: attribute names
└── decode_logical.go, decode_project.go   # EDIT: decode the new attributes

internal/adapters/d2/emit.go   # EDIT: shapes, markdown labels for titles, async dash, tag line, direction
internal/adapters/html/        # EDIT: page headings, view legend (async vs crossing)
internal/adapters/markdown/    # EDIT: headings

cmd/
├── render_fidelity_test.go    # NEW: aspect ratio (FR-006), byte identity (SC-003)
docs/language.md, docs/mcp-integration.md, CHANGELOG.md, docs/adr/0015-rendering-attributes.md
```

**Structure Decision**: Extend the existing layout in place. No new package; one new use-case
file for the new diagnostics, keeping `validate_structure.go` within its budget.

## Complexity Tracking

No constitution violations. The table is intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|---|---|---|
| — | — | — |
