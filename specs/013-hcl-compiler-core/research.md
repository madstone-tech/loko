# Phase 0 Research: HCL Compiler Core

**Feature**: `013-hcl-compiler-core` | **Date**: 2026-09-22
**Spec**: [spec.md](./spec.md) | **Design**: [../012-v1-architecture-dsl/design.md](../012-v1-architecture-dsl/design.md)

This document resolves every NEEDS CLARIFICATION raised in the plan's Technical Context. Each
decision records what was chosen, why, and what was rejected.

---

## R1 — Configuration language and parser library

**Decision**: HashiCorp HCL v2 (`github.com/hashicorp/hcl/v2`), native syntax only. JSON-syntax
HCL (`.loko.hcl.json`) is not accepted in this release.

**Rationale**: The decisive property is `hclwrite`, which edits an HCL file as a token-preserving
AST. A downstream stage lets an agent add a relationship to a hand-written file and leave comments,
ordering, and formatting byte-identical. YAML and TOML both round-trip through a marshaller that
discards comments and reorders keys, which would make machine authoring of the source of truth
destructive rather than merely possible. HCL also carries byte-accurate source ranges through
parsing and evaluation, which FR-030 requires and which a hand-rolled parser would have to
reimplement.

Rejecting JSON syntax keeps one grammar to test and one canonical formatting to define. It costs
nothing today — no known consumer needs generated architecture source — and can be added later
without a breaking change.

**Alternatives considered**:

- *A hand-written parser for a purpose-built grammar.* Total control and no dependency, but source
  ranges, error recovery, and a formatter are each substantial work that HCL already does well. It
  also forfeits editor tooling that recognises HCL.
- *YAML with a comment-preserving library.* Comment preservation in YAML libraries is partial and
  fragile; block-scalar and anchor semantics add surface no architecture model needs.
- *CUE.* A better-typed language, but a heavy dependency, a steep learning curve for the target
  user, and no surgical-edit story.

---

## R2 — Reference resolution: static traversals, never evaluated

**Decision**: A cross-element reference such as `container.orders_db` is extracted with
`hcl.AbsTraversalForExpr` as a **static traversal** and rendered to a dotted address string plus a
source range. It is never passed through expression evaluation and never becomes a `cty` value.
Resolution against the symbol table happens in the core layer, not in the parser.

Scalar attributes (`description`, `technology`, `tags`, `attributes`) are a separate mechanism: they
*are* evaluated, against an `hcl.EvalContext` carrying local values and the five functions of R4.

**Rationale**: Two mechanisms sound like more complexity than one, but the split is what makes the
rest of the system simple:

1. **It puts resolution in the core layer.** If references were evaluated, the parser would need the
   symbol table pre-loaded into an `EvalContext`, so reference resolution and the unresolvable-
   reference diagnostic would live in the adapter — where the constitution forbids business logic
   and where they would be far harder to unit-test. Extracting traversals structurally lets the
   adapter emit a raw address string and lets the core own FR-019 through FR-022 entirely.
2. **Declaration order stops mattering for free.** FR-020 requires two passes. With static
   traversals the first pass is simply "collect every block header" and the second is a map lookup;
   there is no partially-populated evaluation context to sequence.
3. **A later stage needs exactly this.** The rename-declaration block deliberately names an address
   that no longer exists, so its operands *must not* be evaluated. Under this decision that block is
   an ordinary use of the normal reference mechanism rather than a special case carved out of the
   evaluator.

The cost is that a reference cannot be computed — no `container[local.name]`. Since iteration
constructs are excluded by FR-018, nothing in the language can produce a computed reference anyway.

**Alternatives considered**:

- *Populate the `EvalContext` with element values and evaluate references normally.* The obvious
  approach, and the one Terraform uses. Rejected because it drags resolution into the adapter layer
  and forces the rename block to become an exception to the evaluator.
- *Capsule types (`cty.Capsule`) wrapping an element handle.* Gives real type errors from the
  evaluator, but still evaluates, so it inherits the layering problem above, and capsule values
  print poorly in diagnostics.
- *Plain strings (`target = "container.orders_db"`).* No editor support, no structural distinction
  between a reference and a description, and it invites typos that read as valid.

---

## R3 — Two-pass decoding mechanics

**Decision**: Pass 1 walks the parsed file bodies with `hcl.Body.PartialContent` against a
block-header-only schema, recording every declared address, its kind, its labels, and its range.
Pass 2 decodes each block body against a per-kind schema, evaluating scalars and extracting
reference traversals. The adapter emits a flat `SourceModel` of declarations; it does **not** build
the graph.

