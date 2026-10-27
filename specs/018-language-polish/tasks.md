---

description: "Task list for 018-language-polish"
---

# Tasks: Language Polish

**Input**: Design documents from `/specs/018-language-polish/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md),
[data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Test tasks are **included and mandatory**, because Constitution Principle V (Test-First)
is non-negotiable. A test task always comes before the implementation it covers: run it and
**confirm it fails** before implementing.

**Organization**: Phases follow the user stories in priority order. US1 (deployment nodes) and US2
(views) are P1; US3 (docs paths) is P2; US4 (wide-view hint) is P3.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1–US4 from spec.md
- Every task names its exact file paths

## Standing rules for every task

- **No unrequested change.** A project that uses none of the new attributes, and whose docs paths
  resolve from the project root, produces byte-identical exports, D2, SVG and HTML to 1.1.0
  (FR-018, SC-006). The one deliberate change to existing output is the `docs_root_relative`
  information line (plan, "Output change").
- **Exit codes.** Information diagnostics never change an exit code, including under `--strict`
  (FR-012, SC-005).
- **Layering.** Entities stay stdlib-only and do not import each other: `authoring` keeps copies
  held in step by `TestLanguageParity`. Use cases do no file I/O: after US3, `internal/core/usecases`
  imports no `os` (research R9). `internal/mcp` imports no entity package and no render adapter.
- **Budgets.** MCP tool functions ≤ 30 effective lines, CLI handlers ≤ 50, use-case functions ≤ 60,
  use-case and entity files ≤ 200, adapter files ≤ 400 (`arch/schemas.go` exempt as pure data).
- **Goldens.** Any golden change is a review event: regenerate with `-update` only after confirming
  the diff is exactly the intended change.
- **Audit.** Run `make audit-constitution` after every phase.

---

## Phase 1: Setup

**Purpose**: a fixture that exercises every new attribute, used by US1, US2 and the regression
checks.

- [X] T001 Create `testdata/projects/deployment-details/` (README, `main.loko.hcl`) modelled on the report's cloud layout (research R3, R4; spec SC-001): `deployment "prod"` with `node "aws"` → `node "eu_west_1"` → `node "vpc"` holding instances of a Lambda container, a DynamoDB table and an S3 bucket; `deployment "dev"` with one instance of the same container. Use **no** 018 attributes yet: the fixture is filled in by T014 and T024, so its pre-feature build is the byte-identity baseline. Every element has `docs` written from the root file, with the docs files present (so the baseline has no `docs_not_found`).

---

## Phase 2: Foundational (blocking prerequisites)

**Purpose**: the information severity and the three codes (research R8, R11). No behaviour change
on any existing fixture.

**⚠️ CRITICAL**: US3 and US4 depend on this phase. US1 and US2 do not, but run it first anyway so
the codes exist for the rule-coverage test.

### Tests

- [X] T002 [P] Write failing tests in `internal/core/entities/arch/diagnostic_test.go`: `SeverityInfo.String()` is `"info"`; `ExitCode(true)` is `ExitSuccess` for a set holding only information diagnostics and `ExitWarnings` when one warning is added; `CountBySeverity(SeverityWarning)` ignores information; `AllCodes` has 34 entries including `docs_root_relative`, `docs_ambiguous`, `wide_view`; `IgnorableCodes` is the warning and information subset of `AllCodes` in the same order.
- [X] T003 [P] Write failing tests in `internal/core/entities/arch/diagnostic_ignore_test.go`: `Without([]string{"wide_view"})` removes an information diagnostic of that code and keeps errors of any code; `CheckIgnore` accepts `docs_root_relative` and `wide_view`, refuses `syntax_error`, and its message lists `IgnorableCodes`.
- [X] T004 [P] Write failing tests in `internal/core/usecases/validate_filter_test.go`: `FilterSeverity(d, "warning")` keeps errors and warnings and drops information; `"error"` and `""` behave as in 1.1.0; any other value errors with a message naming `error` and `warning`. `Summarize` orders groups errors, warnings, information.
- [X] T005 [P] Write a failing test in `internal/adapters/hclsource/render_diagnostics_test.go`: an information diagnostic prints with an `Info:` label, the summary line reads `0 errors, 0 warnings, 1 info`, and a set with no information prints the 1.1.0 summary line byte for byte.

### Implementation

- [X] T006 Add `SeverityInfo` (value 2, `String()` "info") and the codes `CodeDocsRootRelative` (info), `CodeDocsAmbiguous` (warning), `CodeWideView` (info) to `AllCodes`; rename `WarningCodes` to `IgnorableCodes` (warning and information codes, `AllCodes` order) and update every caller, including `ignoreWarningsNote` in `internal/core/entities/arch/language.go`; in `internal/core/entities/arch/diagnostic.go`. Leave `ExitCode` counting errors and warnings only.
- [X] T007 Make `Without` remove information diagnostics by code as well as warnings, and `CheckIgnore` check against `IgnorableCodes`, in `internal/core/entities/arch/diagnostic_ignore.go`.
- [X] T008 [P] Extend `FilterSeverity` with `"warning"` and order `Summarize` groups by severity (errors, warnings, information) in `internal/core/usecases/validate_filter.go`; add `"warning"` to the MCP `severity` enum and its description in `internal/mcp/tools/schemas.go`, and to the `--severity` help in `cmd/validate_cobra.go`.
- [X] T009 [P] Print `Info:` (dim cyan) for information diagnostics and add `, N info` to the summary line only when N > 0 in `internal/adapters/hclsource/render_diagnostics.go`; any other per-severity count in `cmd/` (validate and build summaries) leaves information out of the warning count.
- [X] T010 Add `info` to the severity enum of `specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`, and run `go test ./...`: T002–T005 pass; every golden is unchanged. `TestValidationRuleCoverage` fails for the three new codes until US3 and US4 land: add a `pendingCodes` set to it (`docs_root_relative`, `docs_ambiguous` removed by T038; `wide_view` removed by T044) that it skips, with a comment naming those tasks. No `t.Skip` of the whole test.

**Checkpoint**: `make test` passes (with the three documented pending codes) and `make audit-constitution` is clean.

---

## Phase 3: User Story 1 - Describe deployment infrastructure (Priority: P1) 🎯 MVP

**Goal**: deployment nodes carry `title`, `technology`, `description` and `tags`; instances carry a
`technology` override; the deployment diagram and page show them (FR-001–FR-004, research R1–R5).

**Independent Test**: build `testdata/projects/deployment-details` with the attributes added; the
deployment diagrams and pages show them, and every other output is byte-identical to its
pre-feature build (quickstart §1).

### Tests for User Story 1

- [X] T011 [P] [US1] Rely on `TestLanguageParity` in `internal/core/usecases/language_parity_test.go` (it already pairs `TargetGroup`↔`node` and `TargetInstance`↔`instance`): confirm it fails once T014 adds the attributes to `arch.Language` but before T016 adds them to `authoring`; add an `authoring` test in `internal/core/entities/authoring/render_attrs_test.go`: `set.title = ""` on `TargetGroup` is refused with "must not be empty".
- [X] T012 [P] [US1] Add a decoder golden `internal/adapters/hclsource/testdata/golden/language_polish/` (input + expected IR): nested nodes with all four attributes, an instance with `technology`, a node with `title = ""` (compile error at its range), and a node `tags` list with a duplicate (the element tag diagnostic).
- [X] T013 [P] [US1] Write failing tests: in `internal/core/usecases/project_deployment_test.go`, group nodes get `Title`, `Technology`, `Description`, `Tags` from the IR group, and an instance node's `Technology` is the instance's when set, else the element's; in `internal/adapters/d2/emit_test.go`, `nodeLabel` for a group gives `<title>\n[<technology>] · <name>`, `<name>\n[<technology>]` without a title, the bare name with neither, and never the description; extend `internal/adapters/d2/testdata/golden/deployment.d2` only via a new golden `deployment_details.d2` so the existing one stays byte-identical.

### Implementation for User Story 1

- [X] T014 [US1] Add to `node` in `internal/core/entities/arch/schemas.go`: `title` (string), `technology` (string), `description` (string), `tags` (list(string)); add `technology` (string, Note "overrides the element's technology in this deployment") to `instance`. Then add the four attributes to `node "vpc"` and `node "aws"` and `technology = "local process"` to the `dev` instance in `testdata/projects/deployment-details/main.loko.hcl`.
- [X] T015 [US1] Add `Title`, `Technology`, `Description`, `Tags`, `AttrRanges` to `GroupDecl` and `Technology` to `InstanceDecl` in `internal/core/entities/arch/source_model_deployment.go`; add `title`, `technology`, `description`, `tags` to IR `Group` and `technology` to IR `Instance` in `internal/core/entities/arch/ir_deployment.go`, all `json:",omitempty" toon:",omitempty"`.
- [X] T016 [P] [US1] Add `title`, `technology`, `description` (`ValueString`) and `tags` (`ValueList`) to `TargetGroup`, and `technology` to `TargetInstance`, in `legalAttrs` in `internal/core/entities/authoring/edit_schema.go`; extend the empty-title rule in `checkEnum` (`internal/core/entities/authoring/render_attrs.go`) to `TargetGroup`.
- [X] T017 [US1] Decode the node and instance attributes in `internal/adapters/hclsource/decode_deployment.go` with the element rules: refuse an empty `title` with the element's message, apply the element tag check to `tags`; record attribute ranges.
- [X] T018 [US1] Copy the new fields into the IR in `internal/core/usecases/build_ir_deployment.go`.
- [X] T019 [US1] Fill `Title`, `Technology`, `Description` and `Tags` on `RoleGroup` nodes, and `cmp.Or(instance.Technology, element.Technology)` on `RoleInstance` nodes, in `internal/core/usecases/project_deployment.go`; add `Tags []string` (`json:"tags,omitempty"`) to `Node` in `internal/core/entities/viewmodel/model.go`.
- [X] T020 [US1] Label groups per research R3 in `nodeLabel`, `internal/adapters/d2/emit.go`: `<wrapTitle(title)>\n[<technology>] · <name>`, `<name>\n[<technology>]`, or `<name>`; never the description.
- [X] T021 [US1] Add the "Nodes" section to deployment view pages in `internal/adapters/html/theme/view.gohtml` (research R4): only when at least one group has a title, technology, description or tags; columns Node (path joined by ` / `), Title, Technology, Description, Tags; add the template helper that builds node paths (`nodeRows`, in `internal/adapters/html/theme.go` beside the other template functions); add an HTML golden for the `deployment-details` prod page under `internal/adapters/html/testdata/golden/`.
- [X] T022 [US1] Run the US1 tests and the full suite; build `testdata/projects/serverless-reference` and confirm its outputs are byte-identical to `main`; confirm `apply_edit` can set and remove each new attribute on a group and an instance with an MCP end-to-end case in `cmd/mcp_tools_test.go`, and that MCP `schema` with `block: "node"` lists the four attributes; add the `deployment-details` fixture to the canonical-format check in `internal/adapters/hclsource/format_canonical_test.go` (`fmt` leaves it unchanged).

**Checkpoint**: US1 works on its own; quickstart §1 passes.

---

## Phase 4: User Story 2 - Views that say what they show (Priority: P1)

**Goal**: views carry `title` and `description`; the site and the diagram canvas show the title, the
page shows the description (FR-005–FR-007, research R6, R7).

**Independent Test**: one view with a title and description in `deployment-details`; its page,
navigation and standalone SVG show the title; untitled views' D2 and SVG are byte-identical
(quickstart §2).

### Tests for User Story 2

- [X] T023 [P] [US2] Write failing tests: `authoring` refuses `set.title = ""` on `TargetView` (`internal/core/entities/authoring/render_attrs_test.go`); `resolve_views` gives a declared view `Title = title` and `TitleDeclared = true` when set, and `Title = name`, `TitleDeclared = false` otherwise (`internal/core/usecases/resolve_views_test.go`); `Emit` adds a `loko-view-title` text object with `near: top-center` only when `TitleDeclared`, and never emits the description (`internal/adapters/d2/emit_test.go`); a decoder golden case for a view with both attributes and one with `title = ""` in `internal/adapters/hclsource/testdata/golden/language_polish/`.

### Implementation for User Story 2

- [X] T024 [US2] Add `title` and `description` (string) to `view` in `internal/core/entities/arch/schemas.go`; add `Title`, `Description` to `ViewDecl` (`internal/core/entities/arch/source_model.go`) and IR `View` (`internal/core/entities/arch/ir_element.go`, `omitempty`); add a titled, described view to `testdata/projects/deployment-details/main.loko.hcl`.
- [X] T025 [P] [US2] Add `title`, `description` to `TargetView` in `internal/core/entities/authoring/edit_schema.go` and the empty-title rule for `TargetView` in `internal/core/entities/authoring/render_attrs.go`.
- [X] T026 [US2] Decode `title` (refused when empty) and `description` in the view decoder, `internal/adapters/hclsource/decode_project.go`; copy them to the IR in `internal/core/usecases/build_ir.go`.
- [X] T027 [US2] Set `Title = cmp.Or(d.Title, d.Name)`, `TitleDeclared = d.Title != ""` and `Description` for declared views in `internal/core/usecases/resolve_views.go`; add `Description string` (`json:"description,omitempty"`) and `TitleDeclared bool` (`json:"-"`) to `View` in `internal/core/entities/viewmodel/view.go`.
- [X] T028 [US2] Emit the canvas title in `internal/adapters/d2/emit.go` (research R7): a root-level `loko-view-title` object, `shape: text`, `near: top-center`, bold, larger font, label `wrapTitle(title)`, only when `TitleDeclared`; add D2 goldens for a titled view and confirm with an SVG test that both dagre and ELK render it above the diagram.
- [X] T029 [US2] Show `View.Description` under the page heading in `internal/adapters/html/theme/view.gohtml` (heading, `<title>`, navigation and `alt` already read `View.Title`); update the HTML golden for `deployment-details`; confirm every existing D2, SVG and HTML golden is unchanged; add an MCP end-to-end case in `cmd/mcp_tools_test.go` where `apply_edit` sets and removes `title` and `description` on a view, with preview, and refuses `title = ""` (FR-017); re-run the canonical-format check on the fixture.

**Checkpoint**: US1 and US2 work independently; quickstart §2 passes.

---

## Phase 5: User Story 3 - Docs links that survive moving files (Priority: P2)

**Goal**: `docs` resolves from the declaring file, the project root is a fallback with a hint, and
the use cases stop touching the file system (FR-008–FR-011, research R9, R12).

**Independent Test**: quickstart §3 on `testdata/projects/first-run`: 7 hints, `--strict` exits 0,
moving a file with its docs folder keeps the link.

### Tests for User Story 3

- [X] T030 [P] [US3] Write failing tests for the resolver in `internal/adapters/hclsource/docs_resolve_test.go` (with `t.TempDir()` trees), one case per row of contracts/language.md "docs resolution": `file`, `root`, `both`, `missing`, a path leaving the root (`missing`), and an element declared at the root (`file`, both candidates equal); `DocsPath` is root-relative with forward slashes; `DocsCandidates` holds the file-relative and root-relative locations.
- [X] T031 [P] [US3] Rewrite `TestDocsNotFoundStatesTheRule` and add cases in `internal/core/usecases/validate_warnings_test.go`, building `ElementDecl`s with `DocsVia` set (no files on disk): `root` → one `docs_root_relative` info whose detail names the file-relative path to write; `both` → one `docs_ambiguous` warning naming both locations; `missing` → `docs_not_found` naming both locations; `file` → nothing. This produces the two docs codes for `TestValidationRuleCoverage`.
- [X] T032 [P] [US3] Write a failing test in `internal/core/usecases/docs_path_test.go` (there is no `build_site_test.go`): prose is read by `DocsPath`, so an element whose `docs` is file-relative gets its document on the site.
- [X] T033 [P] [US3] Add an architecture test (in `internal/core/usecases/layering_test.go`, or the audit rule file if the audit already covers imports) that fails when any non-test file in `internal/core/usecases` imports `os`.

### Implementation for User Story 3

- [X] T034 [US3] Add `DocsVia` (string type with constants `DocsFile`, `DocsRoot`, `DocsBoth`, `DocsMissing`), and `DocsPath`, `DocsVia`, `DocsCandidates [2]string` on `ElementDecl`, in `internal/core/entities/arch/source_model.go`; change the `docs` Note in `commonElementAttrs` (`internal/core/entities/arch/schemas.go`) to "relative to the declaring file; the project root is a fallback".
- [X] T035 [US3] Implement the resolver in `internal/adapters/hclsource/docs_resolve.go` (research R9 table), stat-ing at most two candidates, both confined to the root; call it for every element with `docs` from `Load` and `LoadOverlay` in `internal/adapters/hclsource/decode.go`, resolving overlay files against disk for the prose.
- [X] T036 [US3] Rewrite `warnDocs` in `internal/core/usecases/validate_warnings.go` to emit from `DocsVia` (texts as research R11), delete `docsExist` and the `os` and `path/filepath` imports; set IR `Element.Docs` to `DocsPath` in `internal/core/usecases/build_ir.go`; read prose by `DocsPath` in `internal/core/usecases/build_site.go`.
- [X] T037 [US3] Rewrite the docs paths of `testdata/projects/first-run-remodelled/architecture/*.loko.hcl` in file-relative form (`docs/pricing.md`), leave `testdata/projects/first-run` as is, and regenerate `cmd/testdata/validate_summary_first_run.txt` after confirming the only diff is the seven `docs_root_relative` lines.
- [X] T038 [US3] Remove `docs_root_relative` and `docs_ambiguous` from `pendingCodes` in `TestValidationRuleCoverage`; run the full suite; confirm every example's export is byte-identical to `main` and that no example gains a `docs_not_found` warning (SC-003).

**Checkpoint**: US3 works; `internal/core/usecases` imports no `os`; quickstart §3 passes.

---

## Phase 6: User Story 4 - Hint when a view renders too wide (Priority: P3)

**Goal**: one `wide_view` information diagnostic per rendered view that is at least 3,000 px wide
and 2.5 times wider than tall, unless the author chose `direction` or `layout` (FR-012–FR-015,
research R10).

**Independent Test**: build `testdata/projects/first-run` to SVG: one hint for `system-pricing`;
with `layout = "elk"` on the project, none; exit codes unchanged (quickstart §4).

### Tests for User Story 4

- [X] T039 [P] [US4] Write failing table tests for `WideViewHints` in `internal/core/usecases/wide_view_test.go`: 3875×1441 with nothing chosen → one hint naming the size and ratio; 2999×900 → none; 3200×1300 (ratio 2.46) → none; a declared view with `direction` → none; a project with `layout` → none; a zero-size artifact → none; a declared view's detail suggests `direction` and `layout = "elk"` on the view, a generated view's suggests `layout = "elk"` on the project or declaring a view (contracts/cli-mcp.md).
- [X] T040 [P] [US4] Write a failing test in `internal/adapters/d2/svg_backend_test.go`: the SVG artifact's `Width` and `Height` equal the outer `viewBox` size of the rendered SVG; the D2 source artifact has zero size.
- [X] T041 [P] [US4] Write a failing test in `internal/core/usecases/build_artifacts_test.go`: hints are appended after rendering and removed by `BuildRequest.Ignore` and by the project's `ignore_warnings`; `ExitCode(true)` is unchanged by them; a build with only the D2 backend gives none.

### Implementation for User Story 4

- [X] T042 [US4] Add `Width`, `Height int` to `Artifact` in `internal/core/entities/viewmodel/artifact.go`; set them in `renderView` from the laid-out diagram's bounding box in `internal/adapters/d2/svg_backend.go` (cache the size with the SVG so a cache hit reports it too).
- [X] T043 [US4] Implement `WideViewHints` in `internal/core/usecases/wide_view.go` with named constants `wideViewMinWidth = 3000` and `wideViewMinRatio = 2.5`; it needs each view's kind, whether the declared view sets `direction` or `layout`, and whether the project sets `layout` (from the IR).
- [X] T044 [US4] Call `WideViewHints` in `internal/core/usecases/build_artifacts.go` after rendering, then apply `Without(slices.Concat(req.Ignore, projectIgnore))` to the hints; remove `wide_view` and the then-empty `pendingCodes` set from `TestValidationRuleCoverage`.
- [X] T045 [US4] Add an MCP end-to-end case in `cmd/mcp_build_test.go`: `build` on a copy of `testdata/projects/first-run` returns one `wide_view` diagnostic in compact form; `validate` with `severity: "warning"` hides the seven `docs_root_relative` lines.
- [X] T046 [US4] Build every example and fixture to SVG and confirm the hinted views are exactly those listed in research R10 (SC-004); record the list in the test or in `research.md` if it differs.

**Checkpoint**: all four stories work; quickstart §4 passes.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [X] T047 [P] Document the new attributes, the docs rule and the `info` severity in `docs/language.md` (node, instance, view sections; docs paths with a subdirectory example) and `docs/cli-reference.md` (`--severity warning`, information lines); note `wide_view` and `severity: "warning"` in `docs/mcp-integration.md`.
- [X] T048 [P] Write `docs/adr/0017-language-polish.md` (research R13): the information severity, docs resolution from the declaring file with a root fallback, the wide-view rule and its measured thresholds.
- [X] T049 [P] Add a `[Unreleased]` entry to `CHANGELOG.md` (Added: node, instance and view attributes, `info` severity, three codes, `--severity warning`; Changed: docs resolution and the `docs_root_relative` line).
- [X] T050 Update `docs/roadmap.md` (018 done, 019 next) and `specs/012-v1-architecture-dsl/roadmap.md` if it lists 018 as planned.
- [X] T051 Run `make test`, `make lint`, `make audit-constitution` and `go test -cover ./internal/core/...` (> 80%); confirm the determinism job and every export, D2 and HTML golden of a fixture without the new attributes is unchanged (SC-006); confirm exit codes of `validate` and `build`, with and without `--strict`, match `main` on every example and fixture (SC-005).
- [X] T052 Walk `specs/018-language-polish/quickstart.md` end to end against `./loko`; record any deviation in an "Implementation notes" section at the end of this file.

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: none.
- **Foundational (Phase 2)**: after Setup. US3 and US4 need it (the severity and codes).
- **US1 (Phase 3)** and **US2 (Phase 4)**: after Phase 2; independent of each other except that
  both edit `arch/schemas.go`, `authoring/edit_schema.go`, `authoring/render_attrs.go`,
  `d2/emit.go`, `html/theme/view.gohtml` and the `deployment-details` fixture, so run their
  implementation tasks one story at a time.
- **US3 (Phase 5)**: after Phase 2; independent of US1 and US2.
- **US4 (Phase 6)**: after Phase 2; uses the `first-run` fixture, so run after US3 (T037) to avoid
  regenerating the summary golden twice.
- **Polish (Phase 7)**: after the stories it documents.

### Within each story

Tests first and failing → entities → adapters and use cases → rendering → end-to-end check.

### Parallel opportunities

- Phase 2: T002–T005 together; T008 and T009 together after T006–T007.
- US1: T011–T013 together; T016 alongside T015.
- US2: T023 alone, then T025 alongside T024.
- US3: T030–T033 together.
- US4: T039–T041 together.
- Polish: T047–T049 together.
- With two people: one takes US1 → US2, the other US3 → US4, after Phase 2.

### Parallel example: User Story 3

```text
T030 resolver tests (hclsource/docs_resolve_test.go)
T031 warnDocs tests (usecases/validate_warnings_test.go)
T032 prose-by-DocsPath test (usecases/build_site_test.go)
T033 no-os layering test (usecases/layering_test.go)
```

---

## Implementation Strategy

### MVP first (User Story 1)

1. Phase 1 and Phase 2.
2. Phase 3 (US1): deployment diagrams that say what each node is: the report's largest remaining
   gap.
3. Stop and validate with quickstart §1.

### Incremental delivery

1. US1 → deployment nodes.
2. US2 → titled views.
3. US3 → docs paths (the one behaviour change; ships with ADR-0017).
4. US4 → the wide-view hint.

Each story is a reviewable commit on the feature branch; the release is cut once all four land.

---

## Implementation notes

Recorded while implementing (T052 walked quickstart.md against `./loko`; every scenario passes).

- **Quickstart §1** names `testdata/projects/serverless-reference`, which has no `node` blocks. §1
  was walked on `testdata/projects/deployment-details`, which has the report's account → region →
  VPC layout.
- **Quickstart §3.1** says `--strict` exits `0` on `first-run`. It exits `2`, as in 1.1.0: the
  fixture has 32 warnings. What holds, and what SC-005 requires, is that the exit code is unchanged
  by the seven information lines.
- **Node tag duplicates (T012).** Elements have no duplicate-tag diagnostic; their tags are sorted
  and de-duplicated in the IR. Node tags follow the same rule, so the decoder golden has no tag
  diagnostic.
- **Empty titles (T011, T017, T023, T026).** As for elements, `empty_title` is a core rule
  (`ValidateRenderAttributes`), not a decoder one; the decoder records the attribute range. The
  `language_polish` decoder golden therefore expects no parse diagnostics.
- **Node label with a title but no technology** (not covered by R3): `<title>` over `<name>`.
- **`nodeRows`** lives in `internal/adapters/html/theme.go` beside the other template functions
  (T021); the prose-by-path test is `internal/core/usecases/docs_path_test.go` (T032).
- **Machine-readable counts.** The diagnostics JSON `summary`, and `validate --summary --format
  json`, gain an `info` count only when it is above zero, so output without information is
  unchanged. The diagnostics schema allows it.
- **`wide_view` addresses.** A hint for a declared view carries the view's address and range; one
  for a generated view has neither and names the view in its detail.
- **Hinted views (T046)** are exactly research R10's list: microservices `landscape` and
  `deployment-prod`, serverless-reference `deployment-prod` and `deployment-dev`, and
  `system-pricing` in both first-run fixtures. No exit code of any example or fixture changed.
- **Coverage (T051).** `internal/core/usecases` 90.2%, `authoring` 89.8%, `viewmodel` 96.3%;
  `entities/arch` is 48.7% (48.5% on `main`), below the 80% target before this feature.
