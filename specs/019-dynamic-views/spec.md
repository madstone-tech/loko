# Feature Specification: Dynamic Views

**Feature Branch**: `019-dynamic-views`

**Created**: 2026-10-27

**Status**: Draft

**Input**: Item #5 of the first modelling session report
([017 first-run report](../017-mcp-validation-ergonomics/first-run-report.md)): there is no way to
express an ordered interaction (a request flow, a pipeline run, a review loop). `sequence`,
`dynamic`, `scenario` and `flow` blocks are all refused, and relationships have no order. The C4
model's dynamic diagram covers this, and for a pipeline or an event-driven system it is often the
most useful diagram. The report proposes ordered steps that refer to declared elements and
relationships, so the compiler checks every step, drawn as numbered edges in a normal layout or as
a sequence diagram.

## Clarifications

### Session 2026-10-27

- Q: How is a dynamic view declared? → A: `step` blocks inside the existing `view` block; a view
  with steps is a dynamic view. No new top-level block, `dynamic` stays excluded, and a dynamic
  view does not also take `include`, `exclude` or `tags`.
- Q: What happens to a step whose elements have no declared relationship? → A: A warning that
  names the step; the step is still drawn, so a flow can be sketched before every relationship
  exists. `--strict` and `ignore_warnings` apply as for any warning.
- Q: Does each step carry a name label? → A: Yes, a required label, as `uses` blocks have:
  `step "fetch" { … }`. A step's address is `view.<view>.step.<label>` and survives inserting
  steps; its number is still its position.
- Q: How is a reply (a step against a relationship's direction) recognised? → A: The author marks
  it with `reply = true`; it must match a relationship from its `to` to its `from`, and is drawn
  dashed. A step without it must match in its own direction.
- Q: Does the numbered-edge style draw boundaries around participants? → A: Yes, the system only:
  a participating container or component is drawn inside its system's boundary even when the
  system takes no step, as in C4's dynamic diagram. Other ancestors are not drawn.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Describe an ordered flow (Priority: P1)

An architect models a scheduled pipeline: the orchestrator triggers the collector, the collector
searches and fetches from a vendor API, then writes to the raw store. They declare a dynamic view
with three ordered steps, each naming its two elements and optionally the relationship it uses, and
a short description of what happens at that step. The build draws the view with the steps numbered
1, 2, 3 on their edges, in that order, and the view's page lists the steps as a numbered table.

**Why this priority**: an ordered flow cannot be expressed at all today. This is the feature.

**Independent Test**: declare a dynamic view of three steps over a fixture's existing
relationships, build it, and check the diagram's numbered edges and the page's step table against
the declaration.

**Acceptance Scenarios**:

1. **Given** a dynamic view with three steps, **When** the project is built, **Then** its diagram
   draws one edge per step, labelled with the step number and its description, and its page lists
   the steps in order with both elements and the description.
2. **Given** the same two elements in two different steps (a request and a later call), **When**
   the project is built, **Then** both steps are drawn, each with its own number.
3. **Given** a step that matches a relationship, **When** the project is built, **Then** the edge
   is labelled with the step's description (the relationship's description when the step gives
   none) and shows the relationship's technology.
4. **Given** a dynamic view, **When** the model is exported, **Then** the export lists its steps in
   order with resolved addresses.
5. **Given** a model with no dynamic view, **When** it is built, **Then** every output is
   byte-identical to 1.2.0.

---

### User Story 2 - Steps the compiler checks (Priority: P1)

Because steps refer to the model, the compiler catches a flow that no longer matches it: an element
that was renamed or removed, a relationship that does not connect the step's elements, or a step
between two elements that have no relationship at all.

**Why this priority**: an unchecked flow drifts from the model and becomes a drawing. Checking is
what the report asked for, and what separates this from a hand-drawn sequence.

**Independent Test**: break one step at a time in a fixture (an unknown element, a relationship
between other elements, no relationship at all) and check each produces its own diagnostic at the
step's line.

**Acceptance Scenarios**:

1. **Given** a step naming an element that is not declared, **When** the model is validated,
   **Then** the reference is reported as unresolved, at the step's line, with a suggestion.
2. **Given** a step that names a relationship which does not connect its two elements, **When** the
   model is validated, **Then** an error names the relationship, its actual ends and the step's.
3. **Given** a step between two elements with no declared relationship between them, or between
   their ancestors or descendants, **When** the model is validated, **Then** a warning names the
   step and its two elements, and the step is still drawn; `--strict` fails on it and
   `ignore_warnings` can silence it, as for any warning.
