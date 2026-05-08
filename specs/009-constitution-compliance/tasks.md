---
description: "Task list for feature 009 — Constitution Compliance Refactor"
---

# Tasks: Constitution Compliance Refactor

**Input**: Design documents from `/specs/009-constitution-compliance/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: Test tasks ARE included — Constitution Principle V (Test-First) applies to all new code in this feature. Existing tests for refactored files must keep passing without weakening or removal (FR-016, FR-017).

**Organization**: Tasks are grouped by user story per the spec's P1→P3 priorities. The audit tool is built in the Foundational phase because every user story's verification depends on it (per spec "Independent Test" wording).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Maps to user stories from spec.md (US1–US5)
- File paths are repo-root-relative

## Path Conventions

- Repository root: `/Users/andhi/code/mdstn/loko/`
- Audit binary: `tools/archcheck/`
- Rule set (canonical): `specs/009-constitution-compliance/contracts/structural-rules.yaml`
- Constitution: `.specify/memory/constitution.md` (v1.1.0, amended 2026-05-08)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project scaffolding, ADR, and toolchain skeleton needed by every later phase.

- [x] T001 Create the directory `tools/archcheck/` at repo root with an empty `.gitkeep` file
- [x] T002 [P] Create ADR document at `docs/adr/0009-constitution-compliance-tooling.md` recording the decisions from `specs/009-constitution-compliance/research.md` (custom Go AST tool, depguard pairing, exemption model, constitution v1.1.0 amendment)
- [x] T003 [P] Add `audit-constitution` placeholder target to `Makefile` that runs `go run ./tools/archcheck --rules=specs/009-constitution-compliance/contracts/structural-rules.yaml`
- [x] T004 [P] Add `audit-constitution-watch` Makefile target that wraps `audit-constitution` with `entr` or fsnotify-based re-run on `.go` save (best-effort dev convenience)
- [x] T005 [P] Verify `.specify/memory/constitution.md` is at v1.1.0 with the file footer reading `**Version**: 1.1.0 | **Ratified**: 2026-02-06 | **Last Amended**: 2026-05-08`; if not, re-amend per `specs/009-constitution-compliance/research.md` R6
- [x] T006 [P] Create `.golangci.yml` at repo root if absent, or read existing config and add a top-level `linters:` enabling `depguard`. Leave the `depguard` rule body empty for now (filled in T038 under US2)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Build the structural-compliance audit tool (`tools/archcheck`) that every user story uses to verify completion. The tool exposes file-size, function-size, and layer-import checks driven by the YAML rule set.

**⚠️ CRITICAL**: No user story phase can be fully verified until this phase is complete.

### Audit-tool data model and scaffolding

- [ ] T007 Create `tools/archcheck/main.go` with a stub `main()` that loads `--rules=<path>` and `--format=text|json`, walks the module's `.go` files using `golang.org/x/tools/go/packages`, and prints `archcheck: 0 violations`
- [ ] T008 [P] Create `tools/archcheck/types.go` with the Go structs `LayerRule`, `FileSizeRule`, `FunctionSizeRule`, `Exemption`, `Violation`, `Report` exactly as specified in `specs/009-constitution-compliance/data-model.md`
- [ ] T009 [P] Create `tools/archcheck/rules_loader.go` with `LoadRules(path string) (*RuleSet, error)` that parses the YAML rule file (`gopkg.in/yaml.v3`), validates required fields per `data-model.md` "Validation summary", and refuses to run on incomplete rule sets

### Effective-line counter

- [ ] T010 Create `tools/archcheck/lines.go` with `func CountEffectiveLines(src []byte, fromPos, toPos token.Pos) int` that reproduces the rule from `scripts/audit-constitution.sh`: drop blank lines, single- and multi-line comments, the `package` declaration, and `import (...)` blocks; works at any token range (whole file or function body)
- [ ] T011 [P] Create `tools/archcheck/lines_test.go` with table-driven tests covering: blank-line skipping, `// comment`, `/* block */`, multi-line block comments, single-line `import "..."`, `import (...)` block (single and multi-statement), `package` line, mixed cases. Each test asserts the count matches the legacy shell behaviour on the same input

### File-size checker

- [ ] T012 Create `tools/archcheck/filesize.go` with `func CheckFileSize(file *ParsedFile, rules []FileSizeRule, exemptions []Exemption) []Violation` that matches files against `pathPattern` (using `path/filepath.Match` extended for `**`) and returns `Violation{Kind:"file-size",...}` for each over-budget file
- [ ] T013 [P] Create `tools/archcheck/filesize_test.go` with fixture files exercising: under-budget pass, exact-budget pass (200 = allowed), over-budget fail, exempt basename, `*_cobra.go` exempt, `*_test.go` exempt, generated-header exempt

