---

description: "Task list for feature 021: Site Redesign"
---

# Tasks: Site Redesign

**Input**: Design documents from `specs/021-site-redesign/`

**Prerequisites**: plan.md, spec.md (with Clarifications), research.md (R1–R16), data-model.md,
contracts/site-output.md, contracts/theme-blocks.md, contracts/svg-address.md, quickstart.md

**Tests**: Included. Constitution Principle V (Test-First) applies, and plan.md lists the required
tests. Write each test task first and confirm it fails before you do the matching implementation
task.

**Organization**: Tasks are grouped by user story. Each story is an increment you can test on its
own.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: the user story the task serves (US1–US5)
- All paths are relative to the repository root

## Conventions for every task

- **Language and modules.** Go 1.27, single module, no new module. Core changes use the stdlib
  only.
- **File budgets.** Effective-line budgets:
  - `internal/core/entities/**` and `internal/core/usecases/**` files: 200 or fewer.
  - Use-case functions: 60 or fewer.
  - `internal/adapters/**` files: 400 or fewer.
  - Check with `make audit-constitution`.
- **Goldens.** Rewrite them with each package's `-update` flag:
  - `internal/adapters/html/render_test.go`
  - `internal/adapters/d2/emit_test.go`
  - `internal/adapters/markdown/backend_test.go`
  - `cmd/projection_golden_test.go`
  - Review every golden diff before committing. The markdown goldens MUST NOT change (FR-043).
- **Determinism.** Output must be byte-identical across runs (FR-040). Never produce output by
  iterating a map; sort by address first.
- **Template blocks.** Every 014 block name MUST remain defined and called. See
  `contracts/theme-blocks.md`.
- **Accessible names.** Templates show names, using `or .Title .Name`, never a raw id. The one
  exception is the metadata line (FR-005).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: record the baseline, vendor the fonts and brand assets, and open the ADR.

- [X] T001 Record the baseline before you change code:
  - Build `go build -o /tmp/loko-before .` from `HEAD`.
  - For each project in `examples/` with `*.loko.hcl` files, run
    `/tmp/loko-before build --format html,md --out /tmp/before/<name>`.
  - Also run it on the 1,020-element benchmark generated as in `cmd/build_perf_test.go`
    (`writeLargeProject(20,10,4)`).
  - Save the wall time, total output bytes, and the byte counts of `assets/` and `*.html` to
    `/tmp/before/sizes.txt`.
  - T060 compares against this file. The markdown outputs saved here are T061's reference.
- [X] T002 [P] Write `tools/fonts/subset.sh` (bash, `set -euo pipefail`):
  - **Pinned sources.** Download these upstream OFL releases at pinned versions into a temp dir:
    Orbitron variable TTF (google/fonts `ofl/orbitron`), Source Sans 3 (adobe-fonts/source-sans
    release, `SourceSans3-Regular`, `-It`, `-Semibold` TTF) and JetBrains Mono variable TTF
    (JetBrains/JetBrainsMono release). Take Audiowide from `../mdstn-branding/fonts/Audiowide-Regular.ttf`.
  - **Instancing.** Instance the variable fonts with
    `python3 -m fontTools.varLib.instancer` at `wght=600` and `wght=700` for Orbitron, and at
    `wght=400` and `wght=600` for JetBrains Mono.
  - **Subsetting.** Subset each font with `python3 -m fontTools.subset --flavor=woff2
    --layout-features='kern,liga' --unicodes=U+0000-00FF,U+0131,U+0152-0153,U+02C6,U+02DA,U+02DC,U+2000-206F,U+20AC,U+2122,U+2190-2193,U+2212,U+25B8,U+25BE`.
  - **Output files.** Write to `internal/adapters/html/theme/fonts/`:
    `orbitron-600.woff2`, `orbitron-700.woff2`, `source-sans-3-400.woff2`,
    `source-sans-3-400-italic.woff2`, `source-sans-3-600.woff2`, `jetbrains-mono-400.woff2`,
    `jetbrains-mono-600.woff2`, `audiowide-400.woff2`.
  - **Licences.** Copy each family's licence unmodified to
    `internal/adapters/html/theme/fonts/OFL-<family>.txt`.
  - **Final report.** Print each file's size and the total.
  - Research R6. That is 8 files: the 7 in the plan plus the italic.
- [X] T003 Run `tools/fonts/subset.sh` and commit the WOFF2 files and licences under
  `internal/adapters/html/theme/fonts/`. Confirm the printed total is 500 KB or less. If it is
  not, drop `orbitron-700` and `source-sans-3-400-italic` first. Record the sizes for the release
  notes (T063).
- [X] T004 [P] Copy `../mdstn-branding/assets/madstone-glyph.svg` unmodified to
  `internal/adapters/html/theme/brand/madstone-eye.svg`. Add
  `internal/adapters/html/theme/brand/NOTICE.txt` naming the source file in `mdstn-branding` and
  stating that the glyph is a Madstone brand mark, not covered by loko's code licence.
- [X] T005 [P] In `.goreleaser.yaml` under `archives[].files`, add an entry per font licence:
  `src: internal/adapters/html/theme/fonts/OFL-*.txt` with `dst: licenses/`. Model it on the
  Lucide entry 020 added. If `goreleaser` is installed, run `goreleaser check`.
- [X] T006 [P] Create `docs/adr/0020-site-redesign.md` from `docs/adr/template.md`, with status
  Proposed. Its sections, completed in T062:
  - staged render and `RenderOptions.Upstream` (amends 014 FR-012, research R1)
  - the `data-loko-address` attribute (R2)
  - one inline SVG per page (R3)
  - the brand typography exception: Source Sans 3 prose vs the brand's monospace body (spec
    FR-030, R6)
  - the theme block compatibility policy (R10)
  - widened page dependencies, where a top-level rename rewrites every page (R9)
  - WCAG overrides kept in loko and not pushed to the brand tokens (FR-032)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: plumbing that every story needs. This covers the staged render, the SVG address hook,
the extended view-model types and the embedded assets.

**⚠️ CRITICAL**: no user story work can begin until this phase is complete.

### Tests (write first, confirm they fail)

