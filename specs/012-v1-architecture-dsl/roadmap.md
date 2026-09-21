# loko v1.0 — Implementation Roadmap

**Design:** [`design.md`](./design.md)
**Purpose:** decompose v1.0 into spec-kit features. Each stage below carries a paste-ready
seed for `/speckit-specify`. Run one stage at a time; do not seed a later stage before its
predecessor's tasks are complete.

Feature numbering starts at `013` because `012` is this umbrella directory.

---

## Stage order and rationale

| # | Stage | Feature dir | Depends on | Delivers |
|---|---|---|---|---|
| 1 | Compiler core | `013-hcl-compiler-core` | — | `validate`, `fmt`, `export` |
| 2 | Renderers | `014-viewmodel-renderers` | 1 | `build`, `serve` |
| 3 | MCP rewire | `015-mcp-hcl-authoring` | 1, 2 | conversational authoring |
| 4 | Diff engine | `016-semantic-diff` | 1 | `diff`, `changelog` |
| 5 | Observation adapters | `017-observation-adapters` | 1 | `import`, `reconcile` |
| 6 | Policy engine | `018-architecture-policy` | 1, 4 | policies in `validate`, SARIF |

Stages 1–3 restore a complete working tool on the new foundation, so the project can dogfood
its own architecture before new capabilities land. Stages 4–6 are the differentiators and are
independently shippable; 5 and 6 may slip to v1.1 without invalidating the design.

Stage 6 depends on stage 4 only for delta-policies, which are v1.1. Snapshot policies need
stage 1 alone, so 6 can start in parallel with 4 if capacity allows.

---

## Stage 1 — Compiler core (`013-hcl-compiler-core`)

**Seed for `/speckit-specify`:**

> Replace loko's file-tree data model with an HCL-based source of truth and a stateless
> compiler. Introduce `*.loko.hcl` files discovered recursively from the project root and merged,
> with a `project` block replacing `loko.toml`. Support a logical plane (`person`, `system`,
> `container`, `component`, `external`) and a deployment plane (`deployment` containing an optional
> nested `node` tree containing `instance` blocks with `binding` sub-blocks). Relationships are
> `uses` blocks nested inside their source element. Cross-element references are HCL traversals
> such as `container.db`, resolved by two-pass evaluation: decode block headers into a symbol table,
> then evaluate bodies against an EvalContext built from it. Compile to an immutable IR keyed by
> address (`container.api`, `container.api.uses.orders`, `deployment.prod.instance.api`) that
> serializes deterministically to JSON and TOON. Validation reports errors (unresolvable reference,
> wrong parent type, duplicate binding address, containment cycle, unknown block or attribute,
> unsatisfied `loko_version` constraint) and warnings (element with no edges, system with no
> containers, element with no prose, instance with no binding), each with file/line/column ranges.
> Cycles between elements are legal; only containment must be a single-parent acyclic tree. Ship
> `loko validate`, `loko fmt` (canonical formatting via hclwrite), and `loko export --format
> json|toon`. Delete the v0 model layer this replaces: `adapters/filesystem/project_repo.go`,
> `adapters/config`, `adapters/ason`, `internal/api`, all `scaffold_*` and `create_*` use-cases,
> `detect_drift.go`, `drift_issue.go`, and `d2_relationship.go`, along with the tests covering
> them. Exclude `for_each`, `dynamic`, `variable` blocks, and modules; include `locals` and a small
> string-function set. Test with golden-file fixtures pairing `.loko.hcl` input against expected IR
> JSON and expected diagnostics including line ranges.

**Exit criteria:** `loko validate` and `loko export` succeed on a fixture project; `go build` is
green with the v0 model layer deleted; golden fixtures cover every validation rule.

---

## Stage 2 — Renderers (`014-viewmodel-renderers`)

**Seed for `/speckit-specify`:**