### Function-size checker

- [ ] T014 Create `tools/archcheck/funcsize.go` with `func CheckFunctionSize(file *ParsedFile, rules []FunctionSizeRule, exemptions []Exemption) []Violation` that walks `*ast.FuncDecl` nodes, computes effective lines between `funcDecl.Pos()` and `funcDecl.End()`, and emits violations naming the function
- [ ] T015 [P] Create `tools/archcheck/funcsize_test.go` covering: short function (pass), exactly-at-budget function (pass at 50), over-budget function (fail), function with comment-heavy body (effective count drops), method on type (still detected), top-level vs nested closure (only top-level FuncDecls counted), `*_test.go` exempt

### Layer-import checker

- [ ] T016 Create `tools/archcheck/layer.go` with `func CheckLayerImports(file *ParsedFile, rules []LayerRule) []Violation` that resolves each import path against the project's module path, classifies by `pathPattern` (first match wins), and verifies the import is in `allowedImports` and not in `forbiddenImports`. External (non-module) imports unconditionally allowed
- [ ] T017 [P] Create `tools/archcheck/layer_test.go` covering: entities importing usecases (FAIL), usecases importing adapters (FAIL), adapters importing mcp (FAIL), mcp importing api (FAIL), `cmd/` importing entities directly (FAIL via `forbiddenImports` override), `cmd/` importing adapters/mcp/api (PASS), all layers importing stdlib (PASS), all layers importing external module (PASS). Note: mcp and api MAY import entities directly under v1.1.0 — confirm those cases PASS

### Reporter and main loop

- [ ] T018 Create `tools/archcheck/reporter.go` with `func WriteText(w io.Writer, report *Report)` and `func WriteJSON(w io.Writer, report *Report)` matching the format spec in `specs/009-constitution-compliance/contracts/ci-step.contract.md` §3
- [ ] T019 [P] Create `tools/archcheck/reporter_test.go` asserting: deterministic ordering (by kind then file:line:subject), exit-code-mirror, GitHub annotation lines emit when `--annotate=github`
- [ ] T020 Wire `tools/archcheck/main.go` to: load rules, walk module, run all three checkers, sort violations, write `audit-report.json` plus stderr text/annotations, exit 0/1/2 per `ci-step.contract.md` §2 (replaces stub from T007)

### Self-test against fixture tree

- [ ] T021 Create `tools/archcheck/testdata/fixture/` containing a tiny module with one file per intentional violation (oversized handler, oversized usecase file, layer breach in cmd) plus a clean baseline file
- [ ] T022 [P] Create `tools/archcheck/archcheck_test.go` end-to-end test that runs the binary against `testdata/fixture/` and asserts: exit code 1, exact violation count, exact violation messages match the contract format, JSON report parses and matches the `Report` schema

### Smoke test against the live codebase

- [ ] T023 Run `make audit-constitution` against the current codebase, capture `audit-report.json`, commit it to `specs/009-constitution-compliance/baseline-violations.json` (NOT `audit-report.json` at root, which is a CI artefact). This is the baseline that user stories must drive to zero

**Checkpoint**: The audit tool is functional. Every subsequent user story's "Independent Test" can now be run.

---

## Phase 3: User Story 1 — Restore CLI Handler Compliance with Architectural Constitution (Priority: P1) 🎯 MVP

**Goal**: `cmd/new.go` (356 effective lines, multiple oversized handler functions) and `cmd/build.go` (232 effective lines) are decomposed so every handler function is ≤ 50 effective lines, with all scaffolding and build orchestration moved into core use cases.

**Independent Test**: Run `make audit-constitution`. Zero `function-size` violations under `cmd/new.go` or `cmd/build.go`. `make test` passes. Manual smoke: `loko new project foo && loko build` produces the same files and console output as before the refactor.

### Use-case prerequisites for cmd/new.go (split scaffold_entity.go)