- [X] T007 [P] In `internal/core/usecases/build_artifacts_test.go`, add
  `TestRenderAllStagesRequiredFormats`. Use two fake backends: `svg`, with no requirements, emits
  `diagrams/a.svg`; `html` requires `svg`. Assert that:
  - The html fake's `RenderOptions.Upstream` holds exactly `{"diagrams/a.svg": <bytes>}`.
  - A backend with no requirements gets `Upstream == nil`.
  - The combined artefact order is the same across 20 runs.
- [X] T008 [P] In `internal/adapters/d2/svgpass_test.go`, add `TestSVGPassAddressAttribute`
  (contracts/svg-address.md). Every element and instance group gets
  `data-loko-address="<address>"` directly after its `class` attribute. The attribute is
  XML-escaped. Boundary, outside and edge groups get none. Each address appears at most once.
- [X] T009 [P] In `internal/core/entities/viewmodel/nav_test.go`, add `TestNavFor`
  (data-model §NavNode). Build a 3-level tree. Assert that:
  - `NavFor(tree, "")` returns the top level only, with `HasChildren` kept.
  - `NavFor(tree, component)` keeps the top level, the children of each ancestor and the
    component's own children.
  - It marks the component `Current` and its ancestors `Open`.
  - It never mutates the input.
- [X] T010 [P] In `internal/adapters/html/assets_test.go`, add `TestFontBudget`, which fails if
  `theme/fonts/*.woff2` total more than 500 KB and logs the total in KB, plus the size of the
  generated `assets/fonts.css` (FR-031, FR-044, SC-011). Add `TestAssetsPublished`, which checks
  that the build emits:
  - `assets/fonts.css`: one `@font-face` per embedded WOFF2, each `src:
    url(data:font/woff2;base64,…) format("woff2")`, in file-name order.
  - `assets/brand/madstone-eye.svg`: the glyph with the notice comment as its first line.
  - `assets/LICENSES.txt`: every `OFL-*.txt`, the Lucide ISC licence and the brand notice,
    concatenated in name order under headings.

  It also checks that no `.woff2` or licence file is emitted on its own. `cmd/build_notice_test.go`
  must keep passing **unchanged**, so every emitted file opens with the notice (014 FR-020,
  SC-004).

### Implementation

- [X] T011 Add `Upstream map[string][]byte` to `RenderOptions` in
  `internal/core/usecases/ports.go`, with a doc comment: "artefacts of the formats this backend
  requires, keyed by path; nil when it requires none; read-only".
- [X] T012 Change `renderAll` in `internal/core/usecases/build_artifacts.go` to run in two stages:
  - **Stage 1:** the backends whose `Format().Requires()` is empty, run concurrently.
  - **Upstream map:** build it from stage 1's artefacts whose `Format` is in each stage-2
    backend's `Requires()`.
  - **Stage 2:** the remaining backends, run concurrently, each with its own `opts.Upstream`.
  - **Order:** keep the index-ordered collection, so the output order is stage 1 then stage 2,
    each in registry order.
  - **Budget:** keep the function at 60 effective lines or fewer; split out `upstreamFor` if
    needed.
  - Makes T007 pass.
- [X] T013 [P] In `internal/adapters/d2/svgpass.go`, write ` data-loko-address="<escaped address>"`
  after the `class` attribute of each planned element or instance group. Carry the address on the
  label plan's node (`emit_label.go`, `LabelPlan` node gains `Address string`, set from
  `vm.Node.Address` in `emit.go`). Run the d2 goldens with `-update`, and check that the only
  difference is the new attribute. Makes T008 pass.
- [X] T014 [P] Extend `LinkRef` in `internal/core/entities/viewmodel/page.go` with
  `Technology string` and `Description string` (both `json:",omitempty"`).
- [X] T015 [P] Create `internal/core/entities/viewmodel/page_context.go` with these types from
  data-model.md:
  - `SiblingRow{Ref LinkRef; Current, Calls, CalledBy bool}`
  - `NeighbourRow{Other LinkRef; Direction string /* "out", "in" or "both" */; Technology, Description string; Tags []string; Relationships int; Sibling bool}`
  - `SourceRef{File string; Line int}`
  - `Overview{Systems, Externals, Containers, DataStores, Relationships, Async, Warnings, Infos int}`
- [X] T016 Add these fields to `ElementPage` in `internal/core/entities/viewmodel/page.go`, all
  `omitempty` (depends on T015):
  - `Ancestors []LinkRef`
  - `Siblings []SiblingRow`
  - `Neighbours []NeighbourRow`
  - `Prev, Next *LinkRef`
  - `ContextDiagram ViewID`
  - `ProseInherited bool`
  - `Source SourceRef`

  Add `Nav []NavNode` and `Overview Overview` to `Projection`. Add `ElementCount` and
  `RelationshipCount int` to `ViewModel` in `model.go`. Keep each file at 200 effective lines or
  fewer, moving types into `page_context.go` if needed.
- [X] T017 [P] Create `internal/core/entities/viewmodel/nav.go`:
  - `NavNode{Ref LinkRef; Children []NavNode; HasChildren, Current, Open bool}`.
  - `NavFor(nav []NavNode, address string) []NavNode`, which returns a pruned deep copy per
    data-model §NavNode.
  - Makes T009 pass.
- [X] T018 Change `internal/adapters/html/assets.go` to `//go:embed theme` so the `fonts/`,
  `brand/` and `icons/` subdirectories are embedded. Create `internal/adapters/html/fonts.go`
  with three generated artefacts, each opening with the generated-file notice like every other
  file (014 FR-020):
  - **`assets/fonts.css`**: the notice as a CSS comment, then one `@font-face` per
    `theme/fonts/*.woff2` in file-name order.
    - Each rule sets `font-family`, `font-weight`, `font-style` and `font-display: swap`. The
      family, weight and style come from a fixed table keyed by file name.
    - Each `src` is `url(data:font/woff2;base64,<std base64>) format("woff2")`.
    - Base64 output is deterministic, so the file is byte-identical across runs.
  - **`assets/brand/madstone-eye.svg`**: the notice as an XML comment (`<!-- … -->`) on line 1,
    then the glyph bytes unmodified.
  - **`assets/LICENSES.txt`**: the notice on line 1, then each `fonts/OFL-*.txt`, the copied
    Lucide `icons/LICENSE` and `brand/NOTICE.txt` in name order. Each comes under a `== <path> ==`
    heading, with its text unmodified.

  No `.woff2` or licence file is emitted on its own. `fonts.css` is not overridable; it is not in
  `overridable`. Keep `assetNames` for `style.css`, `custom.css` and `site.js`. Makes T010 pass.

