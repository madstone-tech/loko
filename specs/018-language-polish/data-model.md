# Data Model: Language Polish

Types this feature adds or changes, by layer. Field names are indicative; the contracts give the
language and wire shapes. Every new field is optional, and every new JSON field is `omitempty`, so
models that do not use them export as in 1.1.0 (FR-004, FR-018).

## Entities: `internal/core/entities/arch`

### Language table *(changed, `schemas.go`)*

| Block | New attributes |
|---|---|
| `node` | `title` (string), `technology` (string), `description` (string), `tags` (list(string)) |
| `instance` | `technology` (string): "overrides the element's technology in this deployment" |
| `view` | `title` (string), `description` (string) |
| element blocks | `docs` note becomes "relative to the declaring file; the project root is a fallback" |
| `project` | `ignore_warnings` note lists warning and information codes |

### GroupDecl / IR Group *(changed)*

| Field | Type | Notes |
|---|---|---|
| `Title` | string | refused when present and empty |
| `Technology` | string | |
| `Description` | string | |
| `Tags` | []string | element tag rules |
| `AttrRanges` | map[string]SourceRange | decl only, for diagnostics on the new attributes |

IR `Group` exports `title`, `technology`, `description`, `tags`, all `omitempty`.

### InstanceDecl / IR Instance *(changed)*

| Field | Type | Notes |
|---|---|---|
| `Technology` | string | exported as `technology`, `omitempty` |

### ViewDecl / IR View *(changed)*

| Field | Type | Notes |
|---|---|---|
| `Title` | string | refused when present and empty |
| `Description` | string | |

### ElementDecl *(changed)*

| Field | Type | Notes |
|---|---|---|
| `DocsPath` | string | resolved path relative to the root, forward slashes; empty when `docs` is unset |
| `DocsVia` | DocsVia | `file`, `root`, `both` or `missing`; set by the source adapter (R9) |
| `DocsCandidates` | [2]string | file-relative and root-relative locations checked, for diagnostic details |

`DocsVia` is a string type with four constants. IR `Element.Docs` carries `DocsPath` (the resolved
path), which equals the declared string for every model whose paths resolve from the root.

### Severity *(changed, `diagnostic.go`)*

- New: `SeverityInfo` (2), `String()` "info".
- `ExitCode` unchanged: information never changes it.
- `Without(codes)` removes warnings and information with those codes; errors are never removed.

### Diagnostic codes *(changed)*

- New: `CodeDocsRootRelative` (info), `CodeDocsAmbiguous` (warning), `CodeWideView` (info), added to
  `AllCodes` (34).
- `WarningCodes` is renamed `IgnorableCodes`: the warning and information subset of `AllCodes`, in
  the same order. `CheckIgnore` checks against it.

## Entities: `internal/core/entities/authoring`

Edit schema (`legalAttrs`) gains, as `ValueString` unless noted:

| Target | Attributes |
|---|---|
| `TargetGroup` | `title`, `technology`, `description`, `tags` (`ValueList`) |
| `TargetInstance` | `technology` |
| `TargetView` | `title`, `description` |

`checkEnum`'s empty-title rule applies to `TargetGroup` and `TargetView` as well as
`TargetElement`. `TestLanguageParity` covers all three targets already.

## Entities: `internal/core/entities/viewmodel`

| Type | Field | Notes |
|---|---|---|
| `Node` | `Tags []string` | `omitempty`; filled for group nodes, read by the site's Nodes table |
| `Node` (group) | `Title`, `Technology`, `Description` | existing fields, now filled for `RoleGroup` |
| `View` | `Description string` | `omitempty`; declared views only |
| `View` | `Title` | declared views: the declared title, else the name (unchanged when no title) |
| `View` | `TitleDeclared bool` | `json:"-"`; true when the title came from a `title` attribute, so the D2 emitter draws it on the canvas (R7) |
| `Artifact` | `Width`, `Height int` | rendered diagram size in px; zero for non-diagram files and for D2 source |

## Use cases: `internal/core/usecases`

- `resolve_views.go`: declared view title and description.
- `project_deployment.go`: group node title, technology, description and tags; instance
  technology override.
- `validate_warnings.go`: `warnDocs` reads `DocsVia` and emits the three docs diagnostics; no `os`
  import.
- `build_site.go`: reads prose by `DocsPath`.
- `wide_view.go` *(new)*: `WideViewHints(views, artifacts, project) arch.Diagnostics` with the
  3,000 px and 2.5 thresholds as named constants.
- `build_artifacts.go`: appends `WideViewHints`, then applies the run's and project's ignore lists.
- `validate_filter.go`: `FilterSeverity` accepts `warning`; `Summarize` orders information after
  warnings.

## Adapters

- `hclsource`: decode the new attributes; resolve `docs` (stat the two candidates inside the root).
- `d2`: group labels (R3); canvas title (R7); SVG backend fills `Artifact.Width/Height`.
- `html`: view description under the heading; deployment Nodes table (R4).
- `hclsource/render_diagnostics.go`: `Info:` label, `N info` in the summary when N > 0.

## State and lifecycle

No new state. Docs resolution happens on every load, so moving a file changes its outcome at the
next compile; there is no cache to invalidate.
