# Data Model: ViewModel Renderers

**Feature**: `014-viewmodel-renderers` | **Plan**: [plan.md](plan.md) | **Research**: [research.md](research.md)

This feature adds new entity types in a new stdlib-only package,
`internal/core/entities/viewmodel/`. It does **not** change the IR (`internal/core/entities/arch`),
so feature 013's export goldens stay byte-identical.

**Implementation note (2026-09-30)**: the archcheck rules give the entity layer an empty
`allowedImports`, so one entity package may not import another. `viewmodel` therefore carries
addresses, element kinds and declared-view selections as plain strings, where the tables below say
`arch.Address`, `arch.ElementKind` or `*arch.View`; use cases convert at the boundary. For the same
reason, the path function the tables call `viewmodel.ElementPage` is named `ElementPath`, because
the page type already has that name.

Every collection below is an **ordered slice with a documented sort key**, never a map. That is the
same rule `arch.IR` follows (FR-040 of feature 013), and it is what makes FR-021 structural rather
than a matter of each backend remembering to sort.

---

## 1. View

A named slice of the architecture at one level of detail (spec: *View*).

| Field | Type | Notes |
|---|---|---|
| `ID` | `ViewID` (string) | Stable, file-name-safe. Derived from what the view depicts ([R3](research.md#r3-where-views-come-from)). Unique within a projection. |
| `Kind` | `ViewKind` | `landscape` \| `system` \| `container` \| `deployment` \| `declared` |
| `Title` | string | Human title: the project name for landscape, the element or environment name otherwise, the label for declared views. |
| `Subject` | `arch.Address` | The system, container, environment or `view.<label>`. Empty for landscape. |
| `Selection` | `*arch.View` | Set only for declared views. |

**Sort key**: kind order (landscape, system, container, deployment, declared), then `ID`.

**Rules**
- A view whose visible set is empty is not produced (FR-002). A **declared** empty view additionally
  yields a `view_empty` warning (FR-005).
- When a declared view's `ID` equals a derived one's, the derived one is dropped and a
  `view_shadowed` warning names it (FR-004).

---

## 2. ViewModel

The intermediate value one view projects to (spec: *View model*). It is the only thing a per-view
backend reads (FR-011).

| Field | Type | Notes |
|---|---|---|
| `View` | `View` | Identity (above). |
| `Nodes` | `[]Node` | Sorted by `Node.ID`. |
| `Edges` | `[]Edge` | Sorted by `(Source, Target, ID)`. |
| `Sources` | `[]string` | Project-relative source files contributing to this view; sorted, de-duplicated. Feeds the generated-file notice. |
| `DiagramPath` | string | Where this view's diagram lives, e.g. `diagrams/system-payments.svg`. Computed by `Paths`. |
| `PagePath` | string | This view's site page, e.g. `view/system-payments.html`. |

### 2.1 Node

| Field | Type | Notes |
|---|---|---|
| `ID` | `NodeID` (string) | A D2-safe key derived from the address (`container__api`) or the literal `outside`. Unique within the view. |
| `Address` | `arch.Address` | The element, group, instance or environment this node draws. Empty for `outside`. |
| `Role` | `NodeRole` | `element` \| `subject` (the boundary box of a system, container or environment view) \| `group` \| `instance` \| `outside` |
| `Kind` | `arch.ElementKind` | For elements and instances (the instance's `Of` kind). Empty otherwise. |
| `Label` | string | Name. |
| `Technology` | string | Optional. |
| `Description` | string | Optional. |
| `Parent` | `NodeID` | Nesting: the enclosing node's ID, empty at top level. The chain of parents is always acyclic. |
| `Style` | `Style` | Decided in the projection (FR-008). |
| `Link` | string | Site page path of the element the node represents, empty for nodes with no page. Link targets are only ever produced by `Paths`, which is what makes FR-027 checkable. |

### 2.2 Edge

| Field | Type | Notes |
|---|---|---|
| `ID` | string | `<source>--<target>`, plus `--self` for self-loops. |
| `Source`, `Target` | `NodeID` | Both exist in `Nodes`. `Source == Target` only for authored self-connections. |
| `Label` | string | The relationship description, or `"N relationships"` when several were merged. |
| `Technology` | string | Set only when every merged relationship agrees on it. |
| `Relationships` | `[]arch.Address` | The authored relationships this edge represents, sorted. |
| `Crossing` | bool | True when one end is the `outside` node (a *boundary connection*, FR-009). |
| `Style` | `EdgeStyle` | `Dashed` is true for crossing edges. |

**Invariants** (asserted by `viewmodel.(*ViewModel).Validate`, called in tests and on every
projection)
- Every `Edge.Source`, `Edge.Target` and `Node.Parent` names a node that exists in `Nodes`.
- At most one `outside` node exists, and it exists only if at least one edge crosses.
- Nesting follows the containment tree: a node's parent is its nearest visible ancestor
  ([R4](research.md#r4-what-each-view-contains-projection-rules)).

---

## 3. Style / EdgeStyle

| Field | Type | Notes |
|---|---|---|
| `Shape` | `Shape` | `person` \| `rectangle` \| `boundary` \| `oval` |
| `Fill`, `Stroke`, `FontColor` | string | `#rrggbb`, or empty for none. |
| `Dashed` | bool | True for externals, boundaries, and the `outside` node (FR-017). |
| `Classes` | `[]string` | `kind-<kind>`, `external`, `tag-<slug>`…, sorted (FR-018). |

`StyleFor(role, kind, tags)` is a pure function over a fixed table
([R6](research.md#r6-styling-by-convention-fr-008-fr-016019)). Its output depends only on its
inputs (FR-019).

---

## 4. ElementPage

One per logical element: the input to the prose outputs (FR-025).

| Field | Type | Notes |
|---|---|---|
| `Address`, `Kind`, `Name`, `Description`, `Technology`, `Owner` | — | Copied from `arch.Element`. |
| `Tags` | `[]string` | Sorted. |
| `Classes` | `[]string` | The same classes as the element's node style. |
| `Prose` | string | The markdown text of `docs`. Empty when absent. |
| `ProsePath` | string | The authored `docs` value, shown in the "prose missing" note. |
| `ProseMissing` | bool | `docs` was declared but the file does not exist (FR-026). |
| `Parent` | `*LinkRef` | Nil for top-level elements. |
| `Children` | `[]LinkRef` | Sorted by address. |
| `Uses` / `UsedBy` | `[]RelationRow` | Outgoing and incoming relationships, sorted by the other element's address, then the relationship's address. |
| `Diagram` | `ViewID` | The view this page embeds: its own subject view when one exists, otherwise its parent's view when one exists, otherwise `landscape`. It never names a view that was not produced. (A system with no containers, and every person and external, embeds `landscape`.) |
| `PagePath` | string | For example `element/container/api.html`. |

`LinkRef{Address, Name, Kind, PagePath}` and
`RelationRow{Relationship, Other LinkRef, Description, Technology}` carry only resolved links.

Relationship rows deliberately omit the other element's description, so that editing one element's
description changes its own page, its parent's children list, and the views that draw it, but not
every neighbour's page (FR-022, SC-003).

---

## 5. Projection

The complete intermediate value handed to every backend.

| Field | Type | Notes |
|---|---|---|
| `Project` | `ProjectHeader` | `Name` and `Description`. |
| `Views` | `[]ViewModel` | Sorted by the view sort key. |
| `Pages` | `[]ElementPage` | Sorted by address. |
| `Environments` | `[]LinkRef` | For site navigation. Links resolve to deployment view pages. |

The projection is built by `usecases.Project(ir, prose, provenance)`, a pure function (FR-010): the
same inputs always yield a deep-equal `Projection`. A golden test serialises it to JSON and compares
the result against `testdata/golden/<fixture>/projection.json`.

---

## 6. Format / Artifact / Manifest

| Type | Fields | Notes |
|---|---|---|
| `Format` | `d2` \| `svg` \| `md` \| `html` | `ParseFormats(list)` rejects unknown names, listing the supported ones (FR-015). `Requires()`: `html→svg`, `md→svg` ([R5](research.md#r5-the-backend-interface-and-format-dependencies)). |
| `Artifact` | `Path` (forward-slash, relative to the output root), `Format`, `Bytes` | The bytes already carry the generated-file notice (FR-020). |
| `Manifest` | `Paths []string` | Sorted. Serialised as `.loko-manifest` ([R8](research.md#r8-output-ownership-pruning-and-atomicity-fr-022024)). |

Artifacts are sorted by `Path` before `Commit`. A duplicate path, or two paths that are equal under
case folding, is an `output_path_collision` error ([R9](research.md#r9-file-names-safety-uniqueness-case-collisions-fr-006-fr-028)).

---

## 7. Paths (pure functions)

| Function | Example |
|---|---|
| `Segment(name)` | `payments/v2` → `payments~2fv2`, `orders_db` → `orders_db` (an injective escape; amended in 015) |
| `DiagramFile(id, ext)` | `diagrams/landscape.svg` |
| `ViewPage(id)` | `view/landscape.html` |
| `ElementPage(addr, ext)` | `element/container/api.html`, `md/element/container/api.md` |
| `Rel(from, to)` | A relative href between two output paths. The only way links are formed. |

---

## 8. Theme

| Type | Fields | Notes |
|---|---|---|
| `ThemeFile` | `Name` (base name), `Bytes`, `Origin` (a project-relative path, used in error messages) | Loaded by the `ThemeSource` port from `templates/`, sorted by `Name`. |

Built-in names are `layout.gohtml`, `index.gohtml`, `element.gohtml`, `view.gohtml`,
`partials.gohtml`, `style.css`, `custom.css` and `site.js`. An unknown name, a template parse error,
or a `{{define}}` of an unknown block becomes a `theme_invalid` error naming `Origin` (FR-035).

---

## 9. New diagnostic codes

Add these to `arch.AllCodes` and `contracts/diagnostics.schema.json`, in step with
`TestAllCodesAreKnown`. Each gets a fixture.

| Code | Severity | Raised when |
|---|---|---|
| `view_empty` | warning | A declared view selects nothing (FR-005). |
| `view_shadowed` | warning | A declared view replaces a derived view of the same ID (FR-004). |
| `output_path_collision` | error | Two artifacts' paths are equal, or equal under case folding (FR-028). |
| `theme_invalid` | error | A theme override is malformed or unknown (FR-035). |

`view_*` diagnostics carry the view's address. Their `Range` is `provenance.RangeOf(view.Subject)`
(the `view` block's range in the `SourceModel`), so the warning points at the declaration.

---

## 10. Watch state (adapter-internal, not an entity)

`serve` cycles between two states:

`building → ok(artifacts)` or `building → failed(diagnostics)`

Each transition broadcasts `reload` over SSE. `failed` serves the diagnostics page for every HTML
path. `ok` serves the new in-memory artifacts. There is no other state: the command never exits
because a compile failed (FR-031).