> Make every rendered artifact a pure function of the compiled IR. Introduce a ViewModel stage: the
> IR resolves to a set of views — implicit by default (one context view, a container view per
> system, a component view per container, a deployment view per environment) plus optional explicit
> `view` blocks for filtered custom views — and each view projects to a ViewModel carrying the
> element and edge subset at that level of detail with containment boundaries and styling already
> decided. Renderer backends consume ViewModels through a single `Render(ViewModel) ([]byte, error)`
> interface and know nothing about HCL, C4, or each other. Implement backends for D2 source, SVG,
> markdown, and the HTML site. Render D2 in-process using the `oss.terrastruct.com/d2` library that
> is already a direct dependency, removing the `exec.LookPath("d2")` external binary requirement in
> `adapters/d2/renderer.go`, so loko ships as a single static binary. Style by convention: element
> type maps to shape, `external` renders dashed, tags map to CSS classes. Rewire the existing HTML
> site builder to draw from ViewModels instead of the filesystem, and split
> `adapters/html/templates.go` (1,753 lines) during the rewire. Provide a `templates/` override
> directory of Go templates so users can re-theme without forking. Every generated file carries a
> DO-NOT-EDIT banner naming its source. Output must be byte-stable under sorted ordering at every
> stage, because golden-file testing of renderers depends on it and unreviewable generated diffs
> would push users back to hand-editing. Ship `loko build --format d2,svg,md,html` and `loko serve`
> with watch and live reload. Defer the mermaid backend to v1.1.

**Exit criteria:** `loko build` produces a complete site from HCL alone; two consecutive builds are
byte-identical; no external `d2` binary required.

---

## Stage 3 — MCP rewire (`015-mcp-hcl-authoring`)

**Seed for `/speckit-specify`:**

> Rebuild loko's MCP server against the compiled IR and make HCL the only thing it writes. Replace
> the v0 file-scaffolding tools with reads that answer graph questions and writes that edit HCL
> surgically. Read tools: `describe` (project summary at summary/structure/full detail, scoped to an
> address), `query` (dependents, transitive dependencies, path between two elements, orphans,
> coupling), and `validate` (compile and return diagnostics with source ranges). Write tools:
> `apply_edit` (add, update, or remove an element, edge, binding, or deployment instance) and `move`
> (emit a `moved` block and rewrite the address). All writes go through `hclwrite` as AST splices,
> never string concatenation or re-serialization, so comments, ordering, and formatting in
> hand-written files survive byte-identically. Every mutation compiles the result before writing; a
> failed compile returns diagnostics and touches nothing on disk. No MCP tool may write a `.d2`
> file, a diagram, or a markdown page — MCP writes HCL and nothing else, because a tool that can
> author a rendered artifact recreates the two-sources-of-truth defect this release exists to
> remove. Keep TOON as an output format. Add a property test asserting that compile → hclwrite edit
> → compile yields an identical IR; this guards the highest-risk surface in the product, since an
> LLM corrupting a user's source of truth is the failure that would sink the tool. Mirror the query
> engine as `loko query` on the CLI so both front ends call the same use-case functions.

**Exit criteria:** an LLM can build a multi-system architecture end to end over MCP; round-trip
property test passes; hand-written formatting and comments survive every tool-driven edit.

---

## Stage 4 — Diff engine (`016-semantic-diff`)

**Seed for `/speckit-specify`:**

> Add architecture-over-time diff to loko. Introduce `moved` blocks whose operands parse as static
> traversals and are never evaluated, since `from` deliberately names an address that no longer
> exists. Implement a diff engine that compiles two IRs, applies the moved map, and compares by
> address, classifying each change as added, removed, renamed, rewired (edge retargeted),
> attribute_changed, or rebound. Compute blast radius by reverse-reachability from each changed
> node. Load prior revisions with go-git in-process so no `git` binary is required, and support
> `--from-export` to diff against a previously exported IR instead. Ship `loko diff <revA>..<revB>`
> with human-readable and JSON output, and `loko changelog <revA>..<revB>` rendering the same diff
> as markdown. Expose a `diff` MCP tool over the same engine. Test with scenario fixtures covering:
> rename declared with `moved`; rename without `moved` (must read as add plus remove); edge
> retargeted; attribute changed; and an element moved between systems.

**Exit criteria:** every scenario fixture classifies correctly; diff output is deterministic;
changelog renders from the same engine with no duplicated logic.

---

## Stage 5 — Observation adapters (`017-observation-adapters`)

**Seed for `/speckit-specify`:**

