# Data Model: MCP & Validation Ergonomics

Types this feature adds or changes, by layer. Field names are indicative; the contracts give the
wire shapes.

## Entities: `internal/core/entities/arch`

### LanguageBlock *(new, `schemas.go`)*

| Field | Type | Notes |
|---|---|---|
| `Name` | string | `project`, `person`, `system`, `container`, `component`, `external`, `deployment`, `node`, `instance`, `binding`, `uses`, `view`, `reconcile`, `moved`, `locals` |
| `Labels` | []string | e.g. `["name"]`; `binding` is `["kind"]`; `reconcile`, `moved`, `locals` none |
| `LabelValues` | []string | closed set for the label, when there is one: `binding` lists `arch.AllClaimKinds` (`terraform`, `cloudformation`) |
| `In` | []string | where it may appear: `file` for top level, else parent block names (`node` → `deployment`, `node`) |
| `Attrs` | []LanguageAttr | sorted by name |
| `Example` | string | the smallest valid block, formatted canonically |

### LanguageAttr *(new)*

| Field | Type | Notes |
|---|---|---|
| `Name` | string | |
| `Type` | string | `string`, `list(string)`, `ref`, `list(ref)`, `map`, `bool` |
| `Required` | bool | e.g. `uses.target`, `instance.of` |
| `Values` | []string | closed set, when there is one (`layout`, `direction`, `shape`, `kind`) |
| `Note` | string | one line, e.g. "relative to the project root" for `docs` |
| `Kinds` | []string | element kinds it applies to, when narrower than the block (`shape`: container, external) |

**Validation**: an `arch` test asserts the table is sorted, names are unique and every `In` names a
block. An `hclsource` test (adapters may import entities) parses every example and decodes it
without errors, and checks that the decoder accepts exactly the table's attributes per block.

### Language *(new var)*

`var Language []LanguageBlock`, plus `func LanguageBlockNamed(name) (LanguageBlock, bool)`.

### Diagnostic codes *(changed, `diagnostic.go`)*

- New: `CodePlannedInstanceBound`, `CodeUnknownWarningCode` (warnings), added to `AllCodes`.
- New: `var WarningCodes []string`, the warning subset of `AllCodes`, in the same order.
- New: `func CheckIgnore(codes []string) error`: refuses an error code or an unknown code, and the
  message lists `WarningCodes`.
- New: `func (d Diagnostics) Without(codes []string) (kept Diagnostics, ignored int)`: removes
  warnings with those codes. Errors are never removed.

### ProjectDecl *(changed)*

| Field | Type | Notes |
|---|---|---|
| `IgnoreWarnings` | []string | from `ignore_warnings`; order as authored |
| `IgnoreWarningsRange` | SourceRange | for `unknown_warning_code` and `invalid_attribute_value` |

### InstanceDecl / IR instance *(changed)*

| Field | Type | Notes |
|---|---|---|
| `Planned` | bool | from `planned`; exported as `planned` with `omitempty` |
| `PlannedRange` | SourceRange | decl only; for `planned_instance_bound` |

**Rules**: `planned && len(Claims) == 0` → no `unbound_instance`. `planned && len(Claims) > 0` →
`planned_instance_bound`.

## Entities: `internal/core/entities/authoring`

### TargetKind *(changed)*

New `TargetProject`, address `project`. Legal attributes: `description` (string), `layout`
(string, `dagre`|`elk`), `loko_version` (string), `ignore_warnings` (list), and `name` on `add` only.

### legalAttrs *(changed)*

`TargetInstance` gains `planned` (`ValueBool`). `TargetProject` added. Kept in step with
`arch.Language` by `TestLanguageParity` in `usecases`.

### AddressForm *(new)*

| Field | Type | Notes |
|---|---|---|
| `Target` | TargetKind | |
| `Pattern` | string | e.g. `deployment.<env>.node.<node>[.node.<node>…].instance.<name>` |
| `Example` | string | e.g. `deployment.prod.node.web.instance.api`; parsed by a test |

