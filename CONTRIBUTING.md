# Contributing to loko

Thank you for your interest in loko. It is built in public, and contributions of every size are
welcome: bug reports, ideas, documentation fixes, example architectures and code.

## 🎯 Ways to contribute

- 🐛 **Report bugs**: a small `*.loko.hcl` that reproduces the problem is the most useful report.
- 💡 **Suggest features**: start in [Discussions](https://github.com/madstone-tech/loko/discussions).
- 📖 **Improve documentation**: clarify, correct or extend `docs/`.
- 🧪 **Model a real system**: try loko on an architecture you know and report what it could not
  express or draw well. Most of the rendering features came from exactly that.
- 🔧 **Submit code**: bug fixes, features, tests.

## 🚀 Getting started

### Prerequisites

- **Go 1.27+** ([install](https://go.dev/doc/install)). Nothing else is needed to build or test:
  diagrams render in-process, so there is no `d2` to install.
- **git**
- Optional: [Task](https://taskfile.dev) (every target also exists in the `Makefile`),
  `golangci-lint`, and `goreleaser` for release snapshots. `task tools` installs the last two.

### Setup

```bash
git clone https://github.com/madstone-tech/loko
cd loko
go test ./...
go build -o loko .
./loko validate -p testdata/projects/serverless-reference
```

## 🏗️ Architecture

loko is a compiler with **Clean Architecture**, and its
[constitution](.specify/memory/constitution.md) is enforced mechanically by `tools/archcheck` and
golangci-lint's `depguard` rules, not by review.

```
┌───────────────────────────────────────────────────────────────┐
│  cmd/ (cobra CLI)          internal/mcp/ (MCP server, tools)   │  ← thin handlers
├───────────────────────────────────────────────────────────────┤
│  internal/adapters/   hclsource, d2, html, markdown, encoding, │  ← implement ports
│                       projectfs, outputdir, devserver, watch   │
├───────────────────────────────────────────────────────────────┤
│  internal/core/usecases/   compile, resolve, project views,    │  ← define ports
│                            query, describe, edit                │
├───────────────────────────────────────────────────────────────┤
│  internal/core/entities/   arch (IR), viewmodel, authoring     │  ← stdlib only
└───────────────────────────────────────────────────────────────┘
```

The rules that matter most:

- **Entities** import only the standard library, and no entity package imports another.
- **Use cases** hold the business logic and declare the ports they need in
  `internal/core/usecases/ports.go` (`ArchitectureSource`, `SourceEditor`, `Backend`,
  `ArtifactStore` and others).
- **Adapters** implement those ports. HCL, cty, D2 and goldmark may only be imported here.
- **`cmd/` and `internal/mcp/`** never import entities; they parse input, call a use case and
  format its result. Wiring lives in `main.go` or `cmd/`.
- **Size budgets**: CLI handlers ≤ 50 lines, MCP handlers ≤ 30, use-case functions ≤ 60, use-case
  and entity files ≤ 200, adapter files ≤ 400. Tests are exempt.

The one-page [constitution compliance reference](docs/architecture/constitution-compliance.md)
explains each rule and how to read an archcheck failure.

### Where code goes

| I want to... | Where |
|---|---|
| Add a language attribute or block | `internal/adapters/hclsource/` (schema, decode), the field in `internal/core/entities/arch/`, validation in `internal/core/usecases/` |
| Add a diagnostic | the code in `internal/core/entities/arch/diagnostic.go`, emitted from a use case |
| Change what a view contains | `internal/core/usecases/` (`resolve_views.go`, `project_*.go`) |
| Change how a diagram looks | `internal/adapters/d2/` (emit) |
| Change the site or markdown | `internal/adapters/html/`, `internal/adapters/markdown/` |
| Add or change an MCP tool | `internal/mcp/tools/`, calling a use case |
| Add a CLI command | `cmd/`, calling a use case |
| Change what assistants may edit | `internal/core/entities/authoring/` and `internal/adapters/hclsource/edit_*.go` |

A language change is permanent v1.x surface (see
[compatibility commitments](docs/language.md#compatibility-commitments)), so discuss it in an issue
first and document it in `docs/language.md`.

## 🧪 Testing

- **Table-driven unit tests** next to the code, with `t.Parallel()`.
- **Golden files** for projections, D2, the site, markdown and exports. After an intended output
  change, regenerate and review the diff:

  ```bash
  go test ./cmd/ ./internal/adapters/{d2,hclsource,html,markdown}/ -update
  git diff testdata/
  ```

- **Fixture projects** in `testdata/projects/` exercise the compiler end to end. Add one when a
  feature needs a realistic architecture.
- **Property and fuzz tests** cover the HCL editor: any sequence of edits must round-trip.
- **Goroutine leaks fail tests**: packages that start goroutines run the Go 1.27 `goroutineleak`
  check in `TestMain`.

## 🔧 Workflow

1. **Branch** from `main`: `feat/…`, `fix/…`, `docs/…` or `chore/…`.
2. **Larger features** follow the spec workflow in `specs/NNN-name/` (spec, plan, tasks), driven by
   the Spec Kit commands in `.specify/`. Small fixes don't need one.
3. **Check** before pushing:

   ```bash
   task lint                 # gofmt, vet, golangci-lint including depguard layer rules
   task test                 # the full suite
   task audit-constitution   # layer rules and size budgets; a required check on main
   ```

4. **Commit** with [Conventional Commits](https://www.conventionalcommits.org/):
   `feat(render): …`, `fix(mcp): …`, `docs: …`. Release notes are grouped from these prefixes.
5. **Open a PR** with a clear description. Pull requests are rebase-merged.

## 📋 Pull request checklist

- [ ] `task lint`, `task test` and `task audit-constitution` pass
- [ ] New behaviour has tests; changed output has reviewed golden diffs
- [ ] `docs/` updated for user-visible changes, and `docs/language.md` for language changes
- [ ] `CHANGELOG.md` has an entry under `[Unreleased]`
- [ ] Significant design decisions have an ADR in `docs/adr/`

## 🐛 Reporting bugs

Include:

- `loko --version`, your operating system, and how you installed loko
- the smallest `*.loko.hcl` that reproduces the problem
- the command you ran, what you expected, and what happened (`--verbose` output helps)

## 📦 Dependencies

loko keeps dependencies few. Before adding one, check whether the standard library can do it,
whether the module is maintained, and what it adds to the binary, and open an issue to discuss it.
New third-party imports belong in `internal/adapters/**` only.

## 🏛️ Architecture decisions

Significant decisions are recorded in [`docs/adr/`](docs/adr/). The most relevant for v1:

- [ADR-0012](docs/adr/0012-hcl-source-of-truth.md): HCL as the single source of truth
- [ADR-0013](docs/adr/0013-viewmodel-renderers.md): outputs are projections of a view model
- [ADR-0014](docs/adr/0014-hcl-authoring.md): assistants edit the HCL
- [ADR-0015](docs/adr/0015-rendering-attributes.md): rendering attributes and layout engines

## 🚢 Releases

Maintainers release by pushing a `vX.Y.Z` tag; GoReleaser publishes archives, the
`ghcr.io/madstone-tech/loko` image and the Homebrew cask. `task release-snapshot` builds everything
locally without publishing.

## ❓ Questions

- **General questions** → [GitHub Discussions](https://github.com/madstone-tech/loko/discussions)
- **Bug reports** → [GitHub Issues](https://github.com/madstone-tech/loko/issues)
- **Security issues** → email <security@madstone.tech>

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By
participating, you agree to uphold it.

---

**Thank you for contributing to loko!** 🪇
