# ADR-0018: Dynamic views as steps inside `view`

**Status:** Accepted
**Date:** 2026-10-27
**Feature:** [019-dynamic-views](../../specs/019-dynamic-views/spec.md)
**Related:** [ADR-0015](0015-rendering-attributes.md) (layout engines, trigger relationships),
[ADR-0017](0017-language-polish.md) (view titles)

## Context

The first modelling session ([017 first-run report](../../specs/017-mcp-validation-ergonomics/first-run-report.md),
item #5) could not express an ordered interaction: a request path, a pipeline run. Relationships
have no order, and `sequence`, `dynamic`, `scenario` and `flow` blocks are refused. The C4 model's
dynamic diagram covers this. The report asked for flows that the compiler checks against the model,
drawn with numbered edges or as a sequence diagram.

## Decision

1. **Steps inside `view`.** A `view` with `step` blocks is a dynamic view. It keeps everything a
   view has (address, title, description, page, navigation) and gains `style = "edges" | "sequence"`.
   No top-level block is added, and `dynamic` stays an excluded keyword: it is Terraform's
   dynamic-block keyword, and a reader from Terraform would misread it. A dynamic view's elements
   are its steps' elements, so it takes no `include`, `exclude` or `tags`.

2. **Labelled steps.** Each step has a label, as each relationship does, and an address
   `view.<view>.step.<label>`. Its number is its position. Inserting a step renumbers the later
   ones but changes no address, so edits, `moved` history and the coming semantic diff see the
   inserted step, not a shift of every step after it.

3. **Steps are checked against relationships.** Elements are related when they are the same or one
   contains the other, so flows can be written at container level over component relationships and
   the other way round. Matching follows the drawn direction, so a trigger relationship reads
   queue → function. A reply is explicit (`reply = true`) and matches the relationship reversed:
   inferring replies would be ambiguous where relationships exist both ways. A `via` that does not
   connect the step is an error (`step_mismatch`); a step that matches nothing is a warning
   (`step_unmatched`) and is still drawn, so a flow can be sketched before its relationships exist.

4. **Two styles, one projection.** The projection builds the participants (containers and
   components inside their system's boundary, as C4's dynamic diagram draws them), one edge per
   step, and an ordered step list. The numbered-edge style uses the normal emitter; the sequence
   style uses D2's own `sequence_diagram`, already in the binary, which takes its order from the
   step list because the view model keeps edges sorted. Both number and label every step
   identically, and the page lists the steps in a table.

## Consequences

**Positive:**
- The model can say what happens in what order, and the compiler keeps the flow honest as the model
  changes.
- Every existing view feature (titles, pages, navigation, MCP edits) applies to flows unchanged.

**Negative:**
- A view block now has two modes, distinguished by the presence of steps.
- No parallel or nested steps (1a, 1b) in this release.

**Mitigations:**
- `dynamic_view_selection` refuses mixing the two modes, naming the attribute.
- Parallel steps can be added later as an attribute of `step` without changing addresses.