4. **Given** a step between two containers whose components are related, **When** the model is
   validated, **Then** the step matches that relationship and nothing is reported.
5. **Given** an element or relationship used by a step, **When** it is renamed with the `move`
   tool, **Then** every step naming it in `from`, `to` or `via` is rewritten, as every other
   reference is.

---

### User Story 3 - A sequence diagram of the same flow (Priority: P2)

For a request/response exchange, the architect prefers a sequence diagram: one lifeline per
element, steps as arrows in order from top to bottom. They set the view's style to sequence and
build again; the same steps are drawn as a sequence diagram with the same numbers and labels.

**Why this priority**: numbered edges answer most needs; a sequence diagram reads better for
request/response exchanges with many steps between few elements.

**Independent Test**: set the sequence style on the US1 view and build: the diagram has one
lifeline per element in first-appearance order and one arrow per step, numbered as before; the page
and the export are unchanged apart from the style.

**Acceptance Scenarios**:

1. **Given** a dynamic view with the sequence style, **When** it is built, **Then** the diagram
   shows one lifeline per participating element, in the order each first appears, and one arrow
   per step, top to bottom, numbered and labelled as in the numbered-edge style.
2. **Given** the same view without the style, **When** it is built, **Then** it is drawn with
   numbered edges.
3. **Given** a sequence-style view, **When** its page is built, **Then** the step table is the same
   as for the numbered-edge style.

---

### User Story 4 - Assistants author flows through MCP (Priority: P3)

An assistant using `loko mcp` learns the dynamic view's shape from the `schema` tool, creates a
view and appends, inserts (before a named step), changes and removes steps with `apply_edit`, with the same preview, revision
check and refusals as every other edit, and finds the flow through `describe` and `query`.

**Why this priority**: the report's sessions author through MCP. Without edit support an assistant
would have to hand-write HCL, which 015 set out to avoid.

**Independent Test**: through MCP only, create a dynamic view of three steps on a fixture, insert a
step between the first two, remove the last, and build: the diagram shows the expected numbering.

**Acceptance Scenarios**:

1. **Given** the `schema` tool, **When** an assistant asks about the dynamic view, **Then** it lists
   the step's attributes, which are required, and an example.
2. **Given** a dynamic view, **When** an assistant appends a step, inserts one before a named step,
   edits one and removes one, **Then** each edit previews as a diff, applies atomically and keeps the
   other steps in order.
3. **Given** an edit that would break a step (an element that does not exist), **When** it is
   previewed, **Then** it is refused with the compile diagnostic, as for every other edit.

---

### Edge Cases

- A view without steps is an ordinary declared view. Setting `style` on it is an error naming the
  attribute, since the author asked for a flow and wrote none.
- A step whose two elements are the same is allowed when the element has a self relationship, as
  in the logical model.
- A step may involve an element nested in another participant (a component and its container):
  both are drawn, the component inside its container in the numbered-edge style, as separate
  lifelines in the sequence style.
- A component whose container takes no step is drawn directly inside its system's boundary; the
  sequence style draws no boundaries.
- An element used by several steps is drawn once; only edges repeat.
- A dynamic view is a `view` block, so its name is unique among views as today, and a name equal to
  a generated view's replaces it with the existing shadowing warning.
- A view with steps that also sets `include`, `exclude` or `tags` is an error naming the conflict.
- A step's description is plain text, like a relationship's.
- Two steps with the same label in one view are a duplicate declaration; the same label in two
  views is allowed.