**Checkpoint**: `go test ./...` passes with only the d2 golden updates. The HTML and markdown
outputs are unchanged at this point.

---

## Phase 3: User Story 1 — Understand an element without reading the diagram (Priority: P1) 🎯 MVP

**Goal**: element pages show the header, prose-or-context main column, the ordered rail (Part of,
Siblings, Talks to, In context) and the children rows. The In context view inlines the parent's
diagram with this element highlighted.

**Independent Test**: build `examples/microservices` and open a gateway component page. Check
everything in quickstart §2.

### Tests for User Story 1 (write first, confirm they fail)

- [X] T019 [P] [US1] Create `internal/core/usecases/project_context_test.go` with table tests on a
  small IR. It has a system with 2 containers; the container `api` has 3 components; one external
  system; one person. Relationships:
  - `router→auth` twice, with different technologies.
  - `auth→router` once.
  - `router→db`.
  - `person→api`.

  Assert:
  - **`Siblings`** (invariant 1): `router`'s siblings are its parent's children in address order,
    including itself with `Current`. `auth` has `Calls` and `CalledBy` true.
  - **`Neighbours`** (invariants 2, 3):
    - One row for `auth` with `Direction:"both"` and `Relationships:3`.
    - `Technology` is the sorted, de-duplicated join with `", "`.
    - `Description` is the first non-empty one in relationship-address order.
    - `Sibling` is true.
    - No row for the element itself.
  - **`Ancestors`**: `[system, api]` for `router`, and empty for the system.
  - **`Prev`/`Next`** (invariant 4): consistent across all siblings, and nil at the ends.
  - **Top-level elements**: siblings are the other top-level elements of the same kind (person
    with persons, system with systems).
  - **`Sources`**: grows to include the parent, siblings, neighbours and the context view's
    sources.
- [X] T020 [P] [US1] In `internal/core/usecases/project_pages_test.go`, add:
  - `TestContextDiagram` (research R8; invariant 5):
    - It is the parent's own view when produced.
    - Otherwise it is the first produced view whose nodes include the element.
    - It is the landscape for top-level elements.
    - Otherwise it is empty.
    - A container's `ContextDiagram` is its system's view, while its `Diagram` stays its own
      view.
  - `TestProseInherited` (FR-009, invariant 6): a component whose `docs` equals its parent's
    `docs` gets `ProseInherited:true` and an empty `Prose`, and `Parent` is set. A component with
    its own docs is unaffected.
  - `TestElementSource`: `Source` holds the declaring file and its 1-based line from
    `Provenance.RangeOf`.
  - `TestChildrenCarryDetail`: `Children` and `Parent` `LinkRef`s carry `Technology` and
    `Description`.
- [X] T021 [P] [US1] In `internal/adapters/html/inline_test.go`, add `TestInlineSVG`:
  - The XML declaration and the generated-file comment are stripped.
  - `role="img"` and `aria-labelledby="<caption id>"` are added to the root `<svg>`.
  - When a focus address is given, exactly one `g[data-loko-address="<address>"]` gains the class
    `loko-focus`, appended to its existing classes.
  - An unknown address leaves the SVG unchanged apart from the root attributes.
  - The result is `template.HTML`.
- [X] T022 [P] [US1] In `internal/adapters/html/element_page_test.go`, render the T019 fixture
  through `Backend.Render` with an `Upstream` holding a fake parent SVG. Assert the markup per
  contracts/site-output.md §Element page:
  - The header has `p.kind-label` = `COMPONENT · IN <PARENT TITLE OR NAME>`, `span.icon-tile.kind-component`
    in the `h1`, and `p.meta` containing the technology, the full address and `file:line`.
  - `aside.rail` holds the sections `part-of`, `siblings`, `talks-to` and `in-context`, in that
    order, with empty ones omitted.
  - Talks to has one row per neighbour, with a `↔` arrow for both directions and a `sibling`
    marker.
  - Exactly one inline `<svg` appears on the page, and exactly one `loko-focus`.
  - With no prose, `main` contains the In context figure and the rail has no `in-context`.
  - With `ProseInherited`, `main` links to the parent page.
  - A container page has `section.children` rows and an `<img>` of its own view.

### Implementation for User Story 1

- [X] T023 [US1] Create `internal/core/usecases/project_context.go` with:
  - `ancestorsOf`
  - `siblingsOf`, with the top-level rule: same kind among elements with no parent
  - `neighboursOf`, which rolls up `ir.OutgoingFrom` and `ir.IncomingTo` by other address, sorted
  - `prevNext`
  - `contextSources`

  Each takes `*arch.IR` and returns view-model values sorted by address. Each function is 60
  effective lines or fewer, and the file is 200 or fewer. Makes T019 pass.
- [X] T024 [US1] In `internal/core/usecases/project_pages.go`, keep `elementPage` at 60
  effective lines or fewer (constitution Principle III). It calls one
  `fillContext(&p, ir, e, prov, views)` helper defined in `project_context.go`, which does the
  following:
  - Fill `Ancestors`, `Siblings`, `Neighbours`, `Prev`/`Next` and `Source` (from `prov.RangeOf`).
  - Fill `ContextDiagram` through a new `contextDiagramFor(ir, e, views)`. It needs the produced
    view models' node addresses, so pass the resolved views in.
  - Set `ProseInherited` when the parent's `Docs` equals `e.Docs`. Keep `Prose` so the markdown
    output is unchanged (FR-043); the HTML template ignores it when `ProseInherited` is set.
  - Fill `Technology` and `Description` on `linkTo` for parent and children, but **not** on
    `RelationRow.Other`, so the markdown output is unchanged.
  - Widen `p.Sources` with `contextSources`.

  Makes T020 pass. Run `cmd/projection_golden_test.go -update` and review.
