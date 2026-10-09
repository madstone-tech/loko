# Data Model: MCP HCL Authoring

**Feature**: `015-mcp-hcl-authoring` | **Plan**: [plan.md](plan.md) | **Research**: [research.md](research.md)

There are two sets of additions. The **language and IR** gain the `moved` block (R4). A new
stdlib-only entity package, `internal/core/entities/authoring`, holds edits, batches, revisions and
plans. It imports nothing, including `arch`: one entity package may not import another (archcheck,
established in feature 014), so addresses are plain strings. Results that carry diagnostics
(`EditResult`, query, describe and validate results) are use-case DTOs in `internal/core/usecases`.
Outer layers may not import entities, so MCP and CLI handlers pass plain DTOs (`usecases.EditInput`)
and the use case constructs and validates the entities. Every collection is an ordered slice with a documented sort key,
never a map (FR-009).

---

## 1. Language and IR additions (`internal/core/entities/arch`)

| Type | Fields | Notes |
|---|---|---|
| `MovedDecl` (SourceModel) | `From Reference`, `To Reference`, `Range` | top-level `moved {}`; no label |
| `Move` (IR) | `From Address`, `To Address`, `Range` | `IR.Moves []Move`, sorted by `From`, with JSON and TOON tag `moves,omitempty`, so existing exports are byte-identical |

New diagnostic codes, all errors, each with a fixture:

| Code | Raised when |
|---|---|
| `moved_from_declared` | `from` names an element that is still declared |
| `moved_to_unresolved` | `to` does not resolve to a declared element |
| `moved_duplicate_from` | two `moved` blocks share a `from` (the second is reported, with the first as a related range) |

---

## 2. Edits (`internal/core/entities/authoring`)

### 2.1 `Edit`

Handlers pass the same fields as `usecases.EditInput`, a plain struct with JSON tags matching the
MCP contract. `usecases.Apply` converts each input with `authoring.NewEdit`; a validation failure
becomes a refusal with `reason: invalid_edit` and the batch index.

| Field | Type | Rules |
|---|---|---|
| `Op` | `Op` | `add` \| `update` \| `remove` \| `rename` |
| `Target` | `TargetKind` | `element` \| `relationship` \| `environment` \| `group` \| `instance` \| `binding` \| `view` |
| `Address` | string | The target's address. For an add, the address to create (`container.api`, `container.api.uses.orders`, `deployment.prod.node.vpc.subnet-a`, `deployment.prod.instance.api`). A binding is addressed as its instance plus `Binding` |
| `Binding` | `BindingRef` | For binding targets: `Kind` (`terraform` \| `cloudformation`) and `Index` (0-based among same-kind bindings; required when there is more than one) |
| `Set` | `[]Attr` | Attributes to set (add and update), sorted by name. Each is `{Name, Value}` where `Value` is a string, a list of strings, a reference (address string), or a key/value list |
| `Clear` | `[]string` | Attributes to remove (update only) |
| `Cascade` | bool | Remove only (FR-018a) |
| `To` | string | Rename only: the new element address. The kind and/or the name may change |
| `File` | string | Add only: the target file (R8) |

**Constructor validation** (`NewEdit`, Principle IV), each failure an error naming the field:
- Op/target combinations: rename is element-only; cascade is remove-only; `To` is rename-only;
  `File` is add-only.
- The address has the right form for the target kind (`arch.ValidName` on each segment).
- Attribute names are legal for the kind. The legal-attribute table mirrors the language schema:

| Target | Legal attributes |
|---|---|
| element | `description`, `owner`, `technology`, `tags`, `docs`, plus `system` (container) or `container` (component) |
| relationship | `target` (required on add), `description`, `technology` |
| environment | `provider`, `account`, `region` |
| group | (none) |
| instance | `of` (required on add), `attributes` |
| binding | exactly one of `address`, `addresses`, `tags` |
| view | `include`, `exclude` (lists of element references, written unquoted), `tags` (strings) |

- `File`, when present, is a clean relative path ending `.loko.hcl`.
- An update needs at least one `Set` or `Clear` (a no-op is detected after planning, FR-021).

### 2.2 `Batch`

| Field | Type | Rules |
|---|---|---|
| `Edits` | `[]Edit` | 1..100, applied in order (FR-012a) |
| `Preview` | bool | FR-012b |
| `BaseRevision` | `Revision` | required (FR-016) |

### 2.3 `Revision`

An opaque, deterministic encoding of `[]FileHash{Path, SHA256}`, sorted by path.
`Revision.Hash(path)` returns a file's base hash. `ParseRevision` rejects malformed tokens.

### 2.4 `EditResult` (use-case DTO, `internal/core/usecases`)

| Field | Type | Notes |
|---|---|---|
| `OK` | bool | false means nothing was written |
| `Preview` | bool | |
| `NoOp` | bool | FR-021 |
| `Files` | `[]FileChange{Path, Diff}` | sorted by path; `Diff` is a unified diff (R9) |
| `Removed` | `[]string` | addresses removed by a cascade, sorted (FR-018a) |
| `Refusal` | `*Refusal{Reason, Edit int, Detail, Dependents []string}` | `Reason` is one of `compile_errors`, `stale_revision`, `dangling_references`, `address_in_use`, `not_found`, `invalid_edit`, `path_refused`. `Edit` is the index of the failing edit in the batch |
| `Diags` | `arch.Diagnostics` | compile errors on refusal; warnings on success (FR-019) |
| `Revision` | `Revision` | the revision after the write, or the current one on refusal |

---

## 3. Query and describe DTOs (`internal/core/usecases`)

| Type | Fields |
|---|---|
| `QueryRequest` | `Kind` (dependents \| dependencies \| path \| orphans \| coupling), `Address`, `To`, `Transitive`, `Limit` |
| `QueryResult` | `Kind`, `Elements []ElementHit{Address, Name, Kind, Distance}` (dependents and dependencies, sorted by distance then address), `Path []PathStep{From, To, Relationship}` + `Found` (path), `Coupling []CouplingRow{Address, FanIn, FanOut}` (coupling), `Revision` |
| `DescribeRequest` | `Level` (summary \| structure \| full), `Address` (optional) |
| `DescribeResult` | `Project`, `Counts []KindCount`, `Elements []ElementView{Address, Kind, Name, Technology, Description?, Parent, Children}`, `Relationships []RelationshipView`, `Environments []EnvironmentView`, `Revision`. Fields are populated per level (R6) |
| `ValidateResult` | `Diags`, `ErrorCount`, `WarningCount`, `Revision` |

Every DTO carries JSON and TOON tags and is encoded by the existing `OutputEncoder`.

---

## 4. Ports (`internal/core/usecases/ports.go`)

See [contracts/ports.md](contracts/ports.md). There are two new interfaces:
- `OverlaySource`: compile with in-memory file contents.
- `SourceEditor`: plan edits into in-memory file contents, render diffs, hash files, commit
  atomically.

The `AuthoringService` use case holds the per-project mutex (R3).

---

## 5. State: one write

```
request ──validate edits──▶ plan (in memory, batch in order)
   │ invalid ▶ refused(invalid_edit | not_found | address_in_use | dangling_references)
   ▼
compile with overlay ──errors──▶ refused(compile_errors) + diagnostics
   ▼
no change? ──▶ ok, noOp
   ▼
preview? ──▶ ok, preview, diffs (nothing written)
   ▼
base files changed? ──▶ refused(stale_revision)
   ▼
commit (temp + rename, rollback on failure) ──▶ ok, files, warnings, new revision
```

Every refusal leaves every file on disk byte-identical (SC-004).
