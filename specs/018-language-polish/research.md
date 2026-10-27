# Research: Language Polish

Decisions for [plan.md](plan.md). Each one answers a question the spec left to planning, or a
finding from reading the code.

## R1. New attributes come from the language table

**Decision**: add `title`, `technology`, `description` and `tags` to the `node` block,
`technology` to `instance`, and `title` and `description` to `view` in `arch.Language`
(`internal/core/entities/arch/schemas.go`). The decoder's attribute lists derive from the table
(017), so `hclsource` only decodes the values. `authoring`'s edit schema gains the same
attributes for `TargetGroup`, `TargetInstance` and `TargetView`; `TestLanguageParity` already
compares those targets with the table, so a missing one fails the build.

**Rationale**: FR-016. 017 made the table the single definition; this feature is its first user.

**Alternatives considered**: a shared "descriptive attributes" helper for node, view and element.
Rejected: `elementAttrs` already exists for elements, and the three blocks take different subsets
(a view has no `technology`). Listing them in the table keeps it data.

## R2. Value rules

**Decision**: node `title` and view `title` follow the element rule: a string, refused when empty,
by the decoder (`unknown_attribute`-style compile error with the existing empty-title message) and
by `authoring.checkEnum`'s title check, extended to `TargetGroup` and `TargetView`. `technology`,
`description` are free strings. Node `tags` are a string list with the element tag rules
(`duplicate tag` handling reused).

**Rationale**: the spec's edge cases ask for the same rules as elements. Reusing the existing
checks means one message for one mistake.

## R3. What a deployment node shows

**Decision**: the projection fills `viewmodel.Node.Title` and `.Technology` for `RoleGroup` nodes
from the IR group. `d2.nodeLabel` stops returning the bare name for groups: a group with a title is
labelled `<title>\n[<technology>] · <name>`, without a title `<name>\n[<technology>]`, and with
neither, `<name>` as today. The description is never added to a group label (FR-002).
`viewmodel.Node` gains `Tags []string` (omitempty) so the site page can list them; the D2 emitter
ignores it.

**Rationale**: matches the element label shape from 016 (title, then the C4 tag line with the
name), so a reader sees the same convention on every diagram. Groups with no new attributes produce
the same D2 bytes as 1.1.0 (FR-018).

**Alternatives considered**: a `[Node: technology]` tag like elements' `[Container: …]`. Rejected:
"Node" adds a word to every group label and tells the reader nothing the nesting does not.

## R4. The "Nodes" table on a deployment page

**Decision**: the `view` template renders, for deployment views only, a `Nodes` section listing the
groups that carry any of title, technology, description or tags: path (nested names joined by
` / `), title, technology, description, tags. The data comes from the view model's group nodes, so
no new page model is needed. A deployment with no such group renders no section, so its page is
byte-identical to 1.1.0.

**Rationale**: answer to clarification Q1 (2026-10-27).

## R5. Instance technology

