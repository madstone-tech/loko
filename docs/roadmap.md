# loko Roadmap

## Vision

One authored source for a software architecture, compiled like code: reviewed in pull requests,
diffed between revisions, checked against what is actually deployed, and readable by people and
assistants alike.

The detailed plan, with the specification seed for each stage, is
[specs/012-v1-architecture-dsl/roadmap.md](../specs/012-v1-architecture-dsl/roadmap.md).

---

## Shipped

| Stage | Feature | Delivers |
|---|---|---|
| Compiler core | [013](../specs/013-hcl-compiler-core/) | The HCL language, `validate`, `fmt`, `export` (JSON and TOON), byte-stable output |
| Renderers | [014](../specs/014-viewmodel-renderers/) | `build` and `serve`: D2, SVG, markdown and a site; automatic and declared views; theme overrides |
| Assistants | [015](../specs/015-mcp-hcl-authoring/) | MCP tools (`describe`, `query`, `validate`, `apply_edit`, `move`), `loko query`, the `moved` block |
| Rendering fidelity | [016](../specs/016-rendering-fidelity/) | `title`, `shape`, relationship `kind` and `tags`, view `direction`, and the opt-in ELK `layout` |

---

## Planned

### Semantic diff

`loko diff <revA>..<revB>` compiles two revisions and compares them by address, so a change reads as
*added*, *removed*, *renamed* (through `moved`), *rewired*, *attribute changed* or *rebound*, with a
blast radius for each. `loko changelog` renders the same diff as markdown, and a `diff` MCP tool
exposes it to assistants.

### Observation adapters

One adapter interface over Terraform state and plan JSON and CloudFormation (which covers CDK):

- `loko import` proposes HCL, with bindings filled in, from what is deployed. It always writes a
  new file and never merges over existing source.
- `loko reconcile --deployment <name>` compares the bindings against what is observed and reports
  undocumented, phantom, drifted and ambiguous resources, with a coverage percentage that CI can
  gate on. The `reconcile` block's `ignore` patterns keep the noise down.

### Policy engine

`policy` blocks with element rules (`require`, `deny`) and path rules (`deny_path`,
`require_path`), evaluated inside `loko validate` under the same severities and exit codes, with
SARIF output for code scanning. Resource-level configuration checks stay with tools such as
Checkov; `loko export --format json | conftest test -` already covers custom Rego.

---

## Under consideration

These came out of modelling real systems with loko and are not scheduled yet:

- **Dynamic views**: an ordered list of steps on a view, for request flows and sequences.
- **Reusable instance sets**, so that environments sharing the same instances are not typed twice.

Ideas and votes are welcome in
[GitHub Discussions](https://github.com/madstone-tech/loko/discussions).
