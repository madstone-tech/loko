---

description: "Task list for 016-rendering-fidelity"
---

# Tasks: Rendering Fidelity

**Input**: Design documents from `/specs/016-rendering-fidelity/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md),
[data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Test tasks are **included and mandatory**, because Constitution Principle V (Test-First)
is non-negotiable. A test task always comes before the implementation it covers: run it and
**confirm it fails** before implementing.

**Organization**: Phases follow the user stories in priority order. US1 (readable layout) and US2
(titles) are P1; US3 (shapes) and US4 (relationship kinds) are P2.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1–US4 from spec.md
- Every task names its exact file paths

## Standing rules for every task

- **Additive only.** Every new field is optional and `omitempty`. A project that uses none of the
  new attributes exports byte-identically, and renders byte-identically apart from the `direction`
  line in system, container and declared views (SC-003, research R8).
- **Layering.** Entities stay stdlib-only. D2 shape names, markdown labels and the `direction:` line
  appear only in `internal/adapters/d2`. Outer layers never import entities.
- **Goldens.** Any golden change is a review event: regenerate with `-update` only after confirming
  the diff is exactly the intended change.
- **Audit.** Run `task audit-constitution` after every phase.

---

## Phase 1: Setup

- [X] T001 Confirm the fixture `testdata/projects/serverless-reference/main.loko.hcl` compiles with 0 errors (`go run . validate -p testdata/projects/serverless-reference`) and is canonical (`go run . fmt --check -p …`). Add it to the fixtures validated by the `Validate Examples` CI job if that job lists fixtures explicitly (`.github/workflows/`)
- [X] T002 [P] Write `docs/adr/0015-rendering-attributes.md`: why `title`, `shape`, relationship `kind`/`tags` and view `direction` are plain attributes (research R1); trigger semantics (clarification 2, R4); the default direction change (clarification 3, R5); and the markdown-label approach with its fallback (R2)

---

## Phase 2: Foundational (blocking)

**Purpose**: the attributes reach the IR, are validated, and are editable. Every story builds on this.

- [X] T003 [P] Write tests in `internal/core/entities/arch/render_attrs_test.go` for the value sets and membership helpers: shapes "`database`, `queue`, `topic`, `function`, `bucket`"; relationship kinds "`sync` (the default), `async`, `trigger`"; directions "`down`, `right`". Then implement them in `internal/core/entities/arch/render_attrs.go`
- [X] T004 Add the fields to `internal/core/entities/arch`: in `source_model.go`, `ElementDecl.Title`, `ElementDecl.Shape`, `RelationDecl.Kind`, `RelationDecl.Tags` and `ViewDecl.Direction`; in `ir_element.go`, `Element.Title` (`title,omitempty`), `Element.Shape` (`shape,omitempty`), `Relationship.Kind` (`kind,omitempty`, "omitted for `sync` and unset"), `Relationship.Tags` (`tags,omitempty`, sorted and de-duplicated) and `View.Direction` (`direction,omitempty`). Add codes `invalid_attribute_value`, `shape_not_allowed` and `empty_title` to `diagnostic.go` and `AllCodes`, to `specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`, and to the count in `TestAllCodesAreKnown`
- [X] T005 Write parse golden fixtures under `internal/adapters/hclsource/testdata/golden/`: `render_attrs_valid/` (every attribute on every legal kind, with no diagnostics) and `render_attrs_wrong_type/` (a `tags` that is not a list; a `title` that is not a string). Then decode the attributes in `internal/adapters/hclsource/schema.go` (`commonElementAttrs` + `title`; `shape` on container and external through `elementAttrs`; `usesAttrs` + `kind`, `tags`; `viewAttrs` + `direction`), `decode_logical.go` and `decode_project.go`, recording `AttrRanges`
- [X] T006 Write `internal/core/usecases/validate_render_attrs_test.go` over `SourceModel` literals: each invalid value gives `invalid_attribute_value` with the allowed set in the detail and the attribute's range; `shape` on a system or person gives `shape_not_allowed`; `title = ""` gives `empty_title`. Implement `internal/core/usecases/validate_render_attrs.go`, call it from `CompileArchitecture`, and add the three codes to `codesFromCoreTests` in `rule_coverage_test.go`
- [X] T007 Carry the fields through `internal/core/usecases/build_ir.go`, normalising `Kind == "sync"` to empty so it is omitted. Extend `internal/adapters/hclsource/golden_test.go` or `internal/adapters/encoding` tests: `export_v1.json` stays byte-identical, and a project using the attributes exports them
- [X] T008 [P] Extend the authoring legal-attribute table in `internal/core/entities/authoring/edit_schema.go`: element `title` (string), and `shape` (string) "on `container` and `external` only"; relationship `kind` (string) and `tags` (string list); view `direction` (string). `NewEdit` rejects out-of-set values with a field error listing the set. Write the tests first in `internal/core/entities/authoring/render_attrs_test.go`
- [X] T009 Return the attributes from `describe` (`internal/core/usecases/describe.go`): `ElementView.Title` at every level, `ElementView.Shape` at `structure` and `full`, `RelationshipView.Kind` and `Tags`, and `ViewInfo.Direction`. Test first in `describe_test.go`, and re-run `TestAuthoringPerformance` (`cmd/authoring_perf_test.go`) to confirm the structure level stays within 2,000 tokens

**Checkpoint**: the attributes compile, validate, export and edit; nothing renders differently yet.

---

## Phase 3: User Story 1 - Diagrams people can read (Priority: P1) 🎯 MVP

**Goal**: system, container and declared views lay out top-down; a view may choose its direction.

**Independent Test**: quickstart §1. The `serverless-reference` container view renders with an SVG
`viewBox` width ÷ height ≤ 2.

### Tests ⚠️ write first, confirm failing

- [X] T010 [P] [US1] Write `internal/core/usecases/project_direction_test.go`. The landscape and deployment views get `Direction: "right"`; system and container views get `"down"`; a declared view without `direction` gets `"down"`, and with `direction = "right"` gets `"right"`
- [X] T011 [P] [US1] Write `cmd/render_fidelity_test.go` with `TestContainerViewAspectRatio`: build `testdata/projects/serverless-reference` into a temp directory, read `diagrams/system-portal.svg`, parse `viewBox`, and assert width ÷ height ≤ 2, every `font-size` in the SVG's text styles ≥ 16, and that the view is narrower than the same view rendered with `direction = "right"` (FR-006, SC-001; the absolute width limit was dropped by clarification)
- [X] T012 [P] [US1] Add `TestUnchangedProjectsRenderIdentically` to `cmd/render_fidelity_test.go`. Build `testdata/projects/two-systems`; every output file must equal its golden after deleting `direction: …` lines from both sides, and the export must equal `internal/adapters/encoding/testdata/export_v1.json` byte for byte (SC-003)

### Implementation

- [X] T013 [US1] Add `View.Direction` to `internal/core/entities/viewmodel/model.go`, and set it in the projection (`internal/core/usecases/project_view.go`, and `project_deployment.go` for deployment) per the defaults in research R5 and contracts/rendering.md, to pass T010
- [X] T014 [US1] Emit `direction: <View.Direction>` in place of the hard-coded `direction: right` in `internal/adapters/d2/emit.go`. Update `emit_test.go`, regenerate the D2, site, markdown and projection goldens, and confirm the diff is only the direction line (and `"direction"` in `projection.json`). Pass T011 and T012; if T011 fails, record the measured ratio and stop for a decision before touching ELK (plan assumption)

**Checkpoint**: readable layout; unchanged projects differ only by direction.

---

## Phase 4: User Story 2 - Names people recognise (Priority: P1)

**Goal**: an element's `title` is shown prominently everywhere, with the address name smaller beneath.

**Independent Test**: quickstart §2.

### Tests ⚠️ write first, confirm failing

- [X] T015 [P] [US2] Extend `internal/core/usecases/project_view_test.go` and `project_pages_test.go`: a titled element's node has `Title` set and `Label` unchanged (the name); its `ElementPage.Title` is set; untitled elements have no `Title`
- [X] T016 [P] [US2] Extend `internal/adapters/d2/emit_test.go`. A titled node's label is a plain label: the title, then `[Kind: Technology] · name`, then the description, with no markdown label (`|md`) anywhere in the output. An untitled node's label is byte-identical to today's. The whole emitted file is deterministic over 20 renders, and the SVG renders through the in-process D2 library without error. A 120-character title is emitted whole (not truncated) and the SVG renders; its box is no wider than the widest untitled box in the same view plus 50%
- [X] T017 [P] [US2] Extend the html and markdown golden tests (`internal/adapters/html/render_test.go`, `internal/adapters/markdown/backend_test.go`) with a titled element: the page heading is the title, with `<small><code>name</code></small>` (HTML) or `` `name` `` (markdown) beneath; diagrams and tables that link to it show the title. An element page's *Uses* and *Used by* tables, in HTML and markdown, show each relationship's tags (`#write`) in a Tags column (FR-010). Each diagram image links to its SVG (FR-011)