> Let loko ingest and reconcile against real infrastructure through one adapter interface serving
> two verbs. An adapter produces an Observation: a flat set of observed resources each carrying a
> canonical id (ARN where one exists, otherwise `provider:type:logical-id`), a type, an attribute
> map, and optional inferred edges. Adapters know nothing about C4 and the core knows nothing about
> Terraform. Implement adapters for Terraform state JSON, Terraform plan JSON, and CloudFormation
> templates (which covers CDK, since `cdk.out` is CloudFormation). `loko import --from <adapter>
> --out <file>` turns an Observation into proposed HCL, grouping resources into logical elements
> using deterministic heuristics first — Terraform module path, CloudFormation nested stack, and
> Application/Service/Component tags — with bindings pre-filled; it always writes to a new file and
> never merges over existing HCL. `loko reconcile --deployment <name>` diffs an Observation against
> the IR's bindings and reports four findings: undocumented (observed, claimed by no binding),
> phantom (bound, not observed), drifted (claimed, attribute mismatch), and ambiguous (claimed by
> two bindings), plus a coverage percentage gateable with `--min-coverage`. Binding selectors
> support exact addresses, globs, address lists, and tag maps. Noise control is a first-class
> deliverable, not an afterthought: adapters classify incidental resources (IAM attachments, log
> groups, security group rules) and a `reconcile { ignore = [...] }` block lets projects tune it,
> because reconcile output buried in hundreds of noise lines will simply stop being run. Expose a
> `reconcile` MCP tool. Test against checked-in scrubbed fixtures of real `terraform show -json`
> output and real `cdk.out` templates; no live AWS in the test suite.

**Exit criteria:** reconcile against a real scrubbed state fixture produces correct findings and
coverage; default ignore rules keep output for a realistic account under a reviewable length.

---

## Stage 6 — Policy engine (`018-architecture-policy`)

**Seed for `/speckit-specify`:**

> Add architecture-level policy to loko — rules over the compiled graph that resource scanners
> cannot express. Introduce `policy` blocks with a severity and two rule kinds. Element rules
> (`require` / `deny`) pair a `select()` selector with attribute assertions. Path rules
> (`deny_path` / `require_path`) run BFS reachability over the compiled graph: `deny_path` violates
> when a path exists from any `from` element to any `to` element, `require_path` violates when no
> such path exists, and `except` lists intermediate hops that make a path permissible. Fix the
> selector vocabulary (`type`, `tag`, `tag_not`, `technology_matches`, `owner`, `system`) and the
> assertion vocabulary (`not_empty`, `matches`, `one_of`, `equals`). Policies run inside `loko
> validate` alongside compile diagnostics under the same severity model and exit codes — no new
> command — so the `validate` MCP tool returns violations and an agent can iterate until clean.
> Emit SARIF so violations land in GitHub code scanning. Do not implement resource-level
> configuration checks; Checkov and tfsec own that, and `loko export --format json | conftest test -`
> already gives users unlimited Rego power at zero implementation cost, which should be documented
> rather than built. Defer to v1.1: SARIF ingest mapping Checkov and tfsec findings onto elements
> through bindings, and policies evaluated on the diff delta such as "no new cross-boundary edges".
> Test with fixture graphs paired against expected violation sets, including cases where an
> `except` hop makes an otherwise-denied path legal.

**Exit criteria:** fixture graphs produce expected violations; SARIF validates against the schema;
`validate` exit codes unchanged for projects with no policy blocks.

---

## Running a stage

```
/speckit-specify   <paste the stage seed>
/speckit-clarify   # optional, recommended for stages 5 and 6
/speckit-plan
/speckit-tasks
/speckit-analyze   # optional, cross-artifact consistency
/speckit-implement
```

Constraints that apply to every stage, from `.specify/memory/constitution.md`: CLI handler
functions ≤ 50 effective lines, MCP handler functions ≤ 30, use-case files ≤ 200, entity files
≤ 300, and the layer-import rules. Run `task lint`, `task test`, and `task audit-constitution`
before opening a PR. Add one new layer rule during stage 1: no package outside
`internal/adapters/...` may import the HCL or D2 libraries.
