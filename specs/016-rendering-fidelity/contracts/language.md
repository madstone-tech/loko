# Contract: language additions

**Feature**: `016-rendering-fidelity` | **Status**: normative. Permanent v1.x surface (additive).

```hcl
container "orders_db" {
  system     = system.shop
  title      = "Orders database"
  shape      = "database"
  technology = "PostgreSQL"
}

container "sync" {
  system = system.portal
  title  = "Sync Lambda"
  shape  = "function"

  uses "consume" {
    target = container.sync_queue
    kind   = "trigger"        # drawn sync_queue → sync; sync still depends on the queue
  }
  uses "publish" {
    target = container.alerts
    kind   = "async"          # drawn dashed
    tags   = ["on-failure"]
  }
}

view "event-flow" {
  tags      = ["sqs", "sns"]
  direction = "right"         # default for declared views is "down"
}
```

| Rule | Diagnostic |
|---|---|
| `title` is a non-empty string, on any element | `empty_title` |
| `shape` ∈ {database, queue, topic, function, bucket} | `invalid_attribute_value` (lists the set) |
| `shape` only on `container` and `external` | `shape_not_allowed` |
| `kind` ∈ {sync, async, trigger}; omitted means sync | `invalid_attribute_value` |
| `tags` on `uses` is a list of strings | `unknown_attribute` / type error, as for element tags |
| `direction` ∈ {down, right} | `invalid_attribute_value` |

- None of these attributes affects references, resolution, validation of other blocks, queries, or
  `moved`.
- `kind = "trigger"` means: the declaring element is invoked by its target. Queries treat it as
  any other relationship (the declaring element depends on the target); only diagrams reverse the
  arrow.