### Implementation

- [X] T018 [US2] Add `Node.Title` and `ElementPage.Title` in `internal/core/entities/viewmodel/{model,page}.go`, and set them from `Element.Title` in `internal/core/usecases/project_view.go` and `project_pages.go`, to pass T015
- [X] T019 [US2] Render titled labels as plain labels in `internal/adapters/d2/emit.go` (`nodeLabel`), to pass T016 (research R2, amended)
- [X] T020 [US2] Show titles in `internal/adapters/html/theme/element.gohtml` and `partials.gohtml` (headings and links) and in `internal/adapters/markdown/backend.go`; add the Tags column to the relationship tables (FR-010); and wrap each diagram image in `view.gohtml` and `element.gohtml` in a link to its SVG (FR-011), to pass T017
- [X] T021 [US2] Add a `title` case to the MCP tool tests (`internal/mcp/tools/apply_edit_test.go`: set, then `describe` shows it) and title set/clear operations to the property-test generator in `internal/adapters/hclsource/edit_property_test.go` (the model tracks `title`)

**Checkpoint**: P1 complete: readable, recognisable diagrams.

---

## Phase 5: User Story 3 - Shapes that say what a thing is (Priority: P2)

**Goal**: containers and externals can be drawn as a database, queue, topic, function or bucket.

