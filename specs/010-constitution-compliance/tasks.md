---

description: "Task list for feature 010-constitution-compliance"
---

# Tasks: Constitution Compliance Refactor

**Input**: Design documents from `/specs/010-constitution-compliance/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Behaviour-preservation tests are MANDATORY (FR-015, FR-016). Unit tests for new use cases follow constitution Principle V (Test-First). Tests are included throughout — they are not optional for this feature.

**Organization**: Tasks are grouped by user story so that each story can be implemented, tested, and merged independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no incomplete-task dependencies)
- **[Story]**: Maps to user stories from spec.md (`US1`, `US2`, `US3`, `US4`)
- Every task names exact file paths

## Path Conventions

- Go monorepo, paths relative to repo root (`/Users/andhi/code/mdstn/loko/`).
- Production code: `cmd/`, `internal/core/`, `internal/mcp/`, `internal/api/`, `internal/adapters/`, `tools/archcheck/`.
- Tests live alongside production code as `*_test.go`; integration tests under `tests/integration/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project-level scaffolding shared by all stories.

- [X] T001 Verify Taskfile prerequisites: `task --version` succeeds, `task lint` and `task test` pass on `main` baseline → record output in `specs/010-constitution-compliance/baseline-toolchain.txt`
- [X] T002 [P] Create empty suppression file `.archcheck-suppressions.yaml` at repo root (top-level YAML list: `[]`), gitignored from secrets but tracked in git
- [X] T003 [P] Create draft ADR `docs/adr/0010-tighten-outer-layer-entity-import-rule.md` with the v1.2.0 motivation (final content filled in T038)

**Checkpoint**: Toolchain confirmed; ADR + suppression file scaffolded.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Baseline capture + audit-tool extensions that every story depends on.

**⚠️ CRITICAL**: No user-story phase may begin until this phase is complete.

### Baseline capture (behaviour preservation prerequisite — see research.md R11)

- [X] T004 [P] Capture pre-refactor CLI golden files: for each non-trivial CLI subcommand (`new project`, `new system`, `new container`, `new component`, `build`, `init`, `validate`), run with representative inputs in a tmpdir and save `stdout + stderr + exit code + file tree` to `tests/golden/cli/<subcommand>.golden`
- [X] T005 [P] Capture pre-refactor MCP smoke fixtures: replay representative JSON-RPC requests against the running MCP server and store input/expected-output pairs under `tests/golden/mcp/<tool_name>.{request,response}.json`
- [X] T006 [P] Capture pre-refactor per-package coverage baseline: run `go test -coverprofile=cover.out ./...`, parse with `go tool cover -func`, save per-package floor to `specs/010-constitution-compliance/coverage-baseline.txt`
- [X] T007 [P] Capture pre-refactor function/file line counts for the named scope (cmd/, internal/mcp/tools/, internal/core/usecases/, internal/core/entities/) by running the current `tools/archcheck` (size-only mode) and saving to `specs/010-constitution-compliance/baseline-violations.json`
- [X] T063 [P] Capture pre-refactor HTTP API golden fixtures: enumerate routes from `internal/api/openapi.yaml`; for each route fire a representative request through the running server (reusing fixture inputs from `internal/api/handlers/handlers_test.go` where possible) and save `(method, path, status, response headers, response body)` to `tests/golden/api/<route>.golden.json`. Cover at minimum: the routes that touch entity types (i.e., the ones spec FR-008 will affect once v1.2.0 lands).

### Extend `tools/archcheck` (extends existing 009 binary)