- [X] T025 [US1] Run `go test ./internal/adapters/markdown/...` **without** `-update` and confirm
  the markdown goldens are unchanged (FR-043). If they changed, fix the projection rather than the
  goldens.
- [X] T026 [P] [US1] Create `internal/adapters/html/inline.go` with
  `inlineSVG(svg []byte, captionID, focus string) template.HTML`, using byte operations only, as
  in `adapters/d2/svgpass.go`. Makes T021 pass.
- [X] T027 [US1] In `internal/adapters/html/render.go`:
  - Add the `pageData` fields `Inline template.HTML`, `Label string`, `Meta []string` and
    `Crumbs []vm.LinkRef` (contracts/theme-blocks.md §Page data).
  - In `pageJobs`, receive `opts.Upstream`. For element pages, set `Inline` from
    `Upstream[view(ContextDiagram).DiagramPath]` with focus = page address.
  - Compute `Label` (upper-case kind · `IN <PARENT>`) and `Meta`.
  - Move helpers into a new `internal/adapters/html/page_data.go` if `render.go` would exceed
    400 effective lines.
- [X] T028 [P] [US1] Create `internal/adapters/html/theme/rail.gohtml` with blocks `rail`,
  `part-of`, `siblings` and `in-context`. Add `rail.gohtml` to `overridable` in
  `internal/adapters/html/theme.go`.
  - **`part-of`:** the parent's icon tile, name link, kind, technology chip and description. The
    system above it is shown when the parent is a container.
  - **`siblings`:** one row each. The current one is marked `aria-current="true"` with class
    `current`. Related siblings get a `calls`/`called by` chip.
  - **`in-context`:** a `figure` with a `figcaption` ("<Parent> — in context") and `{{.Inline}}`.
- [X] T029 [US1] Rewrite `internal/adapters/html/theme/element.gohtml`:
  - **`element`:** the main column holds the prose plus a source note ("Edit `<ProsePath>`"). With
    no prose or `ProseInherited`, it holds `prose-inherited` (a link to the parent) and the In
    context figure instead. Then `{{template "rail" .}}`, then `element-children`.
  - **`element-tables`:** now renders the rail's Talks to section from `.Element.Neighbours`.
  - **`relation-table`:** receives `rowsData $.Root .Neighbours`, one row per `NeighbourRow`,
    showing:
    - an arrow: `→` for out, `←` for in, `↔` for both
    - a link to `Other`
    - a technology chip
    - the description
    - a `sibling` chip
    - the tags via `hashTags`
  - **`element-children`:** `section.children` rows (icon tile, name, kind, technology, description)
    plus the element's own view as an `<img>` figure with "Open view" when `.Diagram` is set.
  - **Truncated names:** any name or id that CSS may truncate carries a `title` attribute with
    the full text. This covers the rail, children and Talks to rows.
  - Keep `prose-missing`, and add `prose-inherited`.
  - Update `rowsData` and `anyTags` in `theme.go` to accept `[]vm.NeighbourRow`. Field names stay
    compatible (R10).
- [X] T030 [US1] Rewrite the `header` block in `internal/adapters/html/theme/layout.gohtml`. It
  holds:
  - `div.accent-bar`
  - `p.kind-label` (`{{.Label}}`)
  - `h1`, containing `{{template "icon-tile" .}}` and the title
  - `p.meta` (`{{range .Meta}}`, separated by ` · `)
  - `p.lead` (description)

  The subtitle stays in `p.meta` as the full id.

  Add `icon-tile` and `chip` blocks to `internal/adapters/html/theme/partials.gohtml`. The icon
  tile reuses 020's Lucide symbols. Inline the matching icon SVG from
  `internal/adapters/d2/icons/` into an embedded copy under `internal/adapters/html/theme/icons/`.
  Copy the files; adapters must not import each other. Include the ISC licence alongside.

  Add `TestIconCopiesMatch` to `internal/adapters/html/icons_test.go` so the copies cannot drift
  from 020's single icon table. It reads both directories from the repository root and asserts
  that every `internal/adapters/html/theme/icons/*.svg` has a byte-identical twin in
  `internal/adapters/d2/icons/`, and that the two name sets are equal.
- [X] T031 [US1] Add element-page styles to `internal/adapters/html/theme/style.css`:
  - the two-column grid (main + 320px rail)
  - rail cards and row lists
  - the icon tile in the kind colour, using the 020 kind palette values copied as
    `--kind-<kind>` tokens
  - chips
  - `.in-context svg g[data-loko-address]:not(.loko-focus)` and edge groups at `opacity:.3`
  - `.loko-focus` outlined with a 3px Cipher Green stroke on its `rect`/`path`
  - the diagram "paper" panel (`#FFFFFF` with a 1px border)
  - Run `render_test.go -update`, review the element goldens, and make T022 pass.

**Checkpoint**: US1 is complete. Element pages answer "parent, siblings, neighbours, function"
without the diagram (SC-001, SC-003).

---

## Phase 4: User Story 2 — Browse the architecture hierarchy (Priority: P1)

**Goal**: a pruned tree on every page, a breadcrumb up to the project, a sibling pager and arrow
keys.

**Independent Test**: quickstart §3.

### Tests for User Story 2 (write first, confirm they fail)

- [X] T032 [P] [US2] In `internal/core/usecases/project_overview_test.go`, add `TestNavTree`:
  - `Projection.Nav` lists people, then systems (each with containers, each with components), then
    externals.
  - Each group is sorted by address.
  - `Ref.Technology` and `Ref.Description` are empty.
  - `HasChildren` is set wherever children exist.
- [X] T033 [P] [US2] In `internal/adapters/html/nav_test.go`, add:
  - `TestPrunedNavPerPage`: render the T019 fixture.
    - A component page's `nav.tree` contains the top level, the system's containers and the
      container's components, with the current `li` marked `aria-current="page"` and class
      `active`.
    - It contains no other system's children.
    - The index page's tree contains the top level only.
    - Every tree node renders its name, not its id.
  - `TestBreadcrumb`: `nav[aria-label="Breadcrumb"]` reads
    `project › system › container › component`, and the last item has `aria-current="page"`.
  - `TestPager`: `a[rel=prev]` and `a[rel=next]` are present exactly when `Prev`/`Next` are set.
  - `TestNavOverrideStillWorks`: a 014-style `{{define "nav"}}` override that ranges over
    `.Projection.Views` still renders.

