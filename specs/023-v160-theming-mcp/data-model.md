# Data Model: 1.6.0 Theming API, MCP Edit Feedback, Listing, and Layout Defaults

**Branch**: `023-v160-theming-mcp` | **Date**: 2026-11-09

New and changed types, by layer. Every new field is `omitempty` in JSON and TOON
unless stated, so output for projects that do not use a feature stays the same.

## Language (`internal/core/entities/arch`)

| Type | Field | Type | Validation | Notes |
|---|---|---|---|---|
| `ProjectDecl` | `Footer` | string | none | plain text, optional |
| `ProjectDecl` | `Outside`, `OutsideRange` | string, `SourceRange` | one of `OutsideModes` | optional, default `expanded` |
| `EnvironmentDecl` (`deployment` block) | `Outside`, `OutsideRange` | string, `SourceRange` | one of `OutsideModes` | wins over the project value |
| `Project` (IR) | `Footer`, `Outside` | string | | copied from the declaration |
| `Environment` (IR) | `Outside` | string | | copied from the declaration |
| (constant) | `OutsideModes` | `[]string{"expanded","collapsed"}` | | in `render_attrs.go` |
| (constant) | `CodeRedundantRelationship` | `"redundant_relationship"` | | information, ignorable |

`schemas.go` gains `footer` and `outside` on `project` (sorted) and `outside` on
`deployment`. `authoring/edit_schema.go` gains the same, for language parity.
Invalid `outside` values give `invalid_attribute_value`.

## View model (`internal/core/entities/viewmodel`)

| Type | Field | Type | Notes |
|---|---|---|---|
| `View` | `Layout` | string | always `"elk"` or `"dagre"` after projection (was `""` for dagre) |
| `View` | `Outside` | string | deployment views only: `"collapsed"` when chosen, empty when expanded |
| `ProjectHeader` | `Footer` | string | |
| `ThemeFile` | `Rel` | string | path below `templates/`, for example `assets/logo.svg`. Empty for top-level files |
| `ThemeFile` | `Link` | bool | the entry is a symbolic link under `templates/assets/`. Not followed. Validation refuses it with `theme_invalid` |
| `ThemeReport` (new) | `Dir`, `Files[]`, `Assets` | string, `[]ThemeFileReport`, int | nil when there is no `templates/` |
| `ThemeFileReport` (new) | `Name`, `Blocks` | string, `[]string` | `Blocks` sorted, set for `.gohtml` only |
| (function) | `OutsideNodeIDFor(addr)` | string | `"outside__" + NodeIDFor(addr)` |

`ViewModel.Validate` decides a crossing edge by the endpoint's `RoleOutside` and
allows several outside nodes. Each outside node needs at least one crossing
edge.

### Outside endpoint state

```
deployment view
  ├─ cmp.Or(deployment.outside, project.outside, "expanded")
  ├─ expanded  → one node per outside element (component → its container)
  └─ collapsed → one node "outside" (1.5)
other views    → one node "outside" (1.5)
```

## Use cases (`internal/core/usecases`)

| Type | Field | Type | Notes |
|---|---|---|---|
| `EditResult` | `Added`, `Removed` | `arch.Diagnostics` | raw, `json:"-"` |
| `EditResult` | `DiagnosticsAdded`, `DiagnosticsRemoved` | `[]DiagnosticView` | `diagnostics_added`, `diagnostics_removed` |
| `EditResult` | `DiagnosticsUnchanged` | `*DiagnosticCounts` | `diagnostics_unchanged`. nil when the total is 0 |
| `DiagnosticCounts` (new) | `Errors`, `Warnings`, `Info` | int | `errors`, `warnings`, `info` |
| `ApplyRequest` | `ExactDiff` | bool | |
| `MoveRequest` (new) | `From`, `To`, `BaseRevision`, `Preview`, `ExactDiff` | | replaces the positional `Move` arguments |
| `QueryRequest` | `ElementKind` | string | listing filter |
| `QueryResult` | `Views` | `[]ViewRow` | |
| `QueryResult` | `Relationships` | `[]RelationshipRow` | |
| `ElementHit` | `Title` | string | set only by the `elements` listing |
| `ViewRow` (new) | `Name`, `Kind`, `Title` | string | |
| `RelationshipRow` (new) | `Address`, `Source`, `Target`, `Kind`, `Description` | string | `Kind` is `sync` when empty |
| `RenderOptions` | `ThemeReport` | `*viewmodel.ThemeReport` | filled by the html backend |
| `BuildResult`, `ProjectBuildResult` | `Theme` | `*viewmodel.ThemeReport` | `theme` |

### Diagnostic delta

```
before = compile(source before)    // already computed by apply_edit and move
after  = compile(source after)
key(d) = code | severity | address | file | detail
added     = after  − before   (multiset by key, source order)
removed   = before − after
unchanged = counts by severity of the matched pairs
```

`fmt`: before is the after set (one compile, after formatting).

## Authoring (`internal/core/entities/authoring`)

| Type | Field | Type | Notes |
|---|---|---|---|
| `FileChange` | `Whitespace` | `[]int` | new-file line numbers of whitespace-only lines, `whitespace` |
| `DiffMode` (new) | `DiffMeaning`, `DiffExact` | | argument of `SourceEditor.Diff` |

## Port change (`internal/core/usecases/ports.go`)

`SourceEditor.Diff(plan Plan) []FileChange` becomes
`Diff(plan Plan, mode authoring.DiffMode) []FileChange`. The one adapter
(`hclsource.Editor`) and the one test double change with it. No other port
changes. `ThemeSource.LoadTheme` keeps its signature and returns assets as more
`ThemeFile` values.
