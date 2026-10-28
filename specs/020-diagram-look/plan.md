# Implementation Plan: Diagram Look

**Branch**: `020-diagram-look` | **Date**: 2026-10-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/020-diagram-look/spec.md`

## Summary

Every element in a rendered SVG becomes a rounded card (`border-radius: 6`, R13) in a Signal Blue
palette with a Lucide icon in its top-left corner, and its label reads as a hierarchy: bold name,
small kind line (`CONTAINER · Go, gRPC`), description in the regular face. The D2 emitter draws
every element as a rectangle, keeps a declared shape as a `shape-<shape>` class (R12), writes the
new kind line and neutral edge colours, and returns a *label plan* (each node's icon and each label
line's role). After d2 renders, a single pass over the SVG adds one `<defs>` of icon symbols, one
`<use>` per card, and an inline style per label line (R1, R2). A lighter line switches font family,
not weight, because d2 embeds each style as its own family (R3).
Icons are not drawn with d2's `icon:` because d2 pads any node with an icon by the label's height
on both axes, which made the spike's cards oversized and would break the 10% size limit (R1).
Colours change in the core palette table; contrast is tested for every line (R5).

## Technical Context

**Language/Version**: Go 1.27 (toolchain pinned)

**Primary Dependencies**: existing only: in-process d2 v0.7.1 (layout, `d2svg`), `embed` for the
vendored Lucide files (ISC). No new module.

**Storage**: files only, as today.

**Testing**: `go test ./...` with `-race` and the goroutine-leak gate. D2 goldens and projection
goldens regenerated and reviewed; new unit tests for the icon table (completeness), the label plan
(line roles for untitled, titled, wrapped-title, no-technology, no-description and edge cases), the
SVG pass (defs, one `<use>` per card, styles and font family per role, regular face embedded,
nothing on boundaries), contrast and minimum size per kind and role, and byte-identical output
across two runs. Existing size and performance tests cover SC-005.

**Target Platform**: the single `loko` binary (macOS, Linux, Windows).

**Project Type**: CLI and MCP server with a static-site and diagram renderer.

**Performance Goals**: 1,020-element build under 10 s (unchanged, SC-005); the SVG pass is a
linear scan, milliseconds per view.

**Constraints**: byte-identical output; self-contained SVG (no external references, no
`foreignObject`, no script); no view grows more than 10% (SC-004); SVG size grows at most 20%
(SC-005); `kind-*`/`tag-*` classes unchanged.

**Scale/Scope**: one adapter (`internal/adapters/d2`) and one core table
(`viewmodel/style.go`); about 10 vendored icon files.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle / rule | Status | Notes |
|---|---|---|
| I. Clean Architecture | Pass | Colours stay in `core/entities/viewmodel` (stdlib only). Icons, the label plan and the SVG pass live in `adapters/d2`, which alone imports d2. |
| II. Interface-First | Pass | No new port; the SVG backend still implements `Backend`. The label plan is an adapter-internal value between the emitter and the SVG backend, both in `adapters/d2`. |
| III. Thin Handlers | N/A | No CLI, MCP or API handler changes. |
| IV. Entity Validation | Pass | No new user input. |
| V. Test-First | Pass | Each change starts from a failing test (goldens, plan, pass, contrast, completeness). |
| VI. Token Efficiency | Pass | No change to MCP or export output. The `.d2` files lose the `[Kind: Tech] · id` brackets (the same words stay) and draw every element as a rounded `rectangle`. The declared shape stays in the source as a `shape-<shape>` class (R12), so no information is removed and the "previous shape on request" rule does not apply. The glyph change is recorded in ADR-0019 and the changelog. |
| VII. Simplicity & YAGNI | Pass | One byte pass, no XML parser, no customisation hooks beyond keeping three tables in one place each (FR-011). |
| File budgets | Pass | `emit.go` is 235 effective lines; label-line building moves to a new `emit_label.go`. New files `icons.go`, `labelstyle.go`, `svgpass.go` are each well under 400. `style.go` stays under 200. |
| 014 FR-008 (backends decide nothing) | Pass | The icon table translates `Style.Shape` and the kind, which the projection decided, into the SVG backend's glyph (R6). The `viewmodel.Style` comment and ADR-0019 record this. |
| Dependency direction / imports | Pass | `embed` is stdlib; no `os/exec`. |
| ADR for new architectural decisions | Required | ADR-0019: SVG post-render pass and icon injection (R1, R2), font families for line weights (R3), the shape class in `.d2` (R12), and the palette move to the Madstone Signal ramp. |

No violations; Complexity Tracking is empty.

## Project Structure

### Documentation (this feature)

```text
specs/020-diagram-look/
├── spec.md
├── plan.md              # this file
├── research.md          # R1–R13
├── data-model.md        # palette, icon table, line roles, label plan
├── quickstart.md        # validation guide
├── contracts/
│   └── svg-output.md    # what a rendered SVG guarantees
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/entities/viewmodel/
└── style.go               # palette values (R5), boundary neutrals, EdgePalette

