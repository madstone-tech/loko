# Data Model: Site Redesign

These are additions to the view-model projection (`internal/core/entities/viewmodel`), which is the
only input a backend reads (014 FR-011). Every addition is optional for existing consumers: the
markdown backend and the JSON goldens of untouched fields do not change.

## LinkRef (extended)

| Field | Type | Notes |
|---|---|---|
| Address, Name, Title, Kind, PagePath | (existing) | unchanged |
| Technology | string, omitempty | **new**: the referenced element's technology |
| Description | string, omitempty | **new**: the referenced element's description |

Filled for `Parent`, `Children`, `Ancestors`, `Siblings[].Ref` and `Neighbours[].Other`.
`RelationRow.Other` (Uses/UsedBy) is left without them, which keeps the markdown backend
byte-identical.

## ElementPage (extended)

| Field | Type | Rule |
|---|---|---|
| Ancestors | []LinkRef | From the top level down to the parent, excluding the element itself. Empty for top-level elements. Feeds the breadcrumb and "Part of" (the parent plus the system above it). |
| Siblings | []SiblingRow | The parent's other children, sorted by address. For a top-level element: the other top-level elements of the same kind. The current element is included with `Current: true` so the list can mark it. Empty when the element has no siblings. |
| Neighbours | []NeighbourRow | One per other element linked by at least one relationship in either direction, sorted by address. |
| Prev, Next | *LinkRef | The siblings before and after this one in `Siblings` order. Nil at the ends (no wrap). |
| ContextDiagram | ViewID | See research R8: the parent's own view, else the first view containing the element, else the landscape for top-level elements, else empty. Always a produced view. |
| ProseInherited | bool | True when the element's resolved `docs` equals its parent's. When set, the HTML page links to `Parent` and does not show `Prose`. `Prose` itself is kept for the markdown backend (FR-043). |
| Source | SourceRef | Where the element is declared. |
| Diagram, Uses, UsedBy, Children, Parent, Prose… | (existing) | Unchanged. `Children` rows now carry Technology and Description. |

The new fields live in an embedded `PageContext` struct (`page_context.go`), so templates read
them as `.Element.Siblings` and so on. `SiteSources` is the HTML page's generated-file notice: it is
`Sources` plus the sources of the parent, siblings, neighbours and `ContextDiagram` (research R9).
`Sources` itself is unchanged, so the markdown output does not change (FR-043).

## SiblingRow (new)

| Field | Type | Notes |
|---|---|---|
| Ref | LinkRef | The sibling, with technology and description |
| Current | bool | This page's own element |
| Calls | bool | The page's element has a relationship to this sibling |
| CalledBy | bool | This sibling has a relationship to the page's element |

## NeighbourRow (new)

The field names `Other`, `Description`, `Technology` and `Tags` match `RelationRow`, so a 014
`relation-table` override still executes against it (research R10).

| Field | Type | Notes |
|---|---|---|
| Other | LinkRef | The neighbour, with technology and description |
| Direction | string | `out`, `in` or `both` |
| Technology | string | Sorted, de-duplicated relationship technologies joined with `, ` |
| Description | string | The first non-empty relationship description in relationship-address order |
| Tags | []string | The sorted union of relationship tags |
| Relationships | int | How many relationships were rolled up (≥ 1) |
| Sibling | bool | The neighbour shares the element's parent (or is a top-level peer) |

## SourceRef (new)

| Field | Type | Notes |
|---|---|---|
| File | string | Project-relative path, e.g. `services/gateway.loko.hcl` |
| Line | int | 1-based line of the block's start; 0 when unknown |

## NavNode (new) and Projection.Nav

| Field | Type | Notes |
|---|---|---|
| Ref | LinkRef | The element (`Technology` and `Description` empty to keep the tree small) |
| Children | []NavNode | Its children, sorted by address |
| HasChildren | bool | True even when `Children` was pruned away, so the tree can draw `▸` |
| Current | bool | Set only in a pruned copy, on the page's element |
| Open | bool | Set only in a pruned copy, on the page's ancestors and the element itself |

`Projection.Nav []NavNode` is the full tree: people, then systems (containers → components), then
externals, each group sorted by address.

`NavFor(nav []NavNode, address string) []NavNode` returns a pruned copy:

- Every top-level node is kept.
- The children of each ancestor of `address` are kept.
- The children of `address` itself are kept.
- All other `Children` are dropped, but `HasChildren` is kept.
- An empty `address` (index and view pages) returns the top level only.

## Overview (new) and Projection.Overview

| Field | Type | Source |
|---|---|---|
| Systems, Externals | int | Elements of kind `system` / `external` |
| Containers, DataStores | int | Containers; data stores = containers whose shape is `database`, `queue` or `topic` (spec FR-016) |
| Relationships, Async | int | All relationships; those whose kind is `async` |
| Warnings, Infos | int | Open diagnostics after `ignore_warnings`, set by `BuildArtifacts` (research R12) |

## ViewModel (extended)

| Field | Type | Notes |
|---|---|---|
| ElementCount | int | Nodes with role element or instance (excluding boundaries and outside) |
| RelationshipCount | int | Sum of `len(Edge.Relationships)` over non-crossing edges |

### View mode

`View.Mode()` (method, not stored) gives the label and chip for a view (spec FR-018, FR-019). Rules
are applied in order:

| Rule | Mode |
|---|---|
| `Style == "sequence"` | `sequence` |
| `Kind == "declared"` | `declared` |
| `Kind == "deployment"` | `deployment` |
| otherwise (landscape, system, container) | `automatic` |

The page label is `<KIND> VIEW · <MODE>` in upper case, e.g. `SYSTEM VIEW · AUTOMATIC`.

## RenderOptions (extended, `usecases/ports.go`)

| Field | Type | Notes |
|---|---|---|
| Upstream | map[string][]byte | Artefacts of the formats this backend requires, keyed by artefact path. Nil for backends with no requirements. Read-only. |

## Invariants (tested)

1. `Siblings` contains the page's element exactly once, with `Current`, or is empty.
2. Every `Neighbours[].Other.Address` is unique, and the list never contains the page's element.
3. `Neighbours[i].Direction == "both"` exactly when relationships exist in both directions.
4. `Prev` and `Next` are consistent: `pages[x].Next == y` ⇔ `pages[y].Prev == x`.
5. `ContextDiagram` is empty, or names a view in `Projection.Views` whose nodes include the
   element, or is the landscape when the element is top level.
6. `ProseInherited` implies `Parent != nil`. `Prose` keeps the file's text, so the markdown output is
   unchanged (FR-043); the HTML page ignores it and links to the parent.
7. `NavFor` output for an element contains that element with `Current`, and its ancestors with
   `Open`.
8. All lists are deterministic: built from address-sorted inputs, with no map iteration in the
   output order.
