# Contract: MCP tools

**Feature**: `015-mcp-hcl-authoring` | **Status**: normative

The server registers exactly these five tools (FR-027). `tools/list` returns them sorted by name.
Handlers parse arguments, call one use case and encode the result (≤ 30 effective lines,
Principle III). Refusals are **tool results** with `ok: false`; JSON-RPC errors are only for
malformed requests (unknown tool, arguments of the wrong type).

Reads accept `format`: `"toon"` (the default) or `"json"` (FR-008).

---

## `describe`

| Argument | Type | Default | |
|---|---|---|---|
| `level` | `"summary"` \| `"structure"` \| `"full"` | `"summary"` | R6 |
| `address` | string | none | scope to one element (FR-002) |
| `format` | `"toon"` \| `"json"` | `"toon"` | |

Returns a `DescribeResult` (data-model §3), including `revision`. A compile failure returns
`ok: false` with diagnostics (FR-007).

## `query`

| Argument | Type | Required | |
|---|---|---|---|
| `kind` | `"dependents"` \| `"dependencies"` \| `"path"` \| `"orphans"` \| `"coupling"` | yes | |
| `address` | string | dependents, dependencies, path | path: the start |
| `to` | string | path | |
| `transitive` | bool | no (false) | dependents and dependencies |
| `limit` | int | no (20) | coupling |
| `format` | `"toon"` \| `"json"` | no | |

Returns a `QueryResult`. An unknown address returns `ok: false`, `reason: not_found`, and up to 3
suggestions (FR-030).

## `validate`

| Argument | Type | Default |
|---|---|---|
| `format` | `"toon"` \| `"json"` | `"toon"` |

Returns `{ok, errors, warnings, diagnostics[], revision}`. Each diagnostic has `code`, `severity`,
`summary`, `detail`, `address`, and `range{file, startLine, startColumn, endLine, endColumn}`
(FR-006).

## `apply_edit`

| Argument | Type | Required | |
|---|---|---|---|
| `base_revision` | string | yes | from any read result (FR-016) |
| `edits` | array of Edit (below), 1..100 | yes | applied in order, compiled once, saved all-or-nothing (FR-012a) |
| `preview` | bool | no (false) | FR-012b |

Each Edit:

```json
{
  "op": "add | update | remove | rename",
  "target": "element | relationship | environment | group | instance | binding | view",
  "address": "container.api",
  "binding": { "kind": "terraform", "index": 0 },
  "set": { "description": "Storefront API", "system": "system.shop", "tags": ["edge"] },
  "clear": ["owner"],
  "cascade": false,
  "to": "component.api",
  "file": "payments.loko.hcl"
}
```

- An instance's address is `deployment.<env>.instance.<name>`. To place a new one inside a node
  group, an `add` may name the group path: `deployment.prod.node.vpc.subnet-a.instance.api`.
- A rename that changes kind drops the parent attribute the new kind does not take.
- Revisions are short opaque tokens; the server remembers the last 64 it issued.
- Reference-valued attributes (`system`, `container`, `target`, `of`) take an address string and
  are written as bare traversals, never quoted strings.
- `docs` takes a project-relative path and only sets the reference; the file it names is never
  created or written (FR-026a).
- `attributes` (instance) and binding `tags` take a flat string-to-scalar object.

Result: `EditResult` (data-model §2.4). On success, `files[]` lists the changed files with
unified diffs, `diagnostics` holds the warnings, and `revision` is the new revision. On
`preview: true` it returns the same shape with `preview: true`, and nothing is written.

## `move`

| Argument | Type | Required |
|---|---|---|
| `from` | string (element address) | yes |
| `to` | string (element address; the kind and/or name may change) | yes |
| `base_revision` | string | yes |
| `preview` | bool | no |

Equivalent to `apply_edit` with one `rename` edit. It rewrites every reference in every file and
appends `moved { from = …, to = … }` to the declaring file (FR-022..FR-025). Changing the kind
where the new kind needs a different parent requires `apply_edit` with a batch: the rename, then an
update setting the parent attribute.

---

## Refusal reasons

| `reason` | When | Payload |
|---|---|---|
| `compile_errors` | the planned result does not compile | `diagnostics` |
| `stale_revision` | a file the plan would change differs from `base_revision` | the changed paths, with a hint to read again |
| `dangling_references` | remove without cascade, target still referenced | `dependents[]` (addresses with ranges) |
| `address_in_use` | add or rename onto an existing address | the existing declaration |
| `not_found` | the target does not exist | up to 3 suggestions |
| `invalid_edit` | fails `NewEdit` validation | the field and the problem; `edit` is the batch index |
| `path_refused` | `file` is outside the root or not `*.loko.hcl` | the path |