- [ ] T024 [P] [US1] Read `internal/core/usecases/scaffold_entity.go` (364 lines) and identify the four scaffold operations (project, system, container, component); document the split plan in a top-of-file comment in `internal/core/usecases/scaffold_entity.go` before the moves begin
- [ ] T025 [US1] Move the project-scaffolding code from `internal/core/usecases/scaffold_entity.go` into a new file `internal/core/usecases/scaffold_project.go` (cut + paste, single commit). Preserve exported function signatures verbatim. After commit, `go build ./...` and `go vet ./...` MUST pass
- [ ] T026 [P] [US1] Move the system-scaffolding code from `scaffold_entity.go` into `internal/core/usecases/scaffold_system.go`
- [ ] T027 [P] [US1] Move the container-scaffolding code from `scaffold_entity.go` into `internal/core/usecases/scaffold_container.go`
- [ ] T028 [P] [US1] Move the component-scaffolding code from `scaffold_entity.go` into `internal/core/usecases/scaffold_component.go`
- [ ] T029 [US1] Migrate `internal/core/usecases/scaffold_entity_test.go` cases to per-file partners (`scaffold_project_test.go`, `scaffold_system_test.go`, `scaffold_container_test.go`, `scaffold_component_test.go`) so each split file owns its tests
- [ ] T030 [US1] Delete `internal/core/usecases/scaffold_entity.go` and the now-empty `scaffold_entity_test.go`; run `go build ./...` and `go test ./internal/core/usecases/...` to confirm green

### Refactor cmd/new.go

- [ ] T031 [US1] Refactor `cmd/new.go` so each Cobra `RunE` handler ≤ 50 effective lines: extract input parsing into named helpers (still in `cmd/new.go`), call the new `scaffold_*.go` use cases, format output via existing renderers. Preserve the public command surface (flags, exit codes, output text). The refactor is also the natural place to remove the direct `internal/core/entities` import per FR-008 (cmd-only); use the use-case return types or adapter outputs instead
- [ ] T032 [US1] Run `make audit-constitution` and confirm zero violations on `cmd/new.go` (function-size AND layer); run `make test` and confirm `cmd/new_test.go` (if present) and any e2e tests still pass
- [ ] T033 [US1] Manual smoke: `go run . new project /tmp/loko-test-001` and diff output against pre-refactor reference (capture in `specs/009-constitution-compliance/manual-smoke.md`)

### Use-case prerequisites for cmd/build.go (entry-point clean-up of build_docs.go)

- [ ] T034 [US1] In `internal/core/usecases/build_docs.go` (479 lines), identify and document (top-of-file comment) the orchestration entry point that `cmd/build.go` will call. The full file split into `build_docs_d2.go`, `build_docs_markdown.go`, `build_docs_assets.go` is deferred to US4 (T062–T065); US1 only needs the public `BuildDocs(...)` entry point to be stable

### Refactor cmd/build.go

- [ ] T035 [US1] Refactor `cmd/build.go` (232 effective lines) so each Cobra `RunE` handler ≤ 50 effective lines: parse options into a `BuildDocsRequest` struct, call `usecases.BuildDocs(...)`, format the report via existing renderer. Preserve flags, exit codes, and the multi-format output behaviour. Remove the direct `internal/core/entities` import (FR-008) by routing entity types through the use-case result type
- [ ] T036 [US1] Run `make audit-constitution` and confirm zero violations on `cmd/build.go` (function-size AND layer); run `make test` and confirm `cmd/build_test.go` (if present) plus the `build_docs` integration tests still pass
- [ ] T037 [US1] Manual smoke: `go run . build` against the project's own `specs/009-constitution-compliance/` and confirm the produced docs match a pre-refactor capture

**Checkpoint**: User Story 1 is complete. `cmd/new.go` and `cmd/build.go` are constitution-compliant; their behaviour is unchanged.

---

## Phase 4: User Story 2 — Enforce Strict Layer Boundaries Across the Codebase (Priority: P1)

**Goal**: Layer-import violations are detected by `make audit-constitution` (already functional from Foundational) AND by `make lint` (`golangci-lint depguard`); a deliberate violation on a throwaway branch fails both checks with actionable messages. Per FR-008 (cmd-only under v1.1.0), the layer rules forbid `cmd/` from importing `internal/core/entities` directly; mcp and api are not subject to that restriction in this feature.

**Independent Test**: On a scratch branch, add `import "github.com/madstone-tech/loko/internal/core/entities"` to `cmd/root.go`. Run `make audit-constitution` — fails with a `layer` violation message naming `cmd/root.go`, `entities`, and the `cmd` rule. Run `make lint` — `depguard` fails likewise. Revert the change; both pass.

### Wire depguard

- [ ] T038 [US2] Edit `.golangci.yml` and add a `depguard` linter configuration whose rules encode the same six-row layer table as `specs/009-constitution-compliance/contracts/structural-rules.yaml`, including the cmd-only `forbiddenImports: internal/core/entities/**` override. Use `depguard`'s `rules:` map keyed by file glob. Reference: <https://github.com/OpenPeeDeeP/depguard>
- [ ] T039 [P] [US2] Add `tools/archcheck --emit-golangci-config` flag that reads `structural-rules.yaml` and prints a `depguard` config block; document its use in `specs/009-constitution-compliance/quickstart.md` "Amending the rules" section
- [ ] T040 [US2] Add a CI cross-check that fails if `.golangci.yml` `depguard` config is out of sync with `structural-rules.yaml` (compare against the emitter output from T039); implement as `tools/archcheck/sync_check.go` with a paired `sync_check_test.go`

