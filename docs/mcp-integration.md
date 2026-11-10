# MCP Integration

`loko mcp` runs a [Model Context Protocol](https://modelcontextprotocol.io/) server over stdio. An
assistant connected to it can learn the language, read your architecture, ask it questions, edit
the HCL source and render it: safely, through eight tools.

- **The language** (`schema`) lists every block, attribute and allowed value. The server's
  handshake instructions tell clients to call it before their first edit.
- **Reads** (`describe`, `query`, `validate`) answer from the compiled architecture.
- **Writes** (`apply_edit`, `move`, `fmt`) change `*.loko.hcl` files and nothing else. Every change
  is compiled before it is saved; a change that would not compile is refused, and nothing is
  written.
- **Rendering** (`build`) writes diagrams, markdown and a site into a directory inside the project.
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

The server registers eight tools, listed in name order.

## Working with the tools

1. **Learn the language, then read.** Call `schema` once, then `describe`; every read returns a
   `revision`. In an empty directory, `describe` returns `bootstrap`: the smallest valid project
   declaration and a ready-to-send `apply_edit` request that creates it.
2. **Quote the revision when you write.** `apply_edit` and `move` take `base_revision`. If a file
   the write would change has been modified since that read (by a person in an editor, say), the
   write is refused as `stale_revision` and the newer change is kept. Read again and retry. Each
   successful write returns the new revision.
3. **Preview when unsure.** `preview: true` returns unified diffs and writes nothing.
4. **Batch related changes.** `apply_edit` takes up to 100 edits. They apply in order, compile once,
   and save all-or-nothing.

Every tool accepts `format`: `toon` (the default, token-efficient) or `json`.

**Diagnostics** in every tool's result are compact by default: `severity`, `code`, `detail`,
`address` and `at` (`file:line`). `ranges: "full"` returns the full form instead, with `summary`,
`range` (line, column and byte positions) and `related`. ([ADR-0016](adr/0016-mcp-ergonomics.md))

## `schema`

| Argument | Default | |
|---|---|---|
| `block` | | One block, for example `view`; omit for the whole language |

Read-only. Returns every block with its labels (and the allowed label values, such as `binding`'s
`terraform` and `cloudformation`), where it may nest, its attributes with type, required flag,
allowed values and a note, and a minimal example; plus every `apply_edit` address form and the
`docs` path rule. The whole language is under 1,000 tokens of TOON. An unknown `block` returns
`error.reason: unknown_block` with the block names.

Every result also carries `server_version`, the version of the serving binary — `dev` for a
development build — so a client can tell a stale server from the file's `loko_version` constraint.
The `view` block lists the reserved view names (`landscape`, `system-<name>`,
`container-<name>`, `deployment-<environment>`): naming one of these replaces the view the
compiler derives automatically, reported by `validate` as `view_shadowed`.

## `describe`

| Argument | Default | |
|---|---|---|
| `level` | `summary` | `summary`: project, counts by kind, environments, top-level elements. `structure`: the tree down to containers, with technologies and component counts. `full`: every attribute, relationship and placement. Every level lists the declared views (with `include` and `exclude` at `full`) |
| `address` | | Scope to one element: the element, its children and its relationships |
| `format` | `toon` | |

Every result also carries the server's operating context: `server_version`, `root` (the resolved
absolute project root the tools read and write inside) and `files` (the `*.loko.hcl` sources the
compile actually read). An agent compares them against what it thinks it is driving before an
edit lands.

```text
→ describe {}
ok: true
level: summary
server_version: 1.4.0
root: /home/me/architecture
files[#2]: main.loko.hcl, domains/payments.loko.hcl
project:
  name: two-systems
counts[#5]{kind,count}:
  component,4
  container,6
  ...
revision: r1-3f9a0c41d2e8b7a6
```

If the project does not compile, `describe` returns `ok: false` and the diagnostics, not a
partial answer. A dynamic view in `views` also carries its `style` and its `steps` in order (number,
name, from, to, the relationship used, reply, description). `query` with `kind: "flows"` lists the
steps an element takes part in.

## `query`

| Argument | | |
|---|---|---|
| `kind` | required | `dependents`, `dependencies`, `path`, `orphans`, `coupling`, `flows`, `views`, `elements` or `relationships` |
| `address` | dependents, dependencies, path, flows | The element, or a path's start |
| `to` | path | A path's end |
| `transitive` | `false` | Follow relationships to any depth |
| `limit` | `20` | Rows for `coupling` |
| `element_kind` | elements | Only this kind: `person`, `system`, `container`, `component` or `external`. Another value returns `invalid_query` |

