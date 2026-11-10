# Implementation Plan: 1.6.0 Theming API, MCP Edit Feedback, Listing, and Layout Defaults

**Branch**: `023-v160-theming-mcp` | **Date**: 2026-11-09 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/023-v160-theming-mcp/spec.md`

## Summary

Minor release 1.6.0, from feedback items 1, 3, 4, 6, 7, 8, 9, 11, 12, 13 and 18:

1. **Theming API.** A `--font-ui` token, a `brand` block, a project `footer`,
   `templates/assets/**` published under `assets/theme/`, a `fonts.css`
   override, and a theme line in the build report.
2. **MCP write results.** `apply_edit`, `move` and `fmt` return a diagnostic
   delta by default (`diagnostics: "all"` restores 1.5). `apply_edit` and `move`
   return a meaning diff by default (`diff: "exact"` restores 1.5).
3. **Readable defaults.** elk draws every view with no layout. Deployment views
   expand outside endpoints, one node per element (`outside = "collapsed"`
   restores 1.5).
4. **Diagnostics and listing.** `redundant_relationship` (information), and
   `query views`, `query elements [--kind]`, `query relationships`.
5. **Example.** `examples/custom-branding` is committed and uses every theming
   addition.

Three defaults change existing output. ADR-0022 (layout, outside endpoints),
ADR-0023 (MCP write results) and ADR-0024 (theming API, including the
`fonts.css` contract reversal) record them.

## Technical Context

**Language/Version**: Go 1.27+ (`go 1.27.0`, toolchain pinned)

**Primary Dependencies**: hcl/v2, d2 v0.7.1 (elk and dagre in process), cobra,
toon-go, goldmark. No new dependencies.

**Storage**: File system. Output gains `assets/theme/**` when a theme has assets.

**Testing**: `go test`, table-driven with `t.Parallel()`. Golden files under
`testdata/`. Gates: `task test`, `task lint`, `task audit-constitution`, core
coverage above 80%.

**Target Platform**: Single static binary, macOS (arm64) and Linux CI.

**Project Type**: CLI, MCP server and preview server in one binary.

**Performance Goals**: The 1,020-element perf project keeps the 10 s full-build
and 2 s edit-to-browser budgets with `layout = "dagre"`. elk on the same project
takes about 14.7 s (research R9), which ADR-0022 records. MCP writes keep the
2 s budget of `cmd/authoring_perf_test.go`. `fmt` adds one compile.

**Constraints**: Use-case files at most 200 effective lines, functions at most
60. Entity files at most 200. Adapter files at most 400. MCP handler functions
at most 30, CLI at most 50. `validate_warnings.go`, `query.go`,
`project_view.go` and `move.go` (MCP) are near their budgets, so new code goes
into new files (research R8, R12, R13).

**Scale/Scope**: 7 stories, 28 requirements. About 14 packages, 3 ADRs, 4
contracts, and 15 golden files: 1 `style.css`, 6 `projection.json` (3 of them
also change for outside endpoints), 4 html and 2 markdown deployment views.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|---|---|---|
| I. Clean Architecture | PASS | Layout and outside-endpoint resolution, the delta, the redundancy rule and the listing are use cases. Template, CSS, asset publishing and diff classification stay in adapters. |
| II. Interface-First | PASS | One port signature changes: `SourceEditor.Diff` gains a mode (one adapter, one double). `ThemeSource` keeps its signature. No new port. |
| III. Thin Handlers | PASS, with care | `move` (MCP) is at 23 of 30 lines. Shared argument parsing for `diagnostics` and `diff` goes into `helpers.go`. CLI `query` gains subcommands through the existing `querySubcommand` helper. |
| IV. Entity Validation | PASS | `outside` values validated in `validate_render_attrs.go` with `invalid_attribute_value`. Asset paths validated with `viewmodel.SafePath`. |
| V. Test-First | COMMITTED | Each task writes its failing test first. Spec acceptance scenarios map to tests in tasks.md. |
| VI. Token Efficiency | PASS, ADR required | The delta and meaning diff are defaults that remove information. ADR-0023 and the changelog record them, and `diagnostics: "all"` and `diff: "exact"` keep the 1.5 shapes. |
| VII. Simplicity | PASS | No flags beyond the spec. One `Rel` field instead of a new port. No compatibility shims. |
| File budgets | PASS | New files: `validate_redundant.go`, `query_list.go`, `edit_delta.go`, `hclsource/edit_diff_whitespace.go`. |
| ADRs | 3 | 0022, 0023, 0024 (research R14). |

## Project Structure

### Documentation (this feature)

```text
specs/023-v160-theming-mcp/
├── plan.md                              # This file
├── research.md                          # Phase 0 (R1–R14)
├── data-model.md                        # Phase 1
├── quickstart.md                        # Phase 1
├── contracts/
│   ├── theming.md                       # T-1..T-6, amends 014 and 021 theme contracts
│   ├── mcp-write-results.md             # W-1..W-5
│   ├── rendering-defaults.md            # R-1..R-4
│   └── query-and-diagnostics.md         # Q-1, Q-2
├── checklists/requirements.md
└── tasks.md                             # /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/entities/arch/
├── schemas.go, source_model*.go, ir_*.go, render_attrs.go   # footer, outside, OutsideModes
└── diagnostic.go, diagnostic_ignore.go                      # redundant_relationship
internal/core/entities/viewmodel/
├── view.go, page.go, model.go, artifact.go, theme_rules.go  # Layout, Outside, Footer, ThemeFile.Rel, ThemeReport, Validate
internal/core/entities/authoring/
├── plan.go, edit_schema.go                                  # FileChange.Whitespace, DiffMode, footer/outside
internal/core/usecases/
├── resolve_views.go                                         # elk default, Outside
├── project_lift.go, project_view.go, project_deployment.go  # expanded outside nodes
├── wide_view.go                                             # R-2 hint text
├── validate_redundant.go (new), validate_warnings.go        # Q-2
├── query_list.go (new), query.go                            # Q-1
├── edit_delta.go (new), apply_plan.go, apply_edits.go, format_plan.go, authoring_service.go  # W-1..W-3
├── build_ir.go, project.go, build_site.go, build_artifacts.go, build_in_project.go          # footer, theme validation, report
└── ports.go                                                 # Diff mode, RenderOptions.ThemeReport
internal/adapters/hclsource/
├── decode_project.go, decode_deployment.go                  # footer, outside
├── edit_diff.go, edit_diff_whitespace.go (new)              # W-4
internal/adapters/projectfs/theme.go                         # T-4 asset walk
internal/adapters/html/
├── theme/style.css, theme/layout.gohtml                     # T-1, T-2, T-3
├── theme.go, fonts.go, render.go                            # T-4, T-5, T-6, crossing lines
internal/adapters/markdown/backend.go                        # crossing lines
internal/adapters/d2/emit.go                                 # always write layout-engine
internal/mcp/tools/
├── apply_edit.go, move.go, fmt.go, query.go, build.go, schemas.go, helpers.go
cmd/
├── build.go                                                 # theme line
├── query.go, query_cobra.go                                 # listing
└── build_perf_test.go                                       # layout = "dagre"
docs/
├── adr/0022-*.md, 0023-*.md, 0024-*.md (new)
├── language.md, mcp-integration.md, cli-reference.md, guides/site-theming.md
specs/014-viewmodel-renderers/contracts/theme.md, specs/021-site-redesign/contracts/theme-blocks.md  # amended
examples/custom-branding/                                    # committed (FR-011)
CHANGELOG.md                                                 # ## [1.6.0]
```

**Structure Decision**: Every change extends an existing package at its
established place. Four new files keep near-budget files within budget. No new
package, no new port.

## Complexity Tracking

No constitution violations. The one port signature change (`SourceEditor.Diff`)
is the smallest way to give the MCP tools both diff modes.

## Post-Phase-1 Constitution Re-check

| Gate | Status | Notes |
|---|---|---|
| Layer imports | PASS | The design adds no adapter import to core. `ThemeReport` is a view-model type, filled by the html adapter through `RenderOptions`. |
| Budgets | PASS | New files hold the new logic. `move.Call` stays at most 30 lines through a shared helper. |
| Principle VI | PASS | Contract W-1 keeps both 1.5 shapes one argument away. ADR-0023 is listed. |
| Determinism | PASS | elk output was byte-identical across two builds (research R9). CI's Determinism jobs cover macOS and Ubuntu. |