### Refactor cmd/init.go and cmd/validate.go (cmd-only FR-008 violations)

- [ ] T041 [US2] Run `make audit-constitution` against the live codebase. The known starting set of `cmd/` → `entities/` violations under the v1.1.0 cmd-only restriction is exactly two: `cmd/init.go` and `cmd/validate.go`. (`cmd/new.go` and `cmd/build.go` are handled in US1.) All other layer rules should already pass; verify and capture the report alongside the baseline from T023
- [ ] T042 [US2] Refactor `cmd/init.go` so it does not import `internal/core/entities` directly. Where the handler currently constructs or names entity types, obtain them via use-case return values or adapter outputs (e.g., have the relevant use case return an opaque result type, or move construction into a use case). Preserve the public command surface (flags, exit codes, output text). Run `make audit-constitution` and confirm zero `layer` violations on this file; run `make test` and confirm `cmd/init_test.go` (if present) passes
- [ ] T043 [US2] Refactor `cmd/validate.go` likewise: remove the direct `internal/core/entities` import, route entity types through a use case, preserve public surface. Run `make audit-constitution` and `make test` to confirm zero new violations

### Final verification

- [ ] T044 [US2] Run `make lint` and confirm `depguard` reports zero violations on the live codebase
- [ ] T045 [US2] Add an integration test at `tools/archcheck/scenarios_test.go` that programmatically introduces each acceptance scenario from spec US2 (entity importing other internal package; usecase importing adapter; adapter importing mcp; `cmd/` importing entities directly; mcp/api importing entities directly — this last one MUST PASS under v1.1.0) and asserts the audit tool emits the correct verdict for each

**Checkpoint**: Layer-import enforcement is mechanical and redundant. Story 2 is complete.

---

## Phase 5: User Story 3 — Bring Oversized MCP Tool Handlers Within Budget (Priority: P2)

**Goal**: Every MCP tool handler function is ≤ 30 effective lines; protocol-handling code contains no business logic.

**Independent Test**: `make audit-constitution` reports zero `function-size` violations under `internal/mcp/tools/`. The MCP integration tests (`internal/mcp/server_test.go` and tool tests) pass with identical responses to pre-refactor for the supported input set.

### Split internal/mcp/tools/graph_tools.go

- [ ] T046 [P] [US3] Read `internal/mcp/tools/graph_tools.go` and identify the three tool implementations (QueryDependencies, QueryRelatedComponents, AnalyzeCoupling); document split plan in a top-of-file comment
- [ ] T047 [US3] Move the `QueryDependencies` tool handler from `graph_tools.go` into a new file `internal/mcp/tools/query_dependencies.go` (≤ 30 effective lines, delegating to a use case). Cut + paste in a single commit; `go build ./...` MUST pass
- [ ] T048 [P] [US3] Move the `QueryRelatedComponents` handler into `internal/mcp/tools/query_related_components.go`
- [ ] T049 [P] [US3] Move the `AnalyzeCoupling` handler into `internal/mcp/tools/analyze_coupling.go`
- [ ] T050 [US3] Migrate `internal/mcp/tools/graph_tools_cache_test.go` and `relationship_tools_test.go` test cases to the new per-tool files; delete `graph_tools.go`
- [ ] T051 [US3] Run `make audit-constitution` on `internal/mcp/tools/` and confirm zero violations from the graph tools

### Move inline schemas out of MCP tool handlers

- [ ] T052 [P] [US3] In `internal/mcp/tools/create_system.go`, move the inline `InputSchema` declaration to `internal/mcp/tools/schemas.go` (existing data-file). Tool handler ≤ 30 effective lines after the move
- [ ] T053 [P] [US3] Same for `internal/mcp/tools/update_system.go`
- [ ] T054 [P] [US3] Same for `internal/mcp/tools/create_component.go`
- [ ] T055 [P] [US3] Same for `internal/mcp/tools/update_component.go`
- [ ] T056 [P] [US3] Same for `internal/mcp/tools/build_docs.go`
- [ ] T057 [P] [US3] Same for `internal/mcp/tools/validate_diagram.go` (and any other tool with inline schemas surfaced by T058)

### Sweep remaining MCP handlers

