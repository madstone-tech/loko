# Research: Rendering Fidelity

**Feature**: `016-rendering-fidelity` | **Date**: 2026-10-09 | **Plan**: [plan.md](plan.md)

Each entry records the decision, why it was made, and what was rejected.

---

## R1. Where the new attributes live

**Decision**: Four optional attributes, all plain language attributes with no new block:

| Attribute | On | Values |
|---|---|---|
| `title` | every element | any non-empty string |
| `shape` | `container`, `external` | `database`, `queue`, `topic`, `function`, `bucket` |
| `kind` | `uses` (relationship) | `sync` (default), `async`, `trigger` |
| `tags` | `uses` (relationship) | list of strings, like element tags |
| `direction` | `view` | `down`, `right` |

The parser accepts them as optional attributes in the existing schemas. Core validation rejects
bad values and `shape` on the wrong kind, with new error codes (R7). The IR gains matching
`omitempty` fields: `Element.Title`, `Element.Shape`, `Relationship.Kind`, `Relationship.Tags`
and `View.Direction`. So an export of a project that uses none of them is byte-identical, and the
schema version stays 1.

**Alternatives rejected**:
- *Shape from tags* (`tags = ["database"]`): tags are free-form and already mean "selectable by
  views". Overloading them with rendering would make a tag rename change a drawing.
- *A `style` block*: more general, but it opens the door to colours and fonts per element, which
  this feature does not need and which would undermine a consistent C4 look.

---

## R2. Titles in diagrams, pages and markdown

**Decision (amended during implementation)**: A titled node's D2 label is a plain label: the title
on the first line, then `[Kind: Technology] · name`, then the description. The first design used a
D2 markdown label so the name could be smaller. A probe showed that D2 renders markdown labels as
embedded HTML (`<foreignObject>`): they display in browsers and on GitHub, but render as **empty
boxes** in PNG and PDF converters (`rsvg-convert`) and in many SVG tools. Portability won, and the
user chose the fallback. Untitled nodes keep today's plain label, byte for byte, so
SC-003 holds for projects without titles. The site and the markdown backend show the title as the
page heading, with the name in a `<small>` code element (HTML) or an inline code span (markdown)
beneath it.

**Rationale**: Clarification 5 asks for the name "in smaller text beneath" the title. A plain D2
label has one font size; a markdown label is the only way D2 offers to vary it, and it is part of
D2's library API, so no external process is involved.

**Risk and fallback**: markdown labels are measured differently from plain ones, which may change
how titled nodes are laid out. If the determinism or legibility tests show a problem, the fallback
is a plain label with the name on its own line in brackets (`read_api`). That gives the same
information without the size difference, and would be recorded as a deviation.

---

## R3. Shapes

**Decision**: `shape` maps to D2 shapes:

| `shape` | D2 shape | Why |
|---|---|---|
| `database` | `cylinder` | universal data-store symbol |
| `queue` | `queue` | D2's queue shape is a horizontal cylinder, the common queue symbol |
| `topic` | `hexagon` | distinct from a queue; often used for pub/sub hubs |
| `function` | `step` | a chevron reads as "runs and finishes", and stays distinct from containers |
| `bucket` | `stored_data` | the stored-data symbol for object storage |

Fill, stroke and font colours stay the kind's palette, so a shaped container is still recognisably
a container. In a view where a node is collapsed into its parent, the parent's shape is used.

**Alternatives rejected**: AWS service icons (licence and artwork churn; possibly later, as an
opt-in theme); more shapes (YAGNI: these five covered every container in both test
architectures).

---

## R4. Relationship kinds through edge aggregation

**Decision**:
- **Trigger.** Declared on the invoked element and targeting its trigger (clarification 2). During
  edge lifting, a trigger relationship's ends are swapped *before* edges are keyed, so it lands on
  the edge drawn from the trigger to the invoked element. Queries never see the swap: it happens
  in `project_lift.go`, which only rendering uses. A trigger with one end outside the view is
  swapped first and then becomes a crossing edge, so the crossing edge points from the outside
  marker to the invoked element, or from the trigger to the outside marker.