### Implementation for User Story 2

- [X] T034 [US2] Create `internal/core/usecases/project_overview.go` with `navTree(ir)`, which
  builds `Projection.Nav` per data-model §NavNode. Wire it from `Project` in
  `internal/core/usecases/project.go`. Makes T032 pass.
- [X] T035 [US2] Create `internal/adapters/html/nav.go` and replace `renderNavs` in
  `internal/adapters/html/render.go` with per-page rendering:
  - Each page gets `Tree = vm.NavFor(in.Nav, address)`, where `address` is empty for the index and
    views.
  - Each page gets `Crumbs`:
    - project link, then `Ancestors`, then the element itself
    - for views: project, then the subject element's ancestors and the subject
  - Render the `nav` block per page from the theme. Keep `.Projection` and `.Root` in its data for
    old overrides.
  - Rendering stays inside the bounded worker pool in `renderPages`.
- [X] T036 [US2] Rewrite the `nav` block in `internal/adapters/html/theme/layout.gohtml`:
  - **Brand header:** the Eye
    (`<img src="{{.Root}}assets/brand/madstone-eye.svg" alt="" width="28" height="28">`), then the
    `LOKO` wordmark with a `span.cursor`.
  - **Project:** the project name link and description.
  - **Search:** the existing `#search` input and `#search-results`, both unchanged.
  - **Tree:** a recursive `tree-node` template over `.Tree`, inside `nav.tree` with
    `aria-label="Architecture"`. Each node shows:
    - a `button.twisty` (`▸`/`▾`, `aria-expanded`) when `HasChildren`
    - the 020 icon
    - the name link, with a `title` holding the full name when it may be truncated
  - **Views list:** names and view kind.
  - **Footer note:** "Generated from `*.loko.hcl`. Edit the source, not these pages."

  Add `breadcrumb`, `pager` and `theme-toggle` blocks to the same file. Call `breadcrumb` before
  `header`, and `pager` after the main content. `theme-toggle` is a `button#theme-toggle` with
  `aria-pressed` and `hidden`; the script reveals it.
- [X] T037 [US2] Rewrite `internal/adapters/html/theme/site.js` as one IIFE. Keep the existing
  search logic verbatim. Add:
  - Tree twisty toggling (`aria-expanded`, `hidden` on the child `ul`).
  - `keydown` on `document`: `ArrowLeft`/`ArrowRight` follow `a[rel=prev]`/`a[rel=next]`. Skip
    when `e.target` is an `input`, `textarea`, `select` or `[contenteditable]`, and skip when any
    modifier key is held.
  - Menu button toggling `body.nav-open`.

  No library and no network. Theme and pan/zoom are added in T046 and T041.
- [X] T038 [US2] Add sidebar, tree, breadcrumb and pager styles to
  `internal/adapters/html/theme/style.css`:
  - The sidebar is solid `#1E1E1E` in both themes, with a dot grid
    (`radial-gradient` at 16px, 6% white).
  - The active tree item has a 3px Cipher Green left rail and a 10% green tint.
  - The tree and labels use JetBrains Mono.

  Run `render_test.go -update`, review, and make T033 pass.

**Checkpoint**: US1 and US2 work together. Every element is reachable through the tree
(SC-002).

---

## Phase 5: User Story 3 — Read the overview and view pages (Priority: P2)

**Goal**: an index with stat cards, the landscape figure and row lists; view pages with header,
inline zoomable diagram and the "In this view" list.

**Independent Test**: quickstart §4.

### Tests for User Story 3 (write first, confirm they fail)

- [X] T039 [P] [US3] Add `TestOverview` to `internal/core/usecases/project_overview_test.go`. It
  counts:
  - `Systems` and `Externals`
  - `Containers`, and `DataStores` for shape `database`, `queue` or `topic`
  - `Relationships`, and `Async` for kind `async`
  - zero counts as 0

  Add `TestViewCounts`: `ElementCount` excludes boundary and outside nodes, and
  `RelationshipCount` sums the relationships of non-crossing edges. In
  `internal/core/usecases/build_artifacts_test.go`, add `TestOverviewDiagnostics`:
  `Overview.Warnings` and `Overview.Infos` count the diagnostics left after
  `project.ignore_warnings`.
- [X] T040 [P] [US3] In `internal/adapters/html/index_view_test.go`, check the index and view
  pages:
  - **Index:** `section.stats` has 4 cards with the expected numbers. The landscape figure has an
    "Open view" link. `section.systems` and `section.views` hold row lists with chips. There is no
    table of raw ids.
  - **View page:** the meta line has the counts, and there is exactly one inline `<svg` inside
    `figure.diagram` with zoom buttons and links to `.svg` and `.d2`.
  - **View labels:** a table test, one row per mode, checks each view label and the matching index
    view chip (data-model §View mode):
    - a system view → `SYSTEM VIEW · AUTOMATIC`
    - a declared view → `DECLARED VIEW · DECLARED`
    - a declared view with `style = "sequence"` → `DECLARED VIEW · SEQUENCE`
    - a deployment view → `DEPLOYMENT VIEW · DEPLOYMENT`
  - **In this view:** the `section.in-this-view` id matches the SVG's `aria-describedby`.
  - **Dynamic views:** keep `table.steps`.

### Implementation for User Story 3

- [X] T041 [US3] Add `func (v View) Mode() string` to
  `internal/core/entities/viewmodel/view.go`, following data-model §View mode:
  - `"sequence"` when `Style == "sequence"`
  - else `"declared"` when `Kind == "declared"`
  - else `"deployment"` when `Kind == "deployment"`
  - else `"automatic"`

  Unit-test it in `view_style_test.go`. Then compute the overview and view counts in two places:
  - **`project_overview.go`:** add `overview(ir)`, wired from `Project`, and fill
    `ElementCount`/`RelationshipCount` where view models are built in
    `internal/core/usecases/project_view.go`.
  - **`build_artifacts.go`:** set `proj.Overview.Warnings` and `proj.Overview.Infos` from the
    compiled diagnostics after ignore filtering, before `renderAll`.

  Makes T039 pass.