Parsing uses `hclparse.Parser`, whose file cache is retained for the run so that diagnostics can be
rendered with source snippets via `hcl.NewDiagnosticTextWriter`.

**Rationale**: `PartialContent` reports unexpected blocks and attributes as diagnostics rather than
failing, which is what FR-031 (report everything in one run) needs. Keeping the parser's file cache
alive is what lets the human-readable output show the offending line with a caret, which is most of
the perceived quality of a compiler.

**Alternatives considered**:

- *`gohcl.DecodeBody` into tagged structs for everything.* Concise, but it fails fast on the first
  structural error and gives less control over partial recovery; `gohcl` is still used for leaf
  bodies where its behaviour is adequate.
- *`hcldec` specs.* Better suited to dynamically-determined schemas; ours is fixed and known at
  compile time, so the extra indirection buys nothing.

---

## R4 — Function set

**Decision**: Exactly five functions — `join`, `split`, `lower`, `upper`, `replace` — sourced from
`github.com/zclconf/go-cty/cty/function/stdlib` (`JoinFunc`, `SplitFunc`, `LowerFunc`, `UpperFunc`,
`ReplaceFunc`), plus native string interpolation. Any other call is an unknown-function error naming
the five (FR-017a).

**Rationale**: The clarified spec fixes the set; the useful finding is that all five already exist in
`cty`'s standard library, correctly handle unknown and null values, and come with upstream tests. The
implementation is a five-entry map, not five functions. `cty` arrives as a transitive dependency of
HCL regardless, so this adds no new module.

**Alternatives considered**:

- *Hand-written implementations.* Avoids depending on `cty`'s function package, but `cty` is already
  in the build and hand-rolled versions would mishandle null and unknown values.
- *Expose all of `cty`'s stdlib (~40 functions).* Contradicts the clarified decision, and every
  function shipped is permanent surface.

---

## R5 — Version constraint evaluation without a dependency

**Decision**: Implement a small constraint evaluator in `internal/core/entities`, stdlib only,
supporting `=`, `!=`, `>`, `>=`, `<`, `<=`, `~>`, and comma-separated conjunctions over
`major.minor.patch`. Pre-release and build metadata are parsed and compared per semantic-versioning
ordering; ranges with wildcards beyond `~>` are rejected with a clear message.

**Rationale**: `hashicorp/go-version` would do this in one line, but the constraint check belongs to
FR-028's error set and therefore to core validation, and the constitution states `internal/core/`
has zero external dependencies. Pushing it to an adapter to satisfy the letter of the rule would put
a validation rule outside the validation layer. The evaluator is roughly 80–120 effective lines of
pure string and integer comparison with no I/O — well inside the entity budget and trivially
table-testable.

**Alternatives considered**:

- *`github.com/hashicorp/go-version` in an adapter behind a port.* Correct by the rules, but it means
  a port, an adapter, a mock, and wiring for what is arithmetic on three integers. Rejected under
  Principle VII (no abstractions for one-time operations).
- *Accept only exact versions.* Removes the need for the evaluator but makes the field nearly
  useless — every patch release would invalidate every project.

---

## R6 — Determinism strategy

**Decision**: The IR holds **ordered slices, never maps**, in every field that reaches an encoder.
Ordering is assigned once, at IR construction, by sorting on the address string using byte-wise
comparison. Encoders serialise in slice order and perform no sorting of their own. The export carries
a schema version integer and nothing else that varies per run (FR-036a, FR-036c).

**Rationale**: Go map iteration order is deliberately randomised, so any map that reaches an encoder
is a latent nondeterminism bug that tests pass 90% of the time. Making the IR's public shape
map-free removes the failure mode structurally rather than relying on every encoder to remember to
sort. Byte-wise address comparison avoids locale-dependent collation.

`encoding/json` does sort map keys, which is why this is easy to get wrong: JSON output would look
deterministic while the TOON encoder — which serialises arrays positionally — would not. Ordering at
construction makes both correct by the same mechanism.

**Alternatives considered**:

- *Maps internally, sort in each encoder.* Duplicates the ordering rule per backend and fails open:
  a new backend is nondeterministic until someone remembers.
- *Sort at assertion time in tests.* Hides the defect instead of preventing it, and leaves real
  committed artefacts churning.

