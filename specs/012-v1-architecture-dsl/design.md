# loko v1 — HCL Architecture DSL

**Date:** 2026-09-21
**Status:** Approved design, pending implementation plan
**Supersedes:** the v0.2 file-tree data model (markdown frontmatter + D2 arrows, union-merged)

---

## 1. Problem

v0.2 has two sources of truth. Relationships are authored in markdown frontmatter *and* in D2 arrow
syntax, then union-merged at build time. `loko validate --check-drift` exists to report when the two
disagree.

Drift detection is not a feature. It is the symptom of a model with no authoritative layer. Because
no single artifact is authoritative, three things follow:

1. **Edits go stale.** Rename or rewire one element and the diagrams, prose, tables, and relationship
   lists silently diverge.
2. **There is no architecture-level diff.** Nothing can answer "what changed since Q1, and what is
   downstream of it" because there is no comparable compiled value — only a directory of prose.
3. **There is nothing to reconcile against reality.** Comparing a deployed environment to a tree of
   markdown files is not well-defined.

v1 replaces the data model. HCL becomes layer 0 — the sole authored artifact. Everything else is a
projection of a compiled graph.

**Positioning shift:** v0 was a documentation generator with a graph bolted on. v1 is a graph
compiler that emits documentation.

---

## 2. Goals

- One authored source of truth; every rendered artifact is a pure function of it.
- Typed references between elements, resolved at compile time, so a broken relationship is a
  compile error rather than a stale diagram.
- Semantic diff between two revisions: added, removed, renamed, rewired, attribute-changed, re-bound,
  plus blast radius.
- Multi-modal ingestion and reconciliation against real infrastructure.
- MCP-native: an LLM and a human edit the same source of truth without either degrading it.
- Architecture-level policy that resource scanners structurally cannot express.

## 3. Non-goals

- Resource-level configuration scanning. Checkov and tfsec own that; loko integrates, not competes.
- Multi-user server, persisted database, or web editor. HCL plus git is the system of record.
- Automatic migration from v0.2. See §12.
- `for_each`, `dynamic`, `variable` blocks, or modules in the language. Deliberately deferred.

---

## 4. Approach

**Stateless compiler; git is the history.**

`loko` parses HCL, resolves references, and produces an immutable IR. No lock file, no database,
no state file. Identity is the HCL address. Diff compiles two revisions and compares.

Two alternatives were considered and rejected:

- **Committed IR snapshot (`loko.lock.json`).** Buys automatic rename detection and git-free diff.
  Costs a second artifact that goes stale, merge conflicts on every branch, and a regeneration
  ritual. `moved` blocks deliver the same benefit without the state file.
- **Persisted graph database with HCL as one editing surface.** Rejected outright: the moment HCL
  is *a* surface rather than *the* surface, there are two sources of truth again — precisely the
  defect being removed.

**Why HCL specifically.** `hclwrite` supports AST-preserving surgical edits. An MCP tool can add a
relationship to a hand-written file and leave comments, ordering, and formatting byte-identical.
YAML or JSON would round-trip through a marshaller and destroy the file. That property is what makes
MCP-native editing of the source of truth acceptable rather than merely possible.

---

## 5. The DSL

### 5.1 File layout

Any `*.loko.hcl` under the project root, merged recursively. A single `project` block replaces
`loko.toml`; configuration moves into the same language as the model.

```
my-arch/
├── arch.loko.hcl          # or many *.loko.hcl, merged
├── docs/payments.md       # prose, referenced by docs = "..."
├── templates/             # optional theme overrides
└── dist/                  # generated, gitignored by default
```

### 5.2 Two planes

- **Logical plane** (C4, environment-agnostic): `person`, `system`, `container`, `component`,
  `external`.
- **Deployment plane**: `deployment` → optional nested `node` tree → `instance`. Instances carry
  bindings, sizing, placement.

A logical element describes intent once. Deployments instantiate it per environment. Reconciliation
runs per environment against that environment's state.

### 5.3 References

`container.db` is an HCL traversal evaluated against a symbol table. Two-pass evaluation, as
Terraform does it:

1. Decode block headers into a symbol table.
2. Evaluate bodies against an `EvalContext` populated from that table.

A typo is a compile error with a source range.

### 5.4 Example

```hcl
project "acme-payments" {
  description   = "Payment processing platform"
  loko_version  = "~> 1.0"
}

system "payments" {
  description = "Handles authorization, capture, and settlement"
  owner       = "platform-team"
  docs        = "./docs/payments.md"
}

container "api" {
  system     = system.payments
  technology = "AWS Lambda (Go)"
  tags       = ["public", "pci"]

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
    technology  = "PostgreSQL wire protocol"
  }
}

container "orders_db" {
  system     = system.payments
  technology = "Aurora PostgreSQL"
  tags       = ["pci"]
}

deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "vpc-main" {
    instance "api" {
      of         = container.api
      attributes = { memory = 1024, timeout = 30 }

      binding "terraform" {
        state   = "envs/prod/terraform.tfstate"
        address = "module.api.aws_lambda_function.this"
      }
    }
  }
}

reconcile {
  ignore = [
    "aws_iam_role_policy_attachment.*",
    "aws_cloudwatch_log_group.*",
  ]
}
```

### 5.5 Relationships nest inside their source

`uses "orders"` inside `container "api"` yields the stable address
`container.api.uses.orders`. That address is what lets diff report *rewired* rather than
"one edge vanished and another appeared."

The block label is the relationship's local name; `target` is the reference.

### 5.6 Bindings

A `binding` selects physical resources. Supported selectors:

- `address` — exact Terraform/CFN address or a glob (`module.api.*`)
- `addresses` — list of the above
- `tags` — map selector matched against observed resource tags

Binding kinds in v1.0: `terraform`, `cloudformation`.

### 5.7 Views

Implicit by default: one context view, a container view per system, a component view per container,
a deployment view per environment. Optional `view "payment-path" { ... }` blocks define filtered
custom views. Zero-config for the common case; an escape hatch for the real one.

### 5.8 Deliberate YAGNI

No `for_each`, `dynamic`, `variable` blocks, or modules in v1.0. `locals` and a small string-function
set only. Each is additive later; shipping them now triples compiler surface before real usage
indicates which are needed.

---

## 6. Compiler and IR

### 6.1 Pipeline

```
discover → parse → symbol table → evaluate → build IR → validate
```

Every command and every MCP tool consumes a compiled IR. None of them parse HCL themselves. One
front end, many back ends.

### 6.2 The IR

An immutable, fully-resolved value: elements keyed by address, edges with resolved endpoints, the
deployment tree, and derived views. Serializes to JSON and TOON — this is `loko export`, the CI
artifact, and the MCP read payload.

### 6.3 Identity

Addresses: `container.api`, `container.api.uses.orders`, `deployment.prod.instance.api`.

Renames are declared:

```hcl
moved {
  from = container.api
  to   = container.payments_api
}
```

`moved` operands parse as **static traversals and are never evaluated** — `from` deliberately names
something that no longer exists. Diff applies the moved map before comparing.

Builds on the v0.2 graph-qualified-ID work (ADR-0004, `migration-001-graph-qualified-ids`).

### 6.4 Validation

**Errors:** unresolvable reference; component whose parent is not a container; container whose parent
is not a system; two bindings claiming the same physical address; containment cycle; unknown block or
attribute; `loko_version` constraint unsatisfied.

**Warnings:** element with no edges; system with no containers; element with no prose; deployment
instance with no binding.

Diagnostics carry file, line, and column ranges from HCL natively.

### 6.5 Cycles are legal between elements

`api → queue → worker → api` is a normal architecture, not an error. Only *containment* must be a
single-parent acyclic tree. This is the one place the Terraform mental model actively misleads, and
the implementation must not import Terraform's acyclicity assumption.

### 6.6 Diff

Compile two IRs, apply the `moved` map, compare by address. Change classes: `added`, `removed`,
`renamed`, `rewired` (edge retargeted), `attribute_changed`, `rebound`. Reverse-reachability from
each changed node yields blast radius.