- [X] T042 [US3] Rewrite `internal/adapters/html/theme/index.gohtml`:
  - **`index`:** the description lead, then `stat-cards`, then the landscape via `diagram` (an
    `<img>` with caption and "Open view"), then `row-list` for systems (with container counts and
    an `external` chip) and for views (kind chip, plus automatic/declared/sequence/deployment).
  - **`stat-cards`:** numbers in Audiowide, with a sub-line ("2 external", "3 data stores",
    "4 async", "warnings · infos").
  - **`row-list`:** shared rows with icon tile, name, description and chips.

  Keep the environments list.
- [X] T043 [US3] Rewrite `internal/adapters/html/theme/view.gohtml`:
  - **`view`:** set the header data, then `diagram` with `.Inline` (the view's own SVG from
    `Upstream`, no focus) and the existing legend, then `in-this-view`, then the existing Nodes,
    Steps and Leaves-this-view sections, unchanged.
  - **`in-this-view`:** rows for element nodes with icon tile, name link, kind and technology.

  In `internal/adapters/html/theme/partials.gohtml`, extend `diagram`: when `.Inline` is set,
  render `figure.diagram` with:
  - a `div.diagram-panel` holding `{{.Inline}}`
  - a toolbar of `button[data-zoom=in|out|reset]` (`hidden`, revealed by the script)
  - a `figcaption` with links to the SVG and D2 source

  Otherwise keep the 014 `<img>` markup. In `render.go`, set view-page `Inline`, `Label` and
  `Meta`. `Label` is `<UPPER(Kind)> VIEW · <UPPER(View.Mode())>`. The index view-row chips use
  `View.Mode()` too (FR-018, FR-019).
- [X] T044 [US3] Add pan and zoom to `internal/adapters/html/theme/site.js` (research R4):
  - It acts on each `figure.diagram svg`, storing the original `viewBox`.
  - **Wheel:** zoom by 1.2× around the pointer, with `preventDefault`.
  - **Click:** zoom 2× at the point, unless the click is on an `a` element.
  - **Drag:** pointer events pan.
  - **Buttons:** reset, zoom in and zoom out.
  - **Keys:** `+`, `-` and `0` when the figure has focus (`tabindex="0"`).
  - **Limits:** clamp the zoom between 1× and 8×.
  - **Print:** `beforeprint` resets the zoom.
- [X] T045 [US3] Add the stat card, row list, figure, diagram panel and toolbar styles to
  `internal/adapters/html/theme/style.css`. The SVG uses `width:100%; height:auto` so it shows at
  its fitted size without scripts. Run `render_test.go -update`, review, and make T040 pass.

**Checkpoint**: the index and view pages are complete, and a wide view zooms to legible labels
(SC-009).

---

## Phase 6: User Story 4 — Madstone look in light or dark (Priority: P2)

**Goal**: brand tokens, light and dark themes following the system with a light fallback, a
remembered toggle, typography option B and bundled fonts.

**Independent Test**: quickstart §5.

### Tests for User Story 4 (write first, confirm they fail)

- [X] T046 [P] [US4] In `internal/adapters/html/contrast_test.go`, add `TestThemeContrast`
  (research R11):
  - Parse `--name: #RRGGBB` declarations from the `:root` block (light) and the
    `:root[data-theme="dark"]` block (dark) of `theme/style.css`.
  - For each pair in a table, assert a WCAG contrast of 4.5 or more, computed with the
    relative-luminance formula. The pairs are:
    `--txt-1/--bg`, `--txt-2/--bg`, `--txt-3/--bg`, `--txt-1/--card`, `--txt-3/--card`,
    `--link/--bg`, `--link/--card`, `--accent-text/--bg`, `--accent-text/--card`,
    `--chip-text/--chip-bg`, `--side-txt/--side-bg`, `--side-muted/--side-bg`.
  - Assert that `--txt-4` never appears in a `color:` declaration.
  - Assert that the light block's `--link` is `#2D66AE` and its `--accent-text` is `#128200`
    (FR-032).
- [X] T047 [P] [US4] In `internal/adapters/html/theme_select_test.go`, add three tests:
  - `TestThemeDefaultsLight`: `style.css` defines the light values on a bare `:root`. The dark
    values appear only inside `@media (prefers-color-scheme: dark)` under
    `:root:not([data-theme="light"])`, and under `:root[data-theme="dark"]`.
  - `TestNoNetworkReferences`: no page or asset contains `http:`, `https:` or `url(//`. The one
    exception is the footer's "Generated by loko" anchor `href` (SC-007, FR-031).
  - `TestHeadThemeScript`: the `head` block contains an inline script under 400 bytes that reads
    `localStorage["loko-theme"]` inside `try`/`catch` and sets
    `document.documentElement.dataset.theme`.
  - `TestFontStacksFallBack` (SC-011): every `--font-*` custom property in `style.css` ends in a
    generic system family (`system-ui`, `sans-serif`, `ui-monospace` or `monospace`). Characters
    outside the Latin subset then always render.

### Implementation for User Story 4

- [X] T048 [US4] Restructure `internal/adapters/html/theme/style.css` around tokens:
  - **Light block (`:root`):**
    - `--bg:#F9FAFB`
    - `--card:#FFFFFF`
    - `--link:#2D66AE`
    - `--accent:#18C800`
    - `--accent-text:#128200`
    - `--txt-1` to `--txt-4`
    - chip tokens
  - **Dark block (`:root[data-theme="dark"]` and the media query):**
    - `--bg:#0F1117`
    - `--card:#1A1D24`
    - `--link:#4A90E2`
    - `--accent-text:#18C800`
    - the dark text tones
  - **Shared tokens:** `--side-bg:#1E1E1E`, kind colours from 020, and radii (6px cards, 4px
    buttons, 3px tags).
  - **Surfaces:** borders on cards in dark mode; light shadows only in light mode.
  - **Accent use:** Cipher Green only on the accent bar, the active tree rail, the current
    sibling, active toggles, hover and `:focus-visible` rings (2px outline, 2px offset, always
    visible).
  - **Contrast tuning:** adjust the `--txt-*` values until T046 passes. Do not edit the brand
    tokens.
