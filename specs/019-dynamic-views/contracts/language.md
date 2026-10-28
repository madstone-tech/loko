# Contract: Dynamic views in the language

A permanent v1.x addition: the `step` block inside `view`, and `view.style`. A project with no
step compiles, exports and renders exactly as in 1.2.0.

```hcl
view "daily_run" {
  title       = "One scheduled run"
  description = "What happens when the scheduler fires."
  style       = "sequence" # optional; "edges" by default

  step "start" {
    from = container.scheduler
    to   = container.collector
    via  = container.scheduler.uses.start
  }

  step "fetch" {
    from        = container.collector
    to          = system.vendor_api
    description = "Search + fetch"
  }

  step "results" {
    from  = system.vendor_api
    to    = container.collector
    reply = true
  }

  step "store" {
    from = container.collector
    to   = container.raw_store
  }
}
```

## `step`

| Attribute | Type | Required | Rule |
|---|---|---|---|
| label | name | yes | unique in its view; address `view.<view>.step.<label>` |
| `from`, `to` | element reference | yes | must resolve to an element |
| `via` | relationship reference | no | must connect `from` and `to` (see matching) |
| `description` | string | no | edge text; the relationship's description when absent |
| `reply` | bool | no | matches the relationship reversed; drawn dashed |

The step's number is its position in the view, from 1.

## Matching

Two elements are *related* when they are the same or one contains the other. A relationship
matches a step when the step's `from` is related to its source and `to` to its target, in the
direction the relationship is drawn (a `trigger` relationship is drawn from its target). With
`reply = true`, `from` and `to` swap.

| Situation | Result |
|---|---|
| `via` matches | used |
| `via` does not match | `step_mismatch` error naming both pairs of ends |
| no `via`, a relationship matches | used (the first by address); no diagnostic |
| no `via`, nothing matches | `step_unmatched` warning; the step is still drawn |

## `view` with steps

| Rule | Result |
|---|---|
| `include`, `exclude` or `tags` with steps | `dynamic_view_selection` error |
| `style` without steps | `dynamic_view_selection` error |
| `style` value | `edges` (default) or `sequence`; else `invalid_attribute_value` |
| `title`, `description` | as on any view |
| `direction`, `layout` | apply to `edges`; ignored by `sequence` |

## Drawing

| Style | Diagram |
|---|---|
| `edges` | each participant once; containers and components inside their system's boundary; one edge per step labelled `<n>. <text>`; replies dashed |
| `sequence` | one lifeline per participant in first-appearance order; one arrow per step, top to bottom, labelled the same; replies dashed; no boundaries |

Both: the view's page lists the steps (number, from, to, description).

## Codes

| Code | Severity |
|---|---|
| `step_mismatch` | error |
| `step_unmatched` | warning (ignorable) |
| `dynamic_view_selection` | error |
