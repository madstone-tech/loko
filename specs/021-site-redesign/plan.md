# Implementation Plan: Site Redesign

**Branch**: `021-site-redesign` | **Date**: 2026-10-28 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/021-site-redesign/spec.md`

## Summary

Rework the generated HTML site so that an element page tells the reader, without the diagram,
where the element sits, what it relates to and what it does. Then restyle the site with Madstone
branding.

- **Core.** The projection gains the derived data the pages need: ancestors, siblings with their
  relation, neighbours rolled up to one row each, children with technology and description,
  previous/next sibling, the "In context" view, source line, a navigation tree and overview counts.
  All of it is stdlib-only.
- **Rendering.** The HTML backend now receives the SVG backend's output through a staged render
  (R1). It inlines the parent's diagram on element pages and the view's diagram on view pages. It
  highlights the element by marking the node group that the SVG pass now tags with
  `data-loko-address` (R2).
- **Theme.** The built-in theme is rewritten: templates, one stylesheet, one small script, Latin
  WOFF2 fonts under 500 KB served as one generated `fonts.css` (so every output file still
  carries the generated-file notice, 014 FR-020), and the Madstone Eye.
- **Compatibility.** Every 014 block name keeps working (R10).

## Technical Context

**Language/Version**: Go 1.27 (toolchain pinned)

**Primary Dependencies**: existing only. These are stdlib `html/template` and `embed`, and the
in-process d2 v0.7.1 already used by `adapters/d2`. No new Go module. Fonts are subset at
development time with Python `fontTools` (`tools/fonts/subset.sh`, R6). The build never runs it,
and the subset WOFF2 files are committed.

**Storage**: files only, as today. The theme choice lives in the reader's browser
(`localStorage`).

**Testing**:
- `go test ./...` with `-race` and the goroutine-leak gate.
- Projection unit tests for each derived field: siblings, neighbour rollup, breadcrumb, prev/next,
  context view, prose inheritance and counts.
- HTML goldens regenerated and reviewed.
- A theme-compatibility test: every 014 block name still resolves, and the existing override
  fixtures still build.
- A contrast test that reads the theme tokens from `style.css` and checks every declared
  text/background pair at 4.5:1 or better.
- A font-budget test (≤ 500 KB).
- Tests for no network references, one inline SVG per page, byte-identical output across two runs,
  and the existing 1,020-element performance test.
- A manual browser pass per `quickstart.md`.

**Target Platform**: the single `loko` binary (macOS, Linux, Windows). The generated site targets
current evergreen browsers, and also works from `file://` and through `loko serve`.

**Project Type**: CLI and MCP server with a static-site and diagram renderer.

**Performance Goals**:
- The 1,020-element build stays under 10 s. Measured on 2026-10-28 at 2.8 s wall, 1,467 artefacts
  and 37 MB.
- Inlining adds about 24 KB per element page (median parent SVG 24 KB, max 38 KB). That is about
  +25 MB on the benchmark, and it costs only string copies.

**Constraints**:
- Byte-identical output.
- No network at view time.
- No third-party JS or CSS.
- The fonts total 500 KB or less.
- The markdown output, search and MCP/CLI outputs are unchanged.
- `loko serve` live reload keeps working.
- Every 014 block name keeps resolving.

**Scale/Scope**:
- 2 entity files (new) and 3 use-case files (2 new, 1 edited).
- The d2 SVG pass gets one attribute.
- The HTML adapter is rewritten: about 6 templates, the stylesheet, the script, 8 font files (served as
  one generated `fonts.css`) and 1 glyph.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle / rule | Status | Notes |