- [X] T049 [US4] Wire the fonts:
  - In the `head` block of `layout.gohtml`, link `{{.Root}}assets/fonts.css` before `style.css`.
    The `@font-face` rules live only in the generated `fonts.css` (T018), never in `style.css`.
  - In `internal/adapters/html/theme/style.css`, define the stacks:
    - `--font-heading: "Orbitron", system-ui, sans-serif`
    - `--font-prose: "Source Sans 3", system-ui, sans-serif`
    - `--font-mono: "JetBrains Mono", ui-monospace, monospace`
    - `--font-display: "Audiowide", "Orbitron", system-ui, sans-serif`
  - Apply them:
    - headings and `h1` use `--font-heading`
    - prose, the lead and descriptions use `--font-prose`
    - the sidebar, `.kind-label`, `.meta`, ids, chips and `code` use `--font-mono`
    - the wordmark and `.stat-number` use `--font-display`
  - Use the brand glyphs `▸ ▾ →` in markup, with sentence-case labels and no emoji.
- [X] T050 [US4] Add the theme script and toggle:
  - In the `head` block of `internal/adapters/html/theme/layout.gohtml`, add the pre-paint inline
    script per research R5, before the stylesheet links.
  - In `internal/adapters/html/theme/site.js`, wire `#theme-toggle`:
    - Reveal it.
    - Set `aria-pressed` to whether dark is active.
    - On click, flip `data-theme` and store the choice in `localStorage["loko-theme"]` (in
      `try`/`catch`). Remove the key when the choice equals `matchMedia("(prefers-color-scheme: dark)")`.

  Makes T047 pass.
- [ ] T051 [US4] Run `render_test.go -update` and review the asset goldens
  (`internal/adapters/html/testdata/golden/assets/`). Then do the manual check in quickstart §5
  across the four setups in its table.

**Checkpoint**: the brand look is in both themes, and the light fallback works without scripts.

---

## Phase 7: User Story 5 — Accessibility, offline, print, narrow (Priority: P3)

**Goal**: reduced motion, figures with captions, print stylesheet, narrow layout, offline
guarantee.

**Independent Test**: quickstart §6.

### Tests for User Story 5 (write first, confirm they fail)

- [X] T052 [P] [US5] In `internal/adapters/html/a11y_test.go`, add:
  - `TestFiguresHaveCaptions`: every `<figure>` on every golden page has a `<figcaption>`, and
    every inline SVG has `role="img"` and an `aria-labelledby` that resolves on the page.
  - `TestKindNotColourOnly`: every `icon-tile` is followed by visible kind text in the same row or
    header.
  - `TestReducedMotion`: `style.css` has a `@media (prefers-reduced-motion: reduce)` block that
    sets `animation: none` on `.cursor` and `transition: none` globally. `.cursor` uses
    `animation-iteration-count` of 3 (FR-033).
  - `TestFocusVisible` (FR-034, SC-004): parse every interactive element in the golden pages
    (`a`, `button`, `input`, `[tabindex]`). Assert that `style.css` has a `:focus-visible` rule
    covering each one's selector, or a universal `:focus-visible` rule, with a non-`none`
    `outline`. Also assert that no rule sets `outline: none` without a `:focus-visible`
    replacement.
- [X] T053 [P] [US5] In `internal/adapters/html/print_test.go`, add `TestPrintStyles`. `style.css`
  has an `@media print` block that:
  - re-declares the light tokens
  - hides `.sidebar`, `#theme-toggle`, `nav.pager`, the diagram toolbar and the menu button
  - sets `break-inside: avoid` on `.card`, `tr`, `figure`
  - removes the dot-grid background and box shadows

  Assert that print places the rail **after** the main column (header → main → rail, FR-037).

  Add `TestNarrowLayout`: a `@media (max-width: 760px)` block does the following:
  - Stacks the grid to one column in the order header → rail → main. The rail sits under the
    description, before the prose (FR-038), set with `grid-template-areas` or `order`.
  - Hides the sidebar behind `body.nav-open`.
  - Sets `overflow-wrap: anywhere` on `.meta` and ids.

### Implementation for User Story 5

- [X] T054 [US5] Add the cursor animation and reduced-motion rules to
  `internal/adapters/html/theme/style.css`:
  - The wordmark cursor blinks 3 times at 1s, then stays solid.
  - Add the reduced-motion block from T052.
  - Add `aria-describedby` and `aria-labelledby` wiring for figures in
    `internal/adapters/html/theme/partials.gohtml` and `rail.gohtml`. Caption ids derive from the
    page so they are unique and deterministic.

  Makes T052 pass.
- [X] T055 [US5] Add the `@media print` and narrow-screen blocks to
  `internal/adapters/html/theme/style.css` (research R13):
  - Print the inline SVG at `width:100%` with `break-inside: avoid`.
  - Order: print is header → main → rail; narrow screens are header → rail → main, matching
    T053.
  - In the narrow block, add a `button.menu-toggle` style. Add the button itself to the `layout`
    block in `layout.gohtml`, hidden at wide widths.
  - With scripts off, the narrow sidebar renders above the content: use `body:not(.js)` and let
    the script add `js` to `<html>`.

  Makes T053 pass.
- [ ] T056 [US5] Do the manual pass in quickstart §6:
  - Lighthouse accessibility on the index, a component page and a view page, in both themes.
  - Tab-through for focus rings.
  - Print preview.
  - The 375px width check.
  - Network-off reload.

  Record the results in `specs/021-site-redesign/checklists/manual-verification.md`. Fix any
  failure in the relevant template or `style.css` and re-run.

**Checkpoint**: all five stories are complete.

---

## Phase 8: Polish & Cross-Cutting Concerns

- [X] T057 [P] Add the theme compatibility tests to `internal/adapters/html/theme_test.go`:
  - `TestAll014BlocksStillDefined`: every block in the 014 table (`layout`, `head`, `header`,
    `nav`, `footer`, `index`, `element`, `element-tables`, `relation-table`, `element-children`,
    `prose-missing`, `view`, `tag-chips`, `link`, `diagram`) is defined by the built-in theme and
    reached when rendering the golden fixture.
  - `Test014OverridesStillBuild`: overrides written against 014 data still render without error.
    One defines `relation-table` using `.Other`, `.Description`, `.Technology` and `.Tags`. One
    defines `element-children` using `link`. One defines `nav` ranging over `.Projection.Views`.
  - Run the existing `cmd/build_theme_test.go`, which must pass unchanged (SC-010).
