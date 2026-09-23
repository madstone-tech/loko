# Phase 1 Data Model: HCL Compiler Core

**Feature**: `013-hcl-compiler-core` | **Date**: 2026-09-22
**Spec**: [spec.md](./spec.md) | **Research**: [research.md](./research.md)

All types below live in `internal/core/entities/` and depend on the standard library only
(Constitution, Principle I). Field types are shown in Go for precision; the constructor and
validation rules are the normative part.

Two distinct models exist, and keeping them apart is the central structural decision (research R7):

- **`SourceModel`** — flat, unresolved declarations as written, produced by the parser adapter.
  References are still raw dotted strings. Nothing is checked beyond syntax.
- **`IR`** — the resolved, ordered, immutable compiled value. Every reference is an address that is
  known to exist. This is what `export` writes and what every later stage consumes.

---

## 1. Foundational types

### `Address`

The stable identity of everything (FR-024).

```go
type Address string
```

Canonical forms, fixed by this release:

| Subject | Form | Example |
|---|---|---|
| Logical element | `<kind>.<name>` | `container.api` |
| Relationship | `<kind>.<name>.uses.<local>` | `container.api.uses.orders` |
| Environment | `deployment.<name>` | `deployment.prod` |
| Placement group | `deployment.<env>.node.<path…>` | `deployment.prod.node.vpc-main.subnet-a` |
| Instance | `deployment.<env>.instance.<name>` | `deployment.prod.instance.api` |
| View | `view.<name>` | `view.payment-path` |

**Instance addresses omit the placement path** (clarification 2026-09-22, FR-012, FR-024). A
placement group has its own nested address; the instances inside it do not inherit a segment. Moving
an instance between groups therefore preserves its identity.

Rules:

- Name segments match `^[A-Za-z_][A-Za-z0-9_-]*$`. A label failing this is a syntax-level error.
- Comparison and sorting are byte-wise on the string (research R6). No locale collation.
- `Address` is opaque to consumers: it is built by constructors, never concatenated by callers.

### `SourceRange`

```go
type SourceRange struct {
    File        string // project-relative, forward slashes on every platform
    StartLine   int    // 1-based
    StartColumn int    // 1-based
    StartByte   int    // 0-based
    EndLine     int
    EndColumn   int
    EndByte     int
}
```

A faithful copy of the parser's range type, declared locally so that core never imports the parsing
library (research R8, FR-044). Paths are normalised to project-relative with forward slashes so that
exported diagnostics and golden fixtures are identical on every platform (FR-036c).

### `Diagnostic`

```go
type Severity int // SeverityError, SeverityWarning

type Diagnostic struct {
    Severity Severity
    Code     string        // stable machine identifier, e.g. "unresolved_reference"
    Summary  string        // one line
    Detail   string        // optional elaboration
    Range    SourceRange
    Related  []SourceRange // e.g. the other declaration in a duplicate
}
```

Satisfies FR-030 and FR-030a. `Code` is what machine consumers match on; `Summary` and `Detail` are
free to be reworded without breaking them. `Related` is what lets a duplicate-declaration error name
both sites (FR-004, FR-012a).

`Diagnostics []Diagnostic` carries `HasErrors()`, `SortedForOutput()` (by file, then start byte, then
code — FR-033), and `ExitCode(strict bool)` implementing FR-038.

---

## 2. `SourceModel` — parser output, unresolved

```go
type SourceModel struct {
    Project      ProjectDecl
    Elements     []ElementDecl
    Environments []EnvironmentDecl
    Views        []ViewDecl
    Ignores      []IgnorePattern  // FR-015
    Files        []string         // every discovered source file, sorted
}
```

Constructed only by the adapter. Carries no resolved pointers — an unresolvable reference is still
representable here, because detecting it is core's job (FR-021, research R2/R7).

### `Reference`

```go
type Reference struct {
    Raw   string      // "container.orders_db" exactly as traversed
    Range SourceRange
}
```

The single representation of every cross-element pointer before resolution.

### `ElementDecl`

```go
type ElementDecl struct {
    Kind        ElementKind // Person, System, Container, Component, External
    Name        string
    Parent      *Reference  // container→system, component→container; nil otherwise
    Description string
    Owner       string
    Technology  string
    Tags        []string
    Docs        string      // path to a prose file, project-relative
    Relations   []RelationDecl
    Range       SourceRange
    BodyRanges  map[string]SourceRange // per-attribute, for precise diagnostics
}

type RelationDecl struct {
    LocalName   string
    Target      Reference
    Description string
    Technology  string
    Range       SourceRange
}
```

### `EnvironmentDecl`

```go
type EnvironmentDecl struct {
    Name     string
    Provider string
    Account  string
    Region   string
    Groups   []GroupDecl    // nested placement groups, arbitrary depth (FR-012)
    Instances []InstanceDecl // instances declared directly under the environment
    Range    SourceRange
}

type GroupDecl struct {
    Name      string
    Groups    []GroupDecl
    Instances []InstanceDecl
    Range     SourceRange
}

type InstanceDecl struct {
    Name       string
    Of         Reference          // the logical element realised (FR-013)
    Attributes map[string]Value   // environment-specific, order-independent
    Claims     []ClaimDecl
    Range      SourceRange
}

type ClaimDecl struct {
    Kind      string   // e.g. "terraform", "cloudformation"
    Address   string   // exact identifier or glob
    Addresses []string
    Tags      map[string]string
    Range     SourceRange
}
```

