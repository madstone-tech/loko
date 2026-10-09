# Contract: the `moved` block (language addition)

**Feature**: `015-mcp-hcl-authoring` | **Status**: normative. Like the rest of the language, this
is permanent v1.x surface.

```hcl
moved {
  from = container.api
  to   = component.api
}
```

| Rule | Diagnostic |
|---|---|
| The block is top-level, has no label, and both attributes are required | `unknown_block` / `unknown_attribute` / missing-attribute errors, as for other blocks |
| `from` and `to` are bare element-address traversals (`kind.name`), never strings or computed expressions | `wrong_reference_kind` |
| `from` must not name a declared element | `moved_from_declared` |
| `to` must name a declared element; its kind may differ from `from`'s | `moved_to_unresolved` |
| At most one `moved` block per `from` | `moved_duplicate_from` (related range: the first) |

- Moves are carried into the IR as `moves` (sorted by `from`, omitted when empty). They do not
  affect views, rendering or queries; the stage-4 diff engine consumes them.
- `loko fmt` formats `moved` blocks like any other block.
- Chained renames (`a → b`, later `b → c`) are valid: each block is checked against the current
  declarations.
- Renaming back to a former address removes the block that recorded the move away from it;
  removing an element removes the blocks that lead to it.
- `docs/language.md` gains a `moved` section with this table.