The listing kinds take no address. `views` returns `views` (`name`, `kind`, `title`), one per view
that the build produces, generated and declared, sorted by name. `elements` returns `elements`
(`address`, `kind`, `name`, `title`, `distance`), sorted by address. `relationships` returns
`relationships` (`address`, `source`, `target`, `kind`, `description`), sorted by address. A
relationship with no `kind` shows `sync`.

An element stands for itself and everything inside it, so the dependents of `container.orders_db`
include a component in another container that calls it. An unknown address returns
`ok: false`, `error.reason: not_found` and up to three suggestions. `loko query` on the command
line gives the same answers.

## `validate`

Compiles the project and returns every error, warning and information diagnostic, plus the
revision. Information (`severity` `info`) is advice and never counts as a warning.

| Argument | Default | |
|---|---|---|
| `summary` | `false` | Group by code: `summary` holds one entry per code with a count and up to 20 addresses (`more` counts the rest) |
| `severity` | | `error` shows errors only; `warning` shows errors and warnings, not information. The `errors` and `warnings` counts are unchanged |
| `ignore` | | Warning and information codes to leave out, on top of the project's `ignore_warnings`. An error code or an unknown code is refused |
| `ranges` | `line` | `full` for full source ranges |

`ignored` counts the warnings left out.

## `apply_edit`

| Argument | | |
|---|---|---|
| `base_revision` | required | From your last read |
| `edits` | required | 1–100 edits, applied in order |
| `preview` | `false` | Return diffs; write nothing |
| `diagnostics` | `changed` | `changed` returns the diagnostics that the change added and removed, and counts the rest. `all` returns every diagnostic after the change, as in 1.5 |
| `diff` | `meaning` | `meaning` shows a line that changed only in whitespace as a context line. `exact` returns the 1.5 diff |

Each edit:

```json
{
  "op": "add | update | remove | rename",
  "target": "element | relationship | environment | group | instance | binding | view | project",
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
| element | `container.api` | `description`, `owner`, `technology`, `tags`, `docs`, `title`; `shape` (container and external); `system` (container) or `container` (component), required on add |
| relationship | `container.api.uses.orders` | `target` (required on add), `description`, `technology`, `kind` (`sync`, `async`, `trigger`), `tags` |
| environment | `deployment.prod` | `provider`, `account`, `region` |
| group | `deployment.prod.node.vpc.subnet-a` | `title` (not empty), `technology`, `description`, `tags` |
| instance | `deployment.prod.instance.api`; to place a new one in a group, `deployment.prod.node.vpc.instance.api` | `of` (required on add), `attributes` (a flat object), `planned` (bool), `technology` |
| binding | the instance's address, plus `binding.kind` (`terraform` or `cloudformation`) | exactly one of `address`, `addresses`, `tags` |
| view | `view.payments` | `title` (not empty), `description`, `include` and `exclude` (lists of element addresses, written unquoted), `tags`, `direction` (`down`, `right`), `layout` (`dagre`, `elk`), `style` (`edges`, `sequence`; a view with steps only) |
| step | `view.payments.step.charge` | `from` and `to` (required on add), `via`, `description`, `reply` (bool). On add, `before: "<label>"` inserts the step before that one; without it the step is appended |
| project | `project` | on add: `name` (the label) and `file`, both required, and `description`; on update: `description`, `layout`, `loko_version`, `ignore_warnings` |

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
- **The project** is added once, in the file you name; that is how a client starts from an empty
  directory. A second add is refused as `project_exists`. Adding an element before any `project`
  block exists is refused as `no_project`, with the edit to send.

The result holds `ok`, `files`, `removed`, the diagnostics and the new `revision`. A no-op returns
`noop: true` and writes nothing.

Each entry in `files` has a `path`, a unified `diff` and, with `diff: "meaning"`, a `whitespace`
list. An edit can realign the `=` signs of the attributes next to it. The meaning diff shows such a
line as a context line with its new text, and `whitespace` gives its line number in the new file.
A line that also changed in meaning stays a `-` and `+` pair. Whitespace inside a string is
meaning. The file on disk is always written in canonical form.

A successful result (applied, preview or no-op) gives the diagnostics as a change:

| Field | Content |
|---|---|
| `diagnostics_added` | The diagnostics in the model after the change that were not there before |
| `diagnostics_removed` | The diagnostics before the change that are not there now. `at` gives the position before the change |
| `diagnostics_unchanged` | The count of the other diagnostics: `errors`, `warnings` and `info`. A count of 0 is left out. The field is left out when all counts are 0 |

Two diagnostics are the same when their `code`, `severity`, `address`, file and `detail` are the
same. Positions do not count, so a diagnostic that only moved down is unchanged. A rename changes
an address, so the diagnostics of the renamed element show as removed and added.

A refusal always returns `refusal` and the full `diagnostics` list of the refused compile.

To get the 1.5 result, send `diagnostics: "all"` and `diff: "exact"`. `diagnostics: "all"` returns
one `diagnostics` list with every diagnostic after the change.
([ADR-0023](adr/0023-mcp-write-results.md))

## `move`

| Argument | |
|---|---|
| `from`, `to` | Element addresses; the name, the kind, or both may change |
| `base_revision` | required |
| `preview` | `false` |
| `diagnostics` | `changed`. `all` for the 1.5 list, as in `apply_edit` |
| `diff` | `meaning`. `exact` for the 1.5 diff, as in `apply_edit` |

Rewrites every reference in every file, changing only the reference itself, and appends a
`moved { from = …, to = … }` block so history survives the rename (see
[the language reference](language.md#moved)). If the new kind needs a different parent, use
`apply_edit` with a batch: the rename, then an update setting the parent.

## `fmt`

| Argument | | |
|---|---|---|
| `base_revision` | required | From your last read |
| `preview` | `false` | Return diffs; write nothing |
| `diagnostics` | `changed` | `all` returns every diagnostic after formatting |

Formats every `*.loko.hcl` file canonically, exactly as `loko fmt` does. It is a write like any
other: refused as `stale_revision` if a file changed since your read, and queued with `apply_edit`.
A file that does not parse is refused as `compile_errors` and nothing is formatted. The result has
the `apply_edit` shape; nothing to format returns `noop: true`.

`fmt` compiles the project once, after formatting. Formatting changes only whitespace, so
`diagnostics_added` and `diagnostics_removed` are always empty and `diagnostics_unchanged` counts
the diagnostics. With `diagnostics: "all"`, `fmt` lists every diagnostic. In 1.5 it returned no
diagnostics on success. `fmt` always returns the exact diff, because every change it makes is
whitespace.

## `build`

| Argument | Default | |
|---|---|---|
| `formats` | all | `d2`, `svg`, `md`, `html` |
| `out` | `dist` | Inside the project; anything else is refused as `path_refused` |
| `inline` | | A view name, such as `system-shop`, whose diagram to return |
| `inline_format` | `d2` | `d2` or `svg` |
| `strict` | `false` | As `loko build --strict`: files are written, and warnings make `ok` false |
| `ignore` | | As in `validate` |

Writes what `loko build` writes, byte for byte, and returns `paths`, the count `written`, and
`diagnostics`. With `inline`, `inline.content` is that view's diagram, so you can check it without
reading files; an unknown view returns `inline_error` with the view names. A model with errors
writes nothing. Like `loko build`, it removes only files a previous build wrote. A build with SVG
output can return a `wide_view` information diagnostic for a view that renders too wide, naming its
size and the `direction` or `layout` to try; `ignore: ["wide_view"]` leaves it out.

## Refusals

A refusal is a normal tool result with `ok: false`. Nothing was written.

| `refusal.reason` | Meaning |
|---|---|
| `compile_errors` | The edited architecture would not compile; `diagnostics` says why |
| `stale_revision` | A file the write changes was modified since your read, or the revision is neither one the server issued nor current; read again |
| `dangling_references` | The removal would leave references; `dependents` lists them; consider `cascade` |
| `address_in_use` | An add or rename onto an address that is already declared |
| `not_found` | The target does not exist |
| `invalid_edit` | The edit itself is malformed; `edit` is its index in the batch, `detail` names the field |
| `path_refused` | `file`, or `build`'s `out`, is outside the project, or `file` is not a `*.loko.hcl` file |
| `project_exists` | The project is already declared; `detail` names the file |
| `no_project` | An element was added before any `project` block; `detail` ends with the edit that creates it |

Malformed requests, such as an unknown tool or an argument of the wrong type, are JSON-RPC
errors instead.
