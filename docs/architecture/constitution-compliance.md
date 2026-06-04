# Constitution Compliance — Contributor Reference

One page on the structural rules loko enforces mechanically, and how to work with them.
The canonical, machine-readable rule set lives at **`tools/archcheck/rules.yaml`**; the
constitution prose at `.specify/memory/constitution.md` (v1.2.0) is authoritative and the
two are kept in sync by `scripts/check-rules-sync.sh` (run in CI).

## The four budgets

| Budget | Limit (effective lines) | Applies to | Rule name |
|--------|-------------------------|------------|-----------|
| CLI handler function | **50** | `cmd/**/*.go` functions | `cli-handler-func-size` |
| MCP tool handler function | **30** | `internal/mcp/tools/**/*.go` functions | `mcp-tool-func-size` |
| Use-case file | **200** | `internal/core/usecases/**/*.go` files | `usecase-file-size` |
| Entity file | **300** | `internal/core/entities/**/*.go` files | `entity-file-size` |

"Effective lines" excludes blank lines and comment-only lines. If a handler exceeds its
budget, the domain logic belongs in a use case — extract it, don't reformat to squeak under.

## Layer-import rules (dependency direction)

Inner layers never import outer layers. Each file's allowed imports (module-relative):

| Layer (`pathPattern`) | May import | Must NOT import |
|-----------------------|------------|-----------------|
| `internal/core/entities/**` | stdlib only | anything `internal/**` |
| `internal/core/usecases/**` | entities, sibling usecases, stdlib | adapters, mcp, api, cmd |
| `internal/adapters/**` | core, sibling adapters | mcp, api, cmd |
| `internal/mcp/**` | core/usecases, adapters, own sub-pkgs | **`internal/core/entities/**`** (v1.2.0) |
| `internal/api/**` | core/usecases, adapters, own sub-pkgs | **`internal/core/entities/**`** (v1.2.0) |
| `cmd/**` | any internal layer | **`internal/core/entities/**`** |

The outer three entry-points (`cmd`, `mcp`, `api`) must obtain entity types through
**use-case return values or adapter outputs**, never by importing the entity package
directly. This tightening landed with constitution **v1.2.0** (see
`docs/adr/0010-tighten-outer-layer-entity-import-rule.md`). Layer rules are **never**
exemptible — every Go file is subject to them.

## Exemptions

Size budgets (not layer rules) are waived for, per `rules.yaml`:

- `*_test.go` — tests.
- `*_cobra.go` — cobra flag-wiring is declarative setup, not a handler.
- `schemas.go`, `registry.go`, `helpers.go`, `constants.go` — pure-data / shared-helper files.
- Generated files (`// Code generated ... DO NOT EDIT.`).

## Running the check

```bash
task audit-constitution          # build + run archcheck against the repo (CI gate)
go run ./tools/archcheck         # same, ad-hoc
go run ./tools/archcheck --format=json --report archcheck-report.json
go run ./tools/archcheck --emit-golangci-config   # depguard fast-path config
```

Exit codes: `0` clean · `1` violations found · `2` config/IO error · `3` suppression IO
error · `4` invalid suppression file. The check runs in well under a second.

## Suppression workflow (last resort)

For genuinely out-of-scope, pre-existing violations only. Add an entry to
`.archcheck-suppressions.yaml` at the repo root:

```yaml
- rule: entity-file-size          # a rule name from rules.yaml
  file: internal/foo/legacy.go    # repo-relative path or glob (must match a real file)
  function: someHandler           # required only for per-function rules
  owner: "@your-handle"           # GitHub handle, must start with @
  expires_on: "2026-08-01"        # ISO date, ≤ 90 days out (longer needs an ADR)
  reason: "Why this is deferred; reference the follow-up issue/PR (≥ 20 chars)."
```

Rules:
- **Max 90 days.** Longer-lived suppressions are rejected at load (exit 4) and require an ADR.
- **Expiry is a failure.** After `expires_on`, the underlying violation re-fires.
- **No rotting.** If the `file` no longer matches anything, the load fails.
- If your `reason` can't name a specific follow-up, **fix the violation instead.**

## Where things live

- Rules (canonical, machine-readable): `tools/archcheck/rules.yaml`
- Constitution (prose, authoritative): `.specify/memory/constitution.md`
- Audit binary: `tools/archcheck/`
- Suppressions: `.archcheck-suppressions.yaml`
- Redundant lint fast-path: `depguard` block in `.golangci.yml`
- CI gate: the `Audit constitution` step in `.github/workflows/ci.yml`
- ADR for the v1.2.0 tightening: `docs/adr/0010-tighten-outer-layer-entity-import-rule.md`