---

## R7 — Layering: what the adapter owns versus what core owns

**Decision**:

| Concern | Layer | Rationale |
|---|---|---|
| File discovery, parsing, syntax errors | `internal/adapters/hclsource` | Needs the HCL library |
| Scalar evaluation, locals, the five functions | `internal/adapters/hclsource` | Needs `cty` |
| Unknown block / attribute / function errors | `internal/adapters/hclsource` | Detected by the schema walk |
| Traversal extraction to address strings + ranges | `internal/adapters/hclsource` | Needs `hcl.AbsTraversalForExpr` |
| Symbol table, reference resolution, unresolvable-reference error | `internal/core/usecases` | Pure map work over the `SourceModel` |
| Parent-kind rules, containment cycles, duplicate claims, duplicate instance names | `internal/core/usecases` | Business rules |
| Version constraint check | `internal/core/entities` | Domain validation (R5) |
| Warnings | `internal/core/usecases` | Business rules |
| IR construction and ordering | `internal/core/usecases` | Pure |
| Canonical formatting | `internal/adapters/hclsource` | Needs `hclwrite` |
| Encoding to the two output formats | `internal/adapters/encoding` | Existing adapter |

The port is `ArchitectureSource`, returning a `SourceModel` (flat declarations carrying raw reference
strings and source ranges) plus diagnostics. Core turns that into the IR.

**Rationale**: This satisfies FR-044 — no package outside `internal/adapters/...` imports HCL — while
keeping every rule from FR-028 and FR-029 in the layer the constitution designates for validation.
The `SourceModel`/IR split also means the whole validation suite is testable by constructing a
`SourceModel` literal, with no parsing and no fixture files, which is what keeps the golden-file
suite focused on syntax rather than on rules.

**Alternatives considered**:

- *Adapter returns the finished IR.* Fewer types, but resolution and validation would sit in the
  adapter, violating Principles I and IV and making the rules hard to test in isolation.
- *Core imports HCL and there is no adapter.* Simplest code, direct violation of the dependency
  rule, and forfeits the new layer rule the spec requires in FR-044.

---

## R8 — Source ranges without importing HCL into core

**Decision**: `entities.SourceRange` is a plain struct — `File string`, `StartLine`, `StartColumn`,
`StartByte`, `EndLine`, `EndColumn`, `EndByte int`. The adapter converts `hcl.Range` into it at the
boundary. Human-readable rendering with source snippets uses HCL's own diagnostic writer in the
adapter; machine-readable rendering uses the plain struct.

**Rationale**: `hcl.Range` in an entity would drag the HCL library into `internal/core/entities`,
which the constitution forbids and FR-044 turns into an enforced rule. The struct is a faithful
one-to-one copy, so nothing is lost.

**Alternatives considered**:

- *Store an opaque handle and ask the adapter to render it.* Couples diagnostics to a live parser
  instance and breaks the machine-readable output, which must be serialisable on its own.

---

## R9 — Sequencing the deletion (the largest risk in this feature)

**Finding**: `ProjectRepository`, `ConfigLoader`, and `TemplateEngine` are referenced by **44
non-test files**, including 20 of the 22 MCP tools, 8 CLI command files, the whole HTTP interface,
and 13 use cases. FR-041 deletes the implementations of all three, and FR-042 forbids adapting the
tests instead. The build cannot stay green unless everything reachable from those ports goes in the
same change.

**Decision**: Stage 1 removes, in one change:

- the three adapters (`filesystem/project_repo.go`, `adapters/config`, `adapters/ason`) and their ports;
- the HTTP interface (`internal/api`, `cmd/api.go`) and its command;
- every `scaffold_*` and `create_*` use case, `detect_drift.go`, and the drift entity;
- the use cases that read the v0 tree — `init_project`, `find_relationships`, `list_relationships`,
  `delete_relationship`, `search_elements`, `query_architecture*`, `build_docs*`,
  `render_markdown_docs`, `update_diagram`, `validate_architecture` and its helpers;
- the MCP tools backed by them, leaving the MCP server harness and its registry in place with an
  empty or near-empty tool set;
- the CLI commands backed by them — `build`, `serve`, `watch`, `new`, `init`, `api`;
- `adapters/html` usage from the build path (the builder and templates stay on disk, unwired, for the
  renderer stage), and `adapters/d2/parser.go` plus `d2_import_test.go`, which FR-045 explicitly
  parks for the diagram-import stage.

