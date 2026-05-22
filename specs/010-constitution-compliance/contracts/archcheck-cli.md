# Contract: `tools/archcheck` CLI

**Status**: Authoritative for feature 010-constitution-compliance.
**Binary location**: `tools/archcheck/archcheck` (built by `task audit-constitution`).

## Synopsis

```text
archcheck [flags]
```

## Purpose

Walks a Go module's source tree, applies the rule set defined in `structural-rules.yaml`, honours the entries in `.archcheck-suppressions.yaml`, and emits violation reports to stdout (text) or as JSON. Exit code is non-zero if any non-suppressed violation is found.

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--rules <path>` | string | `specs/010-constitution-compliance/contracts/structural-rules.yaml` | Path to the structural-rules YAML file (see contract `structural-rules.yaml`). |
| `--suppressions <path>` | string | `.archcheck-suppressions.yaml` | Path to the suppression file (see contract `suppression-file-schema.yaml`). |
| `--module-root <path>` | string | autodetected via nearest `go.mod` | Root of the Go module to audit. |
| `--include <glob>` | repeated | `**/*.go` | Limit the file scan to matching paths (relative to `--module-root`). |
| `--exclude <glob>` | repeated | `vendor/**`, `**/.git/**` | Skip matching paths. |
| `--format` | enum: `text \| json` | `text` | Output format. |
| `--quiet` | bool | `false` | In `text` mode, suppress the trailing summary line; useful for piping. |
| `--no-suppress` | bool | `false` | Ignore the suppression file entirely; report every violation. Useful for measuring "true" debt. |
| `--baseline <path>` | string | (none) | If set, compare the current violation set against the JSON report at `<path>` and exit non-zero **only** for violations not present in the baseline (used for incremental adoption; not the gating mode). |
| `--help` | bool | `false` | Print usage and exit 0. |
| `--version` | bool | `false` | Print build info and exit 0. |

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | No non-suppressed violations. |
| `1` | At least one non-suppressed violation found. |
| `2` | Configuration error (missing/invalid rules file, malformed suppression file, unknown rule name in a suppression). |
| `3` | I/O error (cannot read module root, cannot parse a Go source file). |
| `4` | Suppression file contains an entry whose `expires_on` is more than `suppression_max_expiry_days` in the future. |

## Output formats

### `text` (default)

One line per violation, sorted by `file`, then by `line` (or `function` name when line is N/A). Format:

```text
<file>[:<line>] [<rule>] <message>
```

Examples:

```text
cmd/new.go:147 [cli-handler-func-size] function `runNewProject` is 68 effective lines, limit is 50
internal/mcp/tools/build_docs.go [outer-no-entities] import "github.com/madstone-io/loko/internal/core/entities" forbidden in mcp_tool files
internal/core/usecases/build_docs.go [usecase-file-size] file is 247 effective lines, limit is 200
```

Trailing summary (unless `--quiet`):

```text
---
14 violations across 9 files. 2 suppressions applied (expire 2026-06-30, 2026-07-15).
```

### `json` (`--format=json`)

A top-level array of `ViolationReport` objects matching the JSON schema in `contracts/violation-report.schema.json`. Suppressed entries are **omitted** from the array but counted in a trailing comment? No — JSON output is pure data, suppressed entries are emitted under a separate top-level key `suppressed`, with the same per-object shape plus an additional `expires_on` field.

```json
{
  "violations": [
    {
      "rule": "cli-handler-func-size",
      "kind": "function_size",
      "file": "cmd/new.go",
      "function": "runNewProject",
      "line": 147,
      "measured": 68,
      "limit": 50,
      "message": "function `runNewProject` is 68 effective lines, limit is 50"
    }
  ],
  "suppressed": [
    {
      "rule": "entity-file-size",
      "kind": "file_size",
      "file": "internal/core/entities/graph_legacy.go",
      "measured": 312,
      "limit": 300,
      "expires_on": "2026-06-30",
      "owner": "@andhi"
    }
  ]
}
```

## Performance

- Single-threaded baseline; goroutine-per-file for AST parsing when the module has > 100 files.
- Memory bounded: AST for at most `N_CPU` files held in memory at once.
- Target end-to-end runtime on the loko repository: **< 10 s** (well within the spec's 30 s budget).

## Compatibility

- `archcheck` is built from the same `go.mod` as the main project (no separate module).
- The CLI surface defined here is stable for the lifetime of constitution v1.x; v2.x is allowed to break it with a corresponding constitution major bump.

## Use cases

### Local pre-commit run

```bash
task audit-constitution           # builds archcheck (cached) + runs with defaults
# or:
./tools/archcheck/archcheck --quiet
```

### CI run

```bash
./tools/archcheck/archcheck --format=json > archcheck-report.json
# CI job uploads archcheck-report.json as an artefact for download.
```

### Baseline mode (transitional only)

```bash
./tools/archcheck/archcheck --format=json --no-suppress > baseline.json   # one-time capture
./tools/archcheck/archcheck --baseline baseline.json                     # in CI: fail only on new violations
```

Baseline mode is **not** how this feature ships — it is documented as an emergency escape hatch in case the suppression mechanism proves insufficient. The default and recommended mode is full suppression-aware checking.