```
$ loko diff main..HEAD
  ~ container.api              technology "Lambda (Node)" → "Lambda (Go)"
  → container.api.uses.orders  retargeted container.orders_db → container.orders_aurora
  + container.dlq_worker
  ! blast radius: 3 elements downstream of container.api
```

Old revisions load via **go-git in-process** (no `git` binary dependency), or from a prior export
with `--from-export`. `loko changelog` renders the same diff as markdown.

### 6.7 Determinism is a hard requirement

Sorted ordering at every stage. Generated diagrams and docs must be byte-stable so a commit diff
shows only what actually changed. Nondeterministic output would make generated artifacts
unreviewable, pushing users back to hand-editing them — which reintroduces a second source of truth.
It is also the precondition for golden-file testing (§13).

---

## 7. Observation adapters

One interface, two verbs.

An **adapter** produces an `Observation`: a flat set of observed resources, each with a canonical id
(ARN where one exists, else `provider:type:logical-id`), a type, an attribute map, and optional
inferred edges. Adapters know nothing about C4; the core knows nothing about Terraform.

### 7.1 `loko import` — observation → proposed HCL

The hard part is grouping, not parsing: several hundred resources into a dozen containers.
Deterministic heuristics run first — Terraform module path, CloudFormation nested stack,
`Application`/`Service`/`Component` tags — producing a candidate grouping with bindings pre-filled.
An LLM refines it over MCP; a human accepts it.

Output is written to `--out` as a new file. Import never merges over existing HCL.

### 7.2 `loko reconcile --deployment prod` — observation vs bindings

| Finding | Meaning |
|---|---|
| `undocumented` | observed, claimed by no binding |
| `phantom` | bound, not observed |
| `drifted` | claimed, attribute mismatch |
| `ambiguous` | claimed by two bindings |

Plus **coverage**: percent of observed resources claimed. This is the metric worth gating CI on
(`--min-coverage 80`).

### 7.3 Noise control is load-bearing

A real account holds hundreds of IAM attachments, log groups, and security group rules nobody will
draw. Adapters classify incidental resources, and `reconcile { ignore = [...] }` lets a project tune
it. Without this, reconcile output is hundreds of lines of noise and users stop running it — which
kills the feature regardless of correctness.

### 7.4 Adapter shipping order

- **v1.0:** Terraform state JSON, Terraform plan JSON, CloudFormation templates (covers CDK, since
  `cdk.out` is CFN).
- **v1.1:** live AWS scan via Resource Explorer / Cloud Control.
- **v1.2:** diagram import (D2, mermaid), explicitly best-effort — a diagram yields nodes and edges
  but no bindings, so it bootstraps the logical plane only.

The conversation path is not an adapter. It is MCP writing HCL, and it ships in v1.0 for free.

---

## 8. Renderers

Every renderer is a pure function of the IR. No generated file is ever hand-edited; each carries a
`DO NOT EDIT` banner naming its source. This is the structural guarantee that replaces drift
detection.

### 8.1 Two stages

The IR resolves to a set of views. Each view projects to a **ViewModel** — the element and edge
subset at that level of detail, with containment boundaries and styling already decided. Backends
consume ViewModels and know nothing about HCL, C4, or each other:

```go
type Renderer interface {
    Render(v ViewModel) ([]byte, error)
}
```

### 8.2 D2 is the default, and embedded

`oss.terrastruct.com/d2` is already a direct dependency, but `adapters/d2/renderer.go` still shells
out via `exec.LookPath("d2")`. v1 renders in-process, making `loko` a single static binary with no
external diagram dependency. ELK layout by default; themes configurable via a `render "d2" { }`
block.

### 8.3 Styling by convention

Element type → shape, `external` → dashed, tags → CSS classes. Deterministic, so diagrams are
byte-stable.

### 8.4 Markdown and HTML site

Backends like any other: a page per element with the sibling prose file inlined, auto-generated
relationship and component tables, and the view's diagram embedded. The v0 site survives; it draws
from the IR instead of from files.

### 8.5 SVG for review workflows

