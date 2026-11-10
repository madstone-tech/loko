# Phase 0 Research: 1.6.0 Theming API, MCP Edit Feedback, Listing, and Layout Defaults

**Branch**: `023-v160-theming-mcp` | **Date**: 2026-11-09

Every decision cites code read in the repository at `9fc7d93` (v1.5.1). Timings
were measured on the maintainer's machine (arm64). Four research slices ran in
parallel: theming, MCP write results, layout, and diagnostics with listing and
outside endpoints.

Three spec statements were wrong about the code and were corrected in
`spec.md` before planning (R-11).

## R1: Interface font variable (FR-001, FR-002, item 7)

**Decision**: Add `--font-ui` to the light `:root` font group in
`internal/adapters/html/theme/style.css` (lines 44-47), with the value of
`--font-mono` (`"JetBrains Mono", ui-monospace, monospace`). Re-point the 15
interface rules that use `var(--font-mono)` to `var(--font-ui)`. Keep
`code, kbd, pre` (line 104) on `--font-mono`.

**Rationale**: 16 rules use `var(--font-mono)`. Only `code, kbd, pre` styles code
and HCL. The default value makes every computed font the same as 1.5 (FR-002).
`contrast_test.go` parses only colour tokens (regex at line 10), so a font token
needs no change there.

**Alternatives**: Re-declare each selector in a theme (the workaround the
feedback rejects). Make `--font-mono` the interface font and add `--font-code`
(breaks the documented `--font-mono` contract).

**Impact**: 1 golden (`internal/adapters/html/testdata/golden/assets/style.css`).

## R2: Brand block (FR-003, item 11)

**Decision**: Move the `<div class="brand">…</div>` line of `{{define "nav"}}`
(`theme/layout.gohtml:42-52`) into `{{define "brand"}}`, byte for byte, and call
`{{template "brand" .}}` from `nav`.

**Rationale**: The block registry is the parsed built-in template set
(`html/theme.go:87-102`). Defining `brand` in the built-in `layout.gohtml`
registers it, so a theme may override it with no registry edit. A 1.5 theme that
overrides `nav` replaces the whole block, so it keeps working. The 014 guards
(`theme_test.go:95-150`, `nav_test.go:83-87`) stay unchanged.

**Alternatives**: Put `brand` in `partials.gohtml` (021 pins block-to-file, and
`layout.gohtml` owns the shell). A project logo attribute (more language surface
for the same result).

**Impact**: 0 goldens, because the markup is byte-identical.

## R3: Project footer (FR-004, item 12)

**Decision**: Add `footer` (string, optional) to the `project` block, carried
through the declaration, IR, projection header and authoring edit schema. The
built-in `footer` block renders `{{with .Project.Footer}}…{{else}}<1.5 markup>{{end}}`.
html/template escapes it, so it is plain text.

**Rationale**: The full path is ten places: `arch/schemas.go:97-105` (sorted
attributes), `arch/source_model.go:83-100`, `hclsource/decode_project.go:41-59`,
`arch/ir_element.go:46-53`, `usecases/build_ir.go:28-32`,
`viewmodel/page.go:62-66` (`omitempty`), `usecases/project.go:29-31`,
`theme/layout.gohtml`, `authoring/edit_schema.go:63-64` (required by
`language_parity_test.go`), and `docs/language.md`. The MCP `schema` tool picks
it up from `arch.Language`.

**Alternatives**: A `copyright` name (narrower than what teams write). A footer
in `templates/` only (the 1.5 problem).

**Impact**: 0 goldens (no fixture sets a footer).

## R4: Theme assets (FR-006, FR-007, FR-009, item 8)

**Decision**: `projectfs.Theme.LoadTheme` walks `templates/` and also returns
every regular file under `templates/assets/`. `viewmodel.ThemeFile` gains
`Rel string` (the path below `templates/`). The html backend publishes each asset
at `assets/theme/<path below templates/assets/>`, bytes unchanged, with no
notice. Validation runs in the build use case before any backend renders:
`SafePath` on every asset path, symbolic links refused, case-insensitive
duplicates refused, each failure a `theme_invalid` naming the file.