- A long flow (40 steps) still renders, numbered 1 to 40.
- External elements and people take part like any other element.
- A reply from the callee is a step marked `reply = true`. It must match a relationship in the
  opposite direction (from the step's `to` to its `from`) and is drawn dashed in both styles. A
  step in the declared direction is never treated as a reply, even when relationships exist both
  ways.
- `title`, `description`, `direction` and `layout` work on a dynamic view as on any declared view;
  `layout` and `direction` do not apply to the sequence style, which has its own layout.
- The wide-view hint applies to the numbered-edge style like any other view.
- `fmt` formats steps in canonical order and spacing; it never reorders steps.

## Requirements *(mandatory)*

### Functional Requirements

**Declaring a flow**

- **FR-001**: The language MUST let an author declare a dynamic view by writing `step` blocks
  inside a `view` block: a view with at least one step is a dynamic view. No new top-level block
  is added, and `dynamic` stays an excluded keyword. A dynamic view MUST NOT also set `include`,
  `exclude` or `tags`: its elements are the steps' elements, and setting them is an error.
- **FR-002**: A step MUST carry a name label, unique within its view, and MUST name its source and
  target elements; it MAY name the relationship it uses, give a description, and be marked as a
  reply. Its address is
  `view.<view>.step.<label>`, which does not change when steps are inserted or removed. Its
  position in the declaration is its number, starting at 1.
- **FR-003**: A dynamic view MUST accept the declared-view attributes `title`, `description`,
  `direction` and `layout`, and a style choosing numbered edges (default) or a sequence diagram.
- **FR-004**: A dynamic view MUST have an address (`view.<name>`), file name and page like any
  declared view.

**Checking**

- **FR-005**: Every element and relationship a step names MUST resolve, follow `moved` renames, and
  be reported at the step's line when it does not.
- **FR-006**: A step that names a relationship MUST have endpoints that match the relationship's,
  directly or through containment (an endpoint may be an ancestor or descendant of the
  relationship's end); a step marked `reply = true` matches the relationship reversed. Otherwise
  it MUST be an error naming both.
- **FR-007**: A step that names no relationship MUST match at least one declared relationship by
  the same rule; when none matches, it MUST be reported as a warning naming the step and its two
  elements, and the step MUST still be drawn.
- **FR-008**: `style` on a view without steps MUST be an error naming the attribute; a view without
  steps is an ordinary declared view.

**Drawing**

- **FR-009**: In the numbered-edge style, the diagram MUST draw each participating element once,
  inside the boundary of its system when it is a container or component (the system is drawn as a
  boundary even when it takes no step; no other ancestor is drawn unless it takes part), and one
  edge per step labelled `<number>. <description>` (the
  relationship's description and technology when the step gives none). Steps marked as replies
  are dashed.
- **FR-010**: In the sequence style, the diagram MUST draw one lifeline per participating element,
  in first-appearance order, and one arrow per step, top to bottom, labelled as in FR-009.
- **FR-011**: The view's page MUST list the steps in a table: number, from, to, description, with
  links to the elements' pages.
- **FR-012**: The compiled export MUST include each dynamic view's steps in order, with resolved
  addresses, the relationship when named or matched, and the description.
- **FR-013**: Models without a dynamic view MUST compile, export and render byte-identically to
  1.2.0.

**Tools**

- **FR-014**: The dynamic view and its steps MUST come from the single language table, so the
  decoder, the edit schema, the `schema` tool and `fmt` agree, and the parity test covers them.
- **FR-015**: `apply_edit` MUST be able to create and remove a dynamic view, and append a step,
  insert one before a named step, update one and remove one, with the same preview, revision and
  refusal behaviour as other edits.
- **FR-016**: `describe` and `query` MUST return dynamic views and their steps.

### Key Entities

- **Dynamic view**: a named, ordered flow through the model, with an optional title, description
  and drawing style. It is a declared view for addressing, pages and navigation.
- **Step**: one interaction in a flow: its label (its stable identity and address), its number
  (position), source and target elements, the relationship it uses (named or matched), whether it
  is a reply, and a description.
- **Participant**: an element that appears in at least one step; drawn once.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The report's `daily_run` flow (orchestrator → collector → vendor API → raw store) can
  be declared against the `first-run` model and builds with numbered edges and as a sequence
  diagram, with no step carried only as free text.
- **SC-002**: Every way a step can drift from the model (unknown element, unknown relationship,
  relationship between other elements, no relationship at all) produces its own diagnostic at the
  step's line, in 100% of the fixture cases.
- **SC-003**: The two styles of the same view number and label every step identically.
- **SC-004**: An assistant can create a three-step flow, insert a step and remove one through MCP
  only, with no refused edit caused by the dynamic view's shape.
- **SC-005**: Projects without a dynamic view produce byte-identical exports and diagrams to 1.2.0
  (the determinism job and the export goldens pass unchanged).
- **SC-006**: The 1,000-element build benchmark (feature 014, SC-006: under 10 seconds) still
  builds in under 10 seconds with a 40-step flow added.

## Assumptions

- Steps are sequential only: no parallel branches (1a, 1b) or nested sub-flows in this release.
- A step's number is its position; authors do not write numbers, so inserting a step renumbers the
  ones after it, while their labels and addresses stay the same.
- The sequence style uses the diagram renderer already in the binary; no new tool.
- Dynamic views appear in navigation and the site like declared views, listed by title.
- Step descriptions are plain strings, like relationship descriptions.
- The release is 1.3.0 (new language block or attributes).
- An ADR records the block form, the matching rule and the two styles.