**Independent Test**: quickstart §3.

### Tests ⚠️ write first, confirm failing

- [X] T022 [P] [US3] Extend `internal/core/entities/viewmodel/style_test.go`: `StyleFor` with an element shape returns that shape and keeps the kind's colours; an empty shape keeps the kind's default. Extend `project_view_test.go`: a node collapsed into its parent carries the parent's shape
- [X] T023 [P] [US3] Extend `internal/adapters/d2/emit_test.go` with the mapping in research R3: `database`→`cylinder`, `queue`→`queue`, `topic`→`hexagon`, `function`→`step`, `bucket`→`stored_data`. Each renders to SVG through the library

### Implementation

- [X] T024 [US3] Add the five shapes to `internal/core/entities/viewmodel/style.go` and an element-shape parameter to `StyleFor`; pass `Element.Shape` in `internal/core/usecases/project_view.go`, to pass T022
- [X] T025 [US3] Map the shapes in `d2Shape` (`internal/adapters/d2/emit.go`), to pass T023. Show `shape` as a field on element pages (`element.gohtml`, markdown backend)
- [X] T026 [US3] Add shape set/clear to the property-test generator (containers and externals only) and a `shape` case to `internal/mcp/tools/apply_edit_test.go`, including the refusal for `shape` on a system (`invalid_edit` naming `set.shape`)

---

## Phase 6: User Story 4 - Connections that say how things talk (Priority: P2)

**Goal**: async edges are dashed, triggers point from the trigger, and relationship tags are shown, while query results are unchanged.

**Independent Test**: quickstart §4.

### Tests ⚠️ write first, confirm failing

- [X] T027 [P] [US4] Write `internal/core/usecases/project_lift_kind_test.go` (`TestLiftTrigger`, `TestLiftAsync`):
  - A trigger relationship `sync.consume → sync_queue` produces an edge from `sync_queue` to `sync`.
  - A trigger and a sync relationship between the same two nodes in opposite authored directions merge into one edge.
  - A trigger whose trigger end is outside the view yields a crossing edge from the outside marker to the invoked element.
  - An edge carrying only async relationships has `Style.Async`; a mix does not.
  - Crossing edges keep `Dashed` and never `Async`.
  - Tags on a single-relationship edge are its tags; a merged edge carries the sorted union.