- [ ] T058 [US3] Run `make audit-constitution` on `internal/mcp/tools/`. For each remaining `function-size` violation: either move presentation logic into a helper in `internal/mcp/tools/helpers.go`, or extract domain logic into a use case (creating `internal/core/usecases/<verb>_<noun>.go` if no existing use case fits)
- [ ] T059 [US3] Move the inline preview-generation logic from `internal/mcp/tools/create_component.go` into a dedicated use case (`internal/core/usecases/preview_component.go`) and call it from the tool handler
- [ ] T060 [US3] Run `make audit-constitution` and confirm zero `function-size` violations under `internal/mcp/tools/`; run `make test` and confirm all MCP tool tests pass
- [ ] T061 [US3] Smoke: drive each MCP tool via `loko mcp` over stdio with a captured before/after request set (write the captures to `specs/009-constitution-compliance/mcp-smoke/`); diff responses byte-for-byte modulo `generated_at` timestamps

**Checkpoint**: MCP tool handlers are thin protocol adapters. Story 3 is complete.

---

## Phase 6: User Story 4 — Keep Use-Case and Entity Files Within Their File-Size Budgets (Priority: P2)

**Goal**: Every file in `internal/core/usecases/` is ≤ 200 effective lines; every file in `internal/core/entities/` is ≤ 300. The pre-existing oversized files are split into cohesive multi-file packages without losing narrative coherence.

**Independent Test**: `make audit-constitution` reports zero `file-size` violations under `internal/core/`. All existing use-case and entity tests pass.

### Split internal/core/usecases/build_docs.go (479 → ≤ 200 each)

- [ ] T062 [P] [US4] Read `internal/core/usecases/build_docs.go` and document the split plan as a top-of-file comment: orchestration in `build_docs.go`, D2 step in `build_docs_d2.go`, markdown step in `build_docs_markdown.go`, asset emission in `build_docs_assets.go`
- [ ] T063 [US4] Move the D2-rendering step from `build_docs.go` into `internal/core/usecases/build_docs_d2.go`; preserve unexported function names (callers stay within the package)
- [ ] T064 [P] [US4] Move the markdown-rendering step into `internal/core/usecases/build_docs_markdown.go`
- [ ] T065 [P] [US4] Move the asset/template emission into `internal/core/usecases/build_docs_assets.go`
- [ ] T066 [US4] Migrate `build_docs_test.go` cases to per-file partners; ensure each new file has its `*_test.go` and the package as a whole still compiles

### Split internal/core/usecases/query_architecture.go (527 → ≤ 200 each)

- [ ] T067 [P] [US4] Document split plan as top-of-file comment: dispatcher in `query_architecture.go`, plus `query_dependencies.go`, `query_related_components.go`, `query_project_overview.go`. Note: these are use-case-package files; the MCP tool files of the same names from US3 (T047, T048) are a different package
- [ ] T068 [US4] Move dependency-query logic from `query_architecture.go` into `internal/core/usecases/query_dependencies.go`
- [ ] T069 [P] [US4] Move related-components-query logic into `internal/core/usecases/query_related_components.go`
- [ ] T070 [P] [US4] Move project-overview-query logic into `internal/core/usecases/query_project_overview.go`
- [ ] T071 [US4] Migrate `query_architecture_test.go` cases to per-file partners; ensure all tests pass

### Split internal/core/usecases/build_architecture_graph.go (469 → ≤ 200 each)

- [ ] T072 [P] [US4] Document split plan: orchestration in `build_architecture_graph.go`, plus `architecture_graph_systems.go`, `architecture_graph_components.go`, `architecture_graph_relationships.go`
- [ ] T073 [US4] Move system-graph construction into `internal/core/usecases/architecture_graph_systems.go`
- [ ] T074 [P] [US4] Move component-graph construction into `internal/core/usecases/architecture_graph_components.go`
- [ ] T075 [P] [US4] Move relationship-edge construction into `internal/core/usecases/architecture_graph_relationships.go`
- [ ] T076 [US4] Migrate `build_architecture_graph_test.go` cases to per-file partners

### Split internal/core/usecases/ports.go (426 → ≤ 200 each)

- [ ] T077 [P] [US4] Read `internal/core/usecases/ports.go` and document split plan: one file per port grouping (`ports_repository.go`, `ports_renderer.go`, `ports_template.go`, `ports_encoder.go`, `ports_clock.go`, `ports_filesystem.go`)
- [ ] T078 [P] [US4] Move `ProjectRepository` interface and related types from `ports.go` into `internal/core/usecases/ports_repository.go`
- [ ] T079 [P] [US4] Move `DiagramRenderer`, `PDFRenderer` interfaces into `internal/core/usecases/ports_renderer.go`
- [ ] T080 [P] [US4] Move `TemplateEngine` interface into `internal/core/usecases/ports_template.go`
- [ ] T081 [P] [US4] Move `OutputEncoder` interface into `internal/core/usecases/ports_encoder.go`
- [ ] T082 [P] [US4] Move `FileWatcher` and other filesystem-shaped ports into `internal/core/usecases/ports_filesystem.go`
- [ ] T083 [P] [US4] Move `Clock` (or other system) interfaces into `internal/core/usecases/ports_clock.go`
- [ ] T084 [US4] Delete `internal/core/usecases/ports.go`; run `go build ./...` to verify all callers still resolve