- [X] T008 [P] Add `tools/archcheck/layer.go`: implement layer-rule engine (parses `${MODULE}` from go.mod, compiles deny patterns into matchers, emits `ViolationReport{Kind: layer_import}` per offence). Unit-tested in `tools/archcheck/layer_test.go`
- [X] T009 [P] Add `tools/archcheck/suppression.go`: load `.archcheck-suppressions.yaml`, validate schema (per `contracts/suppression-file-schema.yaml`), enforce 90-day max expiry, surface stale-suppression warnings after 30 days past expiry. Unit-tested in `tools/archcheck/suppression_test.go`
- [X] T010 [P] Add `tools/archcheck/report.go`: text writer + JSON writer matching `contracts/violation-report.schema.json`. Unit-tested in `tools/archcheck/report_test.go`
- [X] T011 [P] Add `tools/archcheck/category.go`: file categorisation logic (priority-ordered match per `contracts/structural-rules.yaml`). Unit-tested in `tools/archcheck/category_test.go`
- [X] T012 Wire the four new modules into `tools/archcheck/main.go`: load rules YAML → categorise files → run layer + size checks → load suppressions → write report → set exit code per `contracts/archcheck-cli.md`
- [X] T013 [P] Place the rules file at `tools/archcheck/rules.yaml` (copy of `specs/010-constitution-compliance/contracts/structural-rules.yaml`); the `--rules` flag defaults to this stable path so the binary remains useful after the feature ships

### Toolchain wiring

- [X] T014 [P] Add `audit-constitution` task to `Taskfile.yml`: builds `tools/archcheck/archcheck`, runs it against the repo with defaults, exits with archcheck's exit code
- [X] T015 [P] Add `depguard` layer-rule config to `.golangci.yml`: encode entities-pure / usecases-only-entities / adapters-only-core / outer-no-entities / outer-no-cross-talk per `contracts/structural-rules.yaml` (redundant fast-path; archcheck remains authoritative)
- [X] T016 [P] Add coverage-delta script `scripts/coverage-delta.sh`: diffs current per-package coverage against `coverage-baseline.txt` (from T006), exits non-zero on regression > 0.5 pp

**Checkpoint**: Audit tool covers layer rules + suppression + JSON report; baselines captured; toolchain wired but CI gate not yet enabled (that lands in US3).

---

## Phase 3: User Story 1 — CLI Handler Decomposition (Priority: P1) 🎯 MVP

**Goal**: Bring `cmd/new.go` (504 lines) and `cmd/build.go` (251 lines) within budget by extracting their logic into narrowly-scoped use cases under `internal/core/usecases/`.

**Independent Test**: Run `task audit-constitution` against `cmd/`. Zero handler-size violations and zero layer-import violations for the two named commands. CLI golden-file tests from T004 pass byte-for-byte.

### Tests for User Story 1 (write FIRST — must FAIL until extraction lands)

- [X] T017 [P] [US1] Add unit-test scaffolds for the four new scaffold use cases in `internal/core/usecases/scaffold_project_test.go`, `scaffold_system_test.go`, `scaffold_container_test.go`, `scaffold_component_test.go` (concrete mock `ProjectRepository` + `TemplateEngine` ports, table-driven cases mirroring existing `cmd/new.go` paths)
- [X] T018 [P] [US1] Add unit-test scaffold for the extracted build-docs orchestration in `internal/core/usecases/build_docs_test.go` (concrete mock `DiagramRenderer` + `TemplateEngine` ports)
- [X] T019 [US1] Add CLI golden-file regression tests at `tests/integration/cli/golden_test.go` that re-run each captured subcommand from T004 and diff stdout/stderr/exit-code/file-tree against the goldens

### Implementation for User Story 1