- [X] T028 [P] [US4] Write `TestKindsDoNotChangeQueries` in `internal/core/usecases/query_test.go`: run every query kind over the `queryIR` fixture before and after setting `Kind` (trigger, async) and `Tags` on its relationships; the results are identical (FR-004, SC-004)
- [X] T029 [P] [US4] Extend `internal/adapters/d2/emit_test.go`: an async edge gets `style.stroke-dash: 8`; a crossing edge keeps `stroke-dash: 4`; a tagged edge's label ends with a line `#tag1 #tag2`. Extend the html view-page golden with the two-line legend (async vs crossing)

### Implementation

- [X] T030 [US4] Implement the trigger swap, async aggregation and tag union in `internal/core/usecases/project_lift.go`, adding `EdgeStyle.Async` and `Edge.Tags` in `internal/core/entities/viewmodel/{style,model}.go`, to pass T027 and T028
- [X] T031 [US4] Emit the async dash and the tag line in `internal/adapters/d2/emit.go` (`writeEdge`, `edgeLabel`), and add the legend to `internal/adapters/html/theme/view.gohtml`, to pass T029
- [X] T032 [US4] Add relationship `kind` and `tags` operations to the property-test generator (the model tracks both; the IR oracle compares them) and to `internal/mcp/tools/apply_edit_test.go`
- [X] T033 [US4] Remodel the fixture with the new attributes in `testdata/projects/serverless-reference/main.loko.hcl`, through the MCP tools (`apply_edit` batches): titles on every container, shapes on the table, queues, topic, functions and bucket, and triggers for the queue → Lambda and schedule → scheduler relationships, with the arrows restored to the data-flow direction. It still compiles and is canonical; re-run T011 and SC-004's query comparison (quickstart §4)

---

## Phase 7: Polish & Cross-Cutting

- [X] T034 [P] Document the attributes in `docs/language.md` (element `title` and `shape`, `uses` `kind` and `tags`, view `direction`, the three diagnostics, and that kinds do not affect queries), in the targets table of `docs/mcp-integration.md`, and in `CHANGELOG.md` (including the one-time default direction change)
- [X] T035 [P] Confirm the 014 build performance budget (`cmd/build_perf_test.go`) and the 015 authoring budgets (`cmd/authoring_perf_test.go`) still hold
- [X] T036 Run the full gate: `task lint`, `task audit-constitution`, `go test -race ./...`, and every scenario in [quickstart.md](./quickstart.md). Then render the remodelled fixture's container view to PNG and compare it with the original wide strip (SC-001, SC-002); include both images in the PR description

---

## Dependencies & Execution Order

```
Phase 1 ──▶ Phase 2 ──┬──▶ Phase 3 US1 (direction) ──┐
                      ├──▶ Phase 4 US2 (titles)     ├──▶ T033 (fixture remodel) ──▶ Phase 7
                      ├──▶ Phase 5 US3 (shapes)     │
                      └──▶ Phase 6 US4 (kinds) ─────┘
```

- **Phase 2 blocks everything:** the attributes must parse, validate and reach the IR first.
- **The stories are independent of each other**, but all four edit `internal/adapters/d2/emit.go`
  and the D2 goldens. Run their emitter tasks in order: T014 → T019 → T025 → T031.
- **T033 needs all four stories**, because it uses every attribute.
- **Shared files:**
  - `project_view.go`: T013 → T018 → T024.
  - The property-test generator: T021 → T026 → T032.

## Parallel Examples

```
Phase 2: T003 render attrs | T008 authoring schema
US1 tests: T010 direction defaults | T011 aspect ratio | T012 byte identity
US2 tests: T015 projection | T016 D2 labels | T017 html/markdown
US4 tests: T027 lift | T028 queries unchanged | T029 emitter
```

## Implementation Strategy

1. **MVP = Phases 1–4 (US1 + US2).** Diagrams become readable and recognisable. Validate with
   quickstart §1–2 and the aspect-ratio test.
2. **US3 (shapes) and US4 (kinds)** are independent increments after the MVP.
3. **T033 remodels the fixture**, the end-to-end proof against the original UX finding.
4. **Polish**: docs, budgets, the full gate, and before/after images.

Commit after each task or logical group. Golden changes are review events.
