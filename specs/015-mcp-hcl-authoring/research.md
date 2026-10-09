# Research: MCP HCL Authoring

**Feature**: `015-mcp-hcl-authoring` | **Date**: 2026-10-05 | **Plan**: [plan.md](plan.md)

Each entry records the decision, why it was made, and what was rejected. All Technical Context
unknowns are resolved here.

---

## R1. How an edit changes a file

**Decision**: Every write goes through `hclwrite` on whole files: parse a file with
`hclwrite.ParseConfig`, change it through the `Body`, `Block` and `Expression` API, and serialise it
with `File.Bytes()`. Specifically:

- *add*: build the new block in an empty `hclwrite.File`, canonicalise it with `hclwrite.Format`,
  and append it with `AppendNewline` + `AppendBlock`. Nested blocks (`uses`, `node`, `instance`,
  `binding`) go into the parent block's body.
- *update*: call `SetAttributeValue`, `SetAttributeTraversal` (references) or `RemoveAttribute` on
  the target block's body.
- *remove*: `RemoveBlock` on the parent body.
- *rename*: covered in R4.

Existing lines are never reformatted. `hclwrite` keeps every token it was not asked to change, so
bytes outside the edited declaration survive (FR-013). A new declaration is canonical (FR-014)
because it was formatted on its own before being spliced in. A changed attribute line is emitted in
`hclwrite`'s canonical token form, and its neighbours are not realigned. `loko fmt` remains the way
to realign a whole file.

**Serialisation (implementation finding)**: `hclwrite.File.Bytes()` runs the formatter over the
whole file, so it would realign lines nobody asked to change. The editor therefore always writes a
file from its raw token stream (`BuildTokens(nil).Bytes()`), which reproduces hand-written source
byte for byte. Tokens an edit creates carry no layout, so new code is formatted on its own, at its
nesting depth, parsed back, and spliced in.

**Placement of a new attribute**: `hclwrite` can only append, so an attribute added to a block goes
at the end of its body, after any nested blocks. It is valid HCL, and `loko fmt` does not reorder
it. A one-line block (`uses "x" { target = y }`) is expanded to the multi-line form first.

**Guard**: before editing a file, the editor checks that its token stream reproduces `src`. A
file that does not round-trip byte-identically is refused rather than risked. Comparing a file
after the edit with its original outside the edited declaration is the property tested in R10.

**Declaration span** (used by FR-013 and by the byte oracles):
- the block's lines, from its header to its closing brace;
- its attached leading comment lines, meaning contiguous comment lines directly above the header
  with no blank line in between;
- on removal, one adjacent blank line, so no doubled blank line is left behind.

`hclwrite` attaches only `#` and `//` comments to a block; the editor also removes a `/* */`
comment directly above a removed block.

Every byte outside the spans of the edited declarations is unchanged. A comment separated from a
block by a blank line belongs to no declaration and is never touched. For a rename, each
declaration that references the renamed element is also an edited declaration (R4).

**Alternatives considered**:
- *Regenerate files from the compiled model*: compiles, but destroys comments and layout, which
  FR-013 forbids.
- *Raw byte splicing using parser ranges*: equivalent safety, but it re-implements what `hclwrite`
  already does losslessly, including nested-block placement and attribute insertion.
- *Reformat the whole edited declaration*: tidier, but it silently rewrites a person's deliberate
  spacing inside a block they did not ask to change.

---

## R2. Compiling before writing

**Decision**: The editor produces an **edit plan**, which is the new contents of each changed file,
held in memory. The compiler loads the project with those files overlaid on the disk contents (a
new `LoadOverlay` on the source adapter; the parser's file read becomes a lookup that checks the
overlay first). Nothing is written unless the overlaid compile has no errors (FR-012). A batch is
applied in order to the in-memory files and compiled once at the end (FR-012a). Preview returns the
plan's per-file diffs without committing (FR-012b). Commit and preview share the same plan, so
applying after a preview produces exactly the previewed change.

**Alternatives considered**: Writing to a temporary copy of the project and compiling that copy is
slower, needs cleanup, and can leak half-written trees.

---

## R3. Detecting a stale edit (FR-016)

**Decision**: Every read and every write result carries a **revision**: a short, opaque token
(`r1-` plus 16 hex digits) digesting the SHA-256 of each project-relative source file. The server
remembers the per-file hashes behind the last 64 tokens it issued. A token it does not remember
(after a restart, say) is accepted when it equals the current revision's token, since nothing has
changed; otherwise it is refused as `stale_revision`. (Encoding every hash into the token made it grow with the project:
over a thousand tokens per response on a 50-file project.) Write tools require a `base_revision`.
Before committing, the editor recomputes the hashes of the files the plan would change and refuses
if any differs from the base (`stale_revision`). Files the edit does not touch may change freely in
between. The check runs again immediately before the rename into place, which keeps the window for
a racing editor save as small as possible.