### Split internal/core/usecases/validate_architecture.go (368 → ≤ 200 each)

- [ ] T085 [P] [US4] Document split plan: orchestration + `validate_drift.go`, `validate_relationships.go`, `validate_components.go`
- [ ] T086 [US4] Move drift validation into `internal/core/usecases/validate_drift.go`
- [ ] T087 [P] [US4] Move relationship validation into `internal/core/usecases/validate_relationships.go`
- [ ] T088 [P] [US4] Move component validation into `internal/core/usecases/validate_components.go`
- [ ] T089 [US4] Migrate `validate_architecture_test.go` cases to per-file partners

### Split internal/core/entities/graph.go (678 → ≤ 300 each)

- [ ] T090 [P] [US4] Document split plan in `internal/core/entities/graph.go` top-of-file comment: `Graph` type and constructors stay in `graph.go`; edge operations move to `graph_edges.go`; queries/traversal move to `graph_query.go`
- [ ] T091 [US4] Move edge-mutation methods into `internal/core/entities/graph_edges.go`
- [ ] T092 [P] [US4] Move query/traversal methods into `internal/core/entities/graph_query.go`
- [ ] T093 [US4] Migrate `graph_test.go` cases to per-file partners

### Borderline-file sweep and final verification

- [ ] T094 [US4] Run `make audit-constitution` and inspect the report for any `file-size` violations beyond the seven enumerated above (e.g., `search_elements.go` 240 raw lines, `render_markdown_docs.go` 213 raw lines may exceed 200 effective). For each unexpected violation, generate a split task following the same MOVE-semantics pattern (cut/paste, single commit, `go build` green). Document the additional splits in this file as sub-bullets if they arise
- [ ] T095 [US4] Run `make audit-constitution` again and confirm zero `file-size` violations across `internal/core/`; run `make test` and confirm coverage on `internal/core/` is still > 80% via `make coverage`

**Checkpoint**: Use-case and entity files are within budget. Story 4 is complete.

---

## Phase 7: User Story 5 — Prevent Regression Through Continuous Integration Gates (Priority: P3)

**Goal**: The structural-compliance check is a gating step in `.github/workflows/ci.yml`. Deliberate violations on a verification branch fail CI and block merge.

**Independent Test**: Open a draft PR that introduces both a layer violation (`cmd/root.go` importing `internal/core/entities`) and a file-size violation (extend `entities/graph.go` past 300 effective lines). The CI run fails the `audit-constitution` step with both violations reported as inline annotations on the PR Files-changed view. The workflow run summary links to the `audit-report.json` artefact.

### Wire CI gating step

- [ ] T096 [US5] Edit `.github/workflows/ci.yml` to add a job step `Audit constitution compliance` between the existing `Build` step and `Test` step. Step runs `make audit-constitution`, sets `--annotate=github`, uploads `audit-report.json` as an artefact named `audit-report-${{ github.run_id }}` on failure (`if: failure()`)
- [ ] T097 [P] [US5] Add a status-check requirement: `audit-constitution` must pass for merges into `main`. Configure via `.github/branch-protection.yml` if the repo uses one, otherwise document the required GitHub setting in `specs/009-constitution-compliance/quickstart.md` "CI integration"

### Retire the legacy allowlist

- [ ] T098 [US5] Reduce `scripts/audit-constitution.sh` to a `--legacy-shim` wrapper around `tools/archcheck` per `ci-step.contract.md` §7. Remove the `KNOWN_VIOLATIONS` associative array and the per-file allowlist matching logic
- [ ] T099 [P] [US5] Confirm `KNOWN_VIOLATIONS` array in `scripts/audit-constitution.sh` is now empty; the file remains but only exists to print a deprecation notice and forward to `make audit-constitution`

### Verification

- [ ] T100 [US5] Create a draft PR (do NOT merge) on a scratch branch `verify/audit-fail-001` that introduces both deliberate violations described in the Independent Test. Capture the failing CI run URL and the inline annotations to `specs/009-constitution-compliance/ci-failure-evidence.md`. Close the draft PR without merging
- [ ] T101 [US5] Confirm the same scratch branch passes CI after reverting the deliberate violations (capture the passing run URL alongside the failure evidence)

