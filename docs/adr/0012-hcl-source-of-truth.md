# ADR-0012: HCL as the single authored source of truth

**Status:** Accepted
**Date:** 2026-09-23
**Feature:** [013-hcl-compiler-core](../../specs/013-hcl-compiler-core/spec.md)
**Supersedes:** the v0.2 file-tree data model (markdown frontmatter + D2 arrows, union-merged)
**Related:** [ADR-0004](0004-graph-conventions.md) (graph-qualified IDs)

## Context

v0.2 had two sources of truth. Relationships were authored in markdown frontmatter *and* in D2
arrow syntax, then union-merged at build time. `loko validate --check-drift` existed to report when
the two disagreed.

Drift detection is not a feature. It is the symptom of a model with no authoritative layer. Because
nothing was authoritative, three things followed: edits went stale silently, there was no
architecture-level diff because no comparable compiled value existed, and there was nothing
well-defined to reconcile a deployed environment against.

## Decision

HCL becomes layer 0 — the sole authored artefact. Every rendered artefact is a pure function of a
compiled, immutable IR. The v0 model is deleted rather than kept alongside.

Four decisions inside that shape the implementation, each of which had a plausible alternative:

### 1. References are static traversals, never evaluated

`container.orders_db` is extracted with `hcl.AbsTraversalForExpr` as a **static traversal**, rendered
to a dotted string plus a source range, and resolved in the core layer against a symbol table. It is
never passed through expression evaluation and never becomes a `cty` value.

Terraform evaluates references, and that was the obvious thing to copy. Three things follow from not
doing so:

- **Resolution stays in core.** Evaluating references would require the symbol table pre-loaded into
  an `EvalContext`, putting reference resolution — and the unresolvable-reference diagnostic — in the
  adapter, where the constitution forbids business logic and where it would be far harder to test.
- **Declaration order stops mattering for free.** The first pass collects block headers; the second
  is a map lookup. There is no partially-populated evaluation context to sequence.
- **The diff stage's `moved` block stops being a special case.** Its operands deliberately name
  addresses that no longer exist, so they *must not* be evaluated. Under this decision that block is
  an ordinary use of the normal mechanism rather than a carve-out in the evaluator.

The cost is that a reference cannot be computed. Since iteration constructs are excluded from v1.0,
nothing in the language can produce one anyway.

### 2. The parser emits an unresolved `SourceModel`; core builds the IR

Two models, not one. The adapter owns syntax — discovery, parsing, scalar evaluation, unknown
constructs. Core owns the symbol table, resolution, and every validation rule.

The alternative, having the adapter return the finished IR, needs fewer types but puts validation in
the adapter, violating Principles I and IV. The split also means the entire rule suite is testable
from struct literals with no files on disk, which keeps a failing test pointing at one layer.

### 3. The IR holds ordered slices, never maps

Ordering is assigned once, at IR construction, by sorting on the address byte-wise. Encoders
serialise in slice order and sort nothing.

Go randomises map iteration, so any map reaching an encoder is a latent nondeterminism bug that
passes most of the time. `encoding/json` sorts map keys, which is what makes this easy to get wrong:
JSON output would look stable while a positional encoder would not. Determinism is structural rather
than a discipline each backend must remember — and it is the precondition for the golden-file suite.

### 4. The v0 model is deleted, not kept behind a flag

Keeping both models until the renderer stage would have kept every command working throughout.
It was rejected: that *is* the two-sources-of-truth state this release exists to remove, Principle
VII forbids compatibility shims, and it would double the surface the renderer stage then has to
unwind.

## Consequences

### Accepted costs

- **`build`, `serve`, `watch` and `init` stop working** until the renderer stage. Deleting
  `ProjectRepository` reached 44 non-test files; there is no version of this change where those
  commands survive. Not a shortcut — a consequence.
- **No automatic migration.** Writing a D2 arrow parser, a frontmatter reader, TOML handling, and
  v0's union-merge conflict semantics is not justified by the install base, and the D2 half
  duplicates the v1.2 diagram-import adapter. The `v0.2.x` tag stays installable and the migration
  recipe is documented: point an agent at the v0 tree over MCP and let it write HCL.
- **The MCP server registers no tools.** The harness keeps running and answers `tools/list` with
  `[]`, so an editor's configuration survives the gap.

### Gains

- Production code fell from ~21.9k to 8.4k LOC; 35,317 lines were deleted.
- Six direct dependencies were dropped by `go mod tidy`: viper, fsnotify, go-toml, testify,
  mapstructure, and d2. TOML handling left the module entirely.
- `archcheck` reports **0 violations**, including with `--no-suppress`. The 24 pre-existing
  violations left with the code that caused them, so `.archcheck-suppressions.yaml` — whose six
  entries had all expired — was deleted rather than renewed.
- A broken relationship is now a compile error with a file, line, and column.

### Enforcement

A new layer rule confines `github.com/hashicorp/hcl/**`, `github.com/zclconf/go-cty/**`, and
`oss.terrastruct.com/d2/**` to `internal/adapters/**` (FR-044). Making it work needed two fixes to
existing tooling, both of which had been silently ineffective:

- `archcheck` skipped every import not under the module path — *"External imports are always
  allowed"* — so no layer rule could mention a third-party package at all. `LayerRule` gained
  `ForbiddenExternalImports`.
- The `depguard` `files` globs match absolute paths, so the bare `cmd/**/*.go` pattern never matched
  and the pre-existing `cmd-no-direct-entities` rule had never fired on anything. Every pattern now
  carries a `**/` prefix.

Both are proven live: planting an `entities → hcl` import makes archcheck *and* depguard fail, and
removing it makes both pass.

## Parked, not deleted

`internal/_parked/` holds code retained for later stages (FR-045). Go ignores directories beginning
with `_`, so it stays in the tree without compiling. Each entry was written against the deleted model
and must be reworked, not restored: the D2 parser for v1.2 diagram import, the HTML site builder for
the renderer stage, and the MCP tools for the authoring stage.

## Alternatives rejected

| Alternative | Why not |
|---|---|
| A purpose-built grammar with a hand-written parser | Source ranges, error recovery, and a formatter are each substantial work HCL already does. `hclwrite`'s token-preserving edits are what make machine authoring of the source of truth non-destructive, which the authoring stage depends on. |
| YAML with a comment-preserving library | Comment preservation is partial and fragile; anchors and block scalars add surface the model does not need. |
| CUE | Better typed, but a heavy dependency, a steep learning curve for the audience, and no surgical-edit story. |
| A committed IR snapshot (`loko.lock.json`) | Buys rename detection and git-free diff, costs a second artefact that goes stale, merge conflicts on every branch, and a regeneration ritual. `moved` blocks deliver the same benefit without the state file. |
| A persisted graph database with HCL as one editing surface | The moment HCL is *a* surface rather than *the* surface there are two sources of truth again — precisely the defect being removed. |
| `hashicorp/go-version` for the `loko_version` check | The check is one of FR-028's errors and so belongs in the validation layer, and core takes no dependencies. A port, adapter, mock and wiring for arithmetic on three integers fails Principle VII; the evaluator is ~120 lines of stdlib. |