- [X] T058 Update `specs/014-viewmodel-renderers/contracts/theme.md`: add the new blocks and the
  `rail.gohtml` file to the table, and link `specs/021-site-redesign/contracts/theme-blocks.md`
  for the changed meanings. Update `specs/020-diagram-look/contracts/svg-output.md`'s selectors
  table with `g[data-loko-address]`, linking `contracts/svg-address.md`.
- [X] T059 [P] Add `TestByteIdenticalSite` to `internal/adapters/html/render_test.go`: render the
  same projection twice with the same `Upstream` and compare every artefact byte for byte
  (FR-040, SC-006).
- [X] T060 Run `go test ./cmd -run TestPerformanceBuild -v` and rebuild the benchmark from T001.
  Confirm the build takes under 10 s (SC-005). Record the new wall time, the total output bytes
  and the `assets/` bytes next to `/tmp/before/sizes.txt`. If the build takes 10 s or more,
  profile `renderPages` and the inline step before you change anything else.
- [X] T061 Confirm the markdown outputs, MCP and CLI outputs, and search are unchanged (FR-043):
  - `diff -r` the `md` outputs of each example against `/tmp/before/<name>`.
  - Run `go test ./internal/mcp/... ./cmd/...`.
  - Diff `assets/search-index.js` byte-for-byte against the baseline, ignoring the notice line.
  - **Live reload (FR-041):** run `go test ./internal/adapters/devserver/... ./cmd/...` (the serve
    tests rebuild through `BuildArtifacts` and now take the staged render). Then do it by hand:
    1. Run `loko serve` on `examples/microservices`.
    2. Open a component page.
    3. Edit that component's `description` in its `*.loko.hcl`.
    4. Confirm the page reloads with the new description, and with the In context diagram still
       highlighted.
- [X] T062 Finish `docs/adr/0020-site-redesign.md`. Fill every section opened in T006 with its
  decision, its rationale and the rejected alternatives, citing research R1–R16. Set its status to
  Accepted.
- [X] T063 Update `CHANGELOG.md` under `[Unreleased]`:
  - The HTML site redesign.
  - The staged render.
  - The `data-loko-address` attribute in SVGs.
  - Changed block meanings (`header`, `nav`, `element-tables`, `relation-table`,
    `element-children`, `diagram`, `index`), with a pointer to the theme contract.
  - A warning that 014-era whole-file `style.css`/`site.js` overrides meet new markup.
  - Bundled fonts and their licences.
  - The measured added asset weight from T003 and T060.
- [X] T064 [P] Update the user docs for the site and themes (`docs/` pages that describe
  `loko build --format html`, `loko serve` and `templates/` overrides). Cover the new blocks, the
  theme toggle, keyboard paging and the `custom.css` recommendation.
- [X] T065 Run `make lint` and `make audit-constitution`. Fix any file or function that exceeds
  its budget by splitting it, not by exempting it. Then run `go test -race ./...` and the
  goroutine-leak gate.
- [ ] T066 Run every step of `specs/021-site-redesign/quickstart.md` end to end. Mark
  `specs/021-site-redesign/checklists/manual-verification.md` complete.

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: no dependencies. T003 depends on T002.
- **Foundational (Phase 2)**: depends on Setup (fonts embedded in T018 need T003). It blocks all
  stories.
- **US1 (Phase 3)**: depends on Foundational.
- **US2 (Phase 4)**: depends on Foundational. Its tree markup reuses the `icon-tile` partial from
  T030. Do US2 after US1, or move T030's partial work earlier.
- **US3 (Phase 5)**: depends on Foundational, and on T027 (`pageData` fields) and T030
  (`icon-tile`, header).
- **US4 (Phase 6)**: depends on Foundational. The token restructure (T048) touches the same
  `style.css` that US1–US3 add to. Do it after them, or first, but not interleaved.
- **US5 (Phase 7)**: depends on US1–US4 markup.
- **Polish (Phase 8)**: depends on all the stories.

### Within each story

- Tests first. Confirm each fails.
- Core projection (entities, use cases) before adapter templates.
- Templates before `style.css` golden updates.
- `style.css` and `site.js` are shared files. Never run two tasks that edit them in parallel.

### Parallel opportunities

- **Setup:** T002, T004, T005 and T006.
- **Foundational tests:** T007, T008, T009 and T010. Then T013, T014, T015 and T017 in parallel.
  T016 follows T015.
- **US1 tests:** T019, T020, T021 and T022. Then T026 and T028 alongside the core work in T023 and
  T024.
- **US2 tests:** T032 and T033.
- **US3 tests:** T039 and T040.
- **US4 tests:** T046 and T047.
- **US5 tests:** T052 and T053.
- **Polish:** T057, T059 and T064.

---

## Parallel Example: User Story 1

```text
# Tests together (all different files):
T019 internal/core/usecases/project_context_test.go
T020 internal/core/usecases/project_pages_test.go
T021 internal/adapters/html/inline_test.go
T022 internal/adapters/html/element_page_test.go

# Then core and adapter pieces with no shared file:
T023 internal/core/usecases/project_context.go
T026 internal/adapters/html/inline.go
T028 internal/adapters/html/theme/rail.gohtml
```

---

## Implementation Strategy

### MVP first (US1)

1. Phase 1 and Phase 2.
2. Phase 3 (US1). Stop and validate quickstart §2: element pages give full context without the
   diagram. This is the owner's stated goal.

### Incremental delivery

1. Setup + Foundational. The staged render and the SVG address land with no visible site change.
2. US1: element context, the MVP.
3. US2: tree, breadcrumb and pager. Navigation is complete.
4. US3: overview and view pages.
5. US4: the Madstone look and the themes.
6. US5: accessibility, print, narrow and offline.
7. Polish: compatibility tests, ADR, changelog, docs, performance and determinism.

Each step leaves `go test ./...` green and the site buildable.
