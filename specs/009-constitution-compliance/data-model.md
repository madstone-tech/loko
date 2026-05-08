# Data Model: Constitution Compliance Refactor

This feature does not introduce any user-facing data. The "data" of interest is the **rule set** that the audit tool consumes and the **violation report** it emits. This document specifies their shape so that `tools/archcheck`, `.golangci.yml`, and `contracts/structural-rules.yaml` stay in lock-step.

## Logical entities

### `LayerRule`

A constraint stating which packages a layer is allowed to import.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Human-readable layer name, e.g., `"core/entities"`. |
| `pathPattern` | glob | Module-relative path pattern, e.g., `"internal/core/entities/**"`. Multiple patterns allowed; first match wins. |
| `allowedImports` | list of glob | Module-relative globs that files in this layer **may** import (within the project). External imports (anything not under the project module path) are always allowed. |
| `forbiddenImports` | list of glob | Optional explicit deny-list, evaluated *after* `allowedImports`. Empty list means "everything not in `allowedImports` is forbidden." |
| `description` | string | Free-text rule description (used in violation messages). |

**Validation rules**:

- A file matched by exactly one `pathPattern` (first match wins).
- A file matched by no `pathPattern` is unconstrained (no import rule applies). The audit emits an info-level message listing such files; CI does not fail.
- `allowedImports` is module-relative; it never lists external packages.
- The five layer rules collectively MUST encode the constitution table verbatim:

| Layer (`name`) | `pathPattern` | `allowedImports` |
|----------------|---------------|------------------|
| `core/entities` | `internal/core/entities/**` | (empty — stdlib only) |
| `core/usecases` | `internal/core/usecases/**` | `internal/core/entities/**` |
| `adapters` | `internal/adapters/**` | `internal/core/entities/**`, `internal/core/usecases/**` |
| `mcp` | `internal/mcp/**` | `internal/core/**`, `internal/adapters/**` |
| `api` | `internal/api/**` | `internal/core/**`, `internal/adapters/**` |
| `cmd` | `cmd/**` | `internal/core/**`, `internal/adapters/**`, `internal/mcp/**`, `internal/api/**` |

### `FileSizeRule`

A maximum effective-line budget that applies to all files matching a pattern.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Rule label, e.g., `"usecase-file-size"`. |
| `pathPattern` | glob | Files this rule applies to. |
| `maxEffectiveLines` | integer | Inclusive upper bound. |
| `description` | string | Free-text description (used in violation messages). |

**Required instances** (encoded in `structural-rules.yaml`):

| `name` | `pathPattern` | `maxEffectiveLines` |
|--------|---------------|---------------------|
| `usecase-file-size` | `internal/core/usecases/**/*.go` | 200 |
| `entity-file-size` | `internal/core/entities/**/*.go` | 300 |

(CLI and MCP files have no whole-file budget under v1.1.0; they have *function*-level budgets via `FunctionSizeRule`.)

### `FunctionSizeRule`

A maximum effective-line budget applied per function declaration.

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Rule label, e.g., `"cli-handler-func-size"`. |
| `pathPattern` | glob | Files in scope. |
| `maxEffectiveLines` | integer | Inclusive upper bound, applied per function declaration. |
| `description` | string | Free-text description. |

**Required instances**:

| `name` | `pathPattern` | `maxEffectiveLines` |
|--------|---------------|---------------------|
| `cli-handler-func-size` | `cmd/**/*.go` | 50 |
| `mcp-tool-func-size` | `internal/mcp/tools/**/*.go` | 30 |

### `Exemption`

A pattern that excludes a file from a class of rules.

| Field | Type | Description |
|-------|------|-------------|
| `kind` | enum | One of: `"file-size"`, `"function-size"`. (Layer rules cannot be exempted.) |
| `match` | matcher | One of: `{ basename: [...] }`, `{ pathPattern: <glob> }`, `{ generatedHeader: true }`. |
| `reason` | string | Why this category is exempt (used in audit info output). |

**Required instances** (encoded in `structural-rules.yaml`):

| `kind` | `match` | `reason` |
|--------|---------|----------|
| `file-size` | `basename: [schemas.go, registry.go, helpers.go, constants.go]` | Pure-data files; not logic. |
| `file-size` | `pathPattern: "**/*_cobra.go"` | Cobra flag-wiring; data-shaped declarations. |
| `file-size` | `pathPattern: "**/*_test.go"` | Tests are exempt from production-code budgets. |
| `function-size` | `pathPattern: "**/*_test.go"` | Same. |
| `file-size` | `generatedHeader: true` | Files with `// Code generated ... DO NOT EDIT.` header. |

### `Violation`

A single problem found by the audit. Emitted in the JSON report and the human text output.

| Field | Type | Description |
|-------|------|-------------|
| `rule` | string | The `name` of the rule that fired (e.g., `"cli-handler-func-size"`). |
| `kind` | enum | `"layer"`, `"file-size"`, `"function-size"`. |
| `file` | string | Module-relative path. |
| `line` | integer | Line where the violation is anchored (function declaration line for `function-size`; first line of file for `file-size`; import statement line for `layer`). |
| `subject` | string | The offending entity: function name for `function-size`; import path for `layer`; file basename for `file-size`. |
| `actual` | integer or string | Measured value: effective-line count for size rules; the offending import path for layer rules. |
| `limit` | integer or string | The budget or the rule description. |
| `message` | string | Pre-formatted human-readable message: `"<file>:<line>: <subject> exceeds <rule> (<actual> > <limit>)"` for size rules; `"<file>:<line>: layer '<name>' may not import '<subject>' (rule: <description>)"` for layer rules. |

### `Report`

The top-level output document.

```yaml
version: "1.0"
generated_at: "<ISO-8601 UTC>"
audit_tool_version: "<archcheck git SHA>"
rules_path: "specs/009-constitution-compliance/contracts/structural-rules.yaml"
total_files_scanned: <int>
total_functions_scanned: <int>
violations: [<Violation>, ...]
exit_code: 0 | 1
```

`exit_code: 0` iff `violations` is empty. The CI step's exit code mirrors `exit_code`.

## Relationships

```
structural-rules.yaml
  ├── layers[]         → LayerRule
  ├── fileSizes[]      → FileSizeRule
  ├── functionSizes[]  → FunctionSizeRule
  └── exemptions[]     → Exemption

archcheck (binary)
  reads:  structural-rules.yaml + Go source tree
  writes: audit-report.json (Report)

.golangci.yml (depguard)
  derived from: structural-rules.yaml.layers
  enforced by: golangci-lint as redundant secondary check
```

## State transitions

None. The rules are static configuration; there is no entity with lifecycle state. `Violation` instances exist only within a single audit run.

## Validation summary

| Validation | Where enforced |
|------------|----------------|
| Every layer in the constitution table has exactly one `LayerRule`. | `archcheck` startup self-check; refuses to run if rules are incomplete. |
| Every file under `internal/`, `cmd/` is matched by exactly one `LayerRule`. | `archcheck` reports unmatched files at `info` level. |
| `maxEffectiveLines` ≥ 1 for all size rules. | YAML schema validation in `archcheck`. |
| Exemption `match.basename` entries are bare basenames (no path separators). | YAML schema validation. |
| `.golangci.yml` `depguard` rules are a literal subset of `structural-rules.yaml` `layers`. | `make audit-constitution` runs an extra check that fails if the two diverge. |