internal/adapters/d2/
├── emit.go                # every element a rounded rectangle with a shape class; edge colours from EdgePalette; returns the plan
├── emit_label.go          # NEW: node and edge label lines + their roles (kind line text, R8)
├── emit_sequence.go       # participants through the same label/plan path (R9)
├── icons.go               # NEW: icon table, embedded Lucide files → <symbol>s (R6, R7)
├── icons/                 # NEW: 10 Lucide SVGs + LICENSE (ISC), unmodified
├── labelstyle.go          # NEW: line role → inline style, opacity, size (R3, R4)
├── svgpass.go             # NEW: inject <defs>, <use> per card, tspan styles (R1, R2)
├── svg_backend.go         # call the pass; ensure the regular font face (R3); cache key = source + plan (R10)
├── *_test.go              # NEW: emit_label, icons, svgpass, svg_selfcontained, svg_colours, contrast
└── testdata/golden/*.d2   # regenerated

testdata/projects/*/expected/projection.json   # regenerated (palette values)
.goreleaser.yaml           # ships icons/LICENSE as LICENSE-lucide (FR-004)

docs/
├── adr/0019-diagram-look.md
├── language.md            # shapes now shown as icons on cards
CHANGELOG.md               # repo root, under [Unreleased] alongside 019
```

**Structure Decision**: single Go module, existing layout. The change is confined to the d2
adapter plus the core palette table, as the spec's constraints require.

## Phase 0: Research

Complete: [research.md](research.md). No open questions remain. Key outcomes:
- R1: icons injected after rendering (d2 `icon:` inflates every card); FR-016 amended so `.d2`
  files carry no icons.
- R2–R4: line roles from an emitter-side plan; bold only where d2 measured bold; a lighter line
  switches to d2's regular font family, never `font-weight`; sizes and opacities per role, none
  below 11px.
- R5: container fill `#2D66AE`; every card line passes WCAG AA; edge tags at 0.72 opacity.
- R9: sequence participants use the same path; verified during implementation, with FR-018's
  fallback if the icon does not fit.
- R12: a declared shape stays in `.d2` as a `shape-<shape>` class (Principle VI).
- R13: cards are rounded with `border-radius: 6`.

## Phase 1: Design

- [data-model.md](data-model.md): palette, icon table, line roles, label plan.
- [contracts/svg-output.md](contracts/svg-output.md): guarantees of the SVG and `.d2` output,
  including the selectors 021 relies on.
- [quickstart.md](quickstart.md): tests, before/after on the example, size guard, offline and
  rasteriser checks, sequence and deployment, d2 CLI, budgets.

**Post-design Constitution Check**: unchanged, all pass; ADR-0019 is a deliverable.

## Implementation order (for /speckit-tasks)

tasks.md follows the spec's story priorities, so the goldens are rewritten twice (after US1 and
after US3). This costs one extra review and nothing else.

1. Label lines and roles in `emit_label.go`; icon table, vendored files, symbols, completeness
   test; emitter draws rounded rectangles with shape classes and returns the plan (US1).
2. SVG pass: defs, `<use>`, tspan styles, regular font face; cache key; byte-identity test (US1).
3. Sequence and deployment checks (FR-018, R9, R11) (US1).
4. Self-contained and rasteriser checks (US2).
5. Palette values and `EdgePalette` in `style.go`, with the contrast test (US3).
6. Size checks (US4, SC-004), then performance (SC-005), quickstart run, ADR-0019, docs, changelog.

## Complexity Tracking

None.
