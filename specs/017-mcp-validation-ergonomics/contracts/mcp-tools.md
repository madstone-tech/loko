# Contract: MCP tools

Changes to the tools in [015's contract](../../015-mcp-hcl-authoring/contracts/mcp-tools.md). Every
tool keeps `format` (`toon` default, or `json`). Examples are JSON for readability.

## Handshake

`initialize` adds `result.instructions`:

> Call `schema` before your first edit: it lists every block, attribute and allowed value, and the
> address forms `apply_edit` accepts. Model systems your team does not own as `external`. Send
> `preview: true` to see an edit's diff before writing it.

## Diagnostics in every tool *(changed)*

Every tool that returns diagnostics (`validate`, `describe`, `apply_edit`, `move`, `build`, `fmt`)
accepts `ranges` (`"line"` default, or `"full"`).

`ranges: "line"`:

```json
{"severity": "warning", "code": "missing_docs", "summary": "Element has no prose",
 "detail": "system.shop declares no docs file. Prose is what a diagram cannot carry.",
 "address": "system.shop", "at": "architecture/context.loko.hcl:12"}
```

`ranges: "full"`: the 1.0.2 shape, with `range` and `related`.

## `schema` *(new, read-only)*

| Input | Type | Notes |
|---|---|---|
| `block` | string | optional; one block name |

Output:

```json
{
  "blocks": [
    {"name": "instance", "labels": ["name"], "in": ["node"],
     "attributes": [
       {"name": "attributes", "type": "map"},
       {"name": "of", "type": "ref", "required": true},
       {"name": "planned", "type": "bool", "note": "infrastructure not built yet; no binding expected"}
     ],
     "example": "instance \"api\" {\n  of = container.api\n}\n"}
  ],
  "addresses": [
    {"target": "instance", "pattern": "deployment.<env>.node.<node>[.node.<node>…].instance.<name>",
     "example": "deployment.prod.node.web.instance.api"}
  ],
  "docs": "docs paths are relative to the project root, not to the file that declares them"
}
```

Refusal: `block` unknown → `{"reason": "unknown_block", "detail": "…", "suggestions": [<block names>]}`.

Budget: full answer ≤ 3,000 tokens in TOON; one block ≤ 500.

## `validate` *(changed)*

| Input | Type | Notes |
|---|---|---|
| `summary` | bool | group by code |
| `severity` | string | `"error"`, or omitted for all |
| `ignore` | string[] | warning codes; an error code or unknown code is refused |
| `ranges` | string | see above |

Output adds `ignored` (count) and, with `summary`, replaces `diagnostics` with:

```json
"summary": [
  {"code": "unbound_instance", "severity": "warning", "count": 3,
   "addresses": ["deployment.prod.node.web.instance.api", "…"]},
  {"code": "missing_docs", "severity": "warning", "count": 25, "addresses": ["…20 shown…"], "more": 5}
]
```

`errors` and `warnings` count what is left after `ignore`, before `severity`.

## `apply_edit` *(changed)*

New target `project`, address `project`.

```json
{"op": "add", "target": "project", "address": "project", "file": "architecture.loko.hcl",
 "set": {"name": "shop", "description": "Online shop"}}
```

| Op | Rule |
|---|---|
| `add` | `name` and `file` required; refused `project_exists` (naming the file) when one exists |
| `update` | `description`, `layout`, `loko_version`, `ignore_warnings` |
| `remove`, `rename` | refused: a project always has exactly one declaration |

Adding an element when no `project` block exists → refused `no_project`, and the refusal's
`suggestion` is the `add project` request above with placeholders.

`instance` targets accept `planned` (bool).

## `describe` *(changed)*

When no source exists: `ok: false`, the `no_source_found` diagnostic, and:

```json
"bootstrap": {
  "syntax": "project \"name\" {\n  description = \"...\"\n}\n",
  "request": {"base_revision": "<revision>", "edits": [{"op": "add", "target": "project",
    "address": "project", "file": "architecture.loko.hcl",
    "set": {"name": "<name>", "description": "<description>"}}]}
}
```

## `build` *(new)*

| Input | Type | Notes |
|---|---|---|
| `formats` | string[] | default: the `loko build` default set |
| `out` | string | default `dist`; must resolve inside the project root |
| `inline` | string | a view name to return inline |
| `inline_format` | string | `d2` (default) or `svg` |
| `strict` | bool | warnings fail the build, as `--strict` |
| `ignore` | string[] | as in `validate` |

Output:

```json
{"ok": true, "out": "dist", "formats": ["d2", "svg", "md", "html"], "written": 42,
 "paths": ["dist/index.html", "…"], "diagnostics": [],
 "inline": {"view": "system-shop", "format": "d2", "content": "…"}}
```

Refusals: `out` outside the root → `path_refused`; unknown `inline` → `unknown_view` with the view
names; errors in the model → `ok: false`, nothing written, diagnostics returned.

## `fmt` *(new)*

| Input | Type | Notes |
|---|---|---|
| `base_revision` | string | required, as in `apply_edit` |
| `preview` | bool | default `false`; `true` returns the diffs and writes nothing |

Output: the `apply_edit` result shape (`ok`, `diffs`, `revision`, `diagnostics`). Nothing to format →
`ok: true` with no diffs. A file with syntax errors → nothing written, diagnostics returned.
Changed since `base_revision` → `stale_revision`.
