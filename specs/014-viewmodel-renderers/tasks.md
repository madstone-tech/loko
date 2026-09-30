---

description: "Task list for 014-viewmodel-renderers"
---

# Tasks: ViewModel Renderers

**Input**: Design documents from `/specs/014-viewmodel-renderers/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/](./contracts/), [quickstart.md](./quickstart.md)

**Tests**: Test tasks are **included and mandatory**. Constitution Principle V (Test-First) is
non-negotiable in this repository, and SC-002, SC-004, SC-007 and SC-008 are all defined as tests. A
test task always comes before the implementation it covers. Run it and **confirm it fails** before
you write the implementation.

**Organization**: Tasks are grouped by user story in the order P1 → P2 → P3. US1 and US3 are both
P1, and together they form the MVP.

**Revision**: revised after `/speckit-analyze` (2026-09-30) to fix findings C1–C3, I1–I4, U1–U3, K1,
K2, A1, D1, D2 and G1. Four tasks were added with suffixes rather than renumbering (T018a, T023a,
T045a, T087a), taking the total from 102 to 106.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: The user story the task belongs to (US1–US7)
- Every task names its exact file paths

## Path Conventions

A single Go module at the repository root. The layout follows [plan.md](./plan.md) § Project
Structure:

- `internal/core/entities/viewmodel/`: new, stdlib only, ≤ 200 effective lines per file
- `internal/core/usecases/`: ≤ 200 lines per file, ≤ 60 per function
- `internal/adapters/{d2,markdown,html,outputdir,projectfs,watch,devserver}/`: ≤ 400 lines per file
- `cmd/`: ≤ 50 lines per function
- `testdata/projects/<fixture>/`: shared end-to-end fixtures, with expected outputs under `expected/`

**Standing rules for every task**
- Collections crossing any boundary are sorted slices, never maps (see [data-model.md](./data-model.md) preamble).
- Only `viewmodel.Paths` functions produce output paths and hrefs. Never concatenate them at a call site.
- `cmd/` must not import `internal/core/entities/**`. Obtain values through use cases.
- Golden tests support `-update`, as feature 013's golden tests do (see `internal/adapters/hclsource/*_test.go`).

## ⚠️ Story independence in this feature

- **US1 and US3 are the MVP.** US1 draws the diagrams. US3 makes them trustworthy enough to commit.
  Either one alone is not shippable.
- **US2** (declared views) and **US6** (single binary) depend only on the US1 pipeline.
- **US4** (site and markdown) depends on US1. **US5** (serve) and **US7** (themes) depend on US4's HTML backend.
- US6's in-process renderer is *built* in US1, because the SVG backend cannot exist otherwise. The
  US6 phase adds the proof and removes the last traces: the empty-`PATH` test, the Docker image, and
  the enforcement rules.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Get dependencies, package skeletons, enforcement rules, and fixtures in place before any code is written.

- [X] T001 Add `oss.terrastruct.com/d2@v0.7.1` and `github.com/yuin/goldmark@v1.7.4` as direct requirements in `go.mod` (`go get` both), run `go mod tidy`, and confirm `go build ./...` is green. Record in the commit message that d2 is reintroduced as a library per research R1 and that goldmark was already an indirect dependency through d2 (R10)
- [X] T002 [P] Create package skeletons, each with a `doc.go` stating its port and constraints: `internal/core/entities/viewmodel/doc.go` ("stdlib only; the intermediate value every backend consumes"), `internal/adapters/markdown/doc.go`, `internal/adapters/html/doc.go`, `internal/adapters/outputdir/doc.go`, `internal/adapters/projectfs/doc.go`, `internal/adapters/watch/doc.go`, `internal/adapters/devserver/doc.go`. Rewrite the package comment of `internal/adapters/d2` to say it renders in-process and never starts a process (FR-014)
- [X] T003 Write guard tests in `tools/archcheck/layer_test.go` proving that the audit fires for: (a) `github.com/yuin/goldmark` imported from `internal/core/usecases/`; (b) `os/exec` imported from `internal/adapters/d2/`; (c) `internal/adapters/html` imported from `internal/mcp/`. **Run them and confirm they fail.** If `ForbiddenExternalImports` in `tools/archcheck/layer.go` does not match standard-library paths such as `os/exec`, extend it so that it does, and note the change in the test comment
- [X] T004 Add the R15 rules to **all three** rule files. `tools/archcheck/rules.yaml` is the binary default. `specs/009-constitution-compliance/contracts/structural-rules.yaml` is the file `make audit-constitution` loads. `specs/010-constitution-compliance/contracts/structural-rules.yaml` uses the `layer_rules` schema. The rules: add `github.com/yuin/goldmark/**` to `forbiddenExternalImports` of `core/entities`, `core/usecases`, `mcp`, `api` and `cmd`; add `os/exec` to `forbiddenExternalImports` of **every** layer, including `adapters`; add `internal/adapters/d2/**`, `internal/adapters/html/**`, `internal/adapters/markdown/**`, `internal/adapters/outputdir/**` and `internal/adapters/devserver/**` to the `mcp` layer's `forbiddenImports` (FR-041). Confirm T003 now passes. Note: `internal/adapters/d2/renderer.go` imports `os/exec` until T037 deletes it, so run `task audit-constitution` only after T037
- [X] T005 [P] Mirror T004 into the `depguard` section of `.golangci.yml`, using the `**/`-prefixed `files` globs that feature 013's T005b established
- [X] T006 [P] Create the fixture `testdata/projects/two-systems/` (`main.loko.hcl`, `deploy.loko.hcl`, `docs/*.md`). It needs: a `project` block; 1 `person`; systems `shop` and `payments`; 6 containers (3 in each system), where `shop.api` and `payments.gateway` have 2 components each and the others have none; 1 `external` (`bank`); relationships covering person→system, container→container across systems, component→component within a container, and container→external; tags `pci` on `payments.gateway` and `edge` on `shop.web`; `docs` on every element, with one of them pointing at a **missing** file; environments `staging` (flat) and `prod` (nested `node` groups two deep); and one `system "archive"` with no containers
- [X] T007 [P] Create the fixture `testdata/projects/declared-views/main.loko.hcl` from `two-systems`, plus `view` blocks: `payment-path` (include `system.shop` and `system.payments`, exclude `container.ledger`); `pci-only` (tags `["pci"]`); `nothing` (include a system, exclude that same system, so it resolves empty); and `landscape` (include `person.customer`, which shadows the derived view)
- [X] T008 [P] Create the fixture `testdata/projects/deployment-nested/main.loko.hcl`: one environment whose `node` groups nest **five** levels deep, instances at levels 1, 3 and 5, and one relationship whose target has no instance in the environment (a boundary edge)
- [X] T009 [P] Create the fixture `testdata/projects/edge-cases/main.loko.hcl`: a self-relationship; a three-element cycle; an element whose name contains `/`, a space, and `_` (if the HCL label grammar rejects any of these characters, use the most unsafe label it accepts and note which it was in a comment); and a system with no containers. Also create `testdata/projects/case-collision/main.loko.hcl` with containers `Api` and `api`, and `testdata/projects/empty/main.loko.hcl` with only a `project` block. Finally create `testdata/projects/single-element/main.loko.hcl` with one system
- [X] T010 [P] Add `testdata/** -text` to `.gitattributes` (create the file if absent), so golden files are never CRLF-converted on Windows checkouts. The output-determinism guarantee assumes `\n` (FR-021)

**Checkpoint**: `go build ./...` is green, archcheck guard tests pass, and fixtures exist.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The `viewmodel` vocabulary, the new diagnostic codes, the ports, and the lifting algorithm that every view kind shares.

**⚠️ CRITICAL**: No user-story work can begin until this phase is complete.

- [X] T011 [P] Write table tests in `internal/core/entities/viewmodel/paths_test.go` for: `Segment` ("`[A-Za-z0-9.-]` pass through; every other byte, including `_`, becomes `_` followed by two lowercase hex digits": `payments/v2`→`payments_2fv2`, `a_b`→`a_5fb`, `a b`→`a_20b`); injectivity over 1,000 random byte strings; `DiagramFile("landscape","svg")`=`diagrams/landscape.svg`; `ViewPage(id)`=`view/<id>.html`; `ElementPage("container.api","html")`=`element/container/api.html` and `ElementPage(addr,"md")`=`md/element/container/api.md`; and `Rel(from,to)` producing correct relative hrefs between any two output paths (`element/container/api.html`→`diagrams/x.svg` = `../../diagrams/x.svg`)
- [X] T012 Implement `internal/core/entities/viewmodel/paths.go` to pass T011, following [data-model §7](./data-model.md#7-paths-pure-functions) and [contracts/output-layout.md](./contracts/output-layout.md)
- [X] T013 [P] Write tests, then implement `internal/core/entities/viewmodel/view.go`: `ViewKind` (`landscape` | `system` | `container` | `deployment` | `declared`), `ViewID`, and `View{ID, Kind, Title, Subject, Selection}`. Also add `SortViews` ("kind order (landscape, system, container, deployment, declared), then `ID`") with a test in `view_test.go`
- [X] T014 [P] Write tests, then implement `internal/core/entities/viewmodel/model.go`: `ViewModel{View, Nodes, Edges, Sources, DiagramPath, PagePath}`, `Node{ID, Address, Role, Kind, Label, Technology, Description, Parent, Style, Link}`, `NodeRole` (`element` | `subject` | `group` | `instance` | `outside`), and `Edge{ID, Source, Target, Label, Technology, Relationships, Crossing, Style}`. `Validate()` enforces: "every `Edge.Source`, `Edge.Target` and `Node.Parent` names a node that exists in `Nodes`", "at most one `outside` node exists, and it exists only if at least one edge crosses", an acyclic parent chain, `Nodes` sorted by `ID`, and `Edges` sorted by `(Source, Target, ID)`. Edge ID is `<source>--<target>`, plus `--self` for self-loops. Tests in `model_test.go` cover each violation
- [X] T015 [P] Write tests, then implement `internal/core/entities/viewmodel/style.go`: `Style{Shape, Fill, Stroke, FontColor, Dashed, Classes}`, `EdgeStyle{Dashed}`, `Shape` (`person` | `rectangle` | `boundary` | `oval`), and `StyleFor(role, kind, tags)`, returning exactly the [research R6](./research.md#r6-styling-by-convention-fr-008-fr-016019) table (person `#08427b`/`#073b6f`; system `#1168bd`/`#0b4884`; container `#438dd5`/`#3c7fc0`; component `#85bbf0`/`#78a8d8`; external `#999999`/`#8a8a8a`, dashed; subject/group boundary with no fill, stroke `#444444`, dashed; outside oval with no fill, stroke `#999999`, dashed). Classes are `kind-<kind>`, `external` where it applies, and `tag-<Segment(tag)>`, all sorted. `style_test.go` asserts every row and FR-019 (the same input always gives a deep-equal result)
- [X] T016 [P] Write tests, then implement `internal/core/entities/viewmodel/artifact.go`: `Format` (`d2` | `svg` | `md` | `html`); `ParseFormats([]string)`, which accepts comma-separated and repeated values, de-duplicates, sorts, and rejects unknown names with the exact message `unsupported format "pdf": supported formats are d2, svg, md, html` (FR-015), including `ParseFormats(["mermaid"])`, which must be rejected with the same message (FR-040); `Requires()` (`html→svg`, `md→svg`); `Expand(set)`, which returns the expanded set plus the formats it added; `Artifact{Path, Format, Bytes}`; `Manifest{Paths}`; and `ThemeFile{Name, Bytes, Origin}`
- [X] T017 [P] Write tests, then implement `internal/core/entities/viewmodel/notice.go`: `NoticeText(sources []string) string` returns `Code generated by loko from <sources>. DO NOT EDIT.`, where the sources are sorted, de-duplicated and joined with `, `. Tests assert the text is a pure function of its input: it contains no timestamp, no tool version, and no absolute path. Passing an absolute path panics, because provenance always supplies project-relative paths, so an absolute one is a programming error (FR-020)
- [X] T018 [P] Implement `internal/core/entities/viewmodel/page.go` with data-only types: `ElementPage{Address, Kind, Name, Description, Technology, Owner, Tags, Classes, Prose, ProsePath, ProseMissing, Parent *LinkRef, Children, Uses, UsedBy, Diagram, PagePath}`, `LinkRef{Address, Name, Kind, PagePath}`, `RelationRow{Relationship, Other, Description, Technology}`, `ProjectHeader{Name, Description}`, and `Projection{Project, Views, Pages, Environments}`, following [data-model §4–5](./data-model.md#4-elementpage)
- [X] T018a [P] Write tests, then implement `internal/core/entities/viewmodel/commit_plan.go`: `PlanCommit(old Manifest, next []Artifact, onDisk func(path string) ([]byte, bool)) CommitPlan{Write, Unchanged, Remove []string}`. It is pure, and every list is sorted. A path goes in `Unchanged` when its on-disk bytes equal the artifact's. `Remove` = old manifest paths absent from `next`, and never a path the old manifest does not list. A nil or empty `old` removes nothing. This keeps the ownership decision in core, so the `outputdir` adapter only performs I/O (Constitution I; analysis K1)
- [X] T019 Reserve the codes `view_empty` (warning), `view_shadowed` (warning), `output_path_collision` (error) and `theme_invalid` (error) as constants in `internal/core/entities/arch/diagnostic.go` and in the `code` enum of `specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`. Do **not** add them to `arch.AllCodes` yet. `TestValidationRuleCoverage` (`internal/core/usecases/rule_coverage_test.go`) fails for any code in `AllCodes` that no core test or golden produces, so each code joins `AllCodes`, and `codesFromCoreTests`, in the task that first produces it: `output_path_collision` in T049, `view_empty`/`view_shadowed` in T058, and `theme_invalid` in T092. Adjust `TestAllCodesAreKnown` in each of those tasks. If `TestAllCodesAreKnown` compares `AllCodes` against the schema enum, update it here so that reserved-but-unproduced codes are allowed
- [X] T020 Add the ports from [contracts/ports.md](./contracts/ports.md) to `internal/core/usecases/ports.go`, verbatim in shape: `Backend`, `RenderOptions`, `ProseReader`, `ThemeSource`, `ArtifactStore`, `CommitReport`, `ChangeWatcher`, `WatchSpec`, and `PreviewServer`. Also add `ThemeError{File, Line int, Message string}` (an `error` a backend returns for a malformed override, which becomes `theme_invalid`). Create concrete mocks in `internal/core/usecases/render_mocks_test.go`: a `fakeBackend` that records the projection it received and returns canned artifacts, `fakeProse`, `fakeTheme`, `fakeStore`, `fakeWatcher` (a channel driven by the test), and `fakePreview`
- [X] T021 Write tests in `internal/core/usecases/provenance_test.go` using `arch.SourceModel` literals: `BuildProvenance(model)` records, for each element, environment, group, instance and view address, its declaring `arch.SourceRange`. `RangeOf(addr)` returns that range (the view warnings in T054 use it), and `SourcesFor([]arch.Address)` returns the sorted, de-duplicated `Range.File` values. Read `internal/core/entities/arch/source_model.go` and `source_model_deployment.go` for the declaration types and ranges. Implement `internal/core/usecases/provenance.go`. Store the result as a sorted slice with binary search, not an exported map
- [X] T022 Write table tests in `internal/core/usecases/project_lift_test.go` over small `arch.IR` literals (built with `arch.NewIR`) for the four [research R4](./research.md#r4-what-each-view-contains-projection-rules) rules, given an arbitrary visible set: (1) both lifted ends visible and distinct → an edge, where several relationships on the same pair merge into one edge whose `Relationships` are sorted and whose label is the single description or `"N relationships"`, and whose `Technology` is set only when all agree; (2) the same lifted end → kept as a self-loop **only** for an authored `Source == Target`, and dropped when it arises from lifting; (3) exactly one end visible → a crossing edge to the single `outside` node, merged per `(inside element, direction)`; (4) neither visible → absent
- [X] T023 Implement `internal/core/usecases/project_lift.go` (`liftEdges(ir, visible, nodeIDFor) []viewmodel.Edge`, plus the `outside` node when needed) to pass T022. Keep each function ≤ 60 effective lines
- [X] T023a Run `/speckit-constitution` to amend `.specify/memory/constitution.md` to **v1.4.0 (MINOR)**, with a Sync Impact Report. This must happen **before Phase 3**, because T040 depends on it. Changes: Principle I, "Dependencies are injected at startup in `main.go`" becomes "Dependencies are injected in the composition root: `main.go` and `cmd/`, the outermost layer"; Principle II, "Wiring (interface → implementation) happens only in `main.go`" becomes "… happens only in `main.go` or `cmd/`, never in `internal/**`". The rationale to record: `cmd/` is already the outer composition root in the Dependency Direction table, and `cmd/validate.go` and `cmd/export.go` already wire adapters. Fold in the stale-text fixes: Principle VII's "Single binary with no runtime dependencies except d2 (and optionally veve-cli)" becomes "Single binary with no runtime dependencies"; File Organization gains `entities/viewmodel/` and the adapters `markdown`, `html`, `outputdir`, `projectfs`, `watch` and `devserver`; External Dependencies gains goldmark (confined to adapters); the layer-rule note gains `os/exec` and the mcp render-adapter ban; and Quality Gates → "Entity files ≤ 300 effective lines" becomes "≤ 200", matching Architecture Rules. Sync `specs/009-constitution-compliance/contracts/structural-rules.yaml`, so the CI cross-check passes

**Checkpoint**: The entities, ports and lifting are green, and `task audit-constitution` passes for all new files.

---

## Phase 3: User Story 1 - See the architecture without configuring anything (Priority: P1) 🎯 MVP

**Goal**: `loko build` with no arguments writes D2 and SVG for the landscape, each system with containers, each container with components, and each environment.

**Independent Test**: `loko build -p testdata/projects/two-systems --out /tmp/o` produces the expected `diagrams/*.{d2,svg}` set and nothing for `system.archive` ([quickstart §1–2](./quickstart.md#1-zero-config-diagrams-us1-fr-001002-sc-001)).

### Tests for User Story 1 ⚠️ write first, confirm failing

- [X] T024 [P] [US1] Write tests in `internal/core/usecases/resolve_views_test.go` for derived views, following [research R3](./research.md#r3-where-views-come-from): `landscape` exists iff there is at least one element; `system-<name>` exists iff the system has ≥ 1 container; `container-<name>` exists iff it has ≥ 1 component; `deployment-<env>` exists iff the environment has ≥ 1 instance; zero elements gives no views and no diagnostic (FR-002). The output is sorted per `SortViews`. IDs use `viewmodel.Segment` on the name
- [X] T025 [P] [US1] Write tests in `internal/core/usecases/project_view_test.go` for visible sets and nesting, over `arch.IR` literals. **Landscape**: every parentless element, with no crossing edges. **System S**: an S `subject` node, S's containers nested under it, and neighbours lifted to their top-level ancestor placed outside S. **Container C**: a C `subject`, its components nested, neighbours in the same system lifted to sibling containers, and neighbours elsewhere lifted to top-level. Every node's `Style` equals `StyleFor(...)`, `Link` equals `viewmodel.ElementPage(addr,"html")`, and `(*ViewModel).Validate()` returns nil. `Sources` equals `SourcesFor` of the depicted addresses
- [X] T026 [P] [US1] Write tests in `internal/core/usecases/project_deployment_test.go`: the group tree is reproduced exactly as nested `group` nodes (a five-deep case stays five deep); instances are placed in their `PlacedIn` group with role `instance` and the style of their `Of` kind; edges come from logical relationships between the `Of` elements, lifted to the nearest ancestor-or-self **instantiated in this environment**; and a relationship whose other end has no instance in the environment becomes a crossing edge
- [X] T027 [P] [US1] Write golden tests in `internal/adapters/d2/emit_test.go`. Inputs are `ViewModel` JSON files in `internal/adapters/d2/testdata/views/*.json`, and expected output is `testdata/golden/*.d2`. Cover a landscape, a nested subject with neighbours, a deployment five levels deep, a self-loop, a crossing edge to `outside`, labels containing quotes, newlines and `:` (escaped correctly), and tags. Assert that line 1 is `# ` + `NoticeText(vm.Sources)`, that each shape maps (`person`→`shape: c4-person`, `rectangle`→`shape: rectangle`, `boundary`→a container with `style.fill: transparent` and `style.stroke-dash: 4`, `oval`→`shape: oval`), that fill, stroke and font colour come from `Style` only, that every `Style.Classes` value is declared in a `classes: {}` block and referenced by `class: [...]`, and that output is identical across two calls
- [X] T028 [P] [US1] Write tests in `internal/adapters/d2/svg_backend_test.go` with `t.Setenv("PATH", "")`: rendering the `two-systems` landscape view model succeeds; the output starts with `<?xml`, and the XML declaration is followed immediately by `<!-- Code generated by loko from … DO NOT EDIT. -->`; each `tag-*` and `kind-*` class appears in a `class="…"` attribute of the SVG (R6 guard: if d2 v0.7.1 does not emit classes, the test fails loudly and the gap is recorded in research.md rather than papered over); two renders are byte-identical; and a second render of the same D2 source is served from the cache (assert through an exported-for-test hit counter)
- [X] T029 [P] [US1] Write tests in `internal/adapters/outputdir/store_test.go` for basic `Commit`: it creates `outDir` if absent; it writes every artifact at `outDir/Path` with parent directories; it returns an error naming `outDir` when the directory cannot be written (chmod `0o500`, skipped on Windows and when running as root), and leaves **no** file behind (staging under `outDir/.loko-staging-*` is removed); and no staging directory survives a successful commit
- [X] T030 [P] [US1] Write tests in `internal/core/usecases/build_artifacts_test.go` using the Phase 2 mocks. Compile errors mean no backend is called, there are no artifacts, and the diagnostics are returned. `Formats=[html]` runs the `svg` backend too, with `AddedFormats=[svg]`. Artifacts from all backends are merged and sorted by `Path`. Zero elements gives no artifacts and `NothingToDraw=true`. Compile warnings pass through to `Diags`. `BuildResult.ExitCode(strict)` mirrors `CompileResult.ExitCode`. Also write `build_site_test.go`: `Build` calls `ArtifactStore.Commit` exactly once on success and **never** when there are errors (FR-024)
- [X] T031 [US1] Write end-to-end tests in `cmd/build_test.go`, calling `runBuildWith` directly as `cmd/export_test.go` does. For `two-systems` with no flags, the `diagrams/` listing equals the expected set (US1/AC1–3) and exit is `0`. The `prod` view's D2 contains both levels of `node` nesting (AC2). A view with a cross-boundary relationship contains an edge to `outside` (AC4). For a broken copy of the fixture, exit is `1`, diagnostics are on stderr, and a pre-populated output directory is byte-for-byte unchanged (AC5, compared by walking and hashing)

### Implementation for User Story 1

- [X] T032 [US1] Implement derived views in `internal/core/usecases/resolve_views.go` (`ResolveViews(ir) ([]viewmodel.View, arch.Diagnostics)`, derived part only) to pass T024
- [X] T033 [US1] Implement `internal/core/usecases/project_view.go` (`ProjectView(ir, view, prov) viewmodel.ViewModel` for landscape, system and container: visible set, `subject` node, nesting, and nodes built with `StyleFor` and `Paths`, delegating edges to `liftEdges`) to pass T025
- [X] T034 [US1] Implement `internal/core/usecases/project_deployment.go` (group-tree nodes, instance nodes, instance-level lifting that reuses `liftEdges` with a visible set of instantiated elements) to pass T026
- [X] T035 [US1] Implement `internal/core/usecases/project.go`: `Project(ir, prose, prov) (*viewmodel.Projection, arch.Diagnostics)` calls `ResolveViews`, projects each view, sorts, calls `Validate()` on each view model (a failure is a programming error, returned as an `error`), and leaves `Pages` empty until US4. It is a pure function (FR-010): add a test in `project_test.go` asserting that two calls give deep-equal results
- [X] T036 [US1] Implement `internal/adapters/d2/emit.go` (`Emit(vm viewmodel.ViewModel) []byte`, notice included) to pass T027, and `internal/adapters/d2/source_backend.go` (`Backend` for format `d2`, one artifact per view at `vm.DiagramPath` with the `.d2` extension via `viewmodel.DiagramFile`)
- [X] T037 [US1] Implement `internal/adapters/d2/svg_backend.go` to pass T028. `Emit` the D2 source, then call `d2lib.Compile` with `d2elklayout` and a `textmeasure.NewRuler()`, then `d2svg.Render` with d2's default pad and theme 0. Insert the notice after the XML declaration. Keep a SHA-256 → SVG in-memory cache behind a `sync.Mutex`. Render views on a pool of `runtime.GOMAXPROCS(0)` workers, writing each result into its view's index so order never depends on completion. **Delete** `internal/adapters/d2/renderer.go` and its test files (the `exec.LookPath` renderer), then run `task audit-constitution` and confirm 0 violations, including the new `os/exec` ban
- [X] T038 [US1] Implement basic staging and write in `internal/adapters/outputdir/store.go` (`New() *Store`, `Commit`) to pass T029. Manifest and pruning come in US3
- [X] T039 [US1] Implement `internal/core/usecases/build_artifacts.go` (`BuildRequest{Root, BuildVersion, Formats []string, OutDir}`, `BuildDeps{Source, Prose, Theme, Backends map→ sorted []Backend, Store}`, and `BuildResult{Artifacts, Diags, AddedFormats, NothingToDraw, Report}`), plus `internal/core/usecases/build_site.go` (`Build`) to pass T030. Run backends concurrently with results collected by index. Build provenance from `compiled.Model`, and the IR with the existing `BuildIR`. Keep each function ≤ 60 lines. If a file approaches 200, split out `build_formats.go`
- [X] T040 [US1] Create `cmd/wiring.go`: `newBuildDeps(root string) usecases.BuildDeps` constructs `hclsource.New()`, `projectfs` (a stub until US4/US7: return a no-op prose reader and theme source defined in `projectfs`), the `d2` and `svg` backends, and `outputdir.New()`. This is the single composition root for `build` and `serve`, which Constitution v1.4.0 (T023a) permits in `cmd/`. Depends on T023a
- [X] T041 [US1] Implement `cmd/build.go` (`runBuildWith(ctx, BuildOptions) (int, error)`) and `cmd/build_cobra.go` (`--format` as a string slice defaulting to `d2,svg,md,html`, `--out`, where an empty value means `filepath.Join(ProjectRoot, "dist")` and an explicit value is used as typed, resolved against the working directory as `loko export --out` is; `--strict`; `GroupID: "building"`) per [contracts/cli.md](./contracts/cli.md#loko-build). Reuse the diagnostic rendering pattern from `cmd/export.go`. Print the summary line `built <out>: N written, N unchanged, N removed` and the `added svg (required by html)` note. Until US4 registers `md` and `html`, the default set is `d2,svg`: make the default derive from the registered backends, so that US4 widens it without touching this file. Pass T031
- [X] T042 [US1] Update `cmd/root_test.go`: move `build` from the "removed" map to the pinned command set. `serve` stays in "removed" until T083, and `watch` stays removed permanently, with its reason text "folds into serve"

**Checkpoint**: `loko build -p testdata/projects/two-systems` produces every derived D2 and SVG with no configuration, and `task test` and `task audit-constitution` are green.

---

## Phase 4: User Story 3 - Trust that generated output is generated (Priority: P1) 🎯 MVP

**Goal**: Byte-identical rebuilds, a notice on every file, pruning of the tool's own orphans, and localised diffs.

**Independent Test**: Building twice gives identical bytes, and every file's first two lines carry the notice ([quickstart §3–4](./quickstart.md#3-determinism-us3-fr-021-sc-002)).

### Tests for User Story 3 ⚠️ write first, confirm failing

- [X] T043 [P] [US3] Extend `internal/adapters/outputdir/store_test.go`. `.loko-manifest` is written last, with `# ` + the notice on line 1 and then sorted paths one per line. A second commit of identical artifacts rewrites nothing: every path is in `CommitReport.Unchanged` and modification times are preserved. A path that was in the old manifest but is absent from the new set is deleted and listed in `Removed` (FR-023). A user file not named in the old manifest survives untouched (edge case: the output directory is not empty). A missing or corrupt manifest means nothing is pruned
- [X] T044 [P] [US3] Write tests in `internal/core/usecases/artifact_paths_test.go`: two artifacts with equal `Path`, or with paths equal under `strings.ToLower`, produce one `output_path_collision` **error** per pair naming both depicted addresses, and no commit happens (FR-028). Paths derived from unsafe names are stable across calls (FR-006)
- [X] T045 [US3] Write `cmd/build_determinism_test.go`. `TestBuildDeterministic` builds every `testdata/projects/*` fixture (excluding `case-collision`) twice into separate directories through `runBuildWith`, and asserts byte-identical trees (US3/AC1, SC-002). It uses no entity types
- [X] T045a [P] [US3] Write `internal/adapters/hclsource/render_order_test.go` (`TestRenderDeterministicDiscoveryOrder`). Wrap `hclsource.New()` in a test `usecases.ArchitectureSource` that reverses every slice of the returned `*arch.SourceModel` (`Files`, `Elements`, `Environments`, `Views`, and nested relations). For each fixture, run `usecases.BuildArtifacts` with the real `d2`, `markdown` and `html` backends, both normally and reversed, and assert that the artifact sets are byte-identical (US3/AC2, SC-002). This lives in the adapter layer because `cmd/` tests may not import `internal/core/entities/**`: tests remain subject to layer-import rules (analysis C1). Until T069 and T072 land, cover only the backends that exist, and extend this test in T073
- [X] T046 [P] [US3] Write `cmd/build_notice_test.go` (`TestBuildNotice`). For every fixture and every format, walk the output tree and assert each file contains `Code generated by loko from ` and `DO NOT EDIT.` within its first two lines, with every named source existing under the fixture root (FR-020, SC-004)
- [X] T047 [P] [US3] Write `cmd/build_change_test.go`. `TestBuildLocalChange` copies `two-systems`, builds, edits one **component**'s `description`, rebuilds, and asserts that at least 90% of files are byte-unchanged (SC-003) and that every changed file is in the permitted set from [contracts/output-layout.md § Determinism](./contracts/output-layout.md#determinism-fr-021-fr-022) (US3/AC4). `TestBuildPrune` deletes one container, rebuilds, and asserts its files are gone and that a user file `dist/NOTES.txt` survives (US3/AC5)

### Implementation for User Story 3

- [X] T048 [US3] Implement the manifest in `internal/adapters/outputdir/manifest.go` (read and write `.loko-manifest`), and extend `store.go` so it reads the old manifest and on-disk bytes, calls `viewmodel.PlanCommit` (T018a) to decide writes, skips and removals, and **executes** that plan (staging, rename, delete). The adapter makes no ownership decisions of its own. Pass T043. Keep each file ≤ 400 lines
- [X] T049 [US3] Implement `internal/core/usecases/artifact_paths.go` (`checkCollisions(artifacts, owners) arch.Diagnostics`, where `owners` maps each path to the address it depicts, carried alongside artifacts as a sorted slice) and call it from `BuildArtifacts` before returning, to pass T044. Add `output_path_collision` to `arch.AllCodes`, and a `checkCollisions` case to `codesFromCoreTests` in `rule_coverage_test.go` (T019)
- [X] T050 [US3] Fix any nondeterminism that T045–T047 expose. Every fix goes at the source: sort in the projection, not in a backend. Record each finding and its fix in a comment on the test that exposed it
- [X] T051 [US3] Add a `determinism` job to `.github/workflows/ci.yml` with `strategy.matrix.os: [ubuntu-latest, macos-latest]`, running `go test ./cmd/ -run 'TestBuild(Deterministic|Notice)' -count=1`, `go test ./internal/adapters/hclsource/ -run TestRenderDeterministicDiscoveryOrder -count=1`, and the golden suites `go test ./internal/adapters/d2/ ./internal/adapters/markdown/ ./internal/adapters/html/ -count=1` (SC-002)
- [X] T052 [US3] Add `cmd/build_edge_test.go` for `case-collision`: exit `1`, and an `output_path_collision` diagnostic naming `container.Api` and `container.api`, with nothing written. This is the end-to-end counterpart of the core coverage case added in T049

**Checkpoint (MVP)**: US1 and US3 are complete. The output is committable, reviewable, and trustworthy.

---

## Phase 5: User Story 2 - Narrow the picture to one concern (Priority: P2)

**Goal**: Declared `view` blocks render alongside derived views, honour include, exclude and tags, and win on a name collision.

**Independent Test**: Building `declared-views` produces `payment-path` without `ledger`, with its edges reaching `outside`. `nothing` warns and produces no file. `landscape` warns that it shadows the derived view ([quickstart §5](./quickstart.md#5-declared-views-us2-fr-003005)).

### Tests for User Story 2 ⚠️ write first, confirm failing

- [X] T053 [P] [US2] Write table tests in `internal/core/usecases/select_declared_test.go` for "(`include` ∪ descendants of `include`) ∪ (every element carrying any tag in `tags`), minus (`exclude` ∪ descendants of `exclude`). When both `include` and `tags` are absent, the base set is every element." Cover include-only, tags-only (OR semantics across tags), include plus tags, an exclude that removes a subtree, and exclude-everything giving an empty set
- [X] T054 [P] [US2] Extend `internal/core/usecases/resolve_views_test.go`. A declared view gets ID `Segment(label)`, kind `declared`, and subject `view.<label>`. A declared view whose ID equals a derived ID replaces it and emits a `view_shadowed` **warning** naming the derived view, with `Range` taken from the view block's provenance (FR-004). A declared view that selects nothing emits a `view_empty` **warning** naming it, and is not produced (FR-005)
- [X] T055 [P] [US2] Extend `internal/core/usecases/project_view_test.go` for the declared kind: visible elements are nested under their nearest visible ancestor, relationships ending on excluded elements become crossing edges to `outside` (US2/AC2), and there is no `subject` node
- [X] T056 [US2] Write `cmd/build_views_test.go` over `declared-views`: `diagrams/payment-path.{d2,svg}` exists, its D2 has no `ledger` node and has an `outside` edge; `diagrams/pci-only.*` contains every `pci`-tagged element (US2/AC3); `diagrams/nothing.*` is absent and stderr has `view_empty`; `diagrams/landscape.d2` contains only `person.customer`, and stderr has `view_shadowed`; the exit code is `0`, or `2` with `--strict`

### Implementation for User Story 2

- [X] T057 [US2] Implement `internal/core/usecases/select_declared.go` to pass T053
- [X] T058 [US2] Extend `internal/core/usecases/resolve_views.go` with declared views, shadowing and the empty warning to pass T054. Add `view_empty` and `view_shadowed` to `arch.AllCodes`, and a `ResolveViews` case producing each to `codesFromCoreTests` (T019)
- [X] T059 [US2] Extend `internal/core/usecases/project_view.go` with the declared kind to pass T055 and T056. If the file would exceed 200 effective lines, move the declared kind to `project_declared.go`
- [X] T060 [US2] Add projection golden tests in `cmd/projection_golden_test.go` (`TestProjectionGolden`, with one subtest per fixture named after its directory): for each of `two-systems`, `declared-views`, `deployment-nested` and `edge-cases`, compile, run `usecases.Project`, JSON-encode it with sorted, indented output, and compare against `testdata/projects/<fixture>/expected/projection.json`, with `-update` support. This covers every view kind (SC-007)

**Checkpoint**: Declared views work, and SC-007 fixtures cover all five view kinds.

---

## Phase 6: User Story 4 - Read the architecture as prose and pictures together (Priority: P2)

**Goal**: The `md` and `html` backends: one page per element, with prose inlined, uses and used-by tables, children, the embedded diagram, and no broken links.

**Independent Test**: Build `two-systems` with `--format html`, open `index.html`, navigate to `shop.api`, and read its prose. The link checker reports zero broken links ([quickstart §7](./quickstart.md#7-site-navigation-and-links-us4-fr-025028-sc-008)).

### Tests for User Story 4 ⚠️ write first, confirm failing

- [X] T061 [P] [US4] Write tests in `internal/adapters/projectfs/prose_test.go`. `ReadProse(root, docs)` returns the text and `found=true` for an existing file. It returns `found=false` and a nil error for a missing file (FR-026). It returns `found=false` for a path escaping the root (`../x.md`). Containment must match `docsExist` in `internal/core/usecases/validate_warnings.go`, so add a table test that runs the same inputs through both and asserts agreement
- [X] T062 [P] [US4] Write tests in `internal/core/usecases/project_pages_test.go`. Each element gets one page: `Children` sorted by address; `Uses`/`UsedBy` "sorted by the other element's address, then the relationship's address"; rows carrying **no** neighbour description; `Parent` is nil for top-level elements. The `Diagram` rule: its own subject view when one exists, otherwise its parent's view, otherwise `landscape`. Persons, externals, and a system with no containers (`system.archive` in `two-systems`) get `landscape`, and a page never names a view that was not produced. `ProseMissing=true` with `ProsePath` set when the prose reader says the file was not found. `Classes` equal the node style classes. `PagePath` comes from `viewmodel.ElementPage`. `Projection.Environments` links to the deployment view pages
- [X] T063 [P] [US4] Write golden tests in `internal/adapters/markdown/backend_test.go`. Input is a `Projection` JSON in `testdata/projection.json`, and output goes to `testdata/golden/**`. Line 1 is `<!-- ` + notice + ` -->`. There is `md/index.md` (project header, the landscape image via `viewmodel.Rel`, a table of views, a table of elements), `md/view/<id>.md` (image, a node table, and a "Leaves this view" list for crossing edges), and `md/element/<kind>/<name>.md` (prose inlined **verbatim**, a "Prose file `<ProsePath>` not found" note when it is missing, uses and used-by tables with relative links, children, and the diagram image). Two renders are byte-identical
- [X] T064 [P] [US4] Write tests in `internal/adapters/html/prose_test.go`. goldmark renders CommonMark and GFM tables. Raw HTML in prose is escaped, not passed through (`<script>` does not survive). Output is deterministic
- [X] T065 [P] [US4] Write golden tests in `internal/adapters/html/render_test.go` (same projection-JSON input approach as T063). Check `index.html`, `view/<id>.html`, `element/<kind>/<name>.html`, `assets/style.css`, `assets/custom.css` and `assets/site.js`. `<!DOCTYPE html>` is line 1 and the `<!-- notice -->` is line 2 (the notice in `/* */` is line 1 for CSS and JS). Diagrams are embedded as `<img src="…/diagrams/<id>.svg">` via `viewmodel.Rel`. The page `<body>` and element cards carry `Style.Classes` (FR-018). Descriptions containing `<b>` are escaped (html/template). There is no `/_loko/events` or live-reload code anywhere in the output. The prose-missing note is present. Two renders are byte-identical
- [X] T066 [US4] Write `cmd/site_links_test.go` (`TestSiteLinksResolve`). For each of `two-systems`, `declared-views`, `deployment-nested` and `edge-cases`: run `runBuildWith` with `--format html` into a temp directory (which also produces the required `svg`), parse every `href` and `src` in every HTML file, resolve each against the page's path, and assert the target file exists in the output tree. Living in `cmd/` keeps the html adapter's tests free of the svg backend (FR-012; analysis D2) and uses no entity types. External `http(s)://` links and `#fragment`-only links are skipped. Zero broken links (FR-027, SC-008)