**Decision**: `InstanceDecl.Technology` and IR `Instance.Technology` (`json:"technology,omitempty"`).
When projecting a deployment view, an instance node's `Technology` is `cmp.Or(instance.Technology,
element.Technology)`. The element and its other instances are unaffected.

**Rationale**: FR-003. `omitempty` keeps 1.1.0 exports unchanged (FR-004).

## R6. View title and description

**Decision**: `ViewDecl.Title`, `.Description` and IR `View.Title`, `.Description` (omitempty).
`viewmodel.View` gains `Description` (omitempty); `resolve_views` sets the declared view's `Title`
to `cmp.Or(d.Title, d.Name)`. The page heading, the `<title>`, the navigation and the diagram's
`alt` text already read `View.Title`, so they follow. The `view` template shows `Description` under
the heading when set. Addresses, `ViewID`, file names and links keep using the name (FR-006).

**Rationale**: the site already treats `View.Title` as the display name; a declared view simply had
no way to set it.

## R7. The title on the canvas

**Decision**: when a view has a declared title, `d2.Emit` adds a root-level text object
`loko-view-title` with `near: top-center`, `shape: text` and a larger bold font, holding the title
(wrapped by the existing `wrapTitle`). D2 lays out `near` constants outside the graph for both
dagre and ELK, so the diagram itself does not move; only the canvas grows by the heading's height.
Views without a declared title emit nothing new (FR-007). The description is not emitted.

**Rationale**: answer to clarification Q2. A constant-near text object is how D2 documents diagram
titles; it needs no new module.

**Verification needed in implementation**: a golden for a titled view under both layout engines,
and the determinism job.

## R8. An information severity

**Decision**: `arch.SeverityInfo` (value 2, `String()` "info"). It is below warning:

- `ExitCode` is unchanged: it counts errors and warnings only, so information never changes an exit
  code, including under `--strict` (FR-012);
- `CountBySeverity(SeverityWarning)` and every warning total stay as they are; the text renderer
  prints an `Info:` label (dim cyan) and adds `N info` to its summary line only when N > 0, so
  output without information diagnostics is unchanged;
- `Without` drops information diagnostics by code as it drops warnings, and `CheckIgnore` accepts
  information codes (`WarningCodes` becomes `IgnorableCodes`; it is internal, so callers are
  renamed, and the `ignore_warnings` note in the language table lists both kinds);
- `FilterSeverity` accepts `error` (errors only, as today) and `warning` (errors and warnings, no
  information). The MCP `validate` enum gains `warning`;
- `--summary` groups information codes after warnings, under their own severity.

The diagnostics contract (`specs/013-hcl-compiler-core/contracts/diagnostics.schema.json`) gains
`info` in the severity enum. ADR-0017 records the new severity.

**Rationale**: both new hints (R9, R10) are advice, not findings; making them warnings would turn
`--strict` red on every 1.1.0 model that uses root-relative docs paths, which FR-009 forbids.

**Alternatives considered**: warnings that `ignore_warnings` lists by default. Rejected: a default
suppression list is hidden state, and the 017 constitution amendment favours explicit output.

## R9. Docs path resolution moves into the source adapter

**Finding**: `usecases/validate_warnings.go` imports `os` and stats docs files itself
(`docsExist`), which is a layer leak (Principle I), and `build_site.go` reads prose by the declared
string. Resolving from the declaring file needs both the file and the project root, which only the
source adapter has together.

**Decision**: `hclsource` resolves each element's `docs` while loading, since it already knows the
declaring file and the root. It records the outcome on `ElementDecl`:

| `DocsVia` | Meaning | Diagnostic (emitted in `usecases.warnDocs`) |
|---|---|---|
| `file` | found next to the declaring file | none |
| `root` | not there, found from the project root (1.1.0 rule) | `docs_root_relative` (info), naming the file-relative path to write |
| `both` | found at both, different files | `docs_ambiguous` (warning), naming both; the file-relative one is used |
| `missing` | found at neither | `docs_not_found` (warning), naming the file-relative location, root-relative |

`ElementDecl.DocsPath` holds the resolved path, relative to the root, forward slashes. A path that
leaves the root resolves as `missing`, as today. When file and root give the same file (an element
declared at the root), the outcome is `file`.

The IR's `docs` is the resolved root-relative path, so the export keeps one meaning for consumers,
and every 1.1.0 model, whose paths all resolve from the root, exports byte-identically (FR-009,
FR-018). `build_site` reads prose by the resolved path. `usecases` stops importing `os`.

**Rationale**: answer to clarification Q1 (2026-10-26), placed in the layer that owns the file
system.

**Alternatives considered**: a `ProseReader.Exists` port called from `warnDocs`. Rejected: compile
would need a new dependency threaded through `CompileArchitecture`, the overlay compile and every
caller, to answer a question the loader can answer while it has the file open.

## R10. The wide-view hint

**Measurement** (all example and fixture views, rendered to SVG with 1.1.0):

| View | Size (px) | Ratio | Sets direction/layout? |
|---|---|---|---|
| microservices `landscape` | 7052 × 912 | 7.7 | no |
| microservices `order-flow` | 3823 × 513 | 7.5 | view `direction` |
| 3layer-app `request-path` | 5338 × 736 | 7.3 | view `direction`, project `layout` |
| two-systems `landscape` | 2413 × 540 | 4.5 | no |
| serverless-reference `deployment-prod`, `-dev` | 6931 × 1703 | 4.1 | no |
| microservices `deployment-prod` | 7454 × 1914 | 3.9 | no |
| simple-project `request-path` | 2793 × 741 | 3.8 | view `direction` |
| serverless `async-processing`, `order-intake` | ≈3000–3600 wide | 3.1–3.3 | view `direction` |
| microservices `system-order_service` | 2868 × 950 | 3.0 | no |
| first-run `system-pricing` (the report's view) | 3875 × 1441 | 2.7 | no |

With `layout = "elk"` on the project, the report's view becomes 3046 × 1695 (ratio 1.8).

**Finding**: the ratio alone cannot single out the report's view; four views are wider in shape but
narrow enough to read. What made the report's view unreadable is its width at page size: shown 1,200
pixels wide, its labels shrink by more than three times.

**Decision**: a view gets a `wide_view` information diagnostic when its rendered width is at least
**3,000 px** and its width-to-height ratio is at least **2.5**, and neither the view sets `direction`
or `layout` nor the project sets `layout` (clarification Q3). The detail names the size and ratio,
and the settings to try:

- a declared view: `direction` and `layout = "elk"` on the view;
- a generated view (landscape, system, container, deployment), which has no block to set them on:
  `layout = "elk"` on the project, or declaring a view of the same elements.

In the repository today this flags the report's view, its remodelled copy, and four generated
views: microservices `landscape` and `deployment-prod`, and serverless-reference
`deployment-prod` and `deployment-dev`. All are wider than 6,900 px except the report's view. The
spec's SC-004 is reworded to match this measurement.

**How it is measured**: `vm.Artifact` gains `Width`, `Height int` (zero for non-diagram files),
set by the SVG backend from the laid-out diagram's bounding box. `BuildArtifacts` derives the hints
after rendering, then applies the same ignore list as compilation (the request's and the project's
`ignore_warnings`). A build without SVG (for example `--format d2`) lays nothing out and gives no
hint. `validate` renders nothing and gives none; `serve` shows errors only (017) and gives none; the
MCP `build` tool returns them with its other diagnostics.

**Rationale**: FR-013 to FR-015, with no extra layout run (spec assumption).

## R11. New diagnostic codes

`docs_root_relative` (info), `docs_ambiguous` (warning), `wide_view` (info). `AllCodes` goes from
31 to 34; each gets a fixture, as the rule-coverage test requires. Detail texts:

- `docs_root_relative`: `container.api references "architecture/docs/api.md", found from the project
  root. Paths resolve from the declaring file first; write "docs/api.md" instead.`
- `docs_ambiguous`: `… "docs/api.md" exists next to architecture/api.loko.hcl
  (architecture/docs/api.md) and at the project root (docs/api.md); the first is used.`
- `docs_not_found`: `… "docs/api.md" was looked for at architecture/docs/api.md (from the declaring
  file) and docs/api.md (from the project root).`
- `wide_view`: `view.overview renders 3875 × 1441 px (ratio 2.7). Try direction = "right" …`

## R12. Existing fixtures and the hint count

The `first-run` fixtures write root-relative paths from files in `architecture/`, so each of their
seven docs paths now gives a `docs_root_relative` hint. That is the migration path working as
designed; their warning counts (and SC-002 of 017: ≤ 10 warnings for the remodelled one) are
unchanged, since information is not a warning. The `first-run-remodelled` fixture is updated to
the file-relative form so it stays the "clean" reference, and `first-run` keeps the old form so the
hint stays covered.

## R13. ADR and versioning

ADR-0017 records three decisions: the information severity (R8), docs resolution from the declaring
file with a root fallback (R9), and the wide-view hint rule (R10). The new attributes are additive
and documented in `docs/language.md`. This ships as 1.2.0.