- **Async.** A merged edge is drawn dashed only when every relationship it carries is async. A mix
  of async and sync is drawn solid: the edge is at least partly synchronous.
- **Distinct from crossing edges.** Edges leaving the view are already drawn with `stroke-dash: 4`.
  Async edges use `stroke-dash: 8` (long dashes), so the two read differently. Crossing edges are
  unchanged, which keeps SC-003. The site's view pages gain a two-line legend explaining both.
- **Tags.** A single-relationship edge appends its tags to its label as a final line (`#write`). A
  merged edge shows the sorted union of its relationships' tags.

**Alternatives rejected**: restyling crossing edges, which changes every existing diagram;
colouring async edges, which is invisible in monochrome print.

---

## R5. Layout direction

**Decision**: D2 already takes a `direction:` line, currently hard-coded `right` in
`internal/adapters/d2/emit.go`. The view model gains `View.Direction`, set by the projection:

| View | Default |
|---|---|
| landscape | `right` (unchanged) |
| deployment-* | `right` (unchanged) |
| system-*, container-* | `down` (clarification 3) |
| declared views | `down`, unless `direction` is set |

Declared views default to `down` because they are usually container-level slices, the case the
wide layout hurt. The change is recorded in the changelog.

**Measurement (FR-006)**: the aspect ratio is read from the rendered SVG's `viewBox`. A test
renders the `serverless-reference` container view (14 containers, 39 relationships, and some
descriptions longer than the hand-drawn diagram's) and asserts that width ÷ height ≤ 2.

**Measured (implementation)**:

| Layout | Size | Ratio |
|---|---|---|
| left-right, dagre (old default) | 8321 × 2233 | 3.73 |
| top-down, dagre (new default) | 3640 × 2316 | 1.57 |
| hand-drawn reference, ELK | 2207 × 3599 | 0.61 |
| generated view on `main` (before 016), as built by users | 7407 × 1693 | 4.38 |
| generated view after the T033 remodel (titles, shapes, kinds) | 3184 × 2454 | 1.30 |

Before-and-after images: [`images/`](images/).

Width comes from how many nodes share a layer, not from label length: removing every description
only narrowed the view to 3300. An absolute width limit was considered and dropped (clarification
of 2026-10-09): even the hand-drawn reference exceeds 1,600. ELK remains a possible later
improvement.

**Alternatives rejected**: ELK (rejected in 014; revisit only if dagre top-down cannot meet the
ratio); choosing the direction automatically (unpredictable for authors, clarification 3).

---

## R6. Editing through MCP

**Decision**: The new attributes join the authoring legal-attribute table:

- element: `title` (string), plus `shape` (string) on containers and externals only;
- relationship: `kind` (string), `tags` (string list);
- view: `direction` (string).

`NewEdit` validates `kind`, `shape` and `direction` against their sets, so an assistant gets an
`invalid_edit` naming the allowed values before anything is planned. `describe` at `full` returns
them (ElementView, RelationshipView, ViewInfo). The property-test generator sets and clears titles
and shapes, and adds relationships with kinds.

---

## R7. Diagnostics

**Decision**: Three new error codes, each with a fixture in the rule-coverage test:

| Code | When |
|---|---|
| `invalid_attribute_value` | `shape`, `kind` or `direction` not in its set; the detail lists the allowed values |
| `shape_not_allowed` | `shape` on a system or person |
| `empty_title` | `title = ""` (a blank title would render a nameless box) |

These are added to `AllCodes` and the diagnostics schema.

---

## R8. Byte-identity guarantee (SC-003)

A project with none of the new attributes must produce:
- a byte-identical export: every new IR field is `omitempty`;
- byte-identical landscape and deployment diagrams, and byte-identical pages;
- system, container and declared-view diagrams that differ only in `direction: down`.

A test builds `two-systems`, strips the direction line from both the old golden and the new output,
and requires equality. The goldens are then regenerated deliberately (a review event), with the
diff limited to the direction line.