### Implementation for User Story 4

- [X] T067 [US4] Implement `internal/adapters/projectfs/prose.go` to pass T061, and replace the T040 prose stub
- [X] T068 [US4] Implement `internal/core/usecases/project_pages.go` to pass T062, wire it into `Project` in `project.go`, and read prose in `BuildArtifacts` **before** projection, once per distinct `docs` value, in address order. Do not emit a duplicate warning for missing prose: the compiler already reports `docs_not_found`
- [X] T069 [US4] Implement `internal/adapters/markdown/backend.go` (split `page.go` and `view.go` if it approaches 400 lines) to pass T063
- [X] T070 [US4] Split `internal/_parked/html/templates.go` into embedded assets under `internal/adapters/html/theme/`: `layout.gohtml` (blocks `layout`, `head`, `header`, `nav`, `footer`), `index.gohtml` (`index`), `element.gohtml` (`element`, `element-tables`, `element-children`, `prose-missing`), `view.gohtml` (`view`), `partials.gohtml` (`tag-chips`, `link`, `diagram`), `style.css` (from `cssContent`), `custom.css` (empty apart from its notice), and `site.js` (from `jsContent`: search and navigation, with no reload code). Rewrite the templates for **`html/template`** against `PageData{Project, Page, Root}` and the `viewmodel` types, dropping every reference to v0 fields. Load them with `//go:embed theme/*` in `internal/adapters/html/assets.go`
- [X] T071 [US4] Implement `internal/adapters/html/prose.go` (goldmark with `html.WithUnsafe()` **off** and `extension.Table`) to pass T064
- [X] T072 [US4] Implement `internal/adapters/html/render.go` (the `Backend` for format `html`) and `internal/adapters/html/theme.go` (built-in template set only, with overrides in US7) to pass T065 and T066. Keep each file ≤ 400 effective lines
- [X] T073 [US4] Register the `md` and `html` backends in `cmd/wiring.go`, so the default `--format` becomes `d2,svg,md,html` (FR-013). Extend `cmd/build_test.go`: for `two-systems` with `--format html`, stdout contains `added svg (required by html)`, `element/container/api.html` contains the prose text and the uses table, and the page for the element with missing prose exists and contains the not-found note (US4/AC1–3). Extend T045a to include the `markdown` and `html` backends
- [X] T074 [US4] Delete `internal/_parked/html/` and update `internal/_parked/README.md` to remove its row and state that the site builder returned in feature 014