`loko render --out docs/diagrams/` emits SVGs that can be committed and referenced from markdown.
GitHub renders them inline, so a PR that rewires the architecture shows a visibly changed diagram.
Architecture change review becomes ordinary code review.

### 8.6 Mermaid deferred to v1.1

Mermaid *flowchart* is stable; mermaid's native C4 syntax is experimental, renders poorly, and is
unsupported by GitHub's renderer. When it ships, the backend emits `flowchart` with `subgraph`
boundaries, **not** `C4Context`/`C4Container`. The ViewModel boundary makes it a drop-in with no core
changes.

### 8.7 Escape hatch

A `templates/` override directory of Go templates. OSS users must be able to re-theme without
forking.

---

## 9. MCP surface

v0 shipped 17 tools that mostly scaffolded files. With a compiler underneath, the surface gets
smaller and more useful: the model is queried and mutated as a graph, not as a directory.

### 9.1 Reads — against the IR

| Tool | Purpose |
|---|---|
| `describe` | Project summary at detail level (`summary`/`structure`/`full`), scoped to an address |
| `query` | Graph query: dependents, dependencies (transitive), path between elements, orphans, coupling |
| `diff` | Semantic diff between revisions or against an export |
| `reconcile` | Run an adapter; return findings and coverage |
| `validate` | Compile; return diagnostics and policy violations with source ranges |

### 9.2 Writes — through `hclwrite`

| Tool | Purpose |
|---|---|
| `apply_edit` | Add/update/remove an element, edge, binding, or deployment instance |
| `move` | Emit a `moved` block and rewrite the address |
| `propose_import` | Turn an observation into candidate HCL for review |

**The write path must be right.** Every mutation compiles the result before writing; a failed compile
returns diagnostics and touches nothing. Edits are AST splices, not re-serialization, so comments,
ordering, and formatting survive.

### 9.3 Token efficiency

TOON stays as an output format, but the encoding is not the win — v0 measured 9.2%. The win is that
`query` answers "what depends on the orders database" in six lines instead of the model loading the
whole architecture and reasoning over it. Structured graph queries beat compression by an order of
magnitude.

### 9.4 One firm boundary

MCP writes HCL and nothing else. No tool writes a `.d2` file, a diagram, or a markdown page. An MCP
tool that can author a rendered artifact recreates the two-sources-of-truth defect.

---

## 10. Architecture policy

### 10.1 What loko owns

Resource scanners evaluate resources mostly in isolation. loko has a typed graph with boundaries,
tags, ownership, and layers, so it answers questions they cannot express:

- Can anything outside the PCI boundary reach a PCI datastore without passing through the gateway?
- Does any container in the public tier depend directly on a container in the data tier?
- Is there a path from an internet-facing entry point to an unencrypted store?
- Does every datastore have an owner and prose?
- Did this change introduce a cross-system dependency that did not exist on `main`?

The last is only possible because semantic diff exists: policies can be evaluated on the **delta**,
not just the snapshot. (Delta policies ship in v1.1.)

### 10.2 Native `policy` blocks

```hcl
policy "pci-isolation" {
  severity    = "error"
  description = "Only the payment gateway may reach PCI datastores"

  deny_path {
    from   = select(tag_not = "pci")
    to     = select(tag = "pci", type = "container")
    except = [container.payment_gateway]
  }
}

policy "datastores-documented" {
  severity = "warning"
  require {
    select = select(type = "container", technology_matches = "Aurora|DynamoDB|RDS")
    assert = { owner = not_empty, docs = not_empty }
  }
}
```

Two rule kinds in v1.0:

- `require` / `deny` over **elements** — selector plus attribute assertions.
- `deny_path` / `require_path` over **paths** — BFS reachability over the compiled graph.
  `deny_path` violates when a path exists from any `from` element to any `to` element;
  `require_path` violates when no such path exists. `except` lists intermediate hops that, when
  traversed, make a path permissible.

The `select()` vocabulary (`type`, `tag`, `tag_not`, `technology_matches`, `owner`, `system`) and the
assertion vocabulary (`not_empty`, `matches`, `one_of`, `equals`) are fixed in the implementation
plan.

