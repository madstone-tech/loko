# Data Model: Diagram Look

No user-facing data changes: the language, the IR, the export and MCP outputs are untouched. The
entities below are the internal tables and the one value passed from the D2 emitter to the SVG
pass.

## Kind palette (core, `viewmodel/style.go`)

Already exists as `elementPalette`; values change (research R5).

| Field | Type | Notes |
|---|---|---|
| shape | `Shape` | Unchanged meaning: `ShapePerson` for people, the declared shape (`database`, …) when set, else `ShapeRectangle` |
| fill, stroke, font | colour string | Per kind, from R5 |
| dashed | bool | `true` for external |

**Change**: only the colours. `Style.Shape` keeps saying *what* the element is; the d2 adapter now
draws every element as a rounded rectangle card (`style.border-radius: 6`, research R13), uses
`Style.Shape` and the kind to pick the icon, and writes a declared shape as a `shape-<shape>` class
(research R12). No other backend reads `Style.Shape` today.

Boundary style (`RoleSubject`, `RoleGroup`): stroke `#737373`, font `#3D3D3D`, dashed. Outside
style unchanged (`#999999` stroke, `#666666` font).

**New**: `EdgePalette{Stroke: "#737373", Font: "#3D3D3D"}`, read by the emitter instead of d2's
default edge colours.

Validation: a unit test computes the WCAG contrast of the font colour against the fill, at each
role's opacity (R4), for every kind, and fails below 4.5 (FR-010). It also checks the edge font
over the white background at each edge role's opacity.

## Icon table (adapter, `d2/icons.go`)

| Key | Icon |
|---|---|
| shape `database` | `database` |
| shape `queue` | `list-ordered` |
| shape `topic` | `megaphone` |
| shape `function` | `square-function` |
| shape `bucket` | `archive` |
| kind `person` | `user` |
| kind `system` | `boxes` |
| kind `container` | `container` |
| kind `component` | `blocks` |
| kind `external` | `globe` |

Lookup: shape first (`ShapePerson` maps to `user`), then kind; an unknown kind falls back to
`boxes`. A test walks every shape in `arch.Shapes` and every element kind the view model defines
and fails if either has no entry (FR-017).

Each icon name maps to an embedded Lucide file, turned once at start-up into a `<symbol>`.

## Label line roles (adapter, `d2/labelstyle.go`)

```text
LineRole = title | kind | blank | description | edgeLabel | edgeTech | edgeTags
```

Each role has a style and an opacity. `<hash>` is the diagram hash d2 prefixes its font families
with (research R3). No role sets `font-weight`.

| Role | Style (inline on the `<tspan>`) | Opacity | Size in px at 16px |
|---|---|---|---|
| title | unchanged (d2's bold) | 1 | 16 |
| kind | `font-size:0.7em;letter-spacing:0.06em;fill-opacity:0.85` | 0.85 | 11.2 |
| blank | unchanged | 1 | 16 |
| description | `font-family:<hash>-font-regular;font-size:0.9em;fill-opacity:0.9` | 0.9 | 14.4 |
| edgeLabel | unchanged | 1 | 16 |
| edgeTech | `font-size:0.75em;letter-spacing:0.04em;fill-opacity:0.8` | 0.8 | 12 |
| edgeTags | `font-size:0.75em;fill-opacity:0.72` | 0.72 | 12 |

Validation: no role renders below 11px (US4).

## Label plan (emitter → SVG pass)

Produced by the emitter alongside the D2 source, for the SVG backend only.

| Field | Type | Notes |
|---|---|---|
| Nodes | map d2 path → `NodePlan` | Every element and instance node; boundaries and outside nodes are absent (no icon, labels untouched apart from colour) |
| Edges | map d2 edge key → `[]LineRole` | Every edge with a technology or tag line, including a single-line technology-only label. An edge whose lines are all `edgeLabel` is absent |

`NodePlan`: `Icon string`, `Colour string` (the node's text colour, for the icon), `Lines []LineRole`.

The d2 path is the quoted path the emitter already builds (`pathIndex`); d2 writes its base64 as
the node group's first class, which is how the pass finds the group. Edge keys follow d2's
`(<src> -> <dst>)[<index>]` form.

Invariants:
- `len(Lines)` equals the number of lines in the emitted label (one `<tspan>` each).
- The plan is a pure function of the view model, so the same input always yields the same plan
  (FR-015); it is part of the render-cache key (research R10).