The surviving command set after Stage 1 is `validate`, `fmt`, `export`, `version`, `completion`, and
`mcp` (serving a reduced tool set).

**Consequence the reader must accept**: `loko build`, `loko serve`, and `loko init` do not work
between the end of this stage and the end of the renderer stage. This is inherent to replacing the
data model rather than a shortcoming of the plan — the roadmap says as much ("Stages 1–3 restore a
complete working tool on the new foundation") — but it is stated here because it is the single most
consequential fact about this feature and it is not visible from the spec's requirement list.

**Alternatives considered**:

- *Keep the v0 model alongside the new one behind a flag until the renderer stage.* Keeps every
  command working throughout. Rejected: it is precisely the two-sources-of-truth state the release
  exists to remove, Principle VII forbids compatibility shims, and it would double the surface that
  the renderer stage then has to unwind.
- *Delete only `project_repo.go` and stub the port.* A stub that returns "not implemented" leaves
  dead abstractions and failing commands with worse messages than a removed command, and leaves the
  old tests to be adapted, which FR-042 forbids.
- *Defer the deletion to a later stage.* Leaves drift detection reachable, contradicting FR-043, and
  makes the renderer stage carry two models at once.

---

## R10 — `loko init` in this stage

**Decision**: `loko init` is removed in this stage along with `scaffold_project.go`, and returns in
the renderer stage. The quickstart guide has the reader write a short starter file by hand.

**Rationale**: `init` today renders a directory tree of prose and configuration through the template
engine — all of which this feature deletes. A replacement that emits a single starter file is small,
but it is not in the spec's requirement list, and adding an unrequested command is scope the user did
not ask for. Writing eight lines of source by hand is a reasonable first-run experience for a
compiler, and the quickstart doubles as the documentation for it.

**Settled 2026-09-22**: the user confirmed this is a rewrite and that breaking changes are
acceptable. `init` stays out of this stage; no replacement is built here. This is no longer an open
question and should not be reopened during implementation.

---

## R11 — Testing strategy

**Decision**: Three layers.

1. **Golden fixtures for the compiler** (the bulk): `testdata/<case>/` holding input source, expected
   exported IR, and expected diagnostics including ranges. A `-update` flag regenerates them. One
   fixture minimum per rule in FR-028 and FR-029, which is what SC-004 measures.
2. **Table tests in core** constructing `SourceModel` literals directly, with no files, for
   resolution and validation rules. Fast, and they isolate rule bugs from parsing bugs.
3. **Determinism tests**: compile-and-export twice asserting byte equality, plus a run under a
   shuffled file-discovery order asserting identical output, which is what actually catches a map
   that escaped into the IR.

Per Principle V these are written before the implementation they cover. Per Principle VII, mocks are
concrete structs, not a mocking library.

**Rationale**: Golden files are only viable because of R6; if output were not byte-stable the suite
would be unwritable. Splitting rule tests away from fixtures keeps a failure message pointing at one
cause — a fixture diff that changes 400 lines because an unrelated ordering shifted is not a useful
test failure.

---

## R12 — New module dependencies

| Module | Purpose | Layer permitted |
|---|---|---|
| `github.com/hashicorp/hcl/v2` | Parsing, evaluation, formatting | `internal/adapters/hclsource` only |
| `github.com/zclconf/go-cty` | Value model and the five functions | `internal/adapters/hclsource` only |

Both are confined by the new layer rule of FR-044. `go-cty` is a transitive dependency of HCL, so it
is promoted to direct rather than newly introduced. `go-git` is **not** added — it belongs to the
diff stage. No dependency is removed from `go.mod` by hand; `go mod tidy` drops what the deletions
orphan (`ason`, and TOML handling if nothing else imports it).

---

## Open items carried into implementation

**None.**

Two items were previously carried here and are both now closed:

- **R9** — `build`, `serve`, `watch`, and `init` stop working until the renderer stage. This was never
  a decision: it is forced by FR-041 plus the requirement that the build stay green, given the
  44-file dependency base. Recorded so it is not rediscovered mid-implementation.
- **R10** — whether to ship a replacement `init` in this stage. Settled: no. The user confirmed on
  2026-09-22 that this is a rewrite and breaking changes are acceptable.

The project is treated as a rewrite throughout. No requirement, contract, or task in this feature
preserves v0 behaviour, file formats, or command surface for compatibility's sake, and none should be
added on those grounds.