**Rationale**: No port method changes, so no mock churn. The `assets/theme/`
namespace cannot collide with the generated assets (`assets/fonts.css`,
`icons.css`, `style.css`, `custom.css`, `site.js`, `search-index.js`,
`LICENSES.txt`, `brand/madstone-eye.svg`). Assets flow through `Store.Commit` like
every artifact, so the manifest owns them and prunes them (FR-009) with no
extra code. Early validation means an invalid asset fails even a `--formats d2`
build, which matches "applied completely or not at all".

**Alternatives**: Publish at `assets/<path>`, as the feedback proposed (collides
with generated files). A second `LoadThemeAssets` port method (breaks every
implementer and mock for no gain).

**Impact**: New tests in `projectfs`, `viewmodel/theme_rules_test.go`,
`cmd/build_theme_test.go`. 0 goldens.

## R5: Font stylesheet override (FR-008, item 9)

**Decision**: Add `fonts.css` to the overridable files (`html/theme.go:70-72`).
Split `brandArtifacts` (`html/fonts.go:32-56`) into the built-in `fonts.css` and
the brand files. When the theme has `fonts.css`, publish it (with the notice) and
skip the built-in one. Keep `assets/brand/madstone-eye.svg` and
`assets/LICENSES.txt`.

**Rationale**: The woff2 bytes live only inside the built-in `fonts.css`
(`assets_test.go:69-74`). No font file is written on its own, so skipping that
one artifact loads zero built-in fonts (SC-003). The `head` block keeps linking
`assets/fonts.css`.

**Contract change**: 014 `theme.md`, 021 `theme-blocks.md` and the guide say
`fonts.css` is not overridable. ADR-0024 records the reversal. `icons.css` stays
generated.

**Alternatives**: A project flag that turns the built-in fonts off (more
language surface).

## R6: Theme report (FR-005, item 13)

**Decision**: A `ThemeReport{Dir, Files[]{Name, Blocks[]}, Assets int}` in the
view model. The html backend fills it where `applyOverrides` already enumerates
each override's blocks (`html/theme.go:92`). `RenderOptions` carries a pointer
for it. `BuildResult` and `ProjectBuildResult` carry it (`theme`, `omitempty`).
`printBuildSummary` (`cmd/build.go:56-83`) prints it after the `built` line. The
format is in `contracts/theming.md` T-6.

**Rationale**: Only the html backend knows which blocks a file defines.
Re-parsing in the core would duplicate the template parser.

**Alternatives**: A second backend return value (changes the `Backend` port for
one format).

**Known limit**: The block list exists only when html renders. Without html,
the line lists the files and assets, with no blocks.

## R7: MCP diagnostic delta (FR-012, FR-013, FR-015, FR-016, item 3)

**Decision**:
- `apply_edit` and `move` already compile before the change and throw the
  result away (`apply_plan.go:14`). Keep it, and pass before and after sets to
  `finish`. No extra compile.
- `fmt` compiles nothing today and returns no diagnostics on success
  (`format_plan.go:13-34`). It gains one compile after formatting. Formatting is
  lexical (`hclsource/format.go:98-108`), so the delta is empty and only the
  counts are set.
- Matching key: `code | severity | address | file | detail`. Positions are left
  out (an inserted line shifts every diagnostic below it). Equal keys pair in
  source order.
- New fields `diagnostics_added`, `diagnostics_removed`, `diagnostics_unchanged`
  (counts named `errors`, `warnings`, `info`, as in `encoding/diagnostics.go:50-55`).
  New input `diagnostics: "changed" | "all"`, default `changed`. Refusals keep
  the full list.
- Delta code in a new `usecases/edit_delta.go` (about 70 lines). It reuses
  `DiagnosticViews`.

**Known misclassifications**: a rename shows the element's diagnostics as
removed plus added. A changed "Did you mean" suggestion on an error shows as
removed plus added, but errors refuse the change anyway, and refusals show the
full list.