`Value` is a small closed sum over string, number, bool, list, and map — enough for the attribute
maps of FR-013 without importing the evaluator's value type into core.

### `ProjectDecl`, `ViewDecl`, `IgnorePattern`

```go
type ProjectDecl struct {
    Name        string
    Description string
    Version     string // the raw constraint, e.g. "~> 1.0" (FR-003)
    Range       SourceRange
}

type ViewDecl struct {
    Name    string
    Include []Reference
    Exclude []Reference
    Tags    []string
    Range   SourceRange
}

type IgnorePattern struct {
    Pattern string
    Range   SourceRange
}
```

Views are validated here and rendered in a later stage (FR-016). Ignore patterns are carried through
untouched for the reconciliation stage (FR-015).

---

## 3. `IR` — the compiled value

```go
type IR struct {
    SchemaVersion int            // FR-036a; 1 in this release
    Project       Project
    Elements      []Element      // sorted by Address
    Relationships []Relationship // sorted by Address
    Environments  []Environment  // sorted by Address
    Views         []View         // sorted by Address
    Ignores       []string       // sorted
}
```

**Slices, never maps, in every exported field** (research R6, FR-040). Lookup maps may exist as
unexported indexes built at construction, but they never reach an encoder.

```go
type Element struct {
    Address     Address
    Kind        ElementKind
    Name        string
    Parent      Address // "" for person, system, external
    Description string
    Owner       string
    Technology  string
    Tags        []string // sorted, de-duplicated
    Docs        string
    Range       SourceRange
}

type Relationship struct {
    Address     Address // container.api.uses.orders
    Source      Address // resolved
    Target      Address // resolved
    LocalName   string
    Description string
    Technology  string
    Range       SourceRange
}

type Environment struct {
    Address   Address
    Name      string
    Provider  string
    Account   string
    Region    string
    Groups    []Group    // nested, each sorted by Address
    Instances []Instance // sorted by Address, flat — see below
}

type Group struct {
    Address  Address
    Name     string
    Groups   []Group
    Contains []Address // instances placed in this group
}

type Instance struct {
    Address    Address // deployment.prod.instance.api — no group segment
    Name       string
    Of         Address // resolved logical element
    PlacedIn   Address // the group holding it; "" if directly under the environment
    Attributes []Attribute // sorted by Key
    Claims     []Claim     // sorted by Kind then Address
}
```

`Environment.Instances` is **flat and complete**; `Group.Contains` and `Instance.PlacedIn` record
placement as a cross-reference. This is what makes the clarified addressing decision expressible: an
instance is reachable without walking the group tree, so its identity is independent of where it
sits, while the nesting is still fully preserved for the renderer and reconciler.

`IR` exposes `Lookup(Address) (any, bool)`, `ElementsByKind(ElementKind) []Element`, and
`OutgoingFrom(Address) []Relationship` over unexported indexes. It has no mutating methods (FR-025).

---

## 4. Validation rules

Every rule is implemented in `internal/core/`, produces a `Diagnostic` with a range, and has at least
one fixture (SC-004).

### Errors (FR-028)

| Code | Rule | Source |
|---|---|---|
| `no_source_found` | No architecture source beneath the root | FR-005 |
| `syntax_error` | A source file does not parse. Distinct from `unknown_block`/`unknown_attribute`, which apply to files that parse but declare something undefined. Parsing continues on the other files so one run still reports everything (FR-031) | FR-028a |
| `duplicate_declaration` | Same address declared twice; `Related` names both | FR-004 |
| `unresolved_reference` | Reference names an address that does not exist | FR-021 |
| `wrong_reference_kind` | Reference resolves to a kind this position forbids | FR-022 |
| `wrong_parent_kind` | Component's parent is not a container; container's is not a system | FR-028 |
| `containment_cycle` | Containment does not form a single-parent acyclic tree | FR-028 |
| `duplicate_claim` | Two claims over the same physical identifier | FR-028 |
| `duplicate_instance_name` | Two instances share a name in one environment | FR-012a |
| `unknown_block` | Block the language does not define | FR-018, FR-028 |
| `unknown_attribute` | Attribute the language does not define | FR-028 |
| `unknown_function` | Call outside the five of FR-017; detail names the five | FR-017a |
| `version_unsatisfied` | `loko_version` constraint not met by the running tool | FR-028 |

### Warnings (FR-029)

| Code | Rule |
|---|---|
| `orphan_element` | Element with no relationships in either direction |
| `empty_system` | System with no containers |
| `missing_docs` | Element with no prose reference |
| `docs_not_found` | Prose reference points at a file that does not exist |
| `unbound_instance` | Deployment instance with no claim |
| `self_relationship` | Element relates to itself — permitted, flagged |

### Explicitly not diagnosed

Dependency cycles between elements (FR-032). `api → queue → worker → api` is a normal architecture.
The cycle check applies to the `Parent` edge only. This is the one place the Terraform mental model
actively misleads, and the traversal that enforces containment acyclicity must not be reused for
relationships.

---

## 5. Lifecycle

The compiler is stateless (FR-027). There is no persisted state, lock file, or database, and no
entity has a mutable lifecycle. The only sequence is the per-invocation pipeline:

```
discover → parse → SourceModel → symbol table → resolve → validate → IR → encode
             │                                      │         │
             └── syntax diagnostics                 └─────────┴── semantic diagnostics
```

Diagnostics accumulate across every stage and are reported together (FR-031); an error at any stage
suppresses artefact production but not further diagnosis where the model is still coherent enough to
continue.