`var AddressForms []AddressForm`, one per edit target.

### Refusal reasons *(changed)*

New: `project_exists` (names the declaring file), `no_project` (carries the bootstrap request).

## Use cases: `internal/core/usecases`

### CompileRequest *(changed)*

`Ignore []string`: per-run codes, checked with `arch.CheckIgnore` before compiling. Merged with
`ProjectDecl.IgnoreWarnings`; suppression runs last.

### ValidateRequest *(new)* / ValidateResult *(changed)*

| Request field | Notes |
|---|---|
| `Ignore` | per-run codes |
| `Severity` | `""` (all) or `error` |
| `Summary` | group instead of list |
| `FullRanges` | MCP only; CLI always has full ranges |

`ValidateResult` gains `Ignored int` and `Summary []DiagnosticGroup` (when asked); its
`Diags` become `[]DiagnosticView`, compact unless `FullRanges`. With `Summary`, `Diags` is empty.
`Errors`/`Warnings` count after suppression, before the severity filter.

### DiagnosticGroup *(new)*

`{Code, Severity, Count, Addresses []string, More int}`. `Addresses` holds at most 20, sorted;
`More` counts the rest. `func Summarize(d) []DiagnosticGroup`; `func FilterSeverity(d, severity)`.

### DiagnosticView *(new)*

One type for both forms: compact `{Severity, Code, Detail, Address, At}`, with `At` as `file:line`
(`file` alone for a project-wide diagnostic); full `{Severity, Code, Summary, Detail, Address,
Range, Related}`, which encodes exactly as `arch.Diagnostic`. `func DiagnosticViews(d, full)`.

`EditResult`, `DescribeResult` and `QueryResult` keep their raw `Diags` (not encoded, for the CLI's
text renderer) and gain an encoded `Diagnostics []DiagnosticView`, set by `ShapeDiagnostics(full)`
just before encoding.

### LanguageResult *(new)*

`{Blocks []LanguageBlockView, Addresses []AddressFormView, Docs string}`. The views are use-case
DTOs, so `internal/mcp` does not import entities. `func DescribeLanguage(block string)
(*LanguageResult, error)`: an unknown block returns a refusal listing the names.

### BuildRequest *(changed)*

`Ignore []string`. Strict stays a caller decision on `BuildResult.ExitCode(strict)`, as today.

### ProjectBuild *(new, `build_in_project.go`)*

`func BuildInProject(ctx, deps BuildDeps, req ProjectBuildRequest) (*ProjectBuildResult, error)`,
used by the MCP `build` tool:

1. `deps.Store.Confine(ctx, req.Root, req.OutDir)`; a refusal stops here with `path_refused`.
2. `BuildArtifacts` with the confined directory.
3. `PickInlineView(artifacts, view, format)` when `req.Inline` is set: returns the artefact's bytes,
   or an `unknown_view` refusal listing the view names.

`ProjectBuildRequest` is `{Root, OutDir, Formats, Ignore, Strict, Inline, InlineFormat}`.
`ProjectBuildResult` is `{OK, OutDir, Formats, Written, Paths, Diags, Inline *InlineView, Refusal}`,
where `OK` is false when there are errors, or warnings under `Strict`. Under `Strict` the files are
still written, exactly as `loko build --strict` does.

### FormatPlan *(new)*

`func (s *AuthoringService) Format(ctx, base string, preview bool) (*EditResult, error)`: reuses
`EditResult` (diff, revision, written files) so preview and confirm look like `apply_edit`.

### DescribeResult *(changed)*

`Bootstrap *BootstrapHint`, with `{Syntax string, Request map[string]any}`; set only when no source
is found.

## Ports: `usecases/ports.go`

- `SourceFormatter` gains `Canonical(ctx, root) ([]authoring.FileContent, arch.Diagnostics, error)`.
- `ArtifactStore` gains `Confine(ctx, root, outDir string) (string, error)`: resolves `outDir`
  against an absolute `root`, follows symlinks, and refuses anything outside the root. Implemented in
  `adapters/outputdir`.

No new port.
