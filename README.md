# 🪇 loko - Guardian of Architectural Wisdom

> _A compiler for software architecture._

**loko** compiles a [C4 model](https://c4model.com/) architecture from HCL. You describe the
architecture once, in `*.loko.hcl` files. Diagrams, markdown, a browsable site, graph queries, a
machine-readable export and an MCP server for AI assistants are all projections of the compiled
result.

[![Go Version](https://img.shields.io/github/go-mod/go-version/madstone-tech/loko)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/madstone-tech/loko)](https://github.com/madstone-tech/loko/releases)
[![License](https://img.shields.io/badge/license-BUSL--1.1-blue)](LICENSE)
[![CI](https://github.com/madstone-tech/loko/actions/workflows/ci.yml/badge.svg)](https://github.com/madstone-tech/loko/actions/workflows/ci.yml)

---

## ✨ What it does

| Command | What you get |
|---|---|
| `loko validate` | Every problem in the architecture, with file, line and column |
| `loko fmt` | Canonical formatting, like `terraform fmt` |
| `loko build` | D2 and SVG diagrams, markdown and a static site. No `d2` binary needed |
| `loko serve` | A live preview that rebuilds as you edit |
| `loko query` | Dependents, dependencies, paths, orphans and coupling |
| `loko export` | The compiled architecture as JSON or [TOON](https://toonformat.dev) |
| `loko mcp` | An MCP server: assistants describe, query, validate and edit the HCL ([guide](docs/mcp-integration.md)) |

---

## Why a compiler

When an architecture lives in several places (diagram files, wiki pages, frontmatter), they drift
apart, and nothing says which one is right. loko has one authored artefact, the HCL, and every
reference in it is checked at compile time. A broken relationship is an error with a position:

```
Error: Unresolvable reference
  on arch.loko.hcl:28:19:
  28 |     target      = container.ordrs_db
                         ^^^^^^^^^^^^^^^^^^
  No element "container.ordrs_db" is declared. Did you mean
  container.orders_db?
```

Because the source is plain text with stable addresses, an architecture can be reviewed in a pull
request, diffed between revisions, and edited by an assistant without hand-written diagrams going
stale.

---

## 🚀 Quick start

### Install

```bash
brew install --cask madstone-tech/tap/loko          # macOS and Linux
go install github.com/madstone-tech/loko@latest     # from source
docker run --rm -v "$PWD:/workspace" ghcr.io/madstone-tech/loko:latest build
```

Release archives for Linux, macOS and Windows are on the
[releases page](https://github.com/madstone-tech/loko/releases).

### Write an architecture

Create `arch.loko.hcl`:

```hcl
project "acme-payments" {
  description  = "Payment processing platform"
  loko_version = "~> 1.0"
}

person "customer" {
  description = "Pays for orders"

  uses "checkout" {
    target      = container.api
    description = "Pays with a card"
    technology  = "HTTPS"
  }
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = "platform-team"
}

container "api" {
  system     = system.payments
  title      = "Payments API"
  technology = "AWS Lambda (Go)"
  shape      = "function"

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
  }

  uses "settle" {
    target      = container.settlements
    description = "Queues captured payments"
    kind        = "async"
  }
}

container "orders_db" {
  system     = system.payments
  technology = "Aurora PostgreSQL"
  shape      = "database"
}

container "settlements" {
  system     = system.payments
  technology = "Amazon SQS"
  shape      = "queue"
}
```

### Check it, draw it, ask it

```bash
loko validate                              # errors and warnings, with positions
loko build                                 # diagrams, markdown and a site in ./dist
loko serve                                 # preview at http://localhost:8080, live reload
loko query dependents container.orders_db  # who breaks if the database is down?
loko export --format json                  # the compiled model, for programs and CI
```

Exit codes are `0` clean, `1` errors, `2` warnings under `--strict`. There are exactly three, so a
CI pipeline can branch on them.

The [quick start guide](docs/quickstart.md) goes further, and
**[docs/language.md](docs/language.md)** is the full reference: every block, attribute and
function, with a complete worked example. [`examples/`](examples/) has complete projects.

---

## 📚 Core concepts

**Two planes.** The *logical* plane is environment-agnostic C4: `person`, `system`, `container`,
`component`, `external`. The *deployment* plane instantiates it per environment: `deployment`, an
optional nested `node` tree, and `instance` blocks carrying attributes and bindings to real
infrastructure such as Terraform addresses.

**Relationships nest inside their source.** A `uses "orders"` block inside `container "api"` has the
stable address `container.api.uses.orders`. `kind = "async"` or `"trigger"` changes how it is
drawn, never what depends on what.

**References are typed and resolved at compile time.** `container.db` is a reference, not a string.
A typo is an error with a source range and usually a suggestion.

**Views are generated, and you can declare more.** Every system and container with children gets a
diagram automatically. A `view` block selects elements by reference or tag for a focused picture.

**Prose lives beside the model.** `docs = "./docs/api.md"` attaches markdown to an element; it
appears on the element's site page.

**Output is byte-stable.** The same source always produces identical bytes, on any machine, in any
file-system order, so a committed export is reviewable in a diff.

**The compiler is stateless.** No lock file, no state file, no database. Git is the history.

---

## 🤖 Assistants

`loko mcp` serves five tools over stdio: `describe`, `query`, `validate`, `apply_edit` and `move`.
An assistant can model an existing system from its code and infrastructure, or design a new one
conversationally. Every edit compiles before it is written, and saves all-or-nothing, so a broken
reference never reaches disk. Output defaults to [TOON](docs/guides/toon-format-guide.md), which
roughly halves the size of tabular answers such as query results. See
**[docs/mcp-integration.md](docs/mcp-integration.md)** for setup with Claude Code, Claude Desktop
and other MCP clients.

---

## 🔧 Development

```bash
task build              # or: make build
task test               # go test ./...
task lint               # golangci-lint, including depguard layer rules
task audit-constitution # layer rules and size budgets
```

The [constitution](.specify/memory/constitution.md) is enforced mechanically, not by review:
`tools/archcheck` checks layer-import rules and file and function size budgets on every PR (CLI
handler ≤ 50 lines, MCP handler ≤ 30, use-case and entity files ≤ 200, adapter files ≤ 400). The
HCL, cty, D2 and goldmark libraries are confined to `internal/adapters/**`; the compiler core
cannot see them. Start with the one-page
[constitution compliance reference](docs/architecture/constitution-compliance.md).

Architecture decisions live in [`docs/adr/`](docs/adr/).
[ADR-0012](docs/adr/0012-hcl-source-of-truth.md) explains why the source is HCL.

---

## 📖 Documentation

| Document | What it covers |
|---|---|
| [docs/quickstart.md](docs/quickstart.md) | From an empty directory to a rendered site |
| [docs/language.md](docs/language.md) | The complete `*.loko.hcl` language |
| [docs/cli-reference.md](docs/cli-reference.md) | Every command and flag |
| [docs/mcp-integration.md](docs/mcp-integration.md) | The MCP server and its tools |
| [docs/README.md](docs/README.md) | Everything else: guides, ADRs, roadmap |

---

## 🗺️ Roadmap

| Stage | Delivers | Status |
|---|---|---|
| Compiler core | `validate`, `fmt`, `export` | ✅ |
| Renderers | `build`, `serve`: D2, SVG, markdown, site | ✅ |
| Assistants | MCP tools, `loko query`, HCL authoring | ✅ |
| Rendering fidelity | titles, shapes, relationship kinds, layout direction and engine | ✅ |
| Semantic diff | `diff`, `changelog`, blast radius | planned |
| Observation adapters | `import` and `reconcile` against Terraform and CloudFormation | planned |
| Policy engine | architecture-level rules, SARIF output | planned |

Details in [docs/roadmap.md](docs/roadmap.md).

---

## 🤝 Contributing

loko is built in public. Bugs and ideas are welcome in
[issues](https://github.com/madstone-tech/loko/issues) and
[discussions](https://github.com/madstone-tech/loko/discussions); pull requests are described in
[CONTRIBUTING.md](CONTRIBUTING.md). Before opening one, run `task lint`, `task test` and
`task audit-constitution`; the last is a required check on `main` and runs in well under a second.

---

## 📜 License

[Business Source License 1.1](LICENSE) - Copyright (c) 2025-2026 MADSTONE TECHNOLOGY

---

## 🙏 Acknowledgments

**loko** builds on excellent open-source tools:

- [C4 Model](https://c4model.com) - Architecture visualization approach
- [HCL](https://github.com/hashicorp/hcl) - The configuration language the source is written in
- [D2](https://d2lang.com) - Diagram layout and rendering, embedded
- [TOON](https://toonformat.dev) - Token-efficient notation
- [Cobra](https://github.com/spf13/cobra) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) - CLI and terminal styling
- [Goldmark](https://github.com/yuin/goldmark) - Markdown for the site

---

## 🪇 Why "loko"?

**Papa Loko** is the lwa (spirit) in Haitian Vodou who guards sacred knowledge, maintains tradition, and passes down wisdom to initiates. As the first houngan (priest), he is the keeper of the ritual knowledge that connects the physical and spiritual worlds.

Like Papa Loko, this tool acts as the guardian of your architectural wisdom — organizing knowledge, maintaining documentation traditions, and making complex systems understandable.

_"Papa Loko, you're the wind, pushing us, and we become butterflies."_ 🦋

---