|---|---|---|
| I. Clean Architecture | Pass | Derived data (siblings, rollup, tree, counts) lives in `core/entities/viewmodel` and `core/usecases` and uses the stdlib only. Presentation (HTML, CSS, JS, fonts) stays in `adapters/html`. The SVG attribute lives in `adapters/d2`. Adapters never import each other: the SVG bytes reach the HTML backend through `RenderOptions` (R1). |
| II. Interface-First | Pass | No new port. `RenderOptions` gains `Upstream`, a map from artefact path to bytes for the formats a backend requires (R1). |
| III. Thin Handlers | N/A | No CLI, MCP or API handler changes. |
| IV. Entity Validation | Pass | No new user input. `ElementPage.Validate`-style invariants are added for the new fields (data-model §Invariants). |
| V. Test-First | Pass | Each derived field, the staged render, the SVG attribute, the contrast table and the font budget each start from a failing test. |
| VI. Token Efficiency | Pass | No change to MCP, CLI or export output. Markdown pages keep reading `Uses`/`UsedBy`, which are unchanged. |
| VII. Simplicity & YAGNI | Pass | One stylesheet, one script, no framework, no build step at `loko build` time. Highlighting is a class added to one existing group, not a second render. Pruning the tree is a filter over one projected tree. |
| File budgets | Pass | New entity files `page_context.go` and `nav.go` stay under 200 lines. `page.go` stays under 200. New use-case files `project_context.go` and `project_overview.go` stay under 200, with functions of 60 lines or fewer. `adapters/html/render.go` (229 lines) moves the inline and nav logic into `inline.go` and `nav.go`, each under 400. |
| 014 FR-012 (backends independent) | Amended | Today HTML already *requires* SVG (`Format.Requires`). This feature makes that a data dependency: backends with requirements render after the backends they require and receive those artefacts. Backends still never import each other. Recorded in ADR-0020 (R1). |
| 014 FR-022 (unrelated files unchanged) | Pass, with a note | A page now also depends on its parent, siblings, neighbours and the parent's diagram, which are all *related*. Their sources join the page's `Sources`. The pruned tree lists top-level names on every page, so renaming a top-level element rewrites every page. View titles in the nav already behave this way today. Recorded in ADR-0020 (R9). |
| ADR for new architectural decisions | Required | ADR-0020 covers: staged render and upstream artefacts (R1), the `data-loko-address` attribute (R2), the brand typography exception (R6, spec FR-030), the theme block compatibility policy (R10) and the widened page dependencies (R9). |

No violations need justifying, so Complexity Tracking is empty.

## Project Structure

### Documentation (this feature)

```text
specs/021-site-redesign/
├── plan.md              # This file
├── research.md          # Phase 0: decisions R1–R16
├── data-model.md        # Phase 1: projection additions
├── quickstart.md        # Phase 1: validation guide
├── contracts/
│   ├── site-output.md   # what the generated site guarantees (pages, assets, behaviour)
│   ├── theme-blocks.md  # amended 014 theme contract: block names, data, compatibility
│   └── svg-address.md   # additive amendment to 020's svg-output contract
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
internal/core/entities/viewmodel/
├── page.go               # ElementPage, LinkRef (+Technology, Description), Projection (+Nav, Overview)
├── page_context.go       # NEW: SiblingRow, NeighbourRow, SourceRef, Overview
└── nav.go                # NEW: NavNode tree and NavFor(address) pruning

internal/core/usecases/
├── project.go            # wires the new projection steps
├── project_pages.go      # ContextDiagram, ProseInherited, SourceRef
├── project_context.go    # NEW: siblings, neighbour rollup, ancestors, prev/next
├── project_overview.go   # NEW: nav tree, overview counts (diagnostics passed in)
├── build_artifacts.go    # staged renderAll: required formats first, Upstream to dependents
└── ports.go              # RenderOptions.Upstream

internal/adapters/d2/
└── svgpass.go            # data-loko-address on element and instance groups

internal/adapters/html/
├── render.go             # page jobs, per-page nav, view/element data
├── inline.go             # NEW: strip prolog, mark focus group, return template.HTML
├── fonts.go              # NEW: generated fonts.css (base64 WOFF2), Eye SVG and LICENSES.txt, each with the notice
├── icons.go              # NEW: generated icons.css (Lucide files as CSS masks)
├── page_data.go          # NEW: pageData and the per-page label, meta, crumbs and inline SVG
├── nav.go                # NEW: per-page pruned tree and breadcrumb data
├── theme.go              # assetNames (+fonts, brand), overridable (+rail.gohtml)
├── assets.go             # //go:embed theme (now with subdirectories)
└── theme/
    ├── layout.gohtml     # layout, head (+theme pre-paint), header, nav, footer, breadcrumb, pager
    ├── element.gohtml    # element, element-tables (→ Talks to), relation-table, element-children, prose-missing
    ├── rail.gohtml       # NEW: rail, part-of, siblings, in-context
    ├── index.gohtml      # index, stat-cards, row-list
    ├── view.gohtml       # view, in-this-view
    ├── partials.gohtml   # tag-chips, link, diagram (inline when available), icon-tile, chip
    ├── style.css         # tokens, light/dark, layout, print, narrow
    ├── custom.css        # unchanged (empty)
    ├── site.js           # search, theme toggle, tree, keys, menu, pan/zoom, blink stop
    ├── fonts/            # NEW: 8 Latin-subset WOFF2 files + OFL.txt per family (emitted as assets/fonts.css)
    └── brand/            # NEW: madstone-eye.svg + LICENSE note

tools/fonts/subset.sh     # NEW: dev-time subsetting with fontTools (not run by builds)
docs/adr/0020-site-redesign.md   # NEW
CHANGELOG.md              # block meaning changes, staged render, asset weight
```

**Structure Decision**: the existing hexagonal layout. Derived data goes into the core projection.
Presentation goes into `adapters/html`. The only cross-adapter need, SVG bytes, flows through the
use case (R1).

## Complexity Tracking

None.
