# 🪇 loko - Guardian of Architectural Wisdom

> _A compiler for software architecture._

**loko** compiles a [C4 model](https://c4model.com/) architecture from HCL. You describe the
architecture once, in `*.loko.hcl` files; everything else is a projection of the compiled result.

[![Go Version](https://img.shields.io/github/go-mod/go-version/madstone-tech/loko)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/madstone-tech/loko)](https://github.com/madstone-tech/loko/releases)
[![License](https://img.shields.io/badge/license-BUSL--1.1-blue)](LICENSE)
[![CI](https://github.com/madstone-tech/loko/actions/workflows/ci.yml/badge.svg)](https://github.com/madstone-tech/loko/actions/workflows/ci.yml)

---

## ⚠️ v1 is mid-rewrite

v1 replaces the v0.2 data model. The compiler, validation, export and rendering have landed.

| Working now | Returns in the next release |
|---|---|
| `loko validate`, `loko fmt`, `loko export` | `loko init` |
| `loko build` — D2, SVG, markdown and a site, no configuration, no `d2` binary needed | MCP read and write tools |
| `loko serve` — live preview with reload | |
| `loko mcp` (starts; registers no tools yet) | |

`loko new` and `loko api` are gone for good. `loko validate --check-drift` is gone because drift
cannot occur any more — see below.

For a working v0.2 tool, install the `v0.2.x` tag. See
[Migrating from v0.2](docs/language.md#migrating-from-v02).

---

## Why a compiler

v0.2 had two sources of truth. Relationships were authored in markdown frontmatter *and* in D2 arrow
syntax, then merged at build time, with a `--check-drift` flag to report when they disagreed.

Drift detection is not a feature. It is the symptom of a model with no authoritative layer. Because
nothing was authoritative, edits went stale silently, there was no way to diff an architecture
between two revisions, and there was nothing well-defined to compare a deployed environment against.

v1 makes the HCL the only authored artefact. A broken relationship is a compile error with a file, a
line, and a column:

```
Error: Unresolvable reference
  on arch.loko.hcl:19:19:
  19 |     target      = container.ordrs_db
                         ^^^^^^^^^^^^^^^^^^
  No element "container.ordrs_db" is declared. Did you mean container.orders_db?
```

---

## 🚀 Quick start

### Install

```bash
brew tap madstone-tech/tap && brew install loko    # macOS / Linux
go install github.com/madstone-tech/loko@latest    # from source
```

### Write an architecture

There is no `loko init` in this release. Create `arch.loko.hcl`:

```hcl
project "acme-payments" {
  description  = "Payment processing platform"
  loko_version = "~> 1.0"
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = "platform-team"
}

container "api" {
  system     = system.payments
  technology = "AWS Lambda (Go)"

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
  }
}

container "orders_db" {
  system     = system.payments
  technology = "Aurora PostgreSQL"
}
```

### Check it

```bash
loko validate                  # every problem, with file/line/column
loko validate --strict         # warnings fail too
loko validate --format json    # machine-readable, for CI
loko fmt                       # canonical formatting
loko fmt --check               # fail CI on unformatted source
loko export --format json      # the compiled architecture
```

Exit codes are `0` clean, `1` errors, `2` warnings under `--strict`. Exactly three, so a CI pipeline
can branch on them without learning new ones.

The full grammar is in **[docs/language.md](docs/language.md)** — every block, attribute and
function, with a complete worked example.

---

## 📚 Core concepts

**Two planes.** The *logical* plane is environment-agnostic C4: `person`, `system`, `container`,
`component`, `external`. The *deployment* plane instantiates it per environment: `deployment`, an
optional nested `node` tree, and `instance` blocks carrying attributes and bindings to real
infrastructure.

**Relationships nest inside their source.** A `uses "orders"` block inside `container "api"` has the
stable address `container.api.uses.orders`. That address is what will let a future diff report
*rewired* rather than "one edge vanished and another appeared".

**References are typed and resolved at compile time.** `container.db` is a reference, not a string.
A typo is an error with a source range and usually a suggestion.

**Cycles between elements are legal.** `api → queue → worker → api` is a normal architecture, not a
mistake. Only *containment* must form a single-parent tree.

**Instance identity is independent of placement.** `deployment.prod.instance.api` carries no node
path, so moving an instance between subnets preserves its identity.

**Output is byte-stable.** The same source always produces identical bytes, on any machine, in any
file-system order. That is what makes a committed export reviewable in a diff.

**The compiler is stateless.** No lock file, no state file, no database. Git is the history.

---

## 💰 Token efficiency

`loko export --format toon` emits [TOON](https://github.com/toon-format/toon-go), which runs ~26%
smaller than the equivalent JSON on real architectures, using tabular arrays for uniform data. Both
encodings carry equivalent information — a test asserts every key present in one appears in the
other.

---

## 🔧 Development

```bash
task build              # or: make build
task test               # go test ./...
task lint               # golangci-lint
task audit-constitution # layer rules and size budgets
```

The [constitution](.specify/memory/constitution.md) is enforced mechanically, not by review:
`tools/archcheck` checks layer-import rules and file/function size budgets on every PR. The HCL,
cty, and d2 libraries are confined to `internal/adapters/**` — the compiler core cannot see them.

Architecture decisions live in [`docs/adr/`](docs/adr/).
[ADR-0012](docs/adr/0012-hcl-source-of-truth.md) covers the v1 rewrite and the alternatives that
were rejected.

`internal/_parked/` holds code retained for later stages. It does not compile and is not part of the
build; see the README there.

---

## 📖 Documentation

| Document | What it covers |
|---|---|
| [docs/language.md](docs/language.md) | The complete `*.loko.hcl` grammar |
| [ADR-0012](docs/adr/0012-hcl-source-of-truth.md) | Why HCL, and what was rejected |
| [Constitution](.specify/memory/constitution.md) | Architecture rules and budgets |
| [specs/012-v1-architecture-dsl/](specs/012-v1-architecture-dsl/) | The v1 design and roadmap |

---

## 🗺️ Roadmap

| Stage | Delivers | Status |
|---|---|---|
| Compiler core | `validate`, `fmt`, `export` | ✅ landed |
| Renderers | `build`, `serve`, D2/SVG/markdown/HTML | next |
| MCP rewire | conversational authoring against the IR | planned |
| Semantic diff | `diff`, `changelog`, blast radius | planned |
| Observation adapters | `import`, `reconcile` against Terraform and CloudFormation | planned |
| Policy engine | architecture-level rules, SARIF output | planned |

---

## 🤝 Contributing

We welcome contributions! loko is **building in public** — see our [development progress](https://github.com/madstone-tech/loko/issues).

- 🐛 **Bug reports** → [Open an issue](https://github.com/madstone-tech/loko/issues/new?template=bug_report.md)
- 💡 **Feature requests** → [Start a discussion](https://github.com/madstone-tech/loko/discussions/new?category=ideas)
- 🔧 **Pull requests** → See [CONTRIBUTING.md](CONTRIBUTING.md)

### Quality gates

loko enforces its [Clean Architecture constitution](.specify/memory/constitution.md)
mechanically. Before opening a PR, run:

```bash
task lint                 # gofmt, vet, golangci-lint (incl. depguard layer rules)
task test                 # full unit + integration suite
task audit-constitution   # structural-compliance gate (file/function-size + layer-import rules)
```

`task audit-constitution` runs in well under a second and is a **required check** on `main`.
It enforces four budgets (CLI handler ≤ 50 lines, MCP handler ≤ 30, use-case file ≤ 200,
entity file ≤ 300) and the layer-import rules. New contributors: start with the one-page
[Constitution Compliance reference](docs/architecture/constitution-compliance.md) and the
feature [quickstart](specs/010-constitution-compliance/quickstart.md).

---

## 📜 License

[Business Source License 1.1](LICENSE) - Copyright (c) 2025-2026 MADSTONE TECHNOLOGY

---

## 🙏 Acknowledgments

**loko** builds on excellent open-source tools:

- [D2](https://d2lang.com) - Declarative diagramming
- [ason](https://github.com/madstone-tech/ason) - Template scaffolding
- [TOON](https://toonformat.dev) - Token-efficient notation
- [C4 Model](https://c4model.com) - Architecture visualization approach
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework

---

## 🪇 Why "loko"?

**Papa Loko** is the lwa (spirit) in Haitian Vodou who guards sacred knowledge, maintains tradition, and passes down wisdom to initiates. As the first houngan (priest), he is the keeper of the ritual knowledge that connects the physical and spiritual worlds.

Like Papa Loko, this tool acts as the guardian of your architectural wisdom — organizing knowledge, maintaining documentation traditions, and making complex systems understandable.

_"Papa Loko, you're the wind, pushing us, and we become butterflies."_ 🦋

---