- [X] T020 [P] [US1] Create `internal/core/usecases/scaffold_project.go` — top-level project scaffold (create root dir, write `loko.toml`, build top-level layout). Target ≤ 200 effective lines; consumes `ProjectRepository` and `TemplateEngine` ports
- [X] T021 [P] [US1] Create `internal/core/usecases/scaffold_system.go` — add-system flow extracted from `cmd/new.go`
- [X] T022 [P] [US1] Create `internal/core/usecases/scaffold_container.go` — add-container flow extracted from `cmd/new.go`
- [X] T023 [P] [US1] Create `internal/core/usecases/scaffold_component.go` — add-component flow extracted from `cmd/new.go`
- [X] T024 [US1] Refactor `cmd/new.go`: replace inline scaffolding logic with calls to the four use cases (T020–T023). Each subcommand's `RunE` must be ≤ 50 effective lines, doing only parse → call use case → render output
- [X] T025 [P] [US1] Extract a small flag-parsing helper `cmd/new_input.go` (categorically exempt as data/helper file) if needed to keep the `RunE` functions under budget — **NOT NEEDED**: `task audit-constitution` reports 0 function-size violations in `cmd/`, so the `RunE` functions are already within the 50-line budget without a separate helper file.
- [X] T026 [US1] Extend `internal/core/usecases/build_docs.go` to absorb the orchestration logic currently inline in `cmd/build.go` (use existing companion files `build_docs_diagrams.go`, `build_docs_tables.go` for sub-step logic; create `build_docs_render.go` and/or `build_docs_assets.go` if any single file would exceed 200 effective lines)
- [X] T027 [US1] Refactor `cmd/build.go`: replace inline build orchestration with a single call to the extended `BuildDocs` use case. `RunE` must be ≤ 50 effective lines
- [X] T028 [US1] Run `task audit-constitution` against `cmd/new.go` and `cmd/build.go` (other directories may still report violations — that's fine here). Confirm 0 violations in these two files
- [X] T029 [US1] Run `task test` and `tests/integration/cli/golden_test.go`. All tests pass; goldens diff clean

**Checkpoint**: `cmd/new.go` and `cmd/build.go` compliant; all scaffolding + build logic in core use cases; behaviour byte-equivalent. **MVP candidate.**

---

## Phase 4: User Story 2 — MCP Tool Handler Decomposition (Priority: P1)

**Goal**: Bring every MCP tool handler currently exceeding 30 effective lines into compliance by extracting domain logic into use cases (creating new ones or reusing those from US1).

**Independent Test**: Run `task audit-constitution` against `internal/mcp/tools/`. Zero handler-size violations. MCP smoke fixtures from T005 replay with byte-equivalent responses.

### Tests for User Story 2 (write FIRST)

- [X] T030 [P] [US2] Add MCP smoke regression test at `internal/mcp/server_smoke_test.go` (or `tests/integration/mcp/golden_test.go`) that replays each request from `tests/golden/mcp/` and asserts byte-equivalent response
- [X] T031 [P] [US2] For each MCP tool whose use case is genuinely new (not reused from US1), add a use-case unit test under `internal/core/usecases/<tool>_test.go` with concrete mock ports

### Implementation for User Story 2

- [X] T032 [US2] Enumerate oversized MCP handlers from `baseline-violations.json` (T007). For each, decide: reuse a US1 use case (e.g., `BuildDocs` for the `build_docs` tool) or extract a new use case. Record the mapping at the top of `internal/mcp/tools/MIGRATION.md` (this file is categorically exempt as a doc)
- [X] T033 [P] [US2] For each MCP tool needing a new use case, create `internal/core/usecases/<verb_object>.go` (e.g., `analyze_coupling.go`, `query_architecture.go`, `find_relationships.go`). Each ≤ 200 effective lines
- [X] T034 [P] [US2] Move per-tool request/response structs (JSON schema-shaped types) into sibling `internal/mcp/tools/<tool>_schemas.go` files (categorically exempt from size budget per Principle III)
- [X] T035 [US2] Refactor every oversized handler in `internal/mcp/tools/` to be a thin protocol adapter: unmarshal request → call use case → marshal response. Each handler function ≤ 30 effective lines
- [X] T036 [US2] Run `task audit-constitution` against `internal/mcp/tools/`. Confirm 0 violations
- [X] T037 [US2] Run `task test` and the MCP smoke test from T030. All tests pass; goldens diff clean

**Checkpoint**: All MCP handlers compliant; protocol-handling code carries no domain logic; behaviour byte-equivalent.

---

## Phase 5: User Story 3 — Mechanical Enforcement & CI Gate (Priority: P1)

**Goal**: Turn the audit tool into a mandatory CI gate, ship the constitution amendment v1.1.0 → v1.2.0 that this feature operationalises, and give contributors a usable suppression workflow for pre-existing violations outside scope.

**Independent Test**: Open a PR that intentionally introduces (a) a layer-import violation and (b) a function-size violation. CI fails with both diagnostics naming file + offending entity + rule; the merge button is greyed out. Reverting the violations makes CI green.

### Tests for User Story 3 (write FIRST)

- [X] T038 [P] [US3] Add golden-file diagnostic tests for `tools/archcheck` in `tools/archcheck/diagnostics_test.go` covering: (a) per-file size violation, (b) per-function size violation, (c) layer-import violation, (d) cross-outer-layer violation, (e) suppressed violation, (f) expired suppression, (g) over-90-day suppression rejected at load
- [~] T039 [P] [US3] ~~Add a `--baseline` mode test~~ **OUT OF SCOPE**: archcheck never implemented a `--baseline` flag. Whole-repo audit already exits 0; a baseline-diff mode was not built. Dropped — reopen as a separate feature if incremental-diff gating is wanted.

### Implementation for User Story 3

- [X] T040 [US3] Finalise the constitution amendment v1.1.0 → v1.2.0 in `.specify/memory/constitution.md`: update the Dependency Direction table row for `internal/mcp/` and `internal/api/` to forbid direct `internal/core/entities/` imports; update the `SYNC IMPACT REPORT` header; bump version footer to `1.2.0`; update `Last Amended` date
- [X] T041 [US3] Finalise ADR `docs/adr/0010-tighten-outer-layer-entity-import-rule.md`: motivation, rule change, scope, migration cost (must remain "minimal" — measured by remaining outer-layer imports of entities after US2)
- [X] T042 [US3] Update the constitution governance footer's rule-file pointer from `specs/009-constitution-compliance/contracts/structural-rules.yaml` to `tools/archcheck/rules.yaml` (the stable location placed in T013)
- [X] T043 [US3] Wire the gating step in `.github/workflows/ci.yml`: add a job step `Audit constitution` after `Test`, running `task audit-constitution`, with `archcheck-report.json` published as a workflow artefact
- [ ] T044 [US3] Mark the new step as a required check in branch-protection rules for `main` (this is a GitHub UI/API change, not a code change — record the action in the PR description)
- [X] T045 [US3] Add `scripts/check-rules-sync.sh` and a CI step that fails if `tools/archcheck/rules.yaml` drifts from the prose in `.specify/memory/constitution.md` (per governance footer: the two must never diverge)
- [X] T046 [US3] Run `task audit-constitution` against the entire repo. Any remaining violations outside US1/US2 scope are either fixed in T047 or recorded as suppressions in T048
- [X] T064 [US3] Add HTTP API smoke regression test at `tests/integration/api/golden_test.go` that replays each fixture from `tests/golden/api/` (captured in T063) and asserts byte-equivalent status + body + relevant headers. Wire into `task test`. Failures here mean a refactor regressed an externally-observable HTTP response — fix the refactor, do not update the golden.
- [X] T047 [P] [US3] Fix any remaining trivial violations the audit surfaces (e.g., a single function over 50 lines in `cmd/`, an entity file over 300 lines that splits cleanly) inline rather than suppressing
- [X] T048 [P] [US3] Record any genuinely-out-of-scope pre-existing violations in `.archcheck-suppressions.yaml` with `owner: @andhi`, `expires_on` ≤ 90 days from today, and a `reason` referencing the follow-up feature/issue. Each entry mapped to the matching rule name (per `contracts/suppression-file-schema.yaml`)
- [ ] T049 [US3] Verify the gate end-to-end by opening a throwaway "audit-demo" PR that introduces one layer violation and one size violation; capture the CI failure output; close the PR without merging
- [X] T050 [US3] Run `task audit-constitution`, `task lint`, `task test` locally. All exit 0

**Checkpoint**: Constitution at v1.2.0; CI gate is required on `main`; suppression mechanism documented and exercised; demo PR proves the gate blocks merges.

---

## Phase 6: User Story 4 — Core Layer Stays Within Budget After Migration (Priority: P2)

**Goal**: Verify (and, where needed, split) use-case and entity files that absorbed extracted logic so every file in `internal/core/` is within its budget.

**Independent Test**: Run `task audit-constitution` against `internal/core/`. Zero file-size violations. Reviewer confirms no use case has been split below the threshold of cohesion ("would a contributor look for this in one file?").

### Tests for User Story 4

- [X] T051 [P] [US4] Verify per-package coverage for `internal/core/usecases` and `internal/core/entities` is ≥ baseline from T006 by running `scripts/coverage-delta.sh`

### Implementation for User Story 4

- [~] T052 [US4] ~~Run `archcheck --include 'internal/core/**'`~~ **OUT OF SCOPE**: archcheck has no `--include` flag (it always scans the whole repo). The goal — confirm core within budget — is satisfied by the whole-repo run, which reports **0 violations** (covers all `internal/core/usecases` ≤ 200 and `internal/core/entities` ≤ 300). No path-filter flag was built.
- [X] T053 [P] [US4] For each oversized use-case file flagged by T052, split it along its natural sub-step seam (e.g., `build_docs.go` → `build_docs.go` + `build_docs_<step>.go`). Each new file ≤ 200 effective lines. Preserve package layout — no new sub-package introduced
- [X] T054 [P] [US4] For each oversized entity file flagged by T052, split it along its natural type/family seam (e.g., one ID type per file, one validation cluster per file). Each new file ≤ 300 effective lines
- [X] T055 [US4] Re-run `task audit-constitution` against `internal/core/`. Confirm 0 violations
- [X] T056 [US4] Self-review the splits: open each touched package's directory listing and confirm the file names tell a coherent story; revise file names if not

**Checkpoint**: All four user-story phases complete; `task audit-constitution` exits 0 across the whole repo (modulo any active suppressions, which have ≤ 90-day expiry and named owners).

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Loose ends that touch multiple stories.

- [X] T057 [P] Update `CLAUDE.md` (auto-managed) — should already reflect T040 amendment after `.specify/scripts/bash/update-agent-context.sh claude` runs
- [X] T058 [P] Update `README.md` "Quality gates" / "Contributing" section to mention `task audit-constitution` and link to `specs/010-constitution-compliance/quickstart.md` for first-time contributors
- [X] T059 [P] Add a one-page contributor reference at `docs/architecture/constitution-compliance.md` summarising: the four budgets (50/30/200/300), the layer rules table, the suppression workflow, and where the canonical rules file lives (`tools/archcheck/rules.yaml`)
- [X] T060 Run `quickstart.md` end-to-end on a clean checkout: build, run audit, intentionally break + fix one budget, verify gate, verify smoke tests. **Diagnostic-legibility check (SC-010)**: deliberately introduce one example of each of the four violation kinds — `file_size`, `function_size`, `layer_import`, `expired_suppression` — inspect each resulting diagnostic, and confirm it names (a) the repo-relative file path, (b) the offending entity (function name or import path), (c) the rule/budget name, and (d) measured-vs-limit for size kinds. A reviewer unfamiliar with the project must be able to act on each diagnostic without opening any other docs. Record outcome + the four sample diagnostics verbatim in `specs/010-constitution-compliance/quickstart-validation.md`
- [X] T065 [P] Benchmark archcheck wall-clock: run `tools/archcheck/archcheck --format=json > /dev/null` five times on a warm checkout, record min/median/max via `time` to `specs/010-constitution-compliance/quickstart-validation.md`. Fail the validation step (and the whole feature's "done" criterion) if median > 30 s (SC-009 budget) or > 10 s (research.md R6 internal target — treat as soft warning).
- [ ] T061 [P] Open the final PR (or PR stack: one per story) with title prefix `feat(010):`. Each PR body includes (a) story scope, (b) `task audit-constitution` output, (c) per-package coverage delta, (d) any new suppressions with owner + expiry + reason
- [X] T062 Delete the now-redundant `specs/009-constitution-compliance/contracts/structural-rules.yaml` (its content migrated to `tools/archcheck/rules.yaml` in T013) only if branch 009 is being retired — otherwise leave in place. Decision recorded in `specs/010-constitution-compliance/quickstart-validation.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: no dependencies; T001 → T002 + T003 can run in parallel
- **Phase 2 (Foundational)**: depends on Phase 1; T004–T007 run in parallel as baseline capture; T008–T011 run in parallel as audit-tool extensions; T012 depends on T008–T011; T013 depends on the rules file being final; T014–T016 run in parallel after T012
- **Phase 3 (US1)**: depends on Phase 2 (baselines + audit tool ready); within the phase: T017–T019 first (red), then T020–T023 in parallel, then T024 (uses T020–T023), then T025 if needed, then T026 (build use case), T027 (slim cmd/build.go), T028–T029 (verify)
- **Phase 4 (US2)**: depends on Phase 2; can run in parallel with Phase 3 if staffed (different files, no shared deps); within phase: T030–T031 first, then T032 enumeration, then T033–T034 parallel, then T035, then T036–T037 verify
- **Phase 5 (US3)**: depends on Phase 3 and Phase 4 being substantially complete (so the gate lands green); T040 (constitution amendment) ships in the US3 PR; T044 (branch protection) is the last switch
- **Phase 6 (US4)**: depends on Phase 3 and Phase 4 (the migration that puts pressure on core files must be done first); independent of Phase 5 mechanically, but in practice will land alongside or just after
- **Phase 7 (Polish)**: depends on Phases 3–6

### User Story Dependencies

- **US1 (CLI)** ← Foundational only
- **US2 (MCP)** ← Foundational only (independent of US1; may reuse some use cases extracted in US1 — that reuse happens in T032's mapping step, not as a hard dep)
- **US3 (Gate + Amendment)** ← needs US1 + US2 to be substantially done (else the gate lands red and blocks the very PR that enables it)
- **US4 (Core sizes)** ← needs US1 + US2 (they create the file-size pressure on core)

### Parallel Opportunities

- T002, T003 in parallel within Phase 1
- T004–T007 in parallel within Phase 2 (independent files)
- T008–T011 in parallel within Phase 2 (independent files)
- T014–T016 in parallel within Phase 2
- US1 use-case files (T020–T023) in parallel
- US3 trivial-fix + suppression batches (T047, T048) in parallel
- US4 use-case split + entity split (T053, T054) in parallel
- Phase 7 doc updates (T057–T059, T061) in parallel
- **Cross-phase**: US1 and US2 are mechanically independent and can be developed in parallel by different contributors after Phase 2.

---

## Parallel Example: User Story 1 launch

```bash
# After Phase 2 completes, fan out the four scaffold use cases:
Task: "Create internal/core/usecases/scaffold_project.go (T020)"
Task: "Create internal/core/usecases/scaffold_system.go (T021)"
Task: "Create internal/core/usecases/scaffold_container.go (T022)"
Task: "Create internal/core/usecases/scaffold_component.go (T023)"

# Their unit-test scaffolds (T017) can be written first, in parallel, by the same contributors.
```

---

## Implementation Strategy

### MVP (User Story 1 only)

1. Phase 1 (Setup) → Phase 2 (Foundational, capture baselines + extend archcheck) → Phase 3 (US1).
2. **Stop and validate**: `cmd/new.go` and `cmd/build.go` are compliant; CLI golden tests pass byte-for-byte.
3. Demo: same `loko new` and `loko build` outputs as `main`, but with `cmd/new.go` shrunk from 504 → ~120 lines and `cmd/build.go` shrunk from 251 → ~60 lines.

### Incremental delivery (recommended)

1. Land **US1** → demo CLI compliance.
2. Land **US2** → demo MCP compliance.
3. Land **US3** → ship constitution v1.2.0 + turn the CI gate on. This is the irreversible step; everything before is just refactor.
4. Land **US4** → close out any core-layer files that the migration pushed over budget.
5. Land **Phase 7 (Polish)** → docs, contributor reference, validation.

### Parallel team strategy

With two contributors:
- Contributor A: Setup + Foundational + US1.
- Contributor B (after Foundational): US2 in parallel.
- Both converge on US3 (gate + amendment) — single PR, joint review.
- US4 + Polish: either contributor.

---

## Notes

- [P] means independent files; tasks without [P] either edit the same file as a predecessor or depend on its output.
- Every task ends in a runnable verification step (build, test, or audit invocation) — partial states should not survive a task boundary.
- The behaviour-preservation contract (FR-015, FR-016) is verified at T029, T037, T050, T051, and T060 — failures at any of those checkpoints mean the refactor regressed and the offending task must be re-done, not papered over.
- The constitution amendment v1.2.0 and the CI gate are intentionally bundled in the same PR (US3) so the rule and its enforcement land atomically; do not split.
- Suppression entries (T048) are a last resort. If an entry's `reason` does not name a specific follow-up issue or PR, fix the violation instead.