### 10.3 OPA escape hatch — zero code

`loko export --format json | conftest test -` works on day one because export already exists.
Document it; do not build it.

### 10.4 What loko refuses to own

Resource-level configuration checks. Do not reimplement Checkov. Instead **ingest** it (v1.1):
Checkov and tfsec emit SARIF, and bindings already map `aws_lambda_function.api` → `container.api`,
so a resource finding surfaces on the architecture element that owns it.

loko **emits** SARIF in v1.0 so violations land in GitHub code scanning.

### 10.5 No new command

`loko validate` runs policies alongside compile diagnostics — same severity model, same exit codes.
The MCP `validate` tool therefore returns policy violations, so an LLM can be asked to fix them and
iterate until validate is clean.

---

## 11. CLI

One engine, two front ends. Every command is a thin shell over the same use-case functions the MCP
tools call. No logic lives in `cmd/` or `internal/mcp/`.

```
loko init [name]                      scaffold a project
loko fmt                              canonical formatting via hclwrite
loko validate [--strict]              compile, run policies, print diagnostics
loko build [--format d2,svg,md,html]  emit artifacts
loko serve                            watch + live reload
loko diff <revA>..<revB>              semantic diff
loko changelog <revA>..<revB>         same diff, as markdown
loko import --from <adapter> --out    observation → proposed HCL
loko reconcile --deployment prod      observation vs bindings + coverage
loko query <expr>                     CLI mirror of the MCP query tool
loko export --format json|toon        IR artifact
loko mcp                              MCP server
```

**Exit codes:** `0` clean, `1` errors, `2` warnings under `--strict`, `3` reconcile below
`--min-coverage`. CI gates need to distinguish these.

**Version constraint in source:** `project { loko_version = "~> 1.0" }`, like Terraform's
`required_version`. A tool that ships a language needs the language to state which compilers can
read it.

---

## 12. Codebase transplant

Same repository. This design lands on `v1-design`; implementation proceeds on `v1`.
Keep the skeleton and the periphery; replace the organs.
Measured against `main` at `185b8f7`: 21.9k source LOC, 24.2k test LOC, 103 test files.

| Bucket | ~LOC | Packages |
|---|---|---|
| **Keep largely intact** | 6,800 | `adapters/html` (site builder + templates), `adapters/d2` (generation), `adapters/encoding` (JSON/TOON), `mcp` server harness, `cmd` cobra shells, `ui`, `logging`, `adapters/markdown` |
| **Port inward** | 4,700 | `entities/graph.go`, `graph_traversal.go`, `graph_id.go`, `dependency_report.go`; query use-cases (`architecture_graph_queries`, `find_relationships`, `search_elements`); `mcp/tools` (bodies change, response shapes survive) |
| **Delete** | 6,700 | `adapters/filesystem/project_repo.go` (1,082), `adapters/config` (TOML), `adapters/ason`, `internal/api` + swagger, all `scaffold_*` and `create_*` use-cases, `detect_drift.go`, `drift_issue.go`, `d2_relationship.go` |
| **New** | — | HCL parser/compiler/IR, diff engine, observation adapters, deployment plane, policy engine |

`project_repo.go` is the largest deletion and the most important one: that file **is** the
dual-source-of-truth machinery. Removing it makes drift detection unnecessary rather than
unimplemented.

`adapters/d2/parser.go` and `d2_import_test.go` are **parked, not deleted** — they become the v1.2
diagram-import adapter.

`adapters/html/templates.go` is 1,753 lines in one file. Split it during the ViewModel rewire.

**The trap:** most of the 24k test LOC covers use-cases that are going away. Do not try to keep them
green. Delete tests alongside the code they cover. Preserving old abstractions to avoid touching old
tests is what turns a transplant into a compromise.

**Order of work:** compiler core first (parser → IR → validate), then renderers, then diff, then
adapters, then policy, then MCP rewire, then CLI.

### 12.1 No `loko migrate` command

v0.2's install base does not justify writing a D2 arrow parser, a frontmatter reader, TOML handling,
and v0's union-merge conflict semantics — and the D2 parsing half duplicates the v1.2 diagram-import
adapter.

