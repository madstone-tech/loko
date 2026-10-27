# Implementation Plan: Language Polish

**Branch**: `018-language-polish` | **Date**: 2026-10-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/018-language-polish/spec.md`

## Summary

Close the language gaps the first modelling session found. Deployment nodes gain `title`,
`technology`, `description` and `tags`, and instances a `technology` override; the deployment
diagram labels nodes like element boxes, and the deployment page lists node details in a table
(R3–R5). Views gain `title` and `description`; the title becomes the page heading, the navigation
label and a heading drawn on the diagram canvas (R6, R7). All of them come from the 017 language
table, so the decoder, edit schema, `schema` tool, `fmt` and `apply_edit` follow (R1, R2).

`docs` paths resolve from the declaring file, with the project root as a fallback that comes with a
hint. Resolution moves from the use cases (which stat files today) into the source adapter (R9).

A new information severity carries the two advisory findings: the docs fallback hint and a
`wide_view` hint for views at least 3,000 px wide and 2.5 times wider than tall, measured from the
SVG already laid out (R8, R10).

## Technical Context

**Language/Version**: Go 1.27 (`go 1.27.0`, toolchain pinned)

**Primary Dependencies**: existing only: hcl/v2 (decoder), in-process d2 v0.7.1 (labels,
`near: top-center` title), `html/template` (site). No new module.

**Storage**: files only. Docs resolution reads the file system in `hclsource`, inside the root.

**Testing**: `go test ./...` with `-race` and the goroutine-leak gate. Decoder goldens for the new
attributes, D2 and HTML goldens for labels, title and the Nodes table, fixtures for the three new
codes, the parity test, an MCP end-to-end check of `wide_view` from `build`.

**Target Platform**: macOS, Linux and Windows CLI; MCP over stdio.

**Project Type**: CLI compiler with an MCP server.

**Performance Goals**: no extra layout run (the hint reads the SVG backend's bounding box); docs
resolution adds at most two `stat` calls per element with `docs`. Build and validate stay inside
the 014 and 015 budgets.

**Constraints**: byte-stable output (FR-018): models that use none of the new attributes, and
whose docs resolve from the root, produce identical exports, D2, SVG and HTML. Exit codes unchanged
(SC-005). `internal/mcp` imports no entity or render adapter.

**Scale/Scope**: 7 new attributes on 3 blocks, 1 resolution rule, 1 severity, 3 diagnostic codes,
1 `--severity` value, 1 site table, 1 canvas title.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

| Principle / rule | Status | How |
|---|---|---|
| **I. Clean Architecture** | ✅ | New fields, the severity and the codes are entities. Hints and docs diagnostics are use cases. Docs resolution, which touches the file system, moves into `hclsource`; this removes the `os` import from `usecases/validate_warnings.go`, an existing leak (R9). Diagram size is reported by the SVG backend on the artifact; the hint rule is a use case (R10). |
| **II. Interface-First** | ✅ | No new port. `vm.Artifact` gains two fields; `ArchitectureSource.Load` already has the root. |
| **III. Thin Handlers** | ✅ | No handler changes beyond the MCP `severity` enum and the CLI flag help. |
| **IV. Entity Validation** | ✅ | Empty-title rules live in the decoder and `authoring.checkEnum`; ignore rules in `arch.CheckIgnore`. |
| **V. Test-First** | ✅ | Each phase opens with failing tests: parity for the new attributes, goldens, fixtures for each code, the byte-identity check on unchanged fixtures. |
| **VI. Token Efficiency** | ✅ | New fields are `omitempty`; information diagnostics use the compact MCP form; `severity: "warning"` lets a client drop them. |
| **VII. Simplicity** | ✅ | No new module, port or configuration: the width thresholds are constants (spec assumption), the canvas title uses a D2 built-in. |
| Dependency Direction | ✅ | `authoring` keeps its own copies, held to the table by `TestLanguageParity`. |
| File-size budgets | ✅ | New use-case file `wide_view.go` ≤ 200 effective lines; `schemas.go` is pure data (exempt). `hclsource/decode_deployment.go` and `d2/emit.go` stay within the 400-line adapter budget (checked by the audit). |
| Quality gates | ✅ | lint, test (`go test -cover ./internal/core/...` > 80%), audit, race and leak gate. ADR-0017 records the severity, the docs rule and the hint. |

**Language change governance**: seven permanent v1.x attributes and one resolution rule, in
[contracts/language.md](contracts/language.md) and `docs/language.md`.

**Output change**: one deliberate change to existing output: models with root-relative docs paths
from subdirectories now print information lines (`docs_root_relative`). Exit codes and warning
counts are unchanged, and `--ignore docs_root_relative` or `--severity warning` restores the 1.1.0
text. Recorded in ADR-0017 and the changelog (Principle VI's rule on removing information does not
apply: nothing is removed).

**Contract change**: the 013 diagnostics schema's severity enum gains `info`.

**Post-design re-check**: ✅ Passes. The design adds one use-case file and fields on existing
types; it removes one layer leak and adds none.

## Project Structure

### Documentation (this feature)

```text
specs/018-language-polish/
├── spec.md
├── plan.md               # this file
├── research.md           # R1–R13
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── language.md
│   └── cli-mcp.md
├── checklists/
│   └── requirements.md
└── tasks.md              # /speckit-tasks
```

### Source Code (repository root)

```text
internal/core/entities/arch/
├── schemas.go                   # node, instance, view attributes; docs note (R1)
├── diagnostic.go                # SeverityInfo; 3 codes; IgnorableCodes (R8, R11)
├── diagnostic_ignore.go         # Without drops info too
├── language.go                  # ignoreWarningsNote lists IgnorableCodes
├── source_model.go              # ViewDecl Title/Description; ElementDecl DocsPath/DocsVia
├── source_model_deployment.go   # GroupDecl attributes; InstanceDecl.Technology
├── ir_deployment.go             # IR Group and Instance fields
└── ir_element.go                # IR View Title/Description

