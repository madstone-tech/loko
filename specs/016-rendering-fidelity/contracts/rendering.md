# Contract: rendering

**Feature**: `016-rendering-fidelity` | **Status**: normative

## Nodes

| Input | Diagram (D2/SVG) | Site page | Markdown |
|---|---|---|---|
| no `title` | today's label, byte-identical | heading = name | heading = name |
| `title` | plain label: title, then `[Kind: Technology] · name`, then description (no markdown labels: they render blank outside browsers) | heading = title, `<small><code>name</code></small>` beneath | heading = title, `` `name` `` beneath |
| `shape` | D2 shape per research R3; the kind's colours | page shows the shape as a field | as site |

## Edges

| Relationships on the edge | Drawn |
|---|---|
| all `sync` (or unset) | solid, as today |
| all `async` | `stroke-dash: 8` |
| a mix | solid |
| a `trigger` | its ends swapped before merging: the arrow points from the trigger to the invoked element |
| crossing the view boundary | unchanged: `stroke-dash: 4` to the outside marker |
| with `tags` | a final label line, `#tag` per tag, sorted (the union for merged edges) |

## Direction

`direction: down` for system, container and declared views, unless the view sets `direction`;
`direction: right` for the landscape and deployment views.

## Guarantees

- Output stays byte-identical across runs and machines.
- A project with none of the new attributes renders byte-identically, apart from the `direction`
  line in system, container and declared views.
- The `serverless-reference` container view renders with SVG `viewBox` width ÷ height ≤ 2.