**Checkpoint**: Every backend of FR-013 exists, and the site is navigable with zero broken links.

---

## Phase 7: User Story 5 - See changes as they are made (Priority: P2)

**Goal**: `loko serve` rebuilds when the watched set changes, live-reloads the browser, shows diagnostics on failure, and recovers without a restart.

**Independent Test**: Run `loko serve`, edit a description, and see the browser update within 2 s. A syntax error shows its file, line and column. Fixing it recovers ([quickstart §8](./quickstart.md#8-live-preview-us5-fr-029033-sc-006)).

### Tests for User Story 5 ⚠️ write first, confirm failing

- [X] T075 [P] [US5] Write tests in `internal/adapters/watch/poller_test.go` with an injected tick channel and stat function, so no real sleeping happens. One modified `*.loko.hcl` (mtime or size change) produces **one** signal after the set has been stable for one further tick. Three files changed on consecutive ticks give one signal (FR-033). A created or deleted `*.loko.hcl` is detected. A change to `README.md`, to a file under `WatchSpec.Exclude` (the output directory), or to anything under `dist/` produces no signal (FR-032). Files returned by `ExtraFiles()` (prose) are watched, and `ExtraFiles` is re-read on every tick. Files under `templates/` are watched. The channel closes when the context is cancelled
- [X] T076 [P] [US5] Write tests in `internal/adapters/devserver/server_test.go` using `httptest`. After `Publish(artifacts)`, `GET /` serves `index.html` and `GET /assets/style.css` serves `text/css`, both from memory. HTML responses have a `<script>` injected just before `</body>` that opens an `EventSource` on `/_loko/events`, while the artifact bytes themselves are unchanged (assert against the input). An SSE client receives `event: reload` after each `Publish` and `Fail`. After `Fail(text)`, every `.html` path returns a diagnostics page containing the text (HTML-escaped), while `.svg` and `.css` are still served. A later `Publish` clears the error state. The server binds only `127.0.0.1`
- [X] T077 [P] [US5] Write tests in `internal/core/usecases/serve_site_test.go` with mocks. A failing initial compile calls `Fail`, does **not** return an error, and keeps waiting on the watcher. A later successful build after a watcher signal calls `Publish` (FR-031). Each signal gives exactly one build. `WatchSpec.ExtraFiles` returns the prose paths of the last **good** build. `WatchSpec.Exclude` contains the **absolute** output directory, so exclusion holds wherever `--out` points. `WatchSpec` covers `templates/`. Cancelling the context returns nil. Formats are fixed to `html` and its requirements. `ArtifactStore.Commit` is never called: serve does not write to disk

### Implementation for User Story 5

- [X] T078 [US5] Implement `internal/adapters/watch/poller.go` (`New(interval time.Duration) *Poller`, where the production interval is 200 ms and settling takes one tick, per research R12), discovering sources with `hclsource.Discover` so the discovery rules are shared, to pass T075
- [X] T079 [US5] Implement `internal/adapters/devserver/server.go`, `sse.go` and `inject.go` (`New(addr string) *Server`, `ListenAndServe(ctx)`, `Publish` and `Fail` guarded by a `sync.RWMutex`) to pass T076. Each file must stay ≤ 400 lines
- [X] T080 [US5] Implement `internal/core/usecases/serve_site.go` (`Serve(ctx, deps ServeDeps, req ServeRequest) error`, reusing `BuildArtifacts`) to pass T077. The diagnostics text for `Fail` comes from a `DiagnosticsFormatter` func field on `ServeDeps`, which cmd supplies from `hclsource.NewRendererForRoot(root, false)`, so core stays renderer-free. Pass the output directory to `WatchSpec.Exclude` as an absolute path
- [X] T081 [US5] Implement `cmd/serve.go` (`runServeWith`) and `cmd/serve_cobra.go` (`--port` defaulting to 8080, `GroupID: "serving"`) per [contracts/cli.md § serve](./contracts/cli.md#loko-serve). Handle SIGINT and SIGTERM via `signal.NotifyContext`, print `serving http://127.0.0.1:<port> (ctrl-c to stop)`, exit `1` on a bind failure, and do not exit when the first compile fails. Extend `newBuildDeps` in `cmd/wiring.go` with `newServeDeps`
- [X] T082 [US5] Write `cmd/serve_test.go` as an end-to-end test on a free port. Copy `two-systems` to a temp directory, start `runServeWith` in a goroutine, subscribe to `/_loko/events`, edit a description, and assert a `reload` event and changed content within 2 s. Write a syntax error and assert that `/` returns the diagnostics page with `file:line:col`. Fix it and assert recovery. Touch `README.md` and assert no reload within 1 s
- [X] T083 [US5] Update `cmd/root_test.go`: move `serve` into the pinned command set. `watch` stays removed

**Checkpoint**: The edit loop takes seconds, and failures are visible and recoverable.

---

## Phase 8: User Story 6 - Install one thing (Priority: P2)

**Goal**: Prove there is no external process or executable requirement, and remove d2 from the distribution.

**Independent Test**: Build every format with an emptied `PATH` ([quickstart §6](./quickstart.md#6-no-external-process-us6-fr-014-sc-005)).

- [X] T084 [US6] Write `cmd/build_nopath_test.go` (`TestBuildWithEmptyPath`). Call `t.Setenv("PATH", "")` and `t.Setenv("HOME", t.TempDir())`, then call `runBuildWith` **in-process** with every format on `two-systems`. Assert exit `0` and that every expected SVG exists (US6/AC1, SC-005). No subprocess is started: the `os/exec` ban (T004) applies to test files too, and `rules.yaml` forbids exempting layer rules (analysis C2). The binary-level check is T085's container run
- [X] T085 [US6] Remove the `d2-builder` stage and the `COPY --from=d2-builder` line from `Dockerfile`. Make the final image contain only the loko binary (`FROM scratch` or `distroless/static`, whichever matches the existing release conventions in `.goreleaser.yaml`). Drop any d2-related entries from `.goreleaser.yaml`. Run `task docker-build` and, inside the image, `loko build -p /fixture --out /tmp/o` against a mounted `two-systems` fixture. The image has no shell and no other executables, so a successful `loko build` inside it is the proof for US6/AC2
- [X] T086 [US6] Run `rg -n 'exec\.|LookPath|"os/exec"' internal cmd --glob '!*_test.go'` and confirm there are no matches. Run `task audit-constitution` and confirm 0 violations (US6/AC3, FR-014)

**Checkpoint**: A single static binary with no companion installation.

---

## Phase 9: User Story 7 - Change how it looks without forking (Priority: P3)

**Goal**: `templates/` overrides replace built-in presentation block by block or file by file, and a malformed override fails loudly.

**Independent Test**: Override the `header` block and see it used with everything else unchanged. An unknown block fails with `theme_invalid` naming the file ([quickstart §9](./quickstart.md#9-theme-override-us7-fr-034035-sc-009)).

### Tests for User Story 7 ⚠️ write first, confirm failing

- [X] T087 [P] [US7] Write tests in `internal/adapters/projectfs/theme_test.go`. A missing `templates/` directory returns `(nil, nil)` (US7/AC2). Only `.gohtml`, `.css` and `.js` files are returned, sorted by `Name`, with `Origin` set to `templates/<name>` (forward slashes). Other extensions and subdirectories are ignored
- [X] T087a [P] [US7] Write tests, then implement `internal/core/entities/viewmodel/theme_rules.go`: `ValidateTheme(files []ThemeFile, overridable []string) []ThemeProblem{Origin, Message}`. It flags every `.gohtml`, `.css` or `.js` file whose name is not in `overridable`, and each message lists the overridable names. It is pure, and its results are sorted by `Origin`. Template parse errors and `{{define}}` of an unknown block stay in the html adapter (T091), because only the template parser can see them. ADR-0013 (T097) records this split (analysis K1)
- [X] T088 [P] [US7] Write tests in `internal/adapters/html/theme_test.go`, following [contracts/theme.md](./contracts/theme.md). A `layout.gohtml` override defining only `header` changes the header, and every other byte of the output equals the built-in render (US7/AC1). A `style.css` override replaces the file byte for byte, apart from the notice. A `custom.css` override appears at `assets/custom.css`. Each of these returns a `usecases.ThemeError` naming the file: an unknown file name (`sytle.css`, with the error listing the overridable names), a template parse error (with line), and `{{define "nope"}}` (with the block name). Rendering twice with the same overrides gives byte-identical output (US7/AC4)
- [X] T089 [P] [US7] Extend `internal/core/usecases/build_artifacts_test.go`. The theme from `ThemeSource` reaches the backend in `RenderOptions.Theme`. A `ThemeError` from a backend becomes one `theme_invalid` **error** diagnostic whose `Range.File` is the override's `Origin` and whose line is `Line`, and nothing is committed (FR-035).

### Implementation for User Story 7

- [X] T090 [US7] Implement `internal/adapters/projectfs/theme.go` to pass T087, and replace the T040 theme stub
- [X] T091 [US7] Extend `internal/adapters/html/theme.go` with override resolution to pass T088. Call `viewmodel.ValidateTheme` (T087a) for unknown file names, converting each `ThemeProblem` to a `usecases.ThemeError`, then parse the built-ins first, then each `.gohtml` override in name order. Reject a `{{define}}` of any name absent from the built-in set (collect built-in names before applying overrides). Replace `.css` and `.js` files wholesale
- [X] T092 [US7] Wire `ThemeSource` into `BuildArtifacts` and map `ThemeError` to `theme_invalid` in `internal/core/usecases/build_artifacts.go` (split into `build_theme.go` if the file approaches 200 lines) to pass T089. Add `theme_invalid` to `arch.AllCodes`, and a case producing it (a `fakeBackend` returning `ThemeError`) to `codesFromCoreTests` (T019)
- [X] T093 [US7] Write `cmd/build_theme_test.go`. Copy `two-systems` and add `templates/layout.gohtml` with a custom header: the build succeeds and the header is present in every HTML page, while `diagrams/*` are byte-identical to an unthemed build. Add `templates/partials.gohtml` defining `nope`: exit `1`, stderr names `templates/partials.gohtml` and `nope`, and the output directory is unchanged
- [X] T094 [US7] Confirm `TestValidationRuleCoverage` and `TestAllCodesAreKnown` pass with all four new codes in `arch.AllCodes`, and that the T093 failing case is the end-to-end counterpart of the core `theme_invalid` case added in T092

**Checkpoint**: Teams can re-theme from outside the tool (SC-009).

---

## Phase 10: Polish & Cross-Cutting Concerns

- [X] T095 Write `cmd/build_perf_test.go` (`TestPerformanceBuild`, skipped under `-short`), reusing the generator from `internal/core/usecases/performance_test.go`, and move or export the generator into a shared `internal/testutil` helper if needed. A 1,000-element architecture builds all formats in under 10 s. A 200-container single-system view builds and is reported, and all 200 container labels appear as text in its SVG (spec edge case: very wide view). A single-description edit rebuilt through `usecases.Serve` with the same svg backend instance (so its cache is warm) is reflected in under 2 s (SC-006). If a bound fails, profile, record the finding in research.md R2, and fix it before continuing
- [X] T096 [P] Write `cmd/build_edge_test.go` cases (extending T052). `empty` gives exit `0`, no `diagrams/` files, and stdout containing `nothing to draw: the architecture declares no elements`. `single-element` gives only `landscape.*` and no system view. For `edge-cases`: the self-loop appears as a `--self` edge in D2, the cycle's three edges are drawn, the unsafe name maps to its escaped file name identically across two builds, and the system without containers has no view and no error. For `deployment-nested`, the D2 has five nested group levels
- [X] T097 [P] Write `docs/adr/0013-viewmodel-renderers.md` using `docs/adr/template.md`. It records the projection stage and backend port (R5), in-process d2 with ELK (R1), format dependencies, manifest ownership (R8), polling instead of fsnotify (R12), and the loopback preview server as distinct from the removed HTTP API (R13)
- [X] T098 [P] Re-check `.specify/memory/constitution.md` v1.4.0 (amended in T023a) against the final code. If the adapter set or file organization changed during implementation, apply a PATCH (v1.4.1) and re-sync `specs/009-constitution-compliance/contracts/structural-rules.yaml`. Otherwise record "no change" in the PR description
- [X] T099 [P] Rewrite the `loko build`, `loko serve` and `loko watch` sections of `docs/cli-reference.md` per [contracts/cli.md](./contracts/cli.md). `watch` becomes a one-line note that it folded into `serve`. Remove `--format pdf`, `toon` and `--clean`
- [X] T100 [P] Update `docs/quickstart.md` (build, serve, remove `loko watch`), `README.md` (the command table), `docs/configuration.md` (remove `serve_port` and the TOML build keys, and point to CLI flags), and `docs/language.md` § `view` (replace "rendered in a later stage" with the include/exclude/tags semantics from research R4 and a link to the output layout; add a table of derived view ids, `landscape`, `system-<name>`, `container-<name>` and `deployment-<env>`, and state that a declared `view` with the same label replaces the derived view and warns with `view_shadowed`)
- [X] T101 Run `go test -cover ./internal/core/...` and confirm coverage is above 80% (Constitution V). Add use-case tests for any file under 80%
- [X] T102 Run the full gate, `task lint && task test && task audit-constitution`, and then every scenario in [quickstart.md](./quickstart.md) by hand. Record any deviation as a new task rather than silently fixing it

---

## Dependencies & Execution Order

### Phase dependencies

```
Phase 1 Setup ──▶ Phase 2 Foundational ──▶ Phase 3 US1 ──▶ Phase 4 US3 ──┐ (MVP)
                                              │                           │
                                              ├──▶ Phase 5 US2            │
                                              ├──▶ Phase 8 US6            │
                                              └──▶ Phase 6 US4 ──┬──▶ Phase 7 US5
                                                                 └──▶ Phase 9 US7
                                                       all ──▶ Phase 10 Polish
```

- **US3** depends on US1 (it needs a pipeline to make deterministic).
- **US2** depends on US1 only. **US6** depends on US1 only (T037 builds the in-process renderer).
- **US4** depends on US1. **US5** and **US7** depend on US4's HTML backend. US5 and US7 are independent of each other.

### Within each story

Tests are written first and confirmed failing. Then entities → use cases → adapters → cmd → e2e.

### Key cross-task dependencies

- T004 depends on T003. The audit must be run after T037 (the deletion of `renderer.go`).
- **T023a (the constitution amendment) must be complete before Phase 3.** T040 depends on it.
- T012 depends on T011. T023 depends on T022 and T014.
- T033 and T034 depend on T023. T035 depends on T032–T034.
- T036 and T037 depend on T027 and T028 and on T014–T017. T039 depends on T020, T021 and T035. T041 depends on T039 and T040.
- T048 depends on T038 and T018a. T049 depends on T039. T045a depends on T039, and is extended by T073.
- T068 depends on T067 and T035. T072 depends on T070 and T071. T073 depends on T069 and T072.
- T080 depends on T039, T078 and T079. T081 depends on T080.
- T091 depends on T072 and T087a. T092 depends on T090 and T091.

---

## Parallel Examples

### Phase 2

```
T011 paths_test | T013 view.go | T014 model.go | T015 style.go | T016 artifact.go | T017 notice.go | T018 page.go | T018a commit_plan.go
```

### User Story 1: tests first

```
T024 resolve_views_test | T025 project_view_test | T026 project_deployment_test
T027 d2 emit_test       | T028 svg_backend_test  | T029 outputdir store_test | T030 build_artifacts_test
```

After those, the use-case track (T032 → T033/T034 → T035) and the adapter track (T036, T037, T038) can proceed in parallel, joining at T039.

### User Story 4: tests first

```
T061 prose_test | T062 project_pages_test | T063 markdown golden | T064 goldmark test | T065 html golden
```

### User Story 5

```
T075 poller_test | T076 devserver_test | T077 serve_site_test
```

---

## Implementation Strategy

### MVP (Phases 1–4)

1. Setup and Foundational.
2. US1: zero-config D2 and SVG.
3. US3: determinism, notices, manifest, collisions, and the macOS CI job.
4. **Stop and validate** with quickstart §1–4 and §6. This is shippable: `loko build --format d2,svg` produces committable diagrams.

### Incremental delivery

5. US2: declared views, plus the SC-007 projection goldens.
6. US6: the empty-`PATH` proof and the Docker image.
7. US4: markdown and the site. The default `--format` becomes all four.
8. US5: `serve`.
9. US7: themes.
10. Polish: performance, edge cases, ADR, constitution re-check, docs, coverage, the full gate.

### Notes

- Commit after each task or logical group. Every commit keeps `go build ./...` green.
- A golden file changing is a review event. Never run `-update` to make a failing test pass without
  inspecting the diff.
- Any d2 version bump regenerates every SVG golden in the same PR (research R1).

---

## Phase 11: Convergence

- [X] T103 Bring `docs/cli-reference.md` and `docs/quickstart.md` in line with the shipped command set (`validate`, `fmt`, `export`, `build`, `serve`, `mcp`, `version`, `completion`). In `docs/cli-reference.md`, remove the `loko init` and `loko new` sections, remove `--check-drift` and its sample output from `loko validate`, and fix `loko export`'s `--output` to `--out`. In `docs/quickstart.md`, replace the `loko init`/`loko new` walkthrough and the `loko.toml`/`src/` project tree with an HCL example (`arch.loko.hcl` with a `project` block) and the `dist/` layout from `contracts/output-layout.md`, and remove `loko init`, `loko new` and `loko api` from the command table. Per SC-010 (partial)
- [X] T104 Make an unreadable project root fatal when `loko serve` starts. In `internal/core/usecases/serve_site.go`, return the `BuildArtifacts` error from the *initial* build instead of passing it to `Preview.Fail`; compile diagnostics stay non-fatal, and operational errors on later rebuilds keep being shown in the browser. Have `cmd/serve.go` exit `1` on that error, and add a case to `internal/core/usecases/serve_site_test.go` (initial `Load` error → `Serve` returns the error, `Preview` untouched) and to `cmd/serve_test.go` (nonexistent `--project` → exit `1`). Per contracts/cli.md § serve (5), FR-038 (partial)
- [X] T105 Add a feature-014 entry to `CHANGELOG.md`: `build` and `serve` return, `watch` folds into `serve`, there is no `d2` runtime dependency, and a breaking note that the container image now uses `ENTRYPOINT ["/usr/local/bin/loko"]` (`docker run <image> build`, not `docker run <image> loko build`). Per plan: Dockerfile / T085 (unrequested)

---

## Phase 12: Convergence

- [X] T106 Enforce SC-006's wall-clock budgets in CI. Add a race-free step to the `determinism` job in `.github/workflows/ci.yml` (both `ubuntu-latest` and `macos-latest`) running `go test ./cmd/ -run 'TestPerformance(Build|WideView)|TestServe$' -count=1 -v`. The `-race` test job skips or relaxes these budgets by design (`raceEnabled` in `cmd/race_on_test.go`), so today no CI job asserts the 10 s build and 2 s edit bounds. Per SC-006, plan: Performance Goals (partial)
