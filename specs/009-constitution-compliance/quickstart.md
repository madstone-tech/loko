# Quickstart: Running the structural-compliance audit

This is the developer-facing how-to for the constitution-compliance audit
introduced by feature 009.

## TL;DR

```bash
make audit-constitution
```

Exits `0` if the codebase is compliant, `1` if there are violations. JSON
report is written to `audit-report.json`; a human-readable summary goes to
stderr.

## What it checks

Three rule families, all encoded in
`specs/009-constitution-compliance/contracts/structural-rules.yaml`:

| Family | What it enforces | Where it applies |
|--------|------------------|------------------|
| **Layer-import rules** | Onion-architecture: entities ⊂ usecases ⊂ adapters ⊂ outer (cmd/mcp/api). Outer layer cannot import entities directly. | All `.go` files under `cmd/` and `internal/`. |
| **File-size budgets** | Use-case files ≤ 200 effective lines; entity files ≤ 300. | `internal/core/usecases/`, `internal/core/entities/`. |
| **Function-size budgets** | CLI handler functions ≤ 50 effective lines; MCP tool handler functions ≤ 30. | `cmd/`, `internal/mcp/tools/`. |

"Effective lines" = source lines after dropping blank lines, single- and
multi-line comments, the `package` declaration, and `import (...)` blocks.
Same rule the legacy `scripts/audit-constitution.sh` used.

## Reading a failure

```text
cmd/new.go:42 function-size: runScaffold 67 > 50 [rule: cli-handler-func-size]
```

Means: the function `runScaffold` in `cmd/new.go` (declared at line 42) has
67 effective lines, which exceeds the 50-line budget for CLI handler
functions. Fix: extract the body into one or more use cases under
`internal/core/usecases/scaffold_*.go` and have `runScaffold` call them.

```text
internal/core/usecases/build_docs.go:8 layer: github.com/madstone-tech/loko/cmd 'cmd' may not be imported here [rule: core/usecases]
```

Means: the use-case file `build_docs.go` (line 8 = the import block) imports
the `cmd` package. Use cases may import only entities. Fix: invert the
dependency — pass values from `cmd` *into* the use case as parameters or via
a port interface.

## Running locally

Three equivalent invocations:

```bash
# 1) Make target (recommended)
make audit-constitution

# 2) Direct binary
go run ./tools/archcheck \
  --rules=specs/009-constitution-compliance/contracts/structural-rules.yaml \
  --format=text

# 3) Watch mode for tight feedback while refactoring
make audit-constitution-watch   # uses fsnotify; re-runs on save
```

The redundant `depguard` check is part of `make lint`:

```bash
make lint    # runs golangci-lint, including the depguard layer rule
```

## Adding a legitimate exemption

The audit supports two exemption *categories* (encoded in the YAML rule file):

1. **Data-files**: basenames `schemas.go`, `registry.go`, `helpers.go`,
   `constants.go`, plus `*_cobra.go` Cobra flag-wiring and `*_test.go`. Already
   exempted from file-size rules.
2. **Generated files**: any file whose first non-blank line matches
   `// Code generated ... DO NOT EDIT.`. Already exempted from file-size
   rules.

Per-file ad-hoc exemptions are **not** supported. If a new file legitimately
needs to be larger than its budget, the resolution is one of:

- Refactor the file (most of the time).
- Add a new exemption *category* to `structural-rules.yaml` AND write an ADR
  in `docs/adr/` justifying the new category.
- Amend the constitution (`.specify/memory/constitution.md`) to relax the
  budget, with the amendment process described in the constitution itself.

The legacy `KNOWN_VIOLATIONS` allowlist in `scripts/audit-constitution.sh` is
removed at the end of the refactor; do not reintroduce per-file allowlist
entries.

## Amending the rules

The rule set is data, not code. To change a budget or add a layer:

1. Edit `specs/009-constitution-compliance/contracts/structural-rules.yaml`.
2. Run `make audit-constitution` and confirm zero violations on the current
   tree.
3. Sync `.golangci.yml` `depguard` config from the YAML rules
   (`tools/archcheck --emit-golangci-config > .golangci.depguard.yml` and
   merge — the audit step in CI fails if the two diverge).
4. Update `.specify/memory/constitution.md` if the change is consequential
   (any limit change, any layer added/removed).
5. Open a PR with rationale and migration impact assessment per the
   constitution's amendment process.

## CI integration

The audit runs in `.github/workflows/ci.yml` between `build` and `test`.
Failures appear as inline annotations on the PR Files-changed view (via
`::error file=...::` GitHub Actions commands) and as a downloadable
`audit-report.json` artefact.

## Frequently asked

**Q. My handler is 51 effective lines. Can I get a one-line waiver?**
A. No (see "Adding a legitimate exemption"). Either extract a helper or move
the logic to a use case. The 50-line limit is deliberately tight enough to
force the conversation about whether a function is doing too much.

**Q. The audit says my file is 312 effective lines but `wc -l` says 380.**
A. The audit drops blank lines, comments, the `package` decl, and `import
(...)` blocks. Run `tools/archcheck --explain <file>` to see the per-line
breakdown.

**Q. The audit and `golangci-lint depguard` disagree.**
A. `tools/archcheck` is authoritative. The `depguard` config is generated
from the same YAML; if they disagree, regenerate it
(`tools/archcheck --emit-golangci-config`). The CI job fails specifically
when the two disagree, so this never lands on `main`.

**Q. Where does the constitution itself live?**
A. `.specify/memory/constitution.md` (v1.1.0+). The YAML rule file is the
machine-consumable mirror; the markdown is the human-consumable canonical
text. They are kept in sync by review (and by the post-Phase-1 test that
parses both and asserts equivalence).
