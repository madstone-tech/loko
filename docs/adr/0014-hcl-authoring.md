# ADR-0014: Assistants edit HCL surgically, and nothing else

**Status:** Accepted
**Date:** 2026-10-09
**Feature:** [015-mcp-hcl-authoring](../../specs/015-mcp-hcl-authoring/spec.md)
**Related:** [ADR-0011](0011-toon-mcp-default.md) (TOON by default),
[ADR-0012](0012-hcl-source-of-truth.md) (HCL as the single authored source),
[ADR-0013](0013-viewmodel-renderers.md) (outputs are projections)

## Context

ADR-0012 made `*.loko.hcl` the only authored artifact, and feature 013 deleted the v0 MCP tools,
which scaffolded the old file tree. The server kept running with no tools. Giving assistants
tools again raised three risks:

- **Corruption.** An assistant editing a person's source is the failure that would sink the tool:
  a lost comment, a reformatted block, or a half-applied change.
- **Two sources of truth.** A tool that can write a diagram or a page recreates the defect
  ADR-0012 removed.
- **Lost history.** Renaming an element looks, to anything comparing versions, like a deletion and
  an addition.

## Decision

1. **Five tools on the compiled IR.** `describe`, `query` and `validate` read; `apply_edit` and
   `move` write. The v0 tools were not restored: they wrote a model that no longer exists, and
   nothing in them carried over. `loko query` calls the same query use case as the MCP tool, and
   a test holds their JSON byte-identical.

2. **Surgical edits through `hclwrite`, serialised from the token stream.** Files are parsed with
   `hclwrite`, changed through its tree, and written from the raw tokens. `File.Bytes()` is never
   used, because it runs the formatter over the whole file and would realign lines nobody asked to
   change. New code is formatted on its own, at its nesting depth, and spliced in, so it is
   canonical. Every byte outside the edited declaration is unchanged. A declaration includes the
   comments attached directly above it, of any style.

3. **Plan, compile in memory, then commit.** A batch of edits becomes the new contents of the files
   it touches. The compiler loads them through an overlay, and only an error-free result is
   written. Writing stages every file as a temporary sibling, then renames them into place,
   restoring the originals if a rename fails. Refusals are tool results (`ok: false` with a
   reason), and a refused write leaves every file byte-identical.

4. **Revisions.** Every read returns a short revision token; every write quotes one. The server
   remembers the per-file hashes behind the last 64 tokens. A write is refused as `stale_revision`
   if any file it would change differs from that revision. A token the server does not remember,
   for example after a restart, is still accepted if it matches the files as they are now;
   otherwise it is refused. Other files may change freely.

5. **HCL and nothing else.** Writes are limited to `*.loko.hcl` files inside the project root, by
   the edit validation, the editor and the commit. Setting `docs` records a path and never creates
   the file. A test drives every tool and checks that nothing else on disk changed.

6. **The `moved` block.** A rename rewrites the declaring label and every reference, changing only
   the traversal tokens, then appends `moved { from = …, to = … }`. This is a permanent addition to
   the language. The IR carries moves as an `omitempty` field, so projects without renames export
   exactly as before. Renaming back drops the move it undoes; removing an element drops the
   history that leads to it.

7. **Descendant-inclusive queries.** An element stands for itself and everything inside it, the
   lifting rule views already use, so the dependents of a database include a component elsewhere
   that calls it.

Declared views are an edit target too: they are how a design is explained, so an assistant that
could build an architecture but not present it would be half a tool. This was added after testing
on a real architecture showed the gap.

## Consequences

- A seeded property test (2,000 sequences of up to 12 edits, renames included) and a fuzz target
  check that the IR matches an independent model after every edit, and that bytes outside each
  edited span are unchanged. It found three editor bugs during development.
- `hclwrite` can only append, so a new attribute goes at the end of its block, after any nested
  blocks. It is valid HCL; `loko fmt` does not reorder it.
- After an update, the edited block is realigned as the formatter would (aligned `=`, comments kept).
  Blocks the edit did not touch keep their hand-made layout.
- A kind change drops the parent attribute the new kind does not take; when the new kind needs a
  different parent, the caller sets it in the same batch.
- Revision memory lives only in the server process. After a restart, a token still matches if
  nothing has changed since it was issued; otherwise the client reads again.
- `structure`-level descriptions omit `kind` and `name`, which the address already carries, to stay
  within their token budget.

## Alternatives rejected

- **Regenerate files from the IR.** Simple and always canonical, but it destroys comments and
  layout.
- **Byte splicing using parser ranges.** Equally precise, but it re-implements block placement and
  attribute insertion that `hclwrite` already does losslessly.
- **A sidecar file of renames.** The compiler is stateless and the source is the only artifact
  (ADR-0012).
- **Self-contained revision tokens.** Encoding every file's hash made the token grow with the
  project, adding over a thousand tokens to every response on a 50-file project.
