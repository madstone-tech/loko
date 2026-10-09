---

description: "Task list for 015-mcp-hcl-authoring"
---

# Tasks: MCP HCL Authoring

**Input**: Design documents from `/specs/015-mcp-hcl-authoring/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Test tasks are **included and mandatory**. Constitution Principle V (Test-First) is
non-negotiable, and SC-002, SC-003, SC-004 and SC-006 are defined as tests. A test task always comes
before the implementation it covers. Run it and **confirm it fails** before implementing.

**Organization**: Tasks are grouped by user story, P1 → P2. US1 (read), US2 (write) and US3
(never corrupt) are all P1. US3 adds its verification on top of US2's editor.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1–US5 from spec.md
- Every task names its exact file paths

## Standing rules for every task

- `internal/mcp/**` and `cmd/**` must not import `internal/core/entities/**`. They exchange
  `usecases` DTOs (`EditInput`, `EditResult`, `QueryResult`, `DescribeResult`, `ValidateResult`).
- Nothing under `internal/**` constructs an adapter (`hclsource.New…`, `encoding.NewEncoder`, …);
  construction happens only in `cmd/wiring.go` (Constitution II as of v1.4.0). Tests may construct
  adapters.
- `internal/core/entities/authoring` imports only the standard library (not even `arch`).
- `hclwrite` and `hcl` are imported only in `internal/adapters/hclsource`.
- Collections crossing a boundary are sorted slices, never maps (FR-009).
- Limits:
  - MCP handler functions ≤ 30 effective lines; CLI handler functions ≤ 50.
  - Use-case files ≤ 200 and functions ≤ 60; entity files ≤ 200; adapter files ≤ 400.
- Run `task audit-constitution` after every phase.

---

## Phase 1: Setup

- [X] T001 Delete `internal/_parked/mcp_tools/` and `internal/_parked/mcp_graph_cache.go` + `mcp_graph_cache_test.go`. They wrote the deleted v0 file-tree model, and this feature supersedes them (plan § Structure Decision). Update `internal/_parked/README.md` to say the MCP tools were replaced in feature 015 rather than restored
- [X] T002 [P] Create hand-written fixtures in `internal/adapters/hclsource/testdata/handwritten/`. `main.loko.hcl` covers the logical plane, with leading and trailing comments on blocks, `#` and `//` and `/* */` comments inside blocks, deliberately misaligned `=`, blank-line runs, and declaration order that is neither alphabetical nor by kind. `deploy.loko.hcl` covers environments with nested `node` groups, instances, bindings and comments. `views.loko.hcl` holds a `view` block. The fixture must compile with zero errors (verify with `go run . validate -p …`), and must **not** be canonically formatted (`loko fmt --check` exits 1), so byte-preservation tests are meaningful
- [X] T003 [P] Create `internal/core/entities/authoring/doc.go`: "Edits, batches, revisions and plans for authoring HCL through loko. Standard library only; imports no other internal package."

---

## Phase 2: Foundational (blocking)

**Purpose**: the authoring vocabulary, the two ports, compiling with an in-memory overlay, and the MCP tool plumbing that every story uses.

- [X] T004 [P] Write table tests in `internal/core/entities/authoring/edit_test.go` for `NewEdit`, following data-model §2.1:
  - Valid edits for each `Op` × `Target` combination that is allowed.
  - Rejected: `rename` on a non-element; `Cascade` on a non-remove; `To` on a non-rename; `File` on a non-add.
  - An address of the wrong form for its target (each segment must pass the `ValidName` rule: "a letter or underscore, followed by letters, digits, underscores, or hyphens"; re-implement it locally, because `authoring` cannot import `arch`).
  - An illegal attribute for the kind, per the legal-attributes table: element `description, owner, technology, tags, docs` plus `system` (container) or `container` (component); relationship `target` (required on add), `description`, `technology`; environment `provider, account, region`; group none; instance `of` (required on add), `attributes`; binding exactly one of `address`, `addresses`, `tags`.
  - `File` not ending `.loko.hcl`, absolute, or containing `..`.
  - A binding without `Binding.Kind` in `terraform | cloudformation`.
  - An update with neither `Set` nor `Clear`.
  - Each error names the field.
  - Add `internal/core/usecases/naming_parity_test.go`, which may import both packages. Over a table of 40 names (valid, invalid, and edge cases such as `_x`, `a-`, `1a`, `a.b`, empty and unicode), assert that `authoring`'s name check and `arch.ValidName` agree on every one, so the two copies cannot drift.
- [X] T005 Implement `internal/core/entities/authoring/edit.go` (`Edit`, `Op`, `TargetKind`, `BindingRef`, `Attr{Name, Value AttrValue}` where `AttrValue` holds one of string, string list, reference address, key/value list, plus `NewEdit`) and `edit_schema.go` (the legal-attribute table) to pass T004
- [X] T006 [P] Write tests, then implement `internal/core/entities/authoring/batch.go` (`Batch{Edits, Preview, BaseRevision}`; `NewBatch` rejects 0 or more than 100 edits and an empty base revision) and `revision.go`:
  - `Revision` is built from `[]FileHash{Path, SHA256}` sorted by path.
  - `Encode()` gives an opaque, deterministic token (base64url of the sorted `path\x00hex` lines).
  - `ParseRevision` rejects malformed tokens.
  - `Hash(path)` returns a file's hash, or false.
  - Tests cover deterministic encoding and the round trip.
- [X] T007 [P] Write tests, then implement `internal/core/entities/authoring/plan.go`: `FileContent{Path, Old, New []byte}`, `Plan{Files []FileContent (sorted by path), Removed []string (sorted)}`, `FileChange{Path, Diff}`, `EditError{Index int, Reason string, Detail string, Dependents []string}` (an `error`), and reason constants `compile_errors | stale_revision | dangling_references | address_in_use | not_found | invalid_edit | path_refused`
- [X] T008 Add to `internal/core/usecases/ports.go`, per contracts/ports.md:
  - `OverlaySource.LoadOverlay(ctx, root, overlay []authoring.FileContent)`.
  - `SourceEditor` with `Plan`, `Diff`, `Revision` and `Commit`.
  - Concrete fakes in `internal/core/usecases/authoring_mocks_test.go`: `fakeEditor`, which records calls and returns canned plans and errors and counts commits; and an overlay-capable `fakeOverlaySource` built on the existing `stubSource`.
- [X] T009 Write `internal/adapters/hclsource/overlay_test.go`. `LoadOverlay` with one file replaced compiles the replaced text rather than the disk text; with a new path compiles it as an additional file; with an empty overlay equals `Load`. Then implement it in `internal/adapters/hclsource/parse.go`: route the parser's `os.ReadFile` through a lookup that checks the overlay first, and include overlay-only paths in the discovered file list (sorted). Assert in the test that disk files are untouched
- [X] T010 Fix `internal/mcp/server.go` and extend `internal/mcp/server_test.go`:
  - `tools/list` returns tools **sorted by name** (today it iterates a map).
  - `handleToolCall` passes the server's run context to `tool.Call` instead of `context.Background()`.
  - Add a test asserting the sorted order across 20 runs.
- [X] T011 Create `internal/mcp/tools/helpers.go` (a data and helper file, exempt from function budgets). It holds shared argument decoding: `str`, `optStr`, `boolArg`, `intArg`, `stringList`, `object`, and `decodeEdits` → `[]usecases.EditInput`. It also holds `encodeResult(enc usecases.OutputEncoder, v any, format string) (any, error)`, with TOON as the default (FR-008). The encoder is injected, never constructed here: wiring belongs in `cmd/` (Constitution II as of v1.4.0), and `refusal(reason, detail string, extra map[string]any)`, which shapes `ok:false` results. Write `internal/mcp/tools/helpers_test.go` covering bad types, missing required args and the TOON/JSON switch
- [X] T012 Rewrite `internal/mcp/tools/schemas.go` with exactly the five input schemas in contracts/mcp-tools.md (`describe`, `query`, `validate`, `apply_edit`, `move`), deleting every v0 schema, and update `internal/mcp/tools/registry_test.go` and `MIGRATION.md` (or delete `MIGRATION.md` if it only describes v0 tools)

**Checkpoint**: the entities, ports, overlay and MCP plumbing are green, and `task audit-constitution` reports 0 violations.

---

## Phase 3: User Story 1 - Ask the architecture a question (P1) 🎯 MVP

**Goal**: `describe`, `query` and `validate` over MCP, answered from the compiled IR.

**Independent Test**: quickstart §1. Every query kind on `testdata/projects/two-systems` matches answers checked by hand. A broken project returns diagnostics.

### Tests ⚠️ write first, confirm failing

- [X] T013 [P] [US1] Write `internal/core/usecases/query_test.go` over IR literals, using the `irOf`/`el`/`rl` builders from `render_mocks_test.go`. Cover the descendant-inclusion rule (research R5):
  - `dependents(container.db)` includes the source of a relationship from a component inside another container into the db, reported at that component's address.
  - `dependencies` is symmetric to that.
  - `transitive` terminates on a three-element cycle, reports each element once, and gives the correct `Distance`.
  - `path(A,B)` returns the shortest directed chain, with ties broken by relationship address, and `Found:false` when there is none.
  - `orphans` matches the `orphan_element` connectivity rule.
  - `coupling` counts distinct fan-in and fan-out, ignores self-relationships, sorts by fan-in+fan-out descending then address, and honours `Limit`.
  - An unknown address yields a `not_found` error with up to 3 suggestions from `resolve_suggest.go`.
  - Results are sorted and deterministic over two calls.
- [X] T014 [P] [US1] Write `internal/core/usecases/describe_test.go`. `summary` holds project, counts by kind, environment names and top-level names. `structure` holds the tree to containers, with components counted. `full` holds every attribute, relationship and placement. `Address` scopes to the element, its children and its incoming and outgoing relationships. Every level carries a `Revision` from the fake editor. A compile error returns diagnostics and no data (FR-007)
- [X] T015 [P] [US1] Write `internal/core/usecases/validate_project_test.go`: diagnostics with ranges, error and warning counts, and the revision
- [X] T016 [P] [US1] Write `internal/mcp/tools/read_tools_test.go`, driving each tool's `Call` with real `hclsource` and `encoding` adapters against `testdata/projects/two-systems`:
  - `describe` defaults to TOON, and `format:"json"` gives JSON.
  - `query` covers each kind, plus `not_found` with suggestions.
  - `validate` reports positions.
  - A broken copy of the fixture gives `ok:false` with diagnostics.

### Implementation

- [X] T017 [US1] Implement `internal/core/usecases/query.go` (`QueryRequest`, `QueryResult`, `ElementHit`, `PathStep`, `CouplingRow`, `Query`) and `internal/core/usecases/query_graph.go` (the descendant-closed adjacency, BFS closure and shortest path) to pass T013
- [X] T018 [US1] Implement `internal/core/usecases/describe.go` (`DescribeRequest`, `DescribeResult`, the views per level) to pass T014, and `internal/core/usecases/validate_project.go` (`ValidateResult`, `Validate`) to pass T015. `Revision` comes from `SourceEditor.Revision`
- [X] T019 [US1] Implement `SourceEditor.Revision` in `internal/adapters/hclsource/edit_commit.go`: SHA-256 of every discovered `*.loko.hcl` file, project-relative and sorted. Add `edit_commit_test.go` for it
- [X] T020 [US1] Implement `internal/mcp/tools/describe.go`, `query.go` and `validate.go`, each ≤ 30-line handlers calling one use case and `encodeResult`, to pass T016. Register them in `cmd/mcp.go` through a new `newAuthoringDeps(root)` in `cmd/wiring.go` (source: `hclsource.New()`; editor: `hclsource.NewEditor()`; encoder: `encoding.NewEncoder()`), replacing the "registers NO tools" comment. Each tool type takes its dependencies in its constructor (`NewDescribeTool(svc, enc)`, …). `cmd/wiring.go` constructs `encoding.NewEncoder()` once and passes it to every tool; `internal/mcp/**` constructs no adapter

**Checkpoint**: an assistant can read and query any compiled architecture.

---

## Phase 4: User Story 2 - Build an architecture through conversation (P1)

**Goal**: `apply_edit` adds, updates and removes every target kind, with batch, preview and cascade, and compiles before every save.

**Independent Test**: quickstart §2. Starting from a `project` block, MCP edits alone build an architecture equivalent to `two-systems`.

### Tests ⚠️ write first, confirm failing

- [X] T021 [P] [US2] Write `internal/adapters/hclsource/edit_blocks_test.go` against `testdata/handwritten/`, one case per op × target:
  - **add**: element, relationship, environment, group (nested), instance (in a nested group), binding.
  - **update**: set and clear each legal attribute, and set a reference attribute as a bare traversal, never a quoted string.
  - **remove**: one case per target kind.
  - For each case:
    - the new contents parse;
    - an added block is byte-identical to `hclwrite.Format` of that block alone (FR-014);
    - **every byte of every line outside the edited block is unchanged** (diff the line sequences: only lines inside the edited declarations' spans, as defined in research R1 § Declaration span, or the inserted lines, may differ) (FR-013);
    - removing a block with an attached comment removes the comment, while a comment separated from it by a blank line survives;
    - files the edit does not touch are not in the plan.
  - Placement (R8):
    - a container goes into its system's file;
    - a top-level element goes into the project block's file;
    - an explicit `file` that does not exist is created;
    - a `file` outside the root, or not ending `*.loko.hcl`, gives an `EditError{Reason: path_refused}`.
  - Targets:
    - removing or updating a missing target gives `not_found`;
    - adding an existing address gives `address_in_use`.
- [X] T022 [P] [US2] Write `internal/adapters/hclsource/edit_diff_test.go`. The unified diff of known before and after texts matches expected output (3 lines of context, `--- a/<path>` and `+++ b/<path>` headers), is deterministic, and is empty for identical inputs
- [X] T023 [P] [US2] Extend `internal/adapters/hclsource/edit_commit_test.go` for `Commit`:
  - It writes every planned file atomically, with permissions preserved.
  - It refuses with `stale_revision` if any planned file's disk hash differs from the base, leaving disk unchanged.
  - It refuses any path not ending `.loko.hcl` or escaping the root, even if a plan contains it (R12).
  - When the second of two renames fails (simulated with a read-only target directory), the first file is restored to its original bytes.
- [X] T024 [P] [US2] Write `internal/core/usecases/apply_edits_test.go` with fakes, covering the state flow in data-model §5:
  - an invalid input gives `invalid_edit` with the batch index;
  - a compile failure gives `compile_errors` with diagnostics and no commit;
  - preview gives diffs and no commit;
  - a no-op gives `NoOp:true` and no commit;
  - success commits once and returns files, warnings and the new revision;
  - a batch is applied in order with the compile run once (the overlay sees the final content);
  - the use case serialises: 20 concurrent `Apply` calls never interleave, as a fake editor with a reentrancy detector shows.
- [X] T025 [P] [US2] Write `internal/core/usecases/apply_dependents_test.go` over compiled IR literals.
  - Removing a referenced element without cascade gives `dangling_references`, listing every referring declaration: children, relationships into it, instances `of` it, and view include and exclude entries.
  - With cascade, the edit set expands to: relationships into and out of it, its children recursively with their dependents, and instances of it with their bindings. `Removed` is sorted, and the result compiles.
  - Removing a group that still holds instances is refused without cascade.

### Implementation

- [X] T026 [US2] Implement `internal/adapters/hclsource/edit_blocks.go` (locate blocks by address in an `hclwrite.File`; add with canonical formatting through a scratch `hclwrite.File` plus `hclwrite.Format`; update with `SetAttributeValue`/`SetAttributeTraversal`/`RemoveAttribute`; remove with `RemoveBlock`) and `internal/adapters/hclsource/edit_plan.go`:
  - `NewEditor()` and `Plan`, which reads the affected files, applies edits in order to in-memory `hclwrite.File`s, and returns `authoring.Plan`.
  - Before editing, assert `ParseConfig(src).Bytes() == src` for each file, and refuse otherwise (R1).

  Pass T021. Keep each file ≤ 400 lines; split out `edit_locate.go` if needed.
- [X] T027 [US2] Implement `internal/adapters/hclsource/edit_diff.go` (line-based Myers diff, unified format) to pass T022, and `SourceEditor.Diff`
- [X] T028 [US2] Implement `SourceEditor.Commit` in `internal/adapters/hclsource/edit_commit.go` (stale check, path guard, temp file + rename with mode preserved, rollback) to pass T023
- [X] T029 [US2] Implement:
  - `internal/core/usecases/authoring_service.go`: `AuthoringService{mu sync.Mutex; deps AuthoringDeps}`, `NewAuthoringService`, `EditInput` (JSON tags matching contracts/mcp-tools.md), `EditResult`, `Refusal`.
  - `internal/core/usecases/apply_edits.go`: `(*AuthoringService).Apply`, which converts inputs with `authoring.NewEdit`/`NewBatch`.
  - `internal/core/usecases/apply_plan.go`: the plan → overlay compile → no-op, preview or commit flow.

  Pass T024. Every function ≤ 60 lines.
- [X] T030 [US2] Implement `internal/core/usecases/apply_dependents.go` (`checkDangling`, `expandCascade`, both on the compiled IR before planning) to pass T025, and wire it into `Apply`
- [X] T031 [US2] Implement `internal/mcp/tools/apply_edit.go` (≤ 30 lines: decode `base_revision`, `edits`, `preview`; call `Apply`; encode) and register it in `cmd/mcp.go`. Add `internal/mcp/tools/apply_edit_test.go`, covering each refusal reason surfacing as `ok:false` and a preview
- [X] T032 [US2] Write `internal/mcp/build_e2e_test.go` (`TestBuildArchitectureEndToEnd`, SC-001):
  - Start a real `mcp.Server` over in-memory pipes on a temp project containing only a `project` block.
  - Send JSON-RPC `tools/call` requests to `apply_edit` (single edits and batches, each using the revision from the previous result) that build two systems, six containers, components, relationships, two environments with nested groups, instances and bindings.
  - Assert the project compiles with zero errors.
  - Assert its exported IR, with ranges stripped, equals that of `testdata/projects/two-systems` with ranges and docs stripped.

**Checkpoint**: an assistant can build and change an architecture end to end.

---

## Phase 5: User Story 3 - Never corrupt what a person wrote (P1)

**Goal**: prove the editor's guarantees on hand-written source, under randomized edits and through every tool.

**Independent Test**: quickstart §3, §4 and §6.

- [X] T033 [US3] Write `internal/adapters/hclsource/edit_property_test.go` (`TestRoundTripProperty`, SC-003). A seeded generator (`math/rand/v2`, seeds 1..2000 in CI, plus a random seed from `-seed` printed on failure) produces sequences of up to 12 valid edits against copies of `testdata/handwritten/`. After each edit:
  1. **IR oracle**: compile and compare with a test-local model that applies the same edit to the previous IR (`applyToIR`), with ranges stripped.
  2. **Byte oracle**: every line outside the edited declarations' spans (research R1 § Declaration span), in every file, is identical. For a rename, the referencing declarations count as edited, and within them each changed line differs from the original only in the renamed traversal.
  3. Applying the plan's `Revision` → `Commit` → `Revision` changes exactly the planned files.

  Add `FuzzApplyEdits`, with the same generator driven by fuzz bytes and seeds 1..20 as its corpus. Any divergence fails with the seed and the edit printed.
- [X] T034 [US3] Write `TestRefusedWritesChangeNothing` in `internal/mcp/tools/safety_test.go` (SC-004; moved from `internal/core/usecases`, whose tests may not import real adapters) with the real `hclsource` adapters on a temp copy of `testdata/handwritten/`. For a write that would not compile, one with a stale revision, a dangling removal, a refused path, and a write while another source file in the project has a syntax error (refused as `compile_errors` with that file's parse diagnostics), hash the whole tree before and after, and assert they are byte-identical
- [X] T035 [US3] Write `TestToolsWriteOnlyHCL` in `internal/mcp/tools/safety_test.go` (`TestToolsWriteOnlyHCL`, SC-005, FR-026, FR-026a):
  - Drive all five tools, including every edit kind, cascade, rename, and an edit setting `docs` to a path that does not exist, against a temp project.
  - Snapshot the tree before and after.
  - Assert every created or modified path ends `.loko.hcl`, and that the `docs` target file was not created.
- [X] T036 [US3] Write `TestStaleRevisionIsPerFile` in `internal/mcp/tools/safety_test.go` (US3/AC4). Read a revision, modify a file on disk, then `apply_edit` an edit to that file with the old revision: the result is `stale_revision`, and the disk keeps the external change. An edit to a different, unchanged file with the same old revision succeeds

**Checkpoint**: the write path is proven safe for hand-written source.

---

## Phase 6: User Story 4 - Rename without losing history (P2)

**Goal**: the `moved` block in the language, and the `move` tool.

**Independent Test**: quickstart §5.

### Tests ⚠️ write first, confirm failing

- [X] T037 [P] [US4] Add golden fixtures under `internal/adapters/hclsource/testdata/golden/`:
  - `moved_valid/` (a single move, a chained move, and a kind change);
  - `moved_from_declared/`, `moved_to_unresolved/`, `moved_duplicate_from/`;
  - `moved_wrong_reference_kind/` (a quoted string).

  Each comes with expected diagnostics and ranges, following contracts/language-moved.md.
- [X] T038 [P] [US4] Write `internal/core/usecases/resolve_moved_test.go` over `SourceModel` literals for the three new codes: `moved_duplicate_from` carries the first block as a related range, and a chain `a→b`, `b→c` (only `c` declared) is valid
- [X] T039 [P] [US4] Write `internal/adapters/hclsource/edit_rename_test.go`. Renaming an element referenced from three files (as a parent, a relationship target, an instance `of`, and in a view include list):
  - rewrites the declaring labels;
  - updates every reference;
  - changes only the traversal tokens on each referencing line;
  - appends `moved { from = …, to = … }` (canonical) to the declaring file.

  Also: a kind change (`container.x` → `component.x`) followed in the same batch by `set container = container.y` compiles; renaming onto an existing address gives `address_in_use` naming it.

### Implementation

- [X] T040 [US4] Add `MovedDecl` to `internal/core/entities/arch/source_model.go` and `Move{From, To, Range}` / `IR.Moves` (JSON and TOON tag `moves,omitempty`, sorted by `From`) to `internal/core/entities/arch/ir.go` and `NewIR`. Add the three codes to `diagnostic.go` (constants, `AllCodes`) and to `specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`, and update the `TestAllCodesAreKnown` count. Confirm every feature 013 and 014 export and projection golden is still byte-identical
- [X] T041 [US4] Implement `internal/adapters/hclsource/decode_moved.go` (top-level `moved` with required `from` and `to` static traversals, a schema entry in `schema.go`) to pass the T037 parse goldens. Implement `internal/core/usecases/resolve_moved.go`, called from `CompileArchitecture`, and carry `Moves` through `build_ir.go`, to pass T037 and T038. Add the codes to `codesFromCoreTests` in `rule_coverage_test.go`
- [X] T042 [US4] Implement `internal/adapters/hclsource/edit_rename.go` (label rewrite plus `Expression.RenameVariablePrefix` across every attribute of every block in every file, plus the appended `moved` block) to pass T039. Wire `rename` into `Plan`
- [X] T043 [US4] Implement `(*AuthoringService).Move` in `internal/core/usecases/apply_edits.go` (one rename edit) and `internal/mcp/tools/move.go` (≤ 30 lines). Register it in `cmd/mcp.go` and add `internal/mcp/tools/move_test.go`
- [X] T044 [US4] Add a `moved` section to `docs/language.md` following contracts/language-moved.md (syntax, rules, diagnostics, that kind changes are allowed, and that rendering ignores moves)

**Checkpoint**: renames keep history, and the language change is documented.

---

## Phase 7: User Story 5 - Ask the same questions from the terminal (P2)

**Goal**: `loko query`, calling the same `usecases.Query`.

**Independent Test**: quickstart §7.

- [X] T045 [US5] Write `cmd/query_test.go`. For each kind on `testdata/projects/two-systems`, `runQueryWith(..., format:"json")` output is **byte-identical** to the MCP `query` tool's JSON for the same inputs (SC-006). An unknown address prints the suggestion and exits `1`. A broken project prints diagnostics and exits `1`. The text format renders a table
- [X] T046 [US5] Implement `cmd/query.go` (`runQueryWith`) and `cmd/query_cobra.go` (subcommands `dependents`, `dependencies`, `path`, `orphans` and `coupling`; flags `--transitive`, `--limit` and `--format`; `GroupID: "building"`) per contracts/cli-query.md, to pass T045. Add `query` to the pinned set in `cmd/root_test.go`
- [X] T047 [US5] Add a `loko query` section to `docs/cli-reference.md`, and the command to the quickstart command table

---

## Phase 8: Polish & Cross-Cutting

- [X] T048 Write `cmd/authoring_perf_test.go` (`TestAuthoringPerformance`, skipped under `-short` and `-race`), reusing `writeLargeProject` from `cmd/build_perf_test.go`:
  - `describe` at each level, and each query kind, under 1 s on 1,020 elements.
  - A single `apply_edit`, including its compile, under 2 s.
  - The TOON summary is at most 300 tokens, and the TOON structure description at most 2,000 (research R6), both estimated as whitespace-separated fields; record the estimator in the test.

  This covers SC-007 and SC-008.
- [X] T049 [P] Write `docs/adr/0014-hcl-authoring.md`:
  - edits through `hclwrite` with no reformatting of existing lines;
  - plan, overlay compile and commit;
  - revisions and atomic commit;
  - the `moved` block as a permanent language addition;
  - descendant-inclusive query semantics;
  - the MCP surface of five tools, and why the v0 tools were not restored.
- [X] T050 [P] Rewrite `docs/mcp-integration.md` for the five tools: setup, every tool with an example request and response, refusal reasons, the base-revision workflow, batch, preview, cascade and rename. Remove every v0 tool reference
- [X] T051 [P] Update `CHANGELOG.md` (Unreleased: MCP tools return; the `moved` block; `loko query`) and `README.md` (the "Returns in the next release" table: MCP read and write tools now work)
- [X] T052 Run `go test -cover ./internal/core/...` and keep the aggregate coverage above 80% (Constitution V). Then run the full gate: `task lint && task test && task audit-constitution`, and `go test -race ./...`. Then run every scenario in [quickstart.md](./quickstart.md). Confirm that `rg -n 'encoding\.NewEncoder|hclsource\.New' internal/mcp --glob '!*_test.go'` returns nothing (no adapter construction under `internal/**`). Record any deviation as a new task rather than silently fixing it

---

## Dependencies & Execution Order

```
Phase 1 ──▶ Phase 2 ──┬──▶ Phase 3 US1 (read) ──┬──▶ Phase 7 US5 (CLI query)
                      │                          │
                      └──▶ Phase 4 US2 (write) ──┼──▶ Phase 5 US3 (prove safety)
                                                 └──▶ Phase 6 US4 (rename)
                                       all ──▶ Phase 8
```

- US1 needs only Phase 2. US2 needs Phase 2, and needs `Revision` (T019) from US1.
- US3 verifies US2's editor. US4 extends the editor (`Plan`) and the language. US5 reuses US1's `Query`.
- Within each phase: tests → entities → use cases → adapters → MCP and CLI.
- Same-file serialisation:
  - `edit_commit.go`: T019 → T028.
  - `cmd/mcp.go`: T020 → T031 → T043.
  - `apply_edits.go`: T029 → T043.

## Parallel Examples

```
Phase 2:  T004/T005 edit | T006 batch+revision | T007 plan | T009 overlay | T010 server | T011 helpers
US1 tests: T013 query | T014 describe | T015 validate | T016 MCP read tools
US2 tests: T021 blocks | T022 diff | T023 commit | T024 apply | T025 dependents
US4 tests: T037 goldens | T038 resolve_moved | T039 rename
```

## Implementation Strategy

1. **MVP** is Phases 1–5 (US1, US2, US3): an assistant can read, build and change an architecture,
   and the safety of the write path is proven. Stop and validate with quickstart §1–4 and §6.
2. US4 adds renames and the `moved` block.
3. US5 adds the CLI mirror.
4. Polish: performance, ADR-0014, docs, the full gate.

Commit after each task or logical group. Any change to a golden file is a review event: never
`-update` blindly.

## Phase 9: Convergence

- [X] T053 Allow removing a group or environment whose instances are removed earlier in the same batch: in `internal/core/usecases/apply_dependents.go`, exclude from `instancesWithin` every instance a preceding edit in the batch removes (directly, or by a cascade), and add a test in `internal/core/usecases/apply_dependents_test.go` (remove instance, then its group, in one batch: accepted) per edge case "Removing a placement group that still holds instances" / FR-018 (partial)
- [X] T054 Make setting an attribute to its current value a no-op: in `setAttr` (`internal/adapters/hclsource/edit_blocks.go`), leave the attribute untouched when the new expression's tokens equal the existing ones apart from spacing, and add a case to `internal/adapters/hclsource/edit_blocks_test.go` (set `tags = ["edge", "public"]` on `container.web` in the hand-written fixture: the plan is unchanged) per FR-021 (partial)
