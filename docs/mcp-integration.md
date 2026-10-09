# MCP Integration

`loko mcp` runs a [Model Context Protocol](https://modelcontextprotocol.io/) server over stdio. An
assistant connected to it can read your architecture, ask it questions, and edit the HCL source:
safely, through five tools.

- **Reads** (`describe`, `query`, `validate`) answer from the compiled architecture.
- **Writes** (`apply_edit`, `move`) change `*.loko.hcl` files and nothing else. Every change is
  compiled before it is saved; a change that would not compile is refused, and nothing is written.
- Comments, ordering and spacing you wrote by hand survive every edit: only the declaration being
  changed is touched.

Design notes are in [ADR-0014](adr/0014-hcl-authoring.md).

## Setup

Install loko, then add it to your client's MCP configuration. For Claude Desktop
(`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS,
`%APPDATA%\Claude\claude_desktop_config.json` on Windows):

```json
{
  "mcpServers": {
    "loko": {
      "command": "loko",
      "args": ["mcp", "--project", "/path/to/your/architecture"]
    }
  }
}
```

For Claude Code: `claude mcp add loko -- loko mcp --project /path/to/your/architecture`.

The server registers exactly five tools, listed in name order.

## Working with the tools

1. **Read first.** Start with `describe`; every read returns a `revision`.
2. **Quote the revision when you write.** `apply_edit` and `move` take `base_revision`. If a file
   the write would change has been modified since that read (by a person in an editor, say), the
   write is refused as `stale_revision` and the newer change is kept. Read again and retry. Each
   successful write returns the new revision.
3. **Preview when unsure.** `preview: true` returns unified diffs and writes nothing.
4. **Batch related changes.** `apply_edit` takes up to 100 edits. They apply in order, compile once,
   and save all-or-nothing.

Reads accept `format`: `toon` (the default, token-efficient) or `json`.

## `describe`

| Argument | Default | |
|---|---|---|
| `level` | `summary` | `summary`: project, counts by kind, environments, top-level elements. `structure`: the tree down to containers, with technologies and component counts. `full`: every attribute, relationship and placement |
| `address` | | Scope to one element: the element, its children and its relationships |
| `format` | `toon` | |

```text
→ describe {}
ok: true
level: summary
project:
  name: two-systems
counts[#5]{kind,count}:
  component,4
  container,6
  ...
revision: r1-3f9a0c41d2e8b7a6
```

If the project does not compile, `describe` returns `ok: false` and the diagnostics, not a
partial answer.

## `query`

| Argument | | |
|---|---|---|
| `kind` | required | `dependents`, `dependencies`, `path`, `orphans` or `coupling` |
| `address` | dependents, dependencies, path | The element, or a path's start |
| `to` | path | A path's end |
| `transitive` | `false` | Follow relationships to any depth |
| `limit` | `20` | Rows for `coupling` |

An element stands for itself and everything inside it, so the dependents of `container.orders_db`
include a component in another container that calls it. An unknown address returns
`ok: false`, `error.reason: not_found` and up to three suggestions. `loko query` on the command
line gives the same answers.

## `validate`

Compiles the project and returns every error and warning with its file, line and column, plus the
revision.

## `apply_edit`

| Argument | | |
|---|---|---|
| `base_revision` | required | From your last read |
| `edits` | required | 1–100 edits, applied in order |
| `preview` | `false` | Return diffs; write nothing |

Each edit:

```json
{
  "op": "add | update | remove | rename",
  "target": "element | relationship | environment | group | instance | binding",
  "address": "container.api",
  "set": { "system": "system.shop", "technology": "Go", "tags": ["edge"] },
  "clear": ["owner"],
  "cascade": false,
  "to": "component.api",
  "file": "payments.loko.hcl",
  "binding": { "kind": "terraform", "index": 0 }
}
```

| Target | Address | Attributes you can set |
|---|---|---|
| element | `container.api` | `description`, `owner`, `technology`, `tags`, `docs`; `system` (container) or `container` (component), required on add |
| relationship | `container.api.uses.orders` | `target` (required on add), `description`, `technology` |
| environment | `deployment.prod` | `provider`, `account`, `region` |
| group | `deployment.prod.node.vpc.subnet-a` | none |
| instance | `deployment.prod.instance.api`; to place a new one in a group, `deployment.prod.node.vpc.instance.api` | `of` (required on add), `attributes` (a flat object) |
| binding | the instance's address, plus `binding.kind` (`terraform` or `cloudformation`) | exactly one of `address`, `addresses`, `tags` |

- References (`system`, `container`, `target`, `of`) take an address string and are written as
  bare references, never quoted strings.
- `docs` records a path only; the file it names is never created.
- **Where new declarations go:** an explicit `file` (a new `*.loko.hcl` file is created); otherwise
  containers and components beside their parent, nested declarations inside their parent, and
  everything else in the file that declares the project.
- **Removing** something still referred to is refused as `dangling_references`, listing every
  referrer. `cascade: true` removes the dependents too: relationships into it, its children, its
  instances, and its entries in views. A group or environment that still holds instances needs
  `cascade` as well. The result lists everything removed.
- **Renaming** (`op: rename`, with `to`) works like `move`.

The result holds `ok`, `files` (each with a unified diff), `removed`, `diagnostics` (warnings on
success, errors on a compile refusal) and the new `revision`. A no-op returns `noop: true` and
writes nothing.

## `move`

| Argument | |
|---|---|
| `from`, `to` | Element addresses; the name, the kind, or both may change |
| `base_revision` | required |
| `preview` | `false` |

Rewrites every reference in every file, changing only the reference itself, and appends a
`moved { from = …, to = … }` block so history survives the rename (see
[the language reference](language.md#moved)). If the new kind needs a different parent, use
`apply_edit` with a batch: the rename, then an update setting the parent.

## Refusals

A refusal is a normal tool result with `ok: false`. Nothing was written.

| `refusal.reason` | Meaning |
|---|---|
| `compile_errors` | The edited architecture would not compile; `diagnostics` says why |
| `stale_revision` | A file the write changes was modified since your read, or the revision is unknown; read again |
| `dangling_references` | The removal would leave references; `dependents` lists them; consider `cascade` |
| `address_in_use` | An add or rename onto an address that is already declared |
| `not_found` | The target does not exist |
| `invalid_edit` | The edit itself is malformed; `edit` is its index in the batch, `detail` names the field |
| `path_refused` | `file` is outside the project or is not a `*.loko.hcl` file |

Malformed requests, such as an unknown tool or an argument of the wrong type, are JSON-RPC
errors instead.
