# loko issues (from a first-model session with the MCP server)

Environment for all issues: `loko 1.0.1 (commit 5805e8a, built 2026-10-09)`, macOS, MCP server started with `loko mcp` over stdio, used by an AI coding agent (Claude Code).

Ordered by priority.

---

## 1. MCP: expose the language schema so clients do not learn it by trial and error

**Type:** enhancement · **Area:** MCP

### Problem
The MCP tools (`describe`, `query`, `apply_edit`, `validate`, `move`) do not say which block types exist or which attributes each block takes. A client must guess and read the refusals. In one session it took about eight refused `apply_edit` previews to learn the basics:

| Guess | Refusal |
|---|---|
| `set: { name: "..." }` on an element | `set.name: not an attribute of element; legal: description, docs, owner, tags, technology, title` |
| relationship without `target` | `set.target: required when adding a relationship` |
| `description` on an environment | `not an attribute of environment; legal: account, provider, region` |
| `technology` on a group (node) | `not an attribute of group; legal: (none)` |
| `node` on an instance | `not an attribute of instance; legal: attributes, of` |
| `title` on a view | `not an attribute of view; legal: direction, exclude, include, layout, tags` |

The refusal messages are very good: each one lists the legal values. But this discovery loop costs tokens and turns on every new project.

### Proposal
Add a read-only MCP tool (for example `schema` or `guide`) that returns:
- every top-level block type (`project`, `person`, `system`, `container`, `component`, `deployment`, `view`, …) and its nesting (`uses`, `node`, `instance`, `binding`);
- each block's attributes, with type, required / optional and allowed values (for example `layout`: `dagre | elk`);
- the address grammar used by `apply_edit` (`container.x`, `person.x.uses.y`, `deployment.env.node.n.instance.i`, `view.v`);
- how `docs` paths resolve (see #4);
- one short example per block.

You could also put a one-line pointer to this tool in the MCP server instructions.

---

## 2. Warnings: let external elements opt out of structural checks, and group warning output

**Type:** enhancement · **Area:** validation

### Problem
A typical model includes systems the team does not own: SaaS APIs, cloud-managed services, a customer's system of record. These systems will never have containers or local prose. Each one still raises two warnings:

```
empty_system   system.<x> contains nothing. A system with no containers cannot be drawn at container level.
missing_docs   system.<x> declares no docs file. Prose is what a diagram cannot carry.
```

In a 12-container model, 33 warnings came out:
- about 15 for external systems;
- 8 `unbound_instance` warnings, because the IaC does not exist yet;
- the rest were genuine.

`tags = ["external"]` does not suppress anything. The warnings that matter are hard to find among the others.

The MCP `validate` output adds to this: each warning repeats a full `range` object (start/end line, column, byte), so 33 warnings fill a lot of context.

### Proposal
1. An explicit marker such as `external = true` on `person` / `system`, or a `software_system "x" { external = true }` style. A marked element:
   - skips `empty_system` and `missing_docs`;
   - can render differently (C4 convention: grey).
2. A lifecycle state for instances, for example `planned = true`, which skips `unbound_instance` until a binding is expected.
3. Options for `validate` (CLI and MCP):
   - `--summary`: counts grouped by `code`, with the addresses listed under each code;
   - `--severity error` and `--ignore <code>` filters;
   - in MCP `toon` output, only `file:line` per diagnostic unless the client asks for full ranges.
4. Optional: a project-level setting to turn named warning codes off.

---

## 3. `apply_edit`: cannot place a declaration in a new file; the refusal message is misleading

**Type:** bug · **Area:** MCP / apply_edit

### Repro
Project root `.` contains `architecture/workspace.loko.hcl` with a `project` block. `architecture/context.loko.hcl` does not exist.

```json
{ "op": "add", "target": "element", "address": "person.analyst",
  "file": "architecture/context.loko.hcl",
  "set": { "title": "Analyst" } }
```

### Actual
```
refusal: path_refused — architecture/context.loko.hcl resolves outside the project
```

### Expected
- The declaration is written to a new file `architecture/context.loko.hcl`. The path is inside the project root; the file just does not exist yet.
- If creating files is not allowed on purpose, the message should say that ("file does not exist; apply_edit only edits existing files"). It should not say the path is outside the project.

### Impact
A client cannot split a model across files through MCP. It must write the files by hand, which skips the preview and revision safety of `apply_edit`.

---

## 4. `docs` paths resolve from the project root, not from the declaring file

**Type:** bug or docs · **Area:** compiler

### Repro
```
architecture/containers.loko.hcl   ->  system "s" { docs = "docs/s.md" }
architecture/docs/s.md             (exists)
```

### Actual
```
docs_not_found  system.s references "docs/s.md", which does not exist.
```
It works only as `docs = "architecture/docs/s.md"`.

### Expected
One of these:
- (preferred) resolve relative to the file that declares the element, like HCL `path.module` or Markdown links; or
- keep root-relative resolution, but print the resolved path in the warning (`looked for <root>/docs/s.md`) and document the rule.

Root-relative paths also break when someone moves a `.loko.hcl` file together with its docs folder.

---

## 5. Dynamic / sequence views for ordered interactions

**Type:** feature · **Area:** language, rendering

### Problem
There is no way to express an ordered interaction (a request flow, a pipeline run, a review loop):
- `sequence`, `dynamic`, `scenario` and `flow` blocks are all "Unsupported block type";
- `view.layout` accepts only `dagre | elk`;
- `uses` relationships have no order.

C4 includes the dynamic diagram for this. It is often the most useful diagram for a pipeline or an event-driven system.

### Proposal
A `dynamic` view that refers to existing elements and relationships, so the compiler checks every step:

```hcl
dynamic "daily_run" {
  description = "One scheduled run"

  step { from = container.orchestrator  to = container.collector  via = container.orchestrator.uses.collect }
  step { from = container.collector     to = system.vendor_api    description = "Search + fetch" }
  step { from = container.collector     to = container.raw_store }
}
```

- Validation: each step's endpoints must exist. If `via` is set, the relationship must exist and connect those endpoints.
- Rendering: numbered edges in a normal layout, or `style = "sequence"` to render with D2's `sequence_diagram` shape (D2 is already in the binary).

---

## 6. Deployment nodes cannot describe cloud infrastructure

**Type:** enhancement · **Area:** language (deployment)

### Problem
- `node` blocks take no attributes (`legal: (none)`): no title, no technology, no description.
- `instance` takes only `of` and free-form `attributes`.

So a common cloud layout ("AWS account → region → VPC → Lambda / DynamoDB / S3") can only be shown by node *names*. The deployment diagram cannot show what each node is.

### Proposal
- Give `node` `title`, `technology`, `description` and `tags`.
- Optionally add `technology` on `instance`, to override the container's technology for a given deployment (for example the container is "Lambda" in prod and "local process" in dev).

---

## 7. Views: add `title` and `description`

**Type:** enhancement · **Area:** language (views)

### Problem
A declared `view` takes only `direction`, `exclude`, `include`, `layout` and `tags`. A view such as `view "phase_1"` cannot say what it shows or why. The rendered page shows only the identifier.

### Proposal
Add `title` and `description`. Render them as the diagram heading and a caption, and in the site navigation.

---

## 8. MCP: add `build` (and optionally `fmt`) tools

**Type:** enhancement · **Area:** MCP

### Problem
Through MCP, a client can edit and validate the model but cannot render it. To check a diagram it must use the shell (`loko build`, `loko fmt`, `loko serve`). Clients without shell access cannot do this at all.

### Proposal
- `build`: inputs `formats`, `out`, optional `view`; returns the written paths (and optionally the SVG or D2 for one view inline).
- `fmt`: format the source files in place, or return a diff in preview mode.

---

## 9. Empty project: `describe` dead-ends instead of helping bootstrap

**Type:** enhancement · **Area:** MCP

### Repro
Run MCP `describe` in a directory with no `*.loko.hcl` files.

### Actual
`ok: false`, `no_source_found`: "create one, for example arch.loko.hcl, with a project block." `apply_edit` cannot create that file (see #3).

### Proposal
Either:
- an `init` tool (`name`, `description`, `file`) that writes the `project` block; or
- let the first `apply_edit` in an empty project create the file named in `file`, with a `project` block.

The message should also show the minimal valid syntax: `project "name" { description = "..." }`. A first guess with a `title` attribute was refused.

---

## 10. Default layout of wide container views

**Type:** enhancement · **Area:** rendering

### Problem
A system with about 10 containers, 3 people / external systems on each side and about 25 relationships rendered as a very wide, short diagram with the default `layout` and `direction` (neither was set). Labels were unreadable at normal zoom.

### Proposal
- Pick `direction` from the aspect ratio. For example, switch to `down` when the width-to-height ratio of the `right` layout goes over a limit.
- Or emit an info-level diagnostic suggesting `direction` or `layout = "elk"` for that view.
