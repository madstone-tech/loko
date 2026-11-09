# Tasks: v1.5.1 Patch Bug-Fix Batch

**Input**: Design documents from `/specs/022-v151-fix-batch/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/, quickstart.md — all present.

**Tests**: Included. Constitution Principle V (Test-First) is a project principle, and the spec's acceptance scenarios are test-shaped; each story's phase writes its failing tests first (store, CLI and schema golden tests already exist and change, new tests are added only where none cover the path).

**Organization**: Tasks are grouped by user story. All five fixes are file-disjoint, so every story is independent; the only shared files are `CHANGELOG.md` and the release step (final phase).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Exact file paths included in every description

## Path Conventions

Single Go module at repository root. Fix surfaces (verified anchors in research.md):

- `internal/adapters/outputdir/` — manifest location
- `cmd/` — help text
- `internal/adapters/hclsource/` — `dynamic` hint
- `internal/core/entities/arch/` — `kind` note
- `docs/` — theming guide + quickstart ownership prose

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify the working tree is clean and the baseline is green before any fix lands.

- [X] T001 Confirm clean baseline: `git status` shows only the local test theme and the feedback report untracked; branch `022-v151-fix-batch` checked out; run `go test ./internal/adapters/outputdir/ ./cmd/ ./internal/adapters/hclsource/` and record pass counts

**Checkpoint**: Baseline green; no in-flight edits to the five fix surfaces.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Nothing blocks the stories — this patch has no shared new infrastructure. The state-directory design lives entirely inside US1's files (`internal/adapters/outputdir/`), and no other story reads what it writes. This phase is intentionally empty.

**Checkpoint**: None required. User stories can begin immediately in parallel.

---

## Phase 3: User Story 1 — Deployment output carries no tool state (Priority: P1) 🎯 MVP

**Covers**: FR-001, FR-002, FR-003 (SC-001, SC-002)

**Goal**: The ownership manifest moves out of the output directory into
`.<base>.loko/manifest` beside it, with 1.5.0 fallback and success-only legacy
retirement (contracts/output-state.md).

**Independent Test**: quickstart.md scenarios 1–3: a fresh build leaves `dist/`
without dot-files; a simulated 1.5.0 checkout upgrades (legacy manifest read,
pruned from, then removed); a failed build leaves the legacy manifest untouched.

### Tests for User Story 1 (write FIRST, ensure they FAIL before implementation)

- [X] T002 [P] [US1] Rewrite `TestManifestWrittenLast` in `internal/adapters/outputdir/store_test.go` to expect the manifest at `filepath.Join(filepath.Dir(out), ".dist.loko", "manifest")` with unchanged bytes; expect RED
- [X] T003 [P] [US1] Add `TestLegacyManifestFallback` in `internal/adapters/outputdir/store_test.go`: pre-place `dist/.loko-manifest` listing an owned file, commit a build that no longer writes it, expect the file pruned and the legacy manifest retired; expect RED
- [X] T004 [P] [US1] Add `TestLegacyManifestSurvivesFailedCommit` in `internal/adapters/outputdir/store_test.go`: force a failure after staged writes (unwritable target via a read-only directory injected before retirement), expect `dist/.loko-manifest` still present and `.dist.loko/manifest` absent or incomplete; expect RED
- [X] T005 [P] [US1] Keep coverage of the corrupt-manifest case: update `TestCommitWithoutManifestPrunesNothing` (store_test.go:154) to write its garbage manifest at the legacy path with `.dist.loko` absent, and assert prune-nothing semantics; expect RED. Done as two subtests: corrupt legacy manifest, and a corrupt state manifest that wins over a valid legacy one

### Implementation for User Story 1

- [X] T006 [US1] Implement the state location in `internal/adapters/outputdir/store.go`: add `stateDir(outDir)` returning `filepath.Join(filepath.Dir(outDir), "."+filepath.Base(outDir)+".loko")` and `stateFile(outDir)`; rewrite `readManifest` to read the state file first, falling back to `join(outDir, vm.ManifestFile)`; move the manifest write from the staged path to `os.MkdirAll(stateDir, 0o755)` + `os.WriteFile(stateFile, …, 0o644)` after pruning (depends on T002–T005). Done as `stateManifest` + `writeManifest`, which writes a temporary file and renames it (atomic). Fallback runs only when the state manifest does not exist
- [X] T007 [US1] Implement success-only legacy retirement at the end of `Store.Commit` in `internal/adapters/outputdir/store.go`: after the state manifest is written, `os.Remove(filepath.Join(outDir, vm.ManifestFile))` ignoring `fs.ErrNotExist`, surfacing any other error; keep `vm.ManifestFile` as the refused-artifact-path guard (store.go:33) (depends on T006)
- [X] T008 [P] [US1] Update the package doc in `internal/adapters/outputdir/doc.go` and the manifest const doc in `internal/core/entities/viewmodel/commit_plan.go` to the new location, the legacy-retirement semantics, and the constant's two residual roles after this release (read of a legacy 1.5.0 manifest, and the refused-artifact-path guard in `store.go`); review after T006/T007 (independent of T006/T007 file-wise)
- [X] T009 [US1] Amend the normative contract `specs/014-viewmodel-renderers/contracts/output-layout.md`: drop `.loko-manifest` from the `dist/` tree, document the sibling state directory and the notice row's new location, and update the Ownership section (depends on T006)
- [X] T010 [US1] Update `docs/quickstart.md` output tree (line ~136) to show `.dist.loko/manifest` beside `dist/` (depends on T006)
- [X] T023 [US1] Refuse an output directory that resolves to the project root in `Store.Confine` (`internal/adapters/outputdir/confine.go`), because its state directory would fall outside the project; add the `"."` and absolute-root cases to the refusal loop of `TestConfine` in `internal/adapters/outputdir/confine_test.go` first (RED), then implement (FR-001, contracts/output-state.md confinement rule)

**Checkpoint**: quickstart scenarios 1–3 pass; `go test ./internal/adapters/outputdir/` green.

---

## Phase 4: User Story 2 — Help names the language reference (Priority: P2)

**Covers**: FR-004 (SC-003)

**Goal**: `loko --help` carries one docs pointer naming `docs/language.md` (contracts/cli-language-notes.md, CLI-1..3).

**Independent Test**: quickstart scenario 4.

### Tests for User Story 2 (write FIRST, ensure they FAIL)

- [X] T011 [P] [US2] Extend the root-help test in `cmd/root_test.go` to assert the Long text mentions `docs/language.md` (and the `docs/` directory); expect RED

### Implementation for User Story 2

- [X] T012 [US2] Add one sentence to `rootCmd.Long` in `cmd/root.go` (lines 28–40) pointing at the repository `docs/` directory and naming `docs/language.md` as the full `*.loko.hcl` language reference; do not touch subcommand Longs (depends on T011)

**Checkpoint**: scenario 4 passes; `go test ./cmd/` green (regenerate goldens if the Long text is captured: `go test ./cmd/ -update`).

---

## Phase 5: User Story 3 — The dynamic hint (Priority: P2)

**Covers**: FR-005 (SC-004)

**Goal**: The `dynamic` unknown-block error teaches the supported construct (contracts/cli-language-notes.md, LANG-1).

**Independent Test**: quickstart scenario 5.

### Tests for User Story 3 (write FIRST, ensure they FAIL)

- [X] T013 [P] [US3] Update the `dynamic named` case in `internal/adapters/hclsource/decode_test.go` (line 268) to expect the detail substring `use a view block with step blocks`; expect RED

### Implementation for User Story 3

- [X] T014 [US3] Extend the `"dynamic"` entry of `deliberatelyExcluded` in `internal/adapters/hclsource/schema.go` (line 34) with a second sentence naming `view` + `step` blocks; leave the other excluded-block entries untouched (depends on T013). Final text: `… excluded from v1.0. For a C4 dynamic view, use a view block with step blocks` (no semicolon, and it addresses the HCL-dynamic vs C4-dynamic-view confusion; see contracts/cli-language-notes.md LANG-1)

**Checkpoint**: scenario 5 passes; `go test ./internal/adapters/hclsource/` green.

---

## Phase 6: User Story 4 — The kind note (Priority: P3)

**Covers**: FR-006 (SC-005)

**Goal**: The relationship `kind` note states default direction and trigger reversal (contracts/cli-language-notes.md, LANG-2).

**Independent Test**: quickstart scenario 6.

### Tests for User Story 4 (write FIRST, ensure they FAIL)

- [X] T015 [P] [US4] Extend the note assertion in `internal/core/entities/arch/schemas_test.go` (the `attr("system", "docs").Note` pattern at line 86) with an `attr("uses", "kind").Note` check; expect RED. Final assertion: `sync by default`, `from the declaring element to its target`, `trigger reverses`, and no semicolon

### Implementation for User Story 4

- [X] T016 [US4] Replace the `kind` note in `internal/core/entities/arch/schemas.go` (line 133) (depends on T015). Final text agrees with `docs/language.md:179-181`: `sync by default. The arrow points from the declaring element to its target. trigger reverses the arrow: write it on the invoked element, targeting its trigger`

**Checkpoint**: scenario 6 passes; schema-related goldens regenerated (`go test ./internal/... ./cmd/ -update` where captured).

---

## Phase 7: User Story 5 — Theming class contract docs (Priority: P3)

**Covers**: FR-007 (SC-006, docs half)

**Goal**: `docs/guides/site-theming.md` documents stable classes, variables and the stability commitment (contracts/theme-classes.md).

**Independent Test**: quickstart scenario 7 (read the section).

### Implementation for User Story 5

- [X] T017 [P] [US5] Add a "Stable classes, variables, and blocks" section to `docs/guides/site-theming.md` after "Overridable files": three tables (layout classes; page-content classes; custom properties), populated by reading `internal/adapters/html/theme/*.gohtml` and `style.css`, each table row naming the emitting template; state the stability commitment (rename = changelog-named breaking change, addition = additive, unlisted = no promise) per contracts/theme-classes.md
- [X] T018 [P] [US5] Extend the "Errors" section of `docs/guides/site-theming.md` with one sentence: theme blocks fail the build (`theme_invalid`), CSS classes are covered by the documented class contract, not the build (depends on T017 for placement)

**Checkpoint**: guide reads as one theming contract: files, classes/variables, blocks, errors.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Full-gate verification, changelog, and release prep.

- [X] T019 Run every quickstart.md scenario end-to-end on a real build (`loko build` + `loko validate`), including a build of a local custom-theme project as a theming bench. Scenario 2 needs the simulated 1.5.0 checkout from quickstart §2 (move `.dist.loko/manifest` back to `dist/.loko-manifest`, delete `.dist.loko`); T003 already covers the same path at unit level, so the manual sequence on T019 is the end-to-end confirmation
- [X] T020 Run full gates: `task test && task lint`; coverage target > 80% on `internal/core/` (unchanged by the batch); regenerate goldens with `go test ./cmd/ ./internal/adapters/{d2,hclsource,html,markdown}/ -update` and review `git diff testdata/`
- [X] T021 [P] Write the `## [1.5.1]` CHANGELOG entry at the top of `CHANGELOG.md`: one motivating sentence per the house style (CHANGELOG.md:8–67), a `### Fixed` section with bold one-line summaries for the five fixes, each citing its feedback issue number. **Verify before marking done**: every fix of the spec's five stories appears, and each bullet names its issue (15, 16, 17, 2, 14) — release tagging (T022) waits for this check (SC-006)
- [ ] T022 Prepare release: merge to `main` after gates, cut `release/1.5.1`, tag `v1.5.1`, push tag first per the release convention; confirm the 1.6.0 batch (theming + MCP ergonomics) is NOT included

**Checkpoint**: All spec success criteria SC-001…SC-006 demonstrably pass; release artifacts ready.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies, first.
- **Foundational (Phase 2)**: empty — no blocking prerequisites in this patch.
- **User Stories (Phases 3–7)**: mutually independent; all five fix surfaces are disjoint files. US3, US4, US2 are single-edit fixes; US1 is the largest.
- **Polish (Phase 8)**: after all stories.

### User Story Dependencies

- **US1 (P1)**: independent. Largest: 6 implementation-adjacent tasks. The tests (T002–T005) are file-siblings inside `store_test.go`, so they are written together but marked [P] because they are separable task units touching one file — implement them in one edit session to avoid churn.
- **US2 (P2)**: independent; `cmd/` only.
- **US3 (P2)**: independent; `internal/adapters/hclsource/` only.
- **US4 (P3)**: independent; `internal/core/entities/arch/` (+ possible MCP golden).
- **US5 (P3)**: independent; `docs/` only.

### Within Each User Story

- Tests first (RED) → implementation → checkpoint scenario → goldens where captured.

### Parallel Opportunities

- All five user stories can proceed fully in parallel (different packages + docs).
- Within US1, T002–T005 are test-writing tasks in one file; serialize them in one session but they are separable units.
- T008 (docs) and T009/T010 (contract + quickstart) parallelize with the code change review.
- T017 and T018 are both `docs/guides/site-theming.md` — same file, so order T017 → T018; marked [P] only in the loose sense (distinct from all other tasks).

## Parallel Example: the five stories

```bash
# All five stories have disjoint surfaces — fan them out in one round:
Task: US1 store.go + store_test.go + output-layout contract + quickstart tree
Task: US2 root.go Long text (+ cmd test)
Task: US3 schema.go deliberatelyExcluded map (+ decode test)
Task: US4 schemas.go kind note (+ note test)
Task: US5 site-theming.md class contract section
# After Phase 8: quickstart run-through → gates → CHANGELOG → release
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 baseline verify
2. Phase 3 US1: tests RED → implement → quickstart scenarios 1–3 → green
3. STOP: patch discipline check — `dist/` output diff shows only the manifest removal
4. If shipped alone: still a complete 1.5.1 candidate (the only production-impacting defect)

### Incremental Delivery

1. US1 (P1) → validate → the deployment defect is closed
2. US2 + US3 (P2, independent) → two text-quality fixes
3. US4 + US5 (P3) → schema note + theming contract docs
4. Phase 8: gates, CHANGELOG, release

### Patch-Discipline Guard (applies to every task)

- No rendered output changes beyond the five fixes (spec FR-008): golden diffs must
  be explainable by them — every `testdata/` change names the task that caused it.
- No new language surface, no new diagnostics, no new flags (spec assumptions).

## Notes

- Commit after each story (the optional git hooks fire at each phase boundary).
- `store_test.go` carries the only substantial rewrite; keep the existing table-driven + `t.Parallel()` style.
- The `dynamic` and `kind` text changes are golden-visible; regenerate goldens and include diffs in the story's commit.
- Stop at any checkpoint to validate a story independently.