**Checkpoint**: CI gating is in place. Story 5 is complete.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, agent context refresh, allowlist cleanup, and final SC-001 sweep.

- [ ] T102 [P] Update `CLAUDE.md` "Active Technologies" section so it correctly lists `tools/archcheck` (Go AST audit) and `golangci-lint depguard` as the structural-compliance toolchain (re-run `.specify/scripts/bash/update-agent-context.sh claude` if needed)
- [ ] T103 [P] Update repo-level `README.md` (if present) to mention `make audit-constitution` under a "Development" section
- [ ] T104 [P] Update `Makefile` `help:` target to describe the new `audit-constitution` and `audit-constitution-watch` targets
- [ ] T105 Add cross-references between `docs/adr/0009-constitution-compliance-tooling.md` and `.specify/memory/constitution.md` (the ADR's "Status" should read "Accepted, ratified in constitution v1.1.0")
- [ ] T106 Run `make audit-constitution` one final time and confirm: 0 violations, exit code 0. Capture the report to `specs/009-constitution-compliance/final-clean-audit.json` to demonstrate SC-001
- [ ] T107 Run the full `make test` suite and `make coverage`; confirm coverage on `internal/core/` is still > 80% (Constitution Principle V) and capture the coverage summary to `specs/009-constitution-compliance/final-coverage.txt`
- [ ] T108 Run the spec's full quickstart end-to-end (`specs/009-constitution-compliance/quickstart.md`) as a developer would; record any friction or doc gaps as follow-up issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup. **BLOCKS all user stories** (every story's Independent Test runs the audit tool built here).
- **User Story 1 (Phase 3)**: Depends on Foundational. Internally sequenced because the cmd refactor depends on the use-case split.
- **User Story 2 (Phase 4)**: Depends on Foundational. Independent of US1; the layer-import checker is already in Foundational, this phase wires `depguard` and refactors the two cmd files (init, validate) that have direct entity imports.
- **User Story 3 (Phase 5)**: Depends on Foundational. Largely independent of US1/US2 but T059 (preview use case) interacts with the same `internal/core/usecases/` package; coordinate file moves.
- **User Story 4 (Phase 6)**: Depends on Foundational. **Coordinates with US1**: T034 (US1) deliberately defers the full `build_docs.go` split to T062–T066 (US4). If team is sequential, run US1 entry-point clean-up first, then US4 finishes the split.
- **User Story 5 (Phase 7)**: Depends on US1, US2, US3, US4 (CI gating only makes sense after the codebase is clean — otherwise the gate breaks all PRs from day one).
- **Polish (Phase 8)**: Depends on US1–US5 complete.

### User Story Independence

| Story | Independently testable? | Independent test command |
|-------|--------------------------|--------------------------|
| US1 | Yes | `make audit-constitution` reports 0 violations on `cmd/new.go`, `cmd/build.go`; `make test` green |
| US2 | Yes | Introduce a deliberate cross-layer import in `cmd/`; both `make audit-constitution` and `make lint` fail with actionable messages |
| US3 | Yes | `make audit-constitution` reports 0 `function-size` violations under `internal/mcp/tools/`; MCP smoke captures match |
| US4 | Yes | `make audit-constitution` reports 0 `file-size` violations under `internal/core/`; `make coverage` ≥ 80% on core |
| US5 | Yes | Draft PR with deliberate violations fails CI with inline annotations; same branch passes after reverting |

### Within Each User Story

- Use-case splits MUST land before the cmd/MCP files that depend on them (see T024–T030 → T031, T034 → T035, T062 → US4 internal ordering).
- Test-file migration follows source-file split in the same task or the immediately next one (constitution Principle V: tests stay paired with code).
- `make audit-constitution` MUST be run after each substantive change to confirm progress against the baseline (`baseline-violations.json` from T023) without regression.

### Parallel Opportunities

- Phase 1: T002, T003, T004, T005, T006 all parallelizable (different files, no dependencies).
- Phase 2 audit-tool engines: T008/T009 (types + loader) parallel; T012/T014/T016 (three checkers, three different files) parallel after T010 (effective-line counter) lands; their tests T013/T015/T017 parallel.
- Phase 3 (US1): T026/T027/T028 (three scaffold files) parallel after T025; T024/T034 (analysis tasks) parallel.
- Phase 4 (US2): T039 parallel with T038 (different files); T042 and T043 sequential because review wants to see init.go land first as a pattern, then validate.go follows.
- Phase 5 (US3): T048/T049 parallel; T052–T057 (six different MCP tool files) all parallel.
- Phase 6 (US4): T062/T067/T072/T077/T085/T090 (split-plan analyses) all parallel; the actual moves within each file are sequenced but moves across different files run in parallel.

