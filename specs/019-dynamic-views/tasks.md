---

description: "Task list for 019-dynamic-views"
---

# Tasks: Dynamic Views

**Input**: Design documents from `/specs/019-dynamic-views/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md),
[data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Test tasks are **included and mandatory**, because Constitution Principle V (Test-First)
is non-negotiable. A test task always comes before the implementation it covers: run it and
**confirm it fails** before implementing.

**Organization**: Phases follow the user stories in priority order. US1 (ordered flow) and US2
(checks) are P1; US3 (sequence style) is P2; US4 (MCP) is P3.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1–US4 from spec.md
- Every task names its exact file paths

## Standing rules for every task

- **No unrequested change.** A project with no `step` block produces byte-identical exports, D2,
  SVG, markdown and HTML to 1.2.0 (FR-013, SC-005). Before starting, snapshot every example and
  fixture with the 1.2.0 binary and compare after each phase.
- **Exit codes** are unchanged; `step_unmatched` is a warning like any other.
- **Layering.** Entities stay stdlib-only and do not import each other: `authoring` keeps its copy,
  held by `TestLanguageParity`. Use cases do no file I/O. `internal/mcp` imports no entity package
  and no render adapter.
- **Budgets.** MCP tool functions ≤ 30 effective lines, CLI handlers ≤ 50, use-case functions ≤ 60,
  use-case and entity files ≤ 200, adapter files ≤ 400 (`arch/schemas.go` exempt as pure data).
- **Goldens.** Any golden change is a review event: regenerate with `-update` only after confirming
  the diff is exactly the intended change.
- **Audit.** Run `make audit-constitution` after every phase.

---

## Phase 1: Setup

**Purpose**: the fixture every story uses.

- [X] T001 Create `testdata/projects/dynamic-views/` (README, `main.loko.hcl`, `docs/*.md` for every element so the fixture has no `missing_docs`): a `shop` system with containers `web`, `api`, `queue` (shape `queue`), `worker` (shape `function`) and `db` (shape `database`); `api` has components `handler` and `repo`; a `customer` person and a `payments` external. Relationships: customer → web, web → api, api → payments, `repo` → db (component level), api → queue, worker `uses` queue with `kind = "trigger"`, worker → db. **No** `step` block yet: the fixture's pre-feature build is the byte-identity baseline. Filled in by T013 and T024.

---

## Phase 2: Foundational (blocking prerequisites)

**Purpose**: the language table, entities, codes and authoring copy (research R1, R4, R5;
data-model.md). No behaviour change on any existing fixture.

### Tests

- [X] T002 [P] Extend `internal/core/entities/arch/diagnostic_test.go` and `diagnostic_ignore_test.go` (failing): `AllCodes` has 37 entries including `step_mismatch`, `step_unmatched`, `dynamic_view_selection`; `IgnorableCodes` contains `step_unmatched` and neither error code.
- [X] T003 [P] Write a failing test in `internal/core/entities/arch/address_test.go`: `NewStepAddress("view.daily_run", "fetch")` is `view.daily_run.step.fetch`, and its `Prefix()` is `view`.
- [X] T004 [P] Extend `internal/core/entities/arch/schemas_test.go` (failing): block `step` exists with `In: ["view"]`, label `name`, attributes `description` (string), `from` (ref, required), `reply` (bool), `to` (ref, required), `via` (ref); `view` has `style` with `Values` `["edges", "sequence"]`.

### Implementation

- [X] T005 Add the `step` block and `view.style` to `internal/core/entities/arch/schemas.go` (notes: `via` "the relationship this step uses; checked against from and to", `reply` "a reply along the relationship, drawn dashed", `style` "edges by default; sequence draws a sequence diagram"); add `StepStyles = []string{"edges", "sequence"}` beside `Directions`; give `step` an example from contracts/language.md.
- [X] T006 Move `ViewDecl` from `internal/core/entities/arch/source_model.go` into a new `internal/core/entities/arch/source_model_view.go` and add `Style`, `StyleRange`, `IncludeRange`, `ExcludeRange`, `TagsRange`, `Steps []StepDecl`, and `StepDecl{Name, From, To, Via Reference, Description string, Reply bool, Range SourceRange}` (data-model.md); move IR `View` from `ir_element.go` into a new `internal/core/entities/arch/ir_view.go` and add `Style` and `Steps []Step` with `Step{Address, Name, Number, From, To, Via, Relationship, Reply, Description, Range}`, JSON/TOON tags exactly as data-model.md (`omitempty` on every new optional field); add `NewStepAddress` in `internal/core/entities/arch/address.go`.
- [X] T007 Add `CodeStepMismatch` (error), `CodeStepUnmatched` (warning), `CodeDynamicViewSelection` (error) to `AllCodes` in `internal/core/entities/arch/diagnostic.go`, `CodeStepUnmatched` to `IgnorableCodes` in `diagnostic_ignore.go`, and the three to the `code` enum of `specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`. Add each to a temporary `pendingCodes` set in `internal/core/usecases/rule_coverage_test.go`, removed by T020 and T021.
- [X] T008 [P] Add `TargetStep TargetKind = "step"` and `Edit.Before string` in `internal/core/entities/authoring/edit.go` (refuse `before` unless `op = add` and `target = step`, with a field error naming `before`); add `legalAttrs[TargetStep]` (`from`, `to`, `via` as `ValueRef`, `description` `ValueString`, `reply` `ValueBool`), `requiredOnAdd[TargetStep] = {"from", "to"}` and `style` on `TargetView` in `internal/core/entities/authoring/edit_schema.go`; `StepStyles` and `enumAttrs[TargetView]["style"]` in `internal/core/entities/authoring/render_attrs.go`; add `{TargetStep, "step"}` to the pairs of `TestLanguageParity` in `internal/core/usecases/language_parity_test.go` and the style set to the value-set parity test.
- [X] T009 Run `go test ./...`: T002–T004 pass, parity passes, every golden unchanged; snapshot comparison identical.

**Checkpoint**: `make test` passes and `make audit-constitution` is clean.

---

## Phase 3: User Story 1 - Describe an ordered flow (Priority: P1) 🎯 MVP

**Goal**: steps decode, resolve, compile into the IR, and draw as numbered edges inside system
boundaries, with a Steps table on the view's pages (FR-001–FR-004, FR-009, FR-011–FR-013;
research R5, R6, R8).

**Independent Test**: quickstart §1 on `testdata/projects/dynamic-views`: numbered edges, the
Steps table, the export; every other fixture byte-identical.

### Tests for User Story 1

- [X] T010 [P] [US1] Add a decoder golden `internal/adapters/hclsource/testdata/golden/dynamic_views/` (input + `expected_diagnostics.json` = `[]`) with a view of three steps using every attribute, and a test `internal/adapters/hclsource/decode_steps_test.go` checking `ViewDecl.Steps` keeps declaration order (not label order), each field and range, and `IncludeRange`/`TagsRange`/`StyleRange` on a view that sets them.
- [X] T011 [P] [US1] Write failing tests in `internal/core/usecases/resolve_steps_test.go`: `from`, `to` and `via` resolve through the symbol table; an unknown `from` is `unresolved_reference` at its range with a suggestion; `via = container.api` (not a relationship) is `wrong_reference_kind`; two steps with one label in a view are `duplicate_declaration`, the same label in two views is fine; `BuildIR` gives `Steps` in declaration order, `Number` 1..n, resolved addresses.
- [X] T012 [P] [US1] Write failing tests in `internal/core/usecases/project_dynamic_test.go`: participants in first-appearance order; a container participant inside its system boundary node (`RoleSubject` style) when the system takes no step; a component inside its participating container, else directly inside the system; a participating system drawn as an element; one edge per step with ID `step__0001__<label>`, label `1. <description>` or the relationship's description, technology from the relationship, `Style.Dashed` for a reply; two steps between the same pair give two edges; `ViewModel.Steps` in order with `LinkRef`s and node IDs; the view model passes `Validate`; `ResolveViews` gives no `view_empty` for a view with steps.

### Implementation for User Story 1

- [X] T013 [US1] Decode `step` blocks and `style` in `decodeView`, `internal/adapters/hclsource/decode_project.go` (record `IncludeRange`, `ExcludeRange`, `TagsRange`, `StyleRange`); add `blockStep` and `stepAttrs` in `internal/adapters/hclsource/schema.go`. Then add `view "checkout"` (steps: customer → web, web → api, api → payments with description "Charge card", and `repo` → db via the component relationship, written as `api` → `db`) to `testdata/projects/dynamic-views/main.loko.hcl`.
- [X] T014 [US1] Declare each step address in `internal/core/usecases/symbol_table.go` (kind `step`, noun "Step"); resolve `from`/`to` against `logicalKinds` and `via` against the relationship kind in `internal/core/usecases/resolve_references.go`, storing results in new `Resolved.StepFrom`, `StepTo`, `StepVia` maps keyed by step address.
- [X] T015 [US1] Build IR `View.Steps` (and `Style`) in `buildViews`, `internal/core/usecases/build_ir.go`, in declaration order with `Number` = position; `Relationship` is filled by T021 (empty until then).
- [X] T016 [US1] Add `View.Style` (`json:"style,omitempty"`) in `internal/core/entities/viewmodel/view.go` and `ViewModel.Steps []StepRow` with `StepRow{Number, Name, From, To LinkRef, FromNode, ToNode, Text, Technology, Reply}` (JSON per data-model.md) in `internal/core/entities/viewmodel/model.go`; set `Style` for declared views in `internal/core/usecases/resolve_views.go` and skip the `view_empty` selection check for views with steps.
- [X] T017 [US1] Implement `projectDynamic` in a new `internal/core/usecases/project_dynamic.go` (research R6) and dispatch to it from `ProjectView` in `internal/core/usecases/project_view.go` when the declared view has steps; set `ViewModel.Steps` in `finishView` or alongside it.
- [X] T018 [US1] Add a "Steps" section to `internal/adapters/html/theme/view.gohtml` (`{{with .Steps}}`: columns #, From, To, Description; elements linked; a reply marked "(reply)"), and the same table to the markdown view page in `internal/adapters/markdown/backend.go`; add HTML and markdown goldens for `dynamic-views` (`internal/adapters/html/testdata/golden/dynamic-views/`, `internal/adapters/markdown/testdata/golden/dynamic-views/`), and add `dynamic-views` to `TestProjectionGolden` in `cmd/projection_golden_test.go`.
- [X] T019 [US1] Add a D2 golden for the edges style in `internal/adapters/d2/emit_test.go` (case `dynamic_edges` from the projection of `checkout`), confirm repeated edges render in an SVG test, add `dynamic-views` to the canonical-format check in `internal/adapters/hclsource/format_canonical_test.go` (or `cmd/mcp_language_polish_test.go`'s `TestLanguagePolishFixtureIsCanonical` pattern) together with a case whose steps are declared out of label order and stay in that order after `fmt`, run the full suite and the snapshot comparison: only `dynamic-views` outputs differ.

**Checkpoint**: US1 works on its own; quickstart §1 passes.

---

## Phase 4: User Story 2 - Steps the compiler checks (Priority: P1)

**Goal**: selection conflicts, matching through containment and replies, `step_mismatch` and
`step_unmatched`; `move` rewrites steps (FR-005–FR-008; research R2, R3).

**Independent Test**: quickstart §2: each break produces its own diagnostic at the step's line.

### Tests for User Story 2

- [X] T020 [P] [US2] Write failing tests in `internal/core/usecases/validate_steps_test.go` for selection: a view with steps and `include`, `exclude` or `tags` → `dynamic_view_selection` at that attribute's range; `style` without steps → `dynamic_view_selection` at `style`; `style = "flow"` → `invalid_attribute_value` listing `edges, sequence`. Add a `dynamic_view_selection` case to `codesFromCoreTests` in `internal/core/usecases/rule_coverage_test.go` (the coverage test counts only codes produced there or by golden fixtures), then remove it from `pendingCodes`.
- [X] T021 [P] [US2] Write failing table tests in `internal/core/usecases/validate_steps_test.go` for matching (contracts/language.md "Matching"): direct match; container-level step over a component relationship; component-level step over a container relationship; trigger relationship matched in drawn direction (queue → worker); `reply = true` matched reversed; a non-reply step against the only relationship's direction → `step_unmatched`; `via` not connecting the ends → `step_mismatch` naming both pairs; self step with and without a self relationship; first match in address order recorded as `Step.Relationship`. Add a `step_mismatch` and a `step_unmatched` case to `codesFromCoreTests` in `internal/core/usecases/rule_coverage_test.go`, then remove both from `pendingCodes`, and the then-empty set.
- [X] T022 [P] [US2] Write a failing MCP case in `cmd/mcp_language_test.go` (new file): `move` of `container.api` to `container.orders_api` on a copy of `dynamic-views` rewrites every step's `from`, `to` and `via`, and the model still compiles with no new diagnostic.

### Implementation for User Story 2

- [X] T023 [US2] Implement `ValidateSteps` in a new `internal/core/usecases/validate_steps.go` (selection rules and matching, `related` via `Resolved.Parent`, trigger swap as in `liftEdges`), call it from `CompileArchitecture` in `internal/core/usecases/compile_architecture.go` after resolution and before ignore lists, and the `style` enum check in `ValidateRenderAttributes` (`internal/core/usecases/validate_render_attrs.go`); expose the matched relationship to `BuildIR` (a `Resolved.StepRelationship` map) and set `Step.Relationship`.
- [X] T024 [US2] Add to `testdata/projects/dynamic-views/main.loko.hcl` a view `async_order` (api → queue, queue → worker over the trigger relationship, worker → db) and confirm the fixture validates clean. The error cases stay in the T020 and T021 table tests; no fixture directory is added for them.
- [X] T025 [US2] Run the suite and the snapshot comparison; confirm `TestValidationRuleCoverage` passes with no pending codes.

**Checkpoint**: US1 and US2 work; quickstart §2 passes.

---

## Phase 5: User Story 3 - A sequence diagram of the same flow (Priority: P2)

**Goal**: `style = "sequence"` draws a D2 sequence diagram with the same numbers and labels
(FR-003, FR-010; research R7).

**Independent Test**: quickstart §3: lifelines in first-appearance order, arrows in step order,
same Steps table.

### Tests for User Story 3

- [X] T026 [P] [US3] Write failing tests in `internal/adapters/d2/emit_sequence_test.go`: a view model with `View.Style = "sequence"` emits a root `loko-flow` object with `shape: sequence_diagram`, actors in `Steps` order (first appearance), edges in step order labelled `<n>. <text>` with `stroke-dash` on replies, no boundary nodes, the canvas title when declared, and no `direction` or `layout-engine` lines; a D2 golden `dynamic_sequence.d2`; an SVG test rendering it under dagre and ELK where the step texts appear top to bottom in order.
- [X] T027 [P] [US3] Write a failing test in `internal/core/usecases/wide_view_test.go`: a sequence-style view never gets `wide_view`.

### Implementation for User Story 3

- [X] T028 [US3] Implement `emitSequence` in a new `internal/adapters/d2/emit_sequence.go` and dispatch to it from `Emit` in `internal/adapters/d2/emit.go` when `m.View.Style == "sequence"`.
- [X] T029 [US3] Skip sequence-style views in `WideViewHints`, `internal/core/usecases/wide_view.go`; add `view "checkout_sequence"` (the `checkout` steps plus a `reply = true` step payments → api) with `style = "sequence"` to `testdata/projects/dynamic-views/main.loko.hcl`; regenerate that fixture's goldens after review and confirm the Steps tables of `checkout` and `checkout_sequence` match row for row.

**Checkpoint**: US1–US3 work; quickstart §3 passes.

---

## Phase 6: User Story 4 - Assistants author flows through MCP (Priority: P3)

**Goal**: the `step` edit target with `before`, `describe` steps, `query` `flows` (FR-014–FR-016;
research R9, R10; contracts/cli-mcp.md).

**Independent Test**: quickstart §4 through MCP only.

### Tests for User Story 4

- [X] T030 [P] [US4] Write failing adapter tests in `internal/adapters/hclsource/edit_steps_test.go`: add a step to a view without steps (appended, blank line before it, canonical `fmt`), add with `before` (inserted before that sibling, other steps and their comments unchanged), `before` an unknown label → `not_found`, update `description` and clear `via`, remove a step, all leaving the rest of the file byte-identical.
- [X] T031 [P] [US4] Write failing tests: `describe` returns `style` and `steps` for a view with steps and neither for one without (`internal/core/usecases/describe_test.go`); `Query{Kind: "flows", Address: "container.api"}` lists steps where `api` or one of its components is `from` or `to`, sorted by view then number, with `role` (`internal/core/usecases/query_flows_test.go`).
- [X] T032 [P] [US4] Write a failing MCP end-to-end case in `cmd/mcp_language_test.go`: on a copy of `dynamic-views`, `schema` with `block: "step"` lists the five attributes with `from`/`to` required; `apply_edit` adds `view "refund"`, three steps, inserts one with `before`, removes the last (each previewed first); `build` returns no error and `diagrams/refund.d2` numbers the steps 1–3 in the expected order; a step whose `from` does not exist is refused with the compile diagnostic.

### Implementation for User Story 4

- [X] T033 [US4] Locate steps under their view in `internal/adapters/hclsource/edit_locate.go` (`view.<view>.step.<label>`), and implement add (append or `before`, by removing the later step blocks and re-appending them), update and remove in a new `internal/adapters/hclsource/edit_steps.go`, dispatched from `add`/`update`/`remove` in `internal/adapters/hclsource/edit_blocks.go`.
- [X] T034 [US4] Add `step` to the MCP `target` enum and a `before` string property to the edit schema in `internal/mcp/tools/schemas.go`, parse it in `internal/mcp/tools/helpers.go`, and pass it through `usecases` to the adapter.
- [X] T035 [US4] Add `Style` and `Steps []StepInfo` to `ViewInfo` in `internal/core/usecases/describe_types.go`, filled in `internal/core/usecases/describe.go`; implement kind `flows` in a new `internal/core/usecases/query_flows.go`, dispatched from `internal/core/usecases/query.go`; add `flows` to the MCP `query` kind enum in `internal/mcp/tools/schemas.go` and a `loko query flows <address>` subcommand in `cmd/query_cobra.go`.
- [X] T036 [US4] Run the suite, lint and audit; confirm the snapshot comparison.

**Checkpoint**: all four stories work; quickstart §4 passes.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [X] T037 [P] Document dynamic views in `docs/language.md` (a `view` subsection: steps, matching, replies, styles, the three codes in the diagnostics tables), `docs/cli-reference.md` (`loko query flows`) and `docs/mcp-integration.md` (`step` target, `before`, `describe` steps, `query` `flows`).
- [X] T038 [P] Write `docs/adr/0018-dynamic-views.md` (research R12): steps inside `view` rather than a new block, labels, the matching rule (containment, drawn direction, explicit replies), the two styles.
- [X] T039 [P] Add an `[Unreleased]` entry to `CHANGELOG.md` (Added: `step` blocks and `view.style`, the three codes, the `step` edit target with `before`, `query flows`, steps in `describe`).
- [X] T040 Update `docs/roadmap.md` (019 shipped; semantic diff next).
- [X] T041 Extend `TestPerformanceBuild` in `cmd/build_perf_test.go` (the 1,000-element benchmark, under 10 seconds per feature 014 SC-006) with a generated 40-step flow over the benchmark's elements, and assert the build still completes in under 10 seconds (SC-006).
- [X] T042 Run `make test`, `go test -race ./...`, `make lint`, `make audit-constitution`, `go test -cover ./internal/core/...`; confirm the determinism job and every export, D2, markdown and HTML golden of fixtures without steps is unchanged (SC-005), and exit codes of `validate` and `build` on every example and fixture match 1.2.0.
- [X] T043 Walk `specs/019-dynamic-views/quickstart.md` end to end against `./loko`; record any deviation in an "Implementation notes" section at the end of this file.

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: none.
- **Foundational (Phase 2)**: after Setup; blocks every story (language table, entities, codes).
- **US1 (Phase 3)**: after Phase 2.
- **US2 (Phase 4)**: after US1's resolution (T014, T015); its matching fills `Step.Relationship`.
- **US3 (Phase 5)**: after US1 (needs `ViewModel.Steps`); independent of US2.
- **US4 (Phase 6)**: after US1; independent of US2 and US3 except the fixture.
- **Polish (Phase 7)**: after the stories it documents.

### Within each story

Tests first and failing → entities → adapters and use cases → rendering → end-to-end check.

### Parallel opportunities

- Phase 2: T002–T004 together; T008 alongside T005–T007.
- US1: T010–T012 together.
- US2: T020–T022 together.
- US3: T026–T027 together, and the whole of US3 alongside US2.
- US4: T030–T032 together, and US4 alongside US2/US3 (different files except the fixture).
- Polish: T037–T039 together.

### Parallel example: User Story 4

```text
T030 edit adapter tests (hclsource/edit_steps_test.go)
T031 describe and flows tests (usecases/describe_test.go, query_flows_test.go)
T032 MCP end-to-end (cmd/mcp_language_test.go)
```

---

## Implementation Strategy

### MVP first (User Story 1)

1. Phase 1 and Phase 2.
2. Phase 3 (US1): an ordered flow drawn with numbered edges and listed on its page.
3. Stop and validate with quickstart §1.

### Incremental delivery

1. US1 → ordered flows.
2. US2 → checked flows (the report's main ask; ships with US1 in practice).
3. US3 → sequence diagrams.
4. US4 → MCP authoring and reading.

Each story is a reviewable commit on the feature branch; the release (1.3.0) is cut once all four
land.

---

## Implementation notes

Recorded while implementing (T043 walked quickstart.md against `./loko`; every scenario passes:
§1 and §2.1–§2.4 and §3 by hand on a copy of `first-run`, §2.5 and §4 through
`TestMCPMoveRewritesSteps` and `TestMCPAuthorFlow`).

- **Step labels with no text.** A step whose description and relationship description are both
  empty is drawn with its number alone (`1`), not `1. ` with a trailing space; the `first-run`
  relationships have no descriptions, so the quickstart's flow shows `1`, `2. Search + fetch`, `3`.
- **`via` of the wrong kind** is reported by the resolver as `wrong_reference_kind`, which names the
  required kind as `uses`, the relationship kind in the symbol table.
- **`move` and the `moved` block.** Renaming an element with `move` rewrites `from`, `to` and `via`
  in every step; the `moved` block it adds keeps the old address, as for every rename.
- **Sequence lifelines** carry their element's shape and colours, and the `classes:` they use are
  declared, as in the edges style, so user CSS can target them.
- **Files.** The decoder's step code is in `hclsource/decode_steps.go`; IR steps are built in
  `usecases/build_ir_steps.go`; `StepRow` and `StyleSequence` are in `viewmodel/step.go`; the
  `describe` and `flows` helpers share `usecases/query_flows.go`.
- **Fixture.** `dynamic-views` declares its relationship with the trigger on the worker
  (`worker uses queue, kind = "trigger"`), so the `async_order` flow's queue → worker step checks the
  drawn-direction rule.
- **Performance (T041).** The 1,020-element benchmark with a 40-step flow built in 2.7 s on the
  development machine (budget 10 s), with no step diagnostic.
- **Lint.** `golangci-lint` 2.15.0 locally reports two `errcheck` findings in `cmd/serve.go`, which
  this feature does not touch; `main` reports the same two with this linter version, while CI's
  Lint job passes.
- **Coverage (T042).** `usecases` 91.1%, `authoring` 88.9%, `viewmodel` 96.3%; `entities/arch`
  48.8% (48.5% on `main`), below the 80% target before this feature.