**Serialisation (FR-017)**: one authoring service per project holds a mutex around plan, compile and
commit, so concurrent tool calls apply one at a time, each against the previous one's result.

**Atomic commit**: each changed file is written to a sibling temporary file and renamed into place,
with permissions preserved. If a multi-file commit fails part-way, files already renamed are
restored from the in-memory originals, so a batch is never half-applied (FR-012a). The formatter
writes in place to preserve hard links; the editor accepts the rename trade-off, because
all-or-nothing matters more for multi-file writes.

---

## R4. Renaming and the `moved` block (FR-022..FR-025)

**Decision**: The language gains one top-level block:

```hcl
moved {
  from = container.api
  to   = component.api
}
```

- `from` and `to` are static traversals of element addresses (`kind.name`), extracted the same way
  as other references and never evaluated.
- `from` must **not** be declared (it no longer exists), `to` **must** resolve to a declared
  element, and no address may be the `from` of two blocks. The kind may change (clarification 5).
  Parent rules are enforced by normal validation of the renamed element.
- New error codes: `moved_from_declared`, `moved_to_unresolved`, `moved_duplicate_from`.
- The IR gains `Moves []Move{From, To, Range}`, serialised with `omitempty`, so every export golden
  from features 013 and 014 stays byte-identical and `SchemaVersion` stays 1. Adding an optional
  field is a compatible change. Rendering ignores moves; the stage-4 diff engine consumes them.

A rename edit:
1. rewrites the declaring block's labels (`container "api"` → `component "api"`);
2. calls `Expression.RenameVariablePrefix(["container","api"], ["component","api"])` on every
   attribute expression in every source file. That covers parents, relationship targets, instance
   `of`, view include and exclude lists, and earlier `moved.to` values, and only the traversal
   tokens inside each expression change. The byte oracle checks this with a line-level diff
   restricted to those tokens;
3. appends a `moved` block to the file that declared the element.

Renaming onto an address already in use is refused, naming the existing declaration (FR-025).

Found by the property test and now part of the design:
- A kind change drops the parent attribute the new kind does not take; a batch sets the new one.
- Renaming back to an address the element once had removes the `moved` block that recorded the
  move away from it, which would otherwise claim a declared address.
- Removing an element removes the `moved` blocks leading to it, along chains: a removed element
  has no address for its history to point at.

**Alternatives considered**:
- *Record renames outside the source* (a sidecar file): the compiler is stateless and the source is
  the only artifact (ADR-0012).
- *String replacement of the address text*: would rewrite comments and strings that merely mention
  the name.

---

## R5. Query semantics (FR-003..FR-005)

**Decision**: Queries run on the IR's relationships with **descendant inclusion**: an element
stands for itself and everything it contains, and results are reported at the address of the
element on the other side of each relationship.

| Query | Result |
|---|---|
| `dependents(T)` | every element that has a relationship whose target is T or a descendant of T |
| `dependencies(T)` | every element targeted by a relationship whose source is T or a descendant of T |
| `transitive` | breadth-first closure of either, with a visited set. Cycles terminate, and each element is reported once with its distance |
| `path(A, B)` | shortest directed chain of relationships from A, or a descendant of A, to B, or a descendant of B. Breadth-first search, ties broken by relationship address, so the answer is deterministic. If there is none, `found: false` |
| `orphans` | elements with no relationship touching them or any descendant: the same connectivity rule as the `orphan_element` warning |
| `coupling` | per element, fan-in (distinct elements depending on it) and fan-out (distinct elements it depends on), ranked by fan-in + fan-out descending, then by address |

Self-relationships count as neither fan-in nor fan-out. All results are ordered slices (FR-009).

**Rationale**: "What breaks if the orders database goes down" has to count a component inside
another container that calls the database. Counting descendants is the C4-consistent reading; it
is the same lifting rule the views use.

---

## R6. `describe` levels (FR-001, FR-002, SC-008)

| Level | Content | Budget on 1,000 elements |
|---|---|---|
| `summary` | project name and description, counts by kind, environment names, top-level element names | ≤ 300 tokens |
| `structure` | the containment tree down to containers (components counted, not listed), with kinds and technologies | ≤ 2,000 tokens |
| `full` | every element with every attribute, every relationship, and every environment's placement | unbounded; intended with a scope |