---

## Parallel Example: Foundational Phase 2

```bash
# After T010 (effective-line counter) lands, the three checkers and their
# tests can be built in parallel by three workers:
Task: "Create tools/archcheck/filesize.go with file-size checker (T012)"
Task: "Create tools/archcheck/funcsize.go with function-size checker (T014)"
Task: "Create tools/archcheck/layer.go with layer-import checker (T016)"

# And their test partners:
Task: "Create tools/archcheck/filesize_test.go (T013)"
Task: "Create tools/archcheck/funcsize_test.go (T015)"
Task: "Create tools/archcheck/layer_test.go (T017)"
```

## Parallel Example: User Story 4 file-size sweep

```bash
# All six split-plan analyses are independent and can run in parallel:
Task: "Document split plan for build_docs.go (T062)"
Task: "Document split plan for query_architecture.go (T067)"
Task: "Document split plan for build_architecture_graph.go (T072)"
Task: "Document split plan for ports.go (T077)"
Task: "Document split plan for validate_architecture.go (T085)"
Task: "Document split plan for entities/graph.go (T090)"

# Within each split, the per-step extractions parallelize too once the plan lands.
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1 (Setup): scaffolding + ADR + Makefile target.
2. Complete Phase 2 (Foundational): the audit tool. **Blocks everything else.**
3. Complete Phase 3 (User Story 1): cmd/new.go + cmd/build.go decomposed.
4. **STOP and VALIDATE**: `make audit-constitution` reports 0 violations *in the cmd files US1 touched*. `make test` green. Manual smoke against the two refactored commands.
5. Demo: "the most-trafficked CLI commands are constitution-compliant; the audit tool that proved it is the same one we'll use for everything else."

### Incremental Delivery

1. Setup + Foundational → audit tool exists → can verify any subsequent change. (No customer value yet, but the platform is ready.)
2. + US1 → CLI handlers compliant → first observable success criterion met (SC-002).
3. + US2 → layer rules enforced redundantly + cmd/init.go and cmd/validate.go cleaned → second pillar of the constitution mechanically guaranteed.
4. + US3 → MCP tools compliant → SC-003 met.
5. + US4 → core files within budget → SC-001 within reach (still need US5 to gate it).
6. + US5 → CI gating prevents regressions forever → all SCs met.

Each increment can be merged to `main` without breaking the others, provided `make audit-constitution` has zero NEW violations after each merge (the baseline-violations.json from T023 is the floor; the floor only ever drops).

### Parallel Team Strategy

With 3 developers after Foundational lands:

- **Developer A**: US1 (cmd refactor + scaffold/build_docs entry-point clean-up).
- **Developer B**: US3 (MCP tools).
- **Developer C**: US4 (core file splits — except `build_docs.go` which is shared with A).

US2 (depguard wiring + cmd/init/validate refactor) is small enough that any developer can pick it up between other tasks. US5 lands last, after the codebase is clean.

---

## Notes

- **[P] tasks** = different files, no dependencies on incomplete tasks.
- **[Story] label** maps task to spec.md user story for traceability.
- **Split tasks use MOVE semantics**: When a task says "Move … from `Y` into `X`," interpret this as **cut from Y, paste into X** in a single commit. After each split task, `go build ./...` and `go vet ./...` MUST succeed — no duplicate type or function declarations across `Y` and `X`. The matching deletion task (T030, T050, T084) only deletes the now-empty origin file `Y`; it is **not** the place where contents are removed from `Y`. Per-extraction tasks are responsible for keeping the codebase compilable on every commit.
- **FR-008 is cmd-only under v1.1.0**: the layer rule blocks `cmd/` from importing `internal/core/entities` directly. mcp and api MAY import entities directly under v1.1.0; tightening that is a future-feature concern.
- **Existing tests must pass** at every checkpoint (FR-016). If a refactor breaks a test, fix the refactor — do not weaken the assertion or delete the test.
- **Public surface preserved** at every checkpoint (FR-015): command names, flags, exit codes, MCP tool names, MCP tool input/output schemas, files produced. Use the smoke captures in T033, T037, T061 to verify.
- **Commit cadence**: commit after each task or per logical group (e.g., one file split + its test migration as one commit).
- **Avoid**: vague tasks, same-file conflicts in parallel work, weakening tests to make a refactor fit, ad-hoc allowlist entries in `scripts/audit-constitution.sh` after T098.