internal/core/entities/authoring/
├── edit_schema.go               # group, instance, view attributes
└── render_attrs.go              # empty-title rule for group and view

internal/core/entities/viewmodel/
├── model.go                     # Node.Tags
├── view.go                      # View.Description, TitleDeclared
└── artifact.go                  # Artifact.Width/Height

internal/core/usecases/
├── resolve_views.go             # declared title/description
├── project_deployment.go        # group details; instance technology
├── validate_warnings.go         # docs diagnostics from DocsVia; no os import
├── build_site.go                # prose by DocsPath
├── build_ir.go, build_ir_deployment.go  # IR fields
├── wide_view.go                 # NEW: WideViewHints (R10)
├── build_artifacts.go           # hints after render, then ignore lists
└── validate_filter.go           # FilterSeverity "warning"; summary order

internal/adapters/hclsource/
├── decode.go                    # resolve docs in Load and LoadOverlay
├── decode_deployment.go         # node and instance attributes
├── decode_project.go            # view attributes
├── docs_resolve.go              # NEW: resolve docs from the file, then the root (R9)
└── render_diagnostics.go        # Info label and count

internal/adapters/d2/
├── emit.go                      # group labels (R3); canvas title (R7)
└── svg_backend.go               # Artifact.Width/Height

internal/adapters/html/theme/
└── view.gohtml                  # description; deployment Nodes table (R4)
internal/adapters/html/render.go # node-path template helper

internal/mcp/tools/schemas.go    # severity enum gains "warning"
cmd/validate_cobra.go            # --severity help

testdata/projects/first-run-remodelled/   # docs paths rewritten file-relative (R12)
testdata/projects/deployment-details/     # NEW: nodes, instance technology, titled view
specs/013-hcl-compiler-core/contracts/diagnostics.schema.json  # severity "info"
docs/language.md, docs/cli-reference.md, docs/mcp-integration.md
docs/adr/0017-language-polish.md
CHANGELOG.md
```

**Structure Decision**: the existing single-module layout. Each change lands in the layer the
constitution assigns it; no new package.

## Phases (for `/speckit-tasks`)

1. **Foundation**: `SeverityInfo`, the three codes, `IgnorableCodes`, `Without`, `FilterSeverity`
   `warning`, renderer label. No behaviour change on existing fixtures.
2. **Story 1 (P1)**: node and instance attributes through table, decoder, IR, edit schema;
   projection; D2 group labels; Nodes table; the `deployment-details` fixture.
3. **Story 2 (P1)**: view attributes; declared title and description in the projection; page and
   navigation; canvas title in D2/SVG.
4. **Story 3 (P2)**: `hclsource` docs resolution; `warnDocs` from `DocsVia`; prose by `DocsPath`;
   first-run fixtures (R12).
5. **Story 4 (P3)**: `Artifact.Width/Height`; `WideViewHints`; ignore lists after render; MCP
   `build` check.
6. **Polish**: docs, ADR-0017, changelog, diagnostics schema, quickstart run, budget checks.

Stories 1 and 2 are independent after Phase 1. Stories 3 and 4 need Phase 1 (the severity).

## Complexity Tracking

No constitution violations to justify.
