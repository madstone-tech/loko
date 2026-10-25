---

description: "Task list for 017-mcp-validation-ergonomics"
---

# Tasks: MCP & Validation Ergonomics

**Input**: Design documents from `/specs/017-mcp-validation-ergonomics/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md),
[data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Test tasks are **included and mandatory**, because Constitution Principle V (Test-First)
is non-negotiable. A test task always comes before the implementation it covers: run it and
**confirm it fails** before implementing.

**Organization**: Phases follow the user stories in priority order. US1 (language description) and
US2 (warnings) are P1; US3 (build and fmt through MCP) and US4 (empty-project bootstrap) are P2.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1–US4 from spec.md
- Every task names its exact file paths

## Standing rules for every task

- **No unrequested change.** A project that uses none of the new attributes, and a CLI run that uses
  none of the new flags, produces byte-identical output and exports to 1.0.2 (SC-007). The one
  deliberate change is compact MCP diagnostics (FR-013, research R7).
- **Layering.** Entities stay stdlib-only and do not import each other: `authoring` keeps copies
  held in step by parity tests (research R1). `internal/mcp` imports no entity package and no
  render adapter; the `build` tool's backends are wired in `cmd/wiring.go` (research R8).
- **Thin handlers.** MCP tool functions ≤ 30 effective lines, CLI handler functions ≤ 50, use-case
  functions ≤ 60, use-case and entity files ≤ 200 (`arch/schemas.go` is exempt as pure data).
- **Goldens.** Any golden change is a review event: regenerate with `-update` only after confirming
  the diff is exactly the intended change.
- **Audit.** Run `make audit-constitution` after every phase.

---

## Phase 1: Setup

**Purpose**: the fixture that measures SC-002 to SC-004, and the decision record.

- [X] T001 Create `testdata/projects/first-run/` reproducing the shape of the reported model ([first-run-report.md](./first-run-report.md) #2): one system with 12 containers, 3 people and 6 third-party systems declared as `system` (as the reporter did), 8 instances with no binding, and about 25 relationships. Split it across `architecture/workspace.loko.hcl` (project), `architecture/context.loko.hcl` and `architecture/containers.loko.hcl`. Record in `testdata/projects/first-run/README.md` the 1.0.2 baseline: run `go run . validate -p testdata/projects/first-run` and note the warning count by code (it must be about 33)
- [X] T002 [P] Create `testdata/projects/first-run-remodelled/`: the same model with the third-party systems as `external` and the 8 unbuilt instances marked `planned = true` (every fixture is built by `TestBuildDeterministic`, so the `planned` lines land with T026, when the attribute exists). It is the SC-002 target
- [X] T003 [P] Write `docs/adr/0016-mcp-ergonomics.md` from `docs/adr/template.md`: one language table in `arch` with parity-tested copies (research R1); compact MCP diagnostics as the default and why this 1.x output change is acceptable (clarification Q1, R7); `project.ignore_warnings` and `instance.planned` as permanent attributes (Q2, R4, R5, R12); the `project` edit target instead of an `init` tool (Q3, R10)

---

## Phase 2: Foundational (blocking)

**Purpose**: one language definition that the decoder uses and the edit schema is checked against.
No behaviour change: every existing golden stays byte-identical.

**⚠️ CRITICAL**: US1 serves this table and US2 and US4 add attributes to it.

- [X] T004 Write `internal/core/entities/arch/schemas_test.go`: `Language` is sorted by block name; block and attribute names are unique; every `In` entry is `file` or another block's name; every attribute `Type` is one of "`string`, `list(string)`, `ref`, `list(ref)`, `map`, `bool`"; `layout` lists `arch.Layouts`, `direction` lists `arch.Directions`, `shape` lists `arch.Shapes` with `Kinds` "`container`, `external`", `uses.kind` lists `arch.RelationshipKinds`; `uses.target` and `instance.of` are `Required`; the `binding` block's `LabelValues` equal `arch.AllClaimKinds` ("`terraform`, `cloudformation`") and every other block's `LabelValues` is empty; `docs` has a `Note` saying it is relative to the project root; `LanguageBlockNamed("view")` finds the block and `LanguageBlockNamed("sequence")` does not
- [X] T005 Implement `internal/core/entities/arch/schemas.go` (pure data, exempt from the file-size budget): `LanguageBlock` (with `LabelValues`, data-model.md), `LanguageAttr`, `var Language` covering every block in `hclsource/schema.go` (`project`, `person`, `system`, `container`, `component`, `external`, `deployment`, `node`, `instance`, `binding`, `uses`, `view`, `reconcile`, `moved`, `locals`) with the attributes the decoder accepts today, one canonical minimal `Example` each, and `LanguageBlockNamed`. Pass T004
- [X] T006 Write `internal/adapters/hclsource/language_test.go`: for every block in `arch.Language`, the decoder accepts exactly that block's attribute names (an extra name gives `unknown_attribute`, as today); every `Example` parses and decodes with no error diagnostics. Then derive `commonElementAttrs`, `usesAttrs`, `projectAttrs`, `deploymentAttrs`, `instanceAttrs`, `bindingAttrs`, `viewAttrs` and `reconcileAttrs` in `internal/adapters/hclsource/schema.go` from `arch.Language` instead of literal lists. Run the full `hclsource` golden suite: unchanged
- [X] T007 [P] Write `internal/core/usecases/language_parity_test.go` (`TestLanguageParity`, next to `naming_parity_test.go`): for each `authoring` edit target, the legal attributes, their value shapes and their allowed values match the corresponding `arch.Language` block(s). Use this mapping: `element` → `person`, `system`, `container`, `component`, `external`, each compared with the common element attributes plus that kind's parent attribute (`system` on `container`, `container` on `component`) and with `shape` only on the kinds in its `Kinds`; `relationship` → `uses`; `environment` → `deployment`; `group` → `node`; `instance`, `binding` and `view` → the block of the same name; `project` → `project` once T049 adds the target (skip it until then). Merge `legalAttrs` and `enumAttrs` before comparing. Blocks with no edit target (`reconcile`, `moved`, `locals`) are not compared. Expose what the test needs from `internal/core/entities/authoring/edit_schema.go` as a read-only accessor (for example `LegalAttrs(TargetKind) map[string]ValueKind`), not by exporting the maps. It must pass unchanged today; then delete one attribute from `arch.Language` locally and confirm it fails

**Checkpoint**: `make test` green, goldens unchanged, audit clean.

---

## Phase 3: User Story 1 - Learn the language from the server (Priority: P1) 🎯 MVP

**Goal**: a client asks the server for the language before editing (FR-001 to FR-006).

**Independent Test**: start `loko mcp`, call `schema`, and compare the answer with `docs/language.md`
(quickstart §1).

### Tests for User Story 1

- [X] T008 [P] [US1] Write `internal/core/entities/authoring/edit_address_test.go` cases for `AddressForms`: one form per `TargetKind` that `apply_edit` accepts; every `Example` parses with the existing address parser to that target; the `instance` form's pattern is "`deployment.<env>.node.<node>[.node.<node>…].instance.<name>`"
- [X] T009 [P] [US1] Write `internal/core/usecases/describe_language_test.go`: `DescribeLanguage("")` returns every block, every address form and the `docs` rule; `DescribeLanguage("view")` returns only the `view` block; `DescribeLanguage("sequence")` returns a refusal with reason `unknown_block` and the block names as suggestions; the result encodes byte-identically across 20 calls (FR-006)
- [X] T010 [P] [US1] Write `internal/mcp/tools/schema_test.go`: `schema` is listed with a read-only description; with no input it returns blocks, addresses and docs; `block: "instance"` returns one block; the TOON encoding of the full answer is ≤ 3,000 tokens and of one block ≤ 500 (count tokens the way `cmd/authoring_perf_test.go` does)
- [X] T011 [P] [US1] Extend `internal/mcp/server_test.go`: the `initialize` result has an `instructions` string that names `schema`, `external` and `preview: true`, and is under 80 words

### Implementation for User Story 1

- [X] T012 [US1] Add `AddressForm` and `var AddressForms` to `internal/core/entities/authoring/edit_address.go`, to pass T008
- [X] T013 [US1] Implement `internal/core/usecases/describe_language.go`: `LanguageResult`, the DTO views `LanguageBlockView`, `LanguageAttrView` and `AddressFormView` (with `json` and `toon` tags, `omitempty` on `required`, `values`, `note` and `kinds`), and `DescribeLanguage(block string)`, to pass T009
- [X] T014 [US1] Implement `internal/mcp/tools/schema.go` (`SchemaTool`) and add `SchemaSchema` to `internal/mcp/tools/schemas.go` (`block` string, `format`), to pass T010
- [X] T015 [US1] Add `instructions` to `handleInitialize` in `internal/mcp/server.go` with the text in contracts/mcp-tools.md, to pass T011
- [X] T016 [US1] Register the `schema` tool in `newMCPTools` in `cmd/wiring.go`, update its doc comment's tool count, and update `internal/mcp/tools/read_tools_test.go` or `server_order_test.go` wherever the tool list is asserted

**Checkpoint**: quickstart §1 passes; SC-006 is enforced by T006 and T007.

---

## Phase 4: User Story 2 - Warnings that point at real problems (Priority: P1)

**Goal**: planned instances, suppression by code, summary and severity filters, compact MCP
diagnostics (FR-007 to FR-014).

**Independent Test**: quickstart §2 against `testdata/projects/first-run` and
`first-run-remodelled`.

### Tests for User Story 2

- [ ] T017 [P] [US2] Extend `internal/core/entities/arch/diagnostic_test.go`: `WarningCodes` is exactly the warning subset of `AllCodes`, in order, and includes `planned_instance_bound` and `unknown_warning_code`; `CheckIgnore([]string{"missing_docs"})` is nil; `CheckIgnore` with `syntax_error` or `no_such_code` returns an error whose message lists the warning codes; `Diagnostics.Without([]string{"missing_docs"})` removes those warnings, never removes an error even if its code is passed, and returns the removed count. Update `TestAllCodesAreKnown`
- [ ] T018 [P] [US2] Add parse golden fixtures under `internal/adapters/hclsource/testdata/golden/`: `planned_instance/` (`planned = true` decodes to `InstanceDecl.Planned` with its range), `planned_wrong_type/` (`planned = "yes"` → `invalid_attribute_value`), `ignore_warnings/` (`ignore_warnings = ["missing_docs"]` decodes to `ProjectDecl.IgnoreWarnings` with its range) and `ignore_warnings_wrong_type/` (a string, not a list → `invalid_attribute_value`)
- [ ] T019 [P] [US2] Extend `internal/core/usecases/validate_warnings_test.go`: a planned instance with no claims gives no `unbound_instance`; a planned instance with a claim gives `planned_instance_bound` at its `planned` range; an unplanned instance without claims still gives `unbound_instance`
- [ ] T020 [P] [US2] Write `internal/core/usecases/compile_ignore_test.go`: `CompileRequest.Ignore` and `ProjectDecl.IgnoreWarnings` both suppress; the two lists merge; an error code in `ignore_warnings` gives `invalid_attribute_value`; an unknown code there gives `unknown_warning_code` naming it, and ignoring `unknown_warning_code` suppresses that too; `ExitCode(true)` is `ExitSuccess` when every warning is ignored; an error code in `CompileRequest.Ignore` is refused before compiling. Add both new codes to `codesFromCoreTests` in `rule_coverage_test.go` and to `specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`
- [ ] T021 [P] [US2] Write `internal/core/usecases/validate_filter_test.go`: `Summarize` groups by code, errors first then by code, addresses sorted and capped at 20 with `More` counting the rest; `Severity: "error"` keeps only errors and leaves `ExitCode` unchanged; `BriefDiagnostics` gives `At` as `file:line`, or `file` alone for a project-wide diagnostic; `ValidateResult.Ignored` counts suppressed warnings; `Summarize` of no diagnostics returns no groups
- [ ] T022 [P] [US2] Extend `cmd/validate_test.go`: `--summary` text output matches a golden for `testdata/projects/first-run` and is ≤ 30 lines (SC-003); `--severity error` hides warnings with the same exit code; `--strict --ignore missing_docs` exits `0` when only `missing_docs` warnings remain and `2` without the flag; `--ignore syntax_error` exits `1` listing the warning codes; `--ignore a,b` and repeated `--ignore` both work; `--summary` on a project with no diagnostics prints exactly one line saying so. Extend `cmd/build_test.go`: `build --strict` and `validate --strict` exit the same with the same `--ignore`
- [ ] T023 [P] [US2] Extend `internal/mcp/tools/read_tools_test.go` for `validate`: `summary`, `severity` and `ignore` inputs; by default each diagnostic has `at` and no `range`; `ranges: "full"` returns the 1.0.2 shape. Add `TestCompactDiagnosticsSize`: the default TOON response for `testdata/projects/first-run` is ≤ 50% of the same call with `ranges: "full"` (SC-004)
- [ ] T024 [P] [US2] Add `TestRemodelledWarnings` to `cmd/validate_test.go`: `testdata/projects/first-run-remodelled` compiles with no errors and ≤ 10 warnings, none of them `empty_system`, `missing_docs` on an external, or `unbound_instance` (SC-002)

### Implementation for User Story 2

- [ ] T025 [US2] In `internal/core/entities/arch/diagnostic.go` add `CodePlannedInstanceBound` and `CodeUnknownWarningCode` (to `AllCodes`), `WarningCodes`, `CheckIgnore` and `Diagnostics.Without`, to pass T017
- [ ] T026 [US2] Add `planned` to the `instance` block and `ignore_warnings` (type `list(string)`, with a `Note` naming `WarningCodes`) to the `project` block in `internal/core/entities/arch/schemas.go`; add `ProjectDecl.IgnoreWarnings`/`IgnoreWarningsRange` to `source_model.go`, and `InstanceDecl.Planned`/`PlannedRange` to `source_model_deployment.go`, and the IR instance's `Planned` (`json:"planned,omitempty"`) to `internal/core/entities/arch/ir_deployment.go`. Decode them in `internal/adapters/hclsource/decode_project.go` and `decode_deployment.go`, to pass T018. Add `planned = true` to the 8 instances in `testdata/projects/first-run-remodelled/architecture/deployment.loko.hcl`. Carry `Planned` into the IR in `internal/core/usecases/build_ir_deployment.go`; confirm `export_v1.json` is byte-identical
- [ ] T027 [US2] Add `planned` (`ValueBool`, allowed as a top-level value for this attribute only) to `TargetInstance` in `internal/core/entities/authoring/edit_schema.go`, with a test in `internal/core/entities/authoring/edit_schema_test.go` (or the existing schema test file) that `set: {planned: true}` is accepted and `planned: "yes"` is refused. `TestLanguageParity` (T007) must pass
- [ ] T028 [US2] In `internal/core/usecases/validate_warnings.go`, skip `unbound_instance` for a planned instance and emit `planned_instance_bound` when a planned instance has claims, to pass T019
- [ ] T029 [US2] Add `Ignore []string` to `CompileRequest` in `internal/core/usecases/compile_architecture.go`: refuse invalid per-run codes with `arch.CheckIgnore`; check `ProjectDecl.IgnoreWarnings` (error code → `invalid_attribute_value`, unknown → `unknown_warning_code`) in a new `internal/core/usecases/validate_project_block.go`; apply `Without` with the merged list as the last step, to pass T020
- [ ] T030 [US2] Add `ValidateRequest` (`Ignore`, `Severity`, `Summary`, `FullRanges`) to `internal/core/usecases/validate_project.go` and change `Validate(ctx, deps, req)`; implement `DiagnosticGroup`, `Summarize`, `DiagnosticBrief` and `BriefDiagnostics` in a new `internal/core/usecases/validate_filter.go`; add `Ignored`, `Summary` and `Brief` to `ValidateResult`, to pass T021. Update the existing callers of `Validate`
- [ ] T031 [US2] Add `--summary`, `--severity` and `--ignore` to `cmd/validate_cobra.go` and the text summary renderer to `cmd/validate.go` (layout in contracts/cli.md); add `--ignore` to `cmd/build_cobra.go` and `cmd/serve_cobra.go`, passed through `BuildRequest.Ignore` in `internal/core/usecases/build_artifacts.go`, to pass T022
- [ ] T032 [US2] Add `ranges` handling shared by every MCP tool to `internal/mcp/tools/helpers.go` (compact via `usecases.BriefDiagnostics` unless `ranges: "full"`) and apply it in `validate.go`, `describe.go`, `apply_edit.go` and `move.go`; add `summary`, `severity`, `ignore` and `ranges` to the input schemas in `internal/mcp/tools/schemas.go`, to pass T023. Update the existing MCP tests that assert `range`
- [ ] T033 [US2] Run T024 and fix any remaining gap in the remodelled fixture or the code until it passes; record the final warning count, by code, in `testdata/projects/first-run/README.md` next to the baseline

**Checkpoint**: quickstart §2 passes; SC-002, SC-003 and SC-004 are asserted by tests.

---

## Phase 5: User Story 3 - Render and format without a shell (Priority: P2)

**Goal**: MCP `build` and `fmt` (FR-015 to FR-019).

**Independent Test**: quickstart §3, through MCP only.

### Tests for User Story 3

- [ ] T034 [P] [US3] Write `internal/adapters/outputdir/confine_test.go`: `Store.Confine(ctx, root, "dist")` and `"out/site"` resolve inside the root; `"../x"`, an absolute path outside the root, and a symlink leading outside are refused with `path_refused`; `""` gives `dist`. Use a relative root `"."` (via `t.Chdir`, not parallel) as well as an absolute one, the case 1.0.2 fixed in `checkPath`
- [ ] T035 [P] [US3] Write `internal/adapters/hclsource/format_test.go` cases for `Canonical`: it returns, for each non-canonical file, its path, current bytes and canonical bytes, sorted by path; canonical files are absent; a file with a syntax error is reported in diagnostics and absent from the list; nothing is written
- [ ] T036 [P] [US3] Write `internal/core/usecases/format_plan_test.go` with a concrete `SourceFormatter` and `SourceEditor` stub: `Format(ctx, base, true)` returns diffs and writes nothing; `Format(ctx, base, false)` commits the plan; a stale `base` gives `stale_revision`; no changes gives `ok` with no diffs; syntax errors give diagnostics and no write
- [ ] T037 [P] [US3] Write `internal/core/usecases/build_in_project_test.go` with concrete `ArtifactStore` and `Backend` stubs: `BuildInProject` calls `Store.Confine` first and stops with `path_refused` when it refuses; `PickInlineView` returns the named view's artefact in the requested format, or an `unknown_view` refusal listing the view names; under `Strict` with warnings the result is `OK: false` and the artefacts were still committed (as `loko build --strict`, research R8). Then write `internal/mcp/tools/build_test.go` over `testdata/projects/two-systems` in a temp copy: `build` with `formats: ["d2","svg"]` writes the same files, byte for byte, as `usecases.BuildArtifacts` with the same request (and as `loko build --format d2,svg`, via `cmd/build_test.go` if simpler); `inline` returns the named view's D2, or SVG with `inline_format: "svg"`, equal to the file on disk; an unknown `inline` gives `unknown_view` with the view names; `out: "../x"` gives `path_refused` and writes nothing; a model with an unresolved reference gives `ok: false` and changes no file; `strict` with a warning writes the files and gives `ok: false`; `ignore` behaves as in `validate`; a file the user placed in the output directory survives the build (FR-017)
- [ ] T038 [P] [US3] Write `internal/mcp/tools/fmt_test.go`: a misformatted file with `preview: true` returns a diff and leaves the file unchanged; without `preview` the file equals `loko fmt` output; a stale `base_revision` gives `stale_revision`; a running `apply_edit` and `fmt` on the same service serialise (no lost write)

### Implementation for User Story 3

- [ ] T039 [US3] Add `Confine(ctx, root, outDir string) (string, error)` to `ArtifactStore` in `internal/core/usecases/ports.go` and implement it in a new `internal/adapters/outputdir/confine.go`: make `root` absolute before resolving symlinks (the 1.0.2 `checkPath` lesson), walk up to the nearest existing directory, refuse with `path_refused` anything outside the root. Update every `ArtifactStore` test double. Pass T034. No use case touches the file system for this (Principle II)
- [ ] T040 [US3] Add `Canonical(ctx, root) ([]authoring.FileContent, arch.Diagnostics, error)` to `SourceFormatter` in `internal/core/usecases/ports.go` and implement it in `internal/adapters/hclsource/format.go`, sharing the canonicalising code with `Format` and `Check`, to pass T035. Update any `SourceFormatter` test doubles
- [ ] T041 [US3] Implement `AuthoringService.Format(ctx, base, preview)` in a new `internal/core/usecases/format_plan.go`, returning `EditResult` and queuing on the service's mutex like `Apply`; add `Formatter SourceFormatter` to `AuthoringDeps` in `authoring_service.go` and wire `hclsource.New()` for it in `newAuthoringDeps` in `cmd/wiring.go`, to pass T036
- [ ] T042 [US3] Implement `internal/mcp/tools/fmt.go` (`FmtTool`) with `FmtSchema` (`base_revision` required, `preview`, `ranges`, `format`) in `schemas.go`, to pass T038
- [ ] T043 [US3] Implement `BuildInProject`, `ProjectBuildRequest`, `ProjectBuildResult` and `PickInlineView` in a new `internal/core/usecases/build_in_project.go` (data-model.md), confining through `deps.Store.Confine` and taking `BuildRequest.Ignore` from T031. Then implement `internal/mcp/tools/build.go` (`BuildTool`), taking a `usecases.BuildDeps` and the project root, which only parses input, calls `BuildInProject` and encodes the result (no view lookup or refusal logic in the handler, Principle III); add `BuildSchema` (`formats`, `out`, `inline`, `inline_format` "`d2` (default) or `svg`", `strict`, `ignore`, `ranges`, `format`) to `schemas.go`, to pass T037. The handler stays ≤ 30 effective lines
- [ ] T044 [US3] Register `build` (with `newBuildDeps()`) and `fmt` in `newMCPTools` in `cmd/wiring.go`. Confirm `make audit-constitution` reports no import of a render adapter from `internal/mcp/**`

**Checkpoint**: quickstart §3 passes.

---

## Phase 6: User Story 4 - Start a project from nothing (Priority: P2)

**Goal**: the `project` edit target and the `describe` bootstrap (FR-020 to FR-023).

**Independent Test**: quickstart §4, from an empty directory, through MCP only.

### Tests for User Story 4

- [ ] T045 [P] [US4] Extend `internal/core/entities/authoring` tests (`edit_test.go` or `edit_schema_test.go`): `TargetProject` with address `project`; `add` requires `name` and `file`; `update` accepts "`description`, `layout`, `loko_version`, `ignore_warnings`" and rejects `name`; `remove` and `rename` are refused; `layout` outside `dagre`, `elk` is refused with the set
- [ ] T046 [P] [US4] Extend `internal/adapters/hclsource/edit_blocks_test.go`: `add project` in an empty root creates the named file with exactly one canonical `project` block; `add project` when one exists is refused `project_exists` naming its file; adding `person.user` with no `project` block is refused `no_project`; `update project` sets `ignore_warnings` in place
- [ ] T047 [P] [US4] Extend `internal/core/usecases/describe_test.go`: with no source, `DescribeResult` has `OK: false`, the `no_source_found` diagnostic and `Bootstrap` with the syntax `project "name" {\n  description = "..."\n}\n` and a request whose `base_revision` is the current revision and whose single edit is `add project`; with source, `Bootstrap` is nil
- [ ] T048 [P] [US4] Write `internal/mcp/tools/bootstrap_test.go` (end to end on an empty temp dir): `describe` → send `bootstrap.request` with a name and `preview: true` → a diff creating one file → send without `preview` → add a person and a system → `validate` has no errors → `build` writes a site (SC-005). Also: two services on the same root both send the creation with the same `base_revision`; the second gets `stale_revision`

### Implementation for User Story 4

- [ ] T049 [US4] Add `TargetProject`, its legal attributes and op rules, the `project` `AddressForm`, and the refusal reasons `project_exists` and `no_project` to `internal/core/entities/authoring/edit.go`, `edit_schema.go` and `edit_address.go`, to pass T045. `TestLanguageParity` must pass
- [ ] T050 [US4] Handle the `project` target in `internal/adapters/hclsource/edit_blocks.go` (placement: the edit's `file`, through `checkPath`; `project_exists` naming the declaring file) and refuse element adds with `no_project` when no `project` block exists, carrying the bootstrap request as the refusal's suggestion, to pass T046
- [ ] T051 [US4] Add `BootstrapHint` and `DescribeResult.Bootstrap` to `internal/core/usecases/describe_types.go` and `describe.go`, set only when the diagnostics contain `no_source_found`, to pass T047; carry it through `internal/mcp/tools/describe.go`, and run T048

**Checkpoint**: quickstart §4 passes; SC-005 is asserted by T048.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [ ] T052 [P] Document in `docs/language.md`: `ignore_warnings` in the `project` table, `planned` in the `instance` table, and the two new codes in the diagnostics reference (contracts/language.md)
- [ ] T053 [P] Document in `docs/mcp-integration.md`: the `schema`, `build` and `fmt` tools, the `project` target, `bootstrap`, the `ranges` input and the compact default, and the handshake instructions (contracts/mcp-tools.md)
- [ ] T054 [P] Document in `docs/cli-reference.md`: `validate --summary`, `--severity`, `--ignore`, and `build`/`serve --ignore` (contracts/cli.md)
- [ ] T055 [P] Add the feature to `CHANGELOG.md` under `[Unreleased]`, with a **Changed** entry for compact MCP diagnostics and how to restore full ranges; move semantic diff, observation adapters and policy to 020–022 and add 017–019 in `docs/roadmap.md` and `specs/012-v1-architecture-dsl/roadmap.md`
- [ ] T056 Re-run `TestAuthoringPerformance` (`cmd/authoring_perf_test.go`) and the build performance tests: the existing budgets still hold with suppression and the new tools
- [ ] T057 Run [quickstart.md](./quickstart.md) §1–4 by hand against `./bin/loko` (`make build`), and record any deviation in this file before closing the feature. SC-001 (no refused edits caused by a wrong guess about the language) is checked here, by hand: re-model `testdata/projects/first-run` from an empty directory through an MCP client, guided only by `schema`, and record the number of `apply_edit` refusals
- [ ] T058 Full gate: `make test` (with `-race`), `make lint`, `make audit-constitution`; `go test -cover ./internal/core/...` reports more than 80% (constitution quality gate); the determinism job's command locally; `git diff` of every golden is empty except the deliberate MCP diagnostics changes

---

## Dependencies & Execution Order

```
Phase 1 ──▶ Phase 2 ──┬──▶ Phase 3 US1 (schema) ─────────┐
                      ├──▶ Phase 4 US2 (warnings) ───────┼──▶ Phase 7
                      ├──▶ Phase 5 US3 (build, fmt) ─────┤
                      └──▶ Phase 6 US4 (bootstrap) ──────┘
```

- **Phase 2 blocks everything**: the language table must exist and the decoder must use it first.
- **US1 and US2 are independent**, but both edit `internal/core/entities/arch/schemas.go` (T005,
  T026) and `cmd/wiring.go`: do T026 after T005, and register tools in order (T016, T044).
- **US3 uses T032's `ranges` helper** for its tools' diagnostics: do T032 before T042 and T043, or
  have them return full ranges first and switch when T032 lands.
- **US4 needs T029** if it tests `update project` with `ignore_warnings`; the rest of US4 only
  needs Phase 2.
- **T024 and T033 need T026 to T030**, because the remodelled fixture uses `planned`.

## Parallel Examples

```
Phase 1: T002 remodelled fixture | T003 ADR
Phase 2: T007 parity test (after T005) alongside T006 decoder derivation
US1 tests: T008 address forms | T009 DescribeLanguage | T010 schema tool | T011 handshake
US2 tests: T017 codes | T018 parse goldens | T019 planned | T020 ignore | T021 filters | T022 CLI | T023 MCP | T024 remodel
US3 tests: T034 confine | T035 Canonical | T036 Format plan | T037 build tool | T038 fmt tool
US4 tests: T045 target | T046 edit blocks | T047 describe | T048 end to end
Polish: T052 language doc | T053 MCP doc | T054 CLI doc | T055 changelog and roadmaps
```

## Implementation Strategy

1. **MVP = Phases 1–3 (US1).** A client can learn the language from the server: the friction that
   cost the most turns in the report. Validate with quickstart §1.
2. **US2 next (P1).** Fewer, better warnings and smaller MCP responses; quickstart §2 and the
   SC-002 to SC-004 tests.
3. **US3 and US4 (P2)** are independent increments: shell-less rendering, then the empty-directory
   start. Together they make SC-005 true.
4. **Polish**: docs, changelog, roadmaps, budgets, the full gate.

Commit after each task or logical group. Golden changes are review events.