`address` scopes any level to one element: the element, its children (one level, or the whole
subtree at `full`), and its incoming and outgoing relationships. Every result carries the revision
(R3), so the assistant can base edits on it. Token counts are asserted in a test with a
TOON-encoded 1,000-element fixture.

---

## R7. The MCP tool surface

**Decision**: five tools, exactly as the seed names them.

| Tool | Purpose |
|---|---|
| `describe` | FR-001, FR-002 |
| `query` | `kind` is one of dependents, dependencies, path, orphans, coupling (FR-003..FR-005) |
| `validate` | compile and return diagnostics plus the revision (FR-006) |
| `apply_edit` | `edits[]` (one or more), `preview`, `base_revision` (FR-010..FR-021) |
| `move` | a convenience form of a single rename edit, with `preview` and `base_revision` (FR-022..FR-025) |

Reads take `format: "toon" | "json"`, defaulting to TOON (FR-008, ADR-0011). Refusals (compile
errors, stale revision, dangling references) are **tool results** with `ok: false` and the
diagnostics, not JSON-RPC errors, so the assistant reads and acts on them. JSON-RPC errors are kept
for malformed requests.

Two server fixes come with this:
- `tools/list` is returned sorted by name; it currently iterates a map, so the order is random.
- Tool calls receive the request context instead of `context.Background()`.

Handlers decode arguments with shared helpers in `tools/helpers.go`, which is exempt from function
budgets as a data and helper file. Each handler is ≤ 30 effective lines (Principle III).

---

## R8. Where a new declaration goes

**Decision**: An explicit `file` (project-relative, ending `.loko.hcl`, inside the root) wins; it
may name a new file, which is created. Otherwise:
- containers and components go in the file declaring their parent;
- relationships go inside their source element's block;
- instances and groups go inside their environment or group;
- bindings go inside their instance;
- top-level elements and environments go in the file declaring `project`.

Any other path, a non-source extension, or a path escaping the root is refused (FR-020, FR-026).

---

## R9. Preview diffs

**Decision**: A line-based unified diff (Myers algorithm, 3 lines of context) of each changed file,
implemented in the source adapter (`edit_diff.go`) with no new dependency. Diffs are deterministic,
and files are sorted by path.

**Alternative rejected**: shipping whole new file contents. That costs too many tokens for an
assistant and is harder for a person to review.

---

## R10. The round-trip property test (SC-002, SC-003)

**Decision**: a seeded generator (`math/rand/v2`, fixed seeds in CI, a random seed reported on
failure) produces sequences of valid edits against hand-written fixtures full of comments and
irregular spacing. For every edit, two oracles are checked:

1. **IR oracle**: an independent model applies the edit to the previous IR. Compiling the edited
   source must equal that model, compared after stripping source ranges and moved-block records
   (moves are checked separately). This catches an edit that compiles but means something else.
2. **Byte oracle**: every byte of every file outside the edited declarations, including untouched
   files, is identical to before.

The test runs 2,000 sequences of up to 12 edits on every `go test` run (a few seconds), plus a
`FuzzApplyEdits` fuzz target with the same seeds as its corpus for longer `go test -fuzz` runs. It
lives in the source adapter's test package, because it needs real HCL.

---

## R11. `loko query` (FR-028..FR-031)

**Decision**:
- `loko query dependents|dependencies <address> [--transitive]`
- `loko query path <from> <to>`
- `loko query orphans`
- `loko query coupling [--limit N]`

Every form accepts `--format text|json|toon`. Each subcommand calls the same `usecases.Query`
function as the MCP tool. A test compares the CLI and MCP outputs byte for byte in JSON for every
query kind (SC-006). An unknown address reuses the compiler's closest-match suggester
(`resolve_suggest.go`). Exit codes are the existing three.

---

## R12. Enforcing "HCL and nothing else" (FR-026, FR-026a, SC-005)

**Decision**: it is enforced in three layers.

1. **Static**: archcheck already bans the render adapters from `internal/mcp` (feature 014). The
   only adapters the MCP layer may write through are `hclsource` (the editor) and `encoding`.
2. **Runtime**: the editor's commit refuses any path that is not a `*.loko.hcl` file inside the
   root.
3. **Test**: a tool-surface test drives every tool against a fixture and asserts that the only
   files created or modified are `*.loko.hcl`.

---

## R13. Performance (SC-007)

Compiling 1,000 elements takes milliseconds (feature 013 measured 2 ms). An edit re-parses only the
changed files with `hclwrite`, then compiles once through the overlay. Budget: read under 1 s, edit
including compile under 2 s, asserted in a performance test on the 1,000-element generator from
features 013 and 014.