**Alternatives**: Key without `detail` (two address-less `unresolved_reference`
diagnostics in one file collide). Position matching through a line map (needs
the adapter's diff algorithm in the core).

**Impact**: `internal/mcp/tools/schemas_test.go` (3 rows). New unit tests for the
delta, a size test modelled on `TestCompactDiagnosticsSize`
(`read_tools_test.go:192-206`). No MCP golden files exist.

## R8: Meaning diff (FR-014, item 3)

**Decision**: Classify in the adapter, in a new
`hclsource/edit_diff_whitespace.go`. After the Myers diff
(`edit_diff.go:64-123`), each run of removed and added lines is diffed again on
copies with all whitespace outside string literals removed. A pair that matches
only after that becomes a context line with the new text, and its new line
number goes into `FileChange.Whitespace`. A hunk with only such lines is dropped.
The `SourceEditor.Diff` port gains a mode (`authoring.DiffMeaning`,
`authoring.DiffExact`). New input `diff: "meaning" | "exact"`, default
`meaning`, on `apply_edit` and `move`. `fmt` always uses the exact diff.

**Rationale**: The realigned lines come only from `realign`
(`edit_tokens.go:282-303`) in `update`. Written files stay canonical, so only
the reported diff changes. The default drops the old text of realigned lines,
so it is a default that removes information: same ADR as R7, and `exact` keeps
the 1.5 shape (constitution VI).

**Alternatives**: Keep the exact diff and add line numbers beside it (the noise
stays). Two diffs per file (most tokens).

**Impact**: `hclsource/edit_diff_test.go`, `usecases/authoring_mocks_test.go`
(port signature). `Move` takes a `MoveRequest` struct (a second bool would make
the positional form unreadable). The `move` MCP handler is at 23 of 30 lines, so
argument parsing moves to a helper.

## R9: Default layout elk (FR-017 to FR-020, item 1)

**Decision**: Resolve the default in the core, as `direction` is.
`ResolveViews` sets `View.Layout` to `"elk"` or `"dagre"` for every view and
`layoutFor` (`resolve_views.go:128-135`) goes away. The d2 emitter always writes
`layout-engine` (`emit.go:32-35`), with a `cmp.Or(…, "elk")` fallback like
`direction` at line 31. Update the `"" means dagre` comments (`source_model.go:90`,
`schemas.go:101`, `svg_backend.go:212-217`, `render_attrs.go:50`).

**Rationale**: `direction` is resolved in `resolve_views.go`, with the adapter
default only as a safety net. Doing the same keeps one meaning for every value.
Sequence views already ignore layout (`emit_sequence.go:15`, test at
`emit_sequence_test.go:65`).

**Performance** (user decision: keep elk for every unset view):

| Fixture | dagre | elk |
|---|---|---|
| `build_perf_test.go` project, 1,020 elements | 3.24 / 3.74 / 3.42 s | 14.76 / 14.67 / 14.74 s |
| `first-run` | 0.16 / 0.16 / 0.17 s | 0.88 / 0.85 / 0.88 s |

elk breaks the 10 s full-build budget of feature 014 SC-006 on the large
project. `TestPerformanceBuild` therefore sets `layout = "dagre"` on its
generated project, and ADR-0022 records the cost and the opt-out. The
edit-to-browser 2 s budget was not measured with elk. [INFERENCE] a warm cache
re-lays out only the edited views, so it is likely within budget. Task T-perf
measures it.

**Determinism**: two elk builds of the same input were byte-identical on this
machine (`diff -r`, perf project and `first-run`). CI's Determinism jobs check
other OSes. d2 is `v0.7.1`.

**Wide-view hint**: new text in `contracts/rendering-defaults.md` R-2. 2 cases
in `wide_view_test.go` change.

**Impact**: `project_layout_test.go` `TestViewLayouts` (4 cases),
`emit_layout_test.go`, 6 `testdata/projects/*/expected/projection.json` (each
view gains `"layout"`). d2 goldens unchanged (they build view models directly).

**Alternatives**: Map `""` to elk in the adapter (one value, two meanings). elk
only above a size limit (rejected in clarification).

## R10: Outside endpoints (FR-025, FR-026, item 6)

**Decision**: `outside = "expanded" | "collapsed"` on `project` and `deployment`,
default `expanded`, in `arch.OutsideModes`, validated in
`validate_render_attrs.go`. `View.Outside` (`omitempty`) is set for deployment
views by `derivedViews`. `liftEdges` (`project_lift.go:36-84`) takes a function
that maps an invisible endpoint to an outside node id: `OutsideNodeID` when
collapsed, `outside__<NodeIDFor(addr)>` when expanded (a component endpoint
lifts to its container). Expanded nodes have the outside role, the element's
title as label, and no address. `ViewModel.Validate` (`model.go:116-118,150-152`)
decides crossing by role and no longer rejects more than one outside node. The
html `crossingLines` (`theme.go:222-238`) and markdown "Leaves this view"
(`backend.go:98-107`) find the direction by role. The collapsed text stays
byte-identical.

**Rationale**: Declared views are never deployment views
(`resolve_views.go:46-57,102-111`), so a `view` attribute would do nothing. The
d2 emitter handles any number of nodes. No address keeps the 021 svg-address
contract and leaves `Sources` unchanged.

**Alternatives**: Lift to the top-level ancestor (merges the names item 6 asks
for). Accept `outside` on `view` and ignore it (an attribute that does nothing).

**Impact**: 9 goldens: 3 `projection.json`, 4 html (`view/deployment-{prod,staging}.html`,
`deployment-details/view/deployment-{dev,prod}.html`) and 2 markdown
(`md/view/deployment-{prod,staging}.md`). Tests: `cmd/build_test.go:134`,
`project_deployment_test.go:87-96`, `viewmodel/model_test.go:40`, 9 `liftEdges`
calls in `project_lift*_test.go`.

## R11: Spec corrections found by research

| Spec text | Code fact | Correction made in `spec.md` |
|---|---|---|
| Delta matches "by code, address and position" | Every inserted line shifts positions | Match by code, severity, address, file and detail |
| Story 5 AC2: info "counts as a warning" under `--strict` | Only `missing_docs` escalates (`diagnostic.go:209-224`, ADR-0017) | It stays information and never changes an exit code |
| FR-026: settable "per declared view" | Declared views are never deployment views | Settable on `project` and `deployment` |

## R12: `redundant_relationship` (FR-021, FR-022, item 4)

**Decision**: A new `usecases/validate_redundant.go` with an exported
`ValidateRedundantRelationships`, called from `CompileArchitecture`
(`compile_architecture.go:73`) right after `ValidateWarnings`.
`validate_warnings.go` is at the 200-line budget, so it gets no new line.
Index container relationships by `(container, resolved target)` from
`res.Target`, then match each component's relationships through
`res.Parent[component]`, on exact addresses only. One diagnostic per container
relationship, with a `Related` entry per repeating component relationship.
Register the code in `diagnostic.go` (`AllCodes`), `diagnostic_ignore.go`
(`IgnorableCodes`, same order), the count in `TestAllCodesAreKnown` (37 → 38)
and `rule_coverage_test.go`. Text in `contracts/query-and-diagnostics.md` Q-2.

**Rationale**: Mirrors `warnSelfRelationships` (`validate_warnings.go:142-163`).
Suppression runs once at the end of `CompileArchitecture`
(`compile_architecture.go:80`), so `ignore_warnings` and `--ignore` work with no
extra code.

**Impact**: The pattern occurs in 4 test fixtures. None changes a golden file.

## R13: Listing queries (FR-023, FR-024, item 18)

**Decision**: A new `usecases/query_list.go` (`query.go` is at about 157 lines).
`QueryIR` gains `views`, `elements` and `relationships`. Views come from
`ResolveViews(ir, Provenance{})`, re-sorted by id. Elements reuse `ElementHit`
with a new `title` field (`omitempty`, so existing query kinds are
byte-identical). Relationships use a new `RelationshipRow`. The CLI adds three
`querySubcommand` entries in `cmd/query_cobra.go:41-56` and text tables in
`writeQueryText` (`cmd/query.go:92-123`). MCP adds the kinds to the `kind` enum
and an `element_kind` property (snake case, as `inline_format`).

**Rationale**: Subcommands of `query` leave `TestCommandSurface` unchanged and
reuse the shared encoder. `ResolveViews` is pure over the IR.

**Alternatives**: A separate element row type (a second elements field). Extend
`describe` (covers declared views only).

**Impact**: `cmd/query_test.go` `TestQueryMatchesMCP` (4 new cases),
`usecases/query_test.go` kind loops.

## R14: Decision records

| ADR | Title | Covers |
|---|---|---|
| 0022 | Readable defaults: elk layout and expanded outside endpoints | R9, R10 |
| 0023 | MCP write results: diagnostic delta and meaning diff | R7, R8 |
| 0024 | A supported theming API | R1 to R6, the `fonts.css` contract reversal |

Each follows `docs/adr/0016-mcp-ergonomics.md`: context from the feedback item,
the new default with measurements, the way back, and consequences. The
constitution needs no amendment: principle VI already states the rule.