The capability ships in v1.0 for free via MCP: point an LLM at the v0 tree; it reads the markdown and
`.d2` files and writes HCL through `apply_edit`, compiling as it goes. Ship a "Migrating from v0.2"
doc page with that recipe and keep the `v0.2.x` tag installable. If diagram import lands in v1.2, a
mechanical `migrate` becomes nearly free — build it then, if anyone asks.

---

## 13. Testing

- **Compiler: golden files.** `testdata/` pairs of `.loko.hcl` input with expected IR JSON and
  expected diagnostics including line ranges. Adding a language feature means adding a fixture pair.
  This is the bulk of the suite.
- **Renderers: golden files.** Only possible because output is byte-stable. Determinism is what makes
  the test strategy work, not a nicety.
- **Round-trip property test:** compile → `hclwrite` edit → compile yields an identical IR. This
  single property guards the entire MCP write path — the highest-risk surface, since an LLM mangling
  a user's source of truth is the failure that would sink the tool.
- **Adapters: recorded fixtures.** Real `terraform show -json` output and real `cdk.out` templates,
  checked in and scrubbed. No live AWS in the test suite.
- **Diff: scenario tests.** Rename with `moved`; rename without `moved` (must read as add + remove);
  rewire; attribute change; element moved between systems.
- **Policy: fixture graphs** with expected violation sets, including `except`-hop path cases.
- **Carry over v0's mechanical gates** — `task audit-constitution`, layer-import rules, file and
  function size budgets. Add one rule: no package outside `internal/adapters/...` may import the HCL
  or D2 libraries, keeping the core free of both.

---

## 14. Release scope

**v1.0** — HCL DSL (logical + deployment planes), compiler and IR, `moved` and semantic diff,
Terraform state/plan and CloudFormation adapters, import and reconcile with coverage, D2 (embedded)
and SVG/markdown/HTML renderers, native policy blocks with SARIF output, MCP read and write tools,
full CLI, JSON/TOON export.

**v1.1** — live AWS scan adapter; SARIF ingest from Checkov/tfsec mapped onto elements via bindings;
policies evaluated on the diff delta; mermaid flowchart backend.

**v1.2** — diagram import (D2, mermaid) as best-effort logical-plane bootstrap; mechanical `migrate`
if demanded.

### 14.1 Decision records — deferred, and cheaper than previously estimated

`research/adr-feature-design.md` (Feb 2026) designs ADRs for v0.3.0 as markdown files with YAML
frontmatter and MADR 4.0 templates, configured through `loko.toml`, estimated at 16–21 days.

Those mechanics are obsolete — frontmatter and `loko.toml` are both removed in v1 — but the intent
maps onto the v1 model at a fraction of the cost:

- A `decision` block with **typed references** to the elements it governs, and a typed `supersedes`
  reference, so supersede chains are compile-checked instead of matched on filenames.
- "Require decisions for certain C4 element patterns" is already expressible by the policy engine
  (§10.2): `require { select = select(...) ; assert = { decision = not_empty } }`. No new
  enforcement machinery.

Most of the original estimate was building validation and linking that v1.0 provides regardless.
Deferred because decision-record pain was explicitly not among the problems v1 is chartered to
solve (§1), not because it is expensive.

---

## 15. Risks

- **Grouping quality on import.** Heuristics may produce poor candidate groupings on projects with
  weak tagging and flat Terraform. Mitigation: import always writes to `--out` for human review, and
  the LLM refinement path is the intended default, not a fallback.
- **Reconcile noise.** If default ignore rules are too narrow, output is unusable and adoption fails.
  Mitigation: treat the default ignore list as a first-class deliverable tested against real fixtures,
  not an afterthought.
- **Policy expressiveness ceiling.** `select()` will not cover every rule users want. Mitigation: the
  OPA escape hatch is documented from day one, so the ceiling is never a hard stop.
- **Scope.** v1.0 is substantial. The `moved`/diff engine and the compiler are on the critical path;
  adapters and policy are parallelizable and can slip to v1.1 without invalidating the design.
