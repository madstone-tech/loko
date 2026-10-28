# Research: Dynamic Views

Decisions for [spec.md](spec.md). Each records what was chosen, why, and what was rejected.

## R1. Language shape

**Decision**: `step "<label>"` blocks inside `view`. A view with at least one step is a dynamic
view. The step block takes:

| Attribute | Type | Required | Note |
|---|---|---|---|
| `from` | reference to an element | yes | |
| `to` | reference to an element | yes | |
| `via` | reference to a relationship (`container.a.uses.b`) | no | checked against `from` and `to` |
| `description` | string | no | drawn on the edge; else the relationship's |
| `reply` | bool | no | matches the relationship reversed; drawn dashed |

`view` gains `style` (`edges` or `sequence`; `edges` by default) and the nested `step` block. All
of it is added to the language table (`arch/schemas.go`), so the decoder, the edit schema
(`TestLanguageParity`), the `schema` tool and `fmt` follow. `dynamic` stays an excluded keyword.

**Rationale**: clarification Q1 and Q3 (2026-10-27). A label gives each step a stable address,
`view.<view>.step.<label>`, as `uses` labels do for relationships.

**Alternatives**: a top-level `dynamic` or `flow` block (rejected in clarification); a `steps`
list attribute of objects (no labels, poor `fmt` and diff output).

## R2. Declaration rules

| Rule | Diagnostic |
|---|---|
| Two steps with the same label in one view | `duplicate_declaration` (the step address is in the symbol table) |
| A view with steps also sets `include`, `exclude` or `tags` | `dynamic_view_selection` (error), at the attribute |
| `style` on a view without steps | `dynamic_view_selection` (error), at `style` |
| `style` outside `edges`, `sequence` | `invalid_attribute_value` |
| `from`, `to` or `via` resolve to nothing, or to the wrong kind | `unresolved_reference`, `wrong_reference_kind` (existing resolver, with suggestions) |

A view with steps can never be empty, so FR-008's "dynamic view with no steps" is the
`style`-without-steps case; it is an error because the author asked for a flow and wrote none.

## R3. Matching a step to a relationship

`related(a, b)` holds when `a == b`, or one contains the other (any depth). A relationship `r`
matches a step `s` when, in `r`'s **drawn** direction (a `trigger` relationship is drawn from its
target, 016 R4):

- not a reply: `related(s.from, r.source)` and `related(s.to, r.target)`;
- a reply: `related(s.from, r.target)` and `related(s.to, r.source)`.

| Step | Outcome |
|---|---|
| `via` set and matches | the step uses it |
| `via` set and does not match | `step_mismatch` (error) naming the relationship's ends and the step's |
| no `via`, one or more match | the first match in address order is recorded; nothing reported |
| no `via`, none match | `step_unmatched` (warning) naming the step and its elements; still drawn |

A self step (`from == to`) matches a self relationship like any other.

**Rationale**: FR-006, FR-007 and clarifications Q2 (warning) and Q4 (explicit reply). Matching
in drawn direction keeps a trigger flow (queue → function) natural to write.

**Alternatives**: exact endpoints only (rejected: C4 flows are written at container level while
relationships are often declared between components).

## R4. New diagnostic codes

`step_mismatch` (error), `step_unmatched` (warning, ignorable), `dynamic_view_selection` (error).
`AllCodes` goes from 34 to 37; each gets a fixture case.

## R5. Compiled model

IR `View` gains `Style` and `Steps []Step`, both `omitempty`, so exports of models without steps
are unchanged. `Step` holds `Address`, `Name`, `Number` (1-based position), `From`, `To`, `Via`
(as written, resolved), `Relationship` (the one used: `via` or the match; empty when unmatched),
`Reply` and `Description`. Steps keep declaration order, not address order: order is their meaning.

## R6. Projection

A new `projectDynamic` (use case) builds the view model for a view with steps:

- **Participants**: each distinct `from`/`to`, in first-appearance order.
- **Boundaries** (edges style, clarification Q5): a participating container or component is placed
  inside its system, drawn as a boundary node (the container view's subject style) even when the
  system takes no step. A component whose container participates nests in it; otherwise it sits
  directly in the system boundary. A participating system is drawn as an element, not a boundary.
- **Edges**: one per step, ID `step__<number padded to 4>__<label>`, label `<n>. <text>` where text
  is the description, else the relationship's description, else empty; technology from the
  relationship. A reply is dashed (`EdgeStyle.Dashed`). Edges are not merged, unlike `liftEdges`.
- **Steps**: `ViewModel.Steps []StepRow` in order (number, label, from and to as link references,
  text, reply), for the page table and the sequence emitter, since `Edges` stay sorted by the view
  model invariant.
- **View**: `View.Style` carries `sequence` when set.

## R7. Drawing

- **Edges style**: the existing emitter; edges carry their numbered labels. D2 draws repeated edges
  between the same pair separately (verified in a probe with 0.7.1).
- **Sequence style**: a new `emitSequence` in the d2 adapter: a root object `loko-flow` with
  `shape: sequence_diagram`, one actor per participant in `Steps` order, then one edge per step in
  order, replies with `stroke-dash`. No boundaries. The canvas title (018) is kept. Probed with
  0.7.1: renders in-process; the sequence layout is independent of the engine, so `layout` and
  `direction` do not apply and are ignored for this style.
- The wide-view hint does not apply to the sequence style.

## R8. Pages and exports

The HTML view page and the markdown view page add a "Steps" table (number, from, to, description;
elements linked) for views with steps. Models without steps render byte-identically (FR-013).

## R9. MCP authoring

- New edit target `step`, address `view.<view>.step.<label>`. Legal attributes from the language
  table; `from` and `to` required on add.
- `add` appends by default; `before: "<label>"` inserts before that sibling step (refused as
  `not_found` when it does not exist). Implemented with hclwrite by removing the later step blocks
  and re-appending them after the new one, which keeps each block's tokens and comments.
- `update` and `remove` locate the step under its view, as bindings are located under instances.
- `view` updates accept `style`.
- The `move` tool already rewrites references in nested blocks (`renameIn` walks every block), so a
  renamed element or relationship is rewritten in `from`, `to` and `via`; a test pins it.

## R10. Reading tools

- `describe`: `ViewInfo` gains `style` and `steps` (number, label, from, to, via, relationship,
  reply, description).
- `query` gains kind `flows`: for an element, the dynamic views and steps it takes part in,
  including through containment, sorted by view then number.
- `loko query` mirrors it.

## R11. Fixtures

- `testdata/projects/dynamic-views/`: a small model with a request/response flow (edges style), the
  same flow as a sequence, a trigger step, a reply, an unmatched step and a component-level match.
- The quickstart declares the report's `daily_run` against a copy of `first-run` (SC-001).
- Decoder goldens for each declaration error; a 40-step fixture generated in a test (SC-006).

## R12. ADR and versioning

ADR-0018 records the block form, the matching rule, labels and the two styles. Release 1.3.0.
