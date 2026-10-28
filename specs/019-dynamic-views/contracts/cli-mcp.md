# Contract: Dynamic views in the CLI and MCP

## `apply_edit`

| Target | Address | Attributes |
|---|---|---|
| `step` | `view.<view>.step.<label>` | `from`, `to` (required on add), `via`, `description`, `reply` |
| `view` | `view.<view>` | adds `style` (`edges`, `sequence`) |

- `add` of a `step` appends it to the view; `before: "<label>"` inserts it before that step. An
  unknown `before` is refused as `not_found`; `before` on anything else is a field error.
- `update` sets and clears step attributes; `remove` deletes the step. Preview, revision and
  compile refusals are as for every edit.
- Adding a step to a view that sets `include`, `exclude` or `tags` is refused with the
  `dynamic_view_selection` compile diagnostic.

## `move`

Renaming an element or a relationship rewrites `from`, `to` and `via` in every step that names it.

## `describe`

Each view in `views` gains `style` (when set) and `steps`:

```json
{"address": "view.daily_run", "style": "sequence", "steps": [
  {"number": 1, "name": "start", "from": "container.scheduler", "to": "container.collector",
   "via": "container.scheduler.uses.start", "relationship": "container.scheduler.uses.start"}
]}
```

## `query` and `loko query`

New kind `flows`: `query kind=flows address=container.collector` lists the dynamic views and steps
the element takes part in, directly or through a contained element, sorted by view and number.

```json
{"kind": "flows", "flows": [
  {"view": "view.daily_run", "number": 1, "step": "start", "role": "to"},
  {"view": "view.daily_run", "number": 2, "step": "fetch", "role": "from"}
]}
```

## `schema`

`view` lists `style` and the nested `step` block; `step` lists its attributes, `In: view`, and an
example.

## Exit codes

Unchanged. `step_unmatched` is a warning: it fails `--strict` and can be listed in
`ignore_warnings` or `--ignore`.
