# Contract: CI gating step `audit-constitution`

This contract defines the obligations that the new CI step in
`.github/workflows/ci.yml` must satisfy. It is referenced by FR-013, FR-014,
SC-004, and SC-007.

## 1. Invocation

The step runs `make audit-constitution`, which in turn runs:

```bash
go run ./tools/archcheck \
  --rules=specs/009-constitution-compliance/contracts/structural-rules.yaml \
  --format=json \
  --report=audit-report.json \
  --annotate=github
```

A second sub-step runs `golangci-lint run` so that the redundant `depguard`
import-rule check fires inside the same job. Both must pass for the step to
succeed.

The step is sequenced **after** `make build` (so any compile error fails
earlier and faster) and **before** `make test` (so a structural problem fails
fast without paying the test-suite latency).

## 2. Exit codes

| Exit | Meaning |
|------|---------|
| `0` | Zero violations. CI step passes. |
| `1` | One or more violations. CI step fails and blocks merging. |
| `2` | Internal error: rules file missing, malformed YAML, AST parse failure, etc. CI step fails. |

## 3. Output

### 3.1 Structured report (machine-readable)

Written to `audit-report.json` at the repository root. Schema = `Report` from
`data-model.md`. Uploaded as a GitHub Actions artefact named
`audit-report-${{ github.run_id }}` whenever the step exits non-zero.

### 3.2 Human-readable text (stderr)

For each violation, exactly one line is written to stderr in the format:

```text
<file>:<line>:<col?> <kind>: <subject> <actual> > <limit> [rule: <rule-name>]
```

Examples:

```text
cmd/new.go:42 function-size: runScaffold 67 > 50 [rule: cli-handler-func-size]
internal/core/entities/graph.go file-size: graph.go 678 > 300 [rule: entity-file-size]
internal/core/usecases/build_docs.go:8 layer: github.com/madstone-tech/loko/cmd 'cmd' may not be imported here [rule: core/usecases]
```

A trailing summary line is always emitted, regardless of exit code:

```text
archcheck: <N> file(s) scanned, <M> function(s) scanned, <V> violation(s) found.
```

### 3.3 GitHub Actions annotations (UI overlay)

When `--annotate=github` is passed, each violation also produces a GitHub
workflow command on stderr:

```text
::error file=<file>,line=<line>::[<rule>] <subject>: <actual> > <limit>
```

This makes violations appear as inline comments on the PR Files-changed view
without needing a separate review-bot integration.

## 4. Local equivalence (SC-007)

Running `make audit-constitution` locally MUST produce:

- Identical exit code to the CI step on the same source tree.
- Identical `audit-report.json` (modulo `generated_at` and `audit_tool_version`
  fields).
- Identical text output to stderr (modulo the GitHub annotation lines, which
  are still emitted but harmless when not running under Actions).

This is verified by a smoke test in `tools/archcheck/archcheck_test.go` that
runs the binary against a fixture tree containing intentional violations and
asserts the produced report.

## 5. Determinism guarantees

- File scan order is sorted (lexicographic, module-relative). Two runs over the
  same tree produce identical reports byte-for-byte (excluding the two
  metadata fields above).
- Violation order in the report is grouped by `kind` (`layer`, `file-size`,
  `function-size`) then sorted by `(file, line, subject)` within each group.

## 6. Failure-mode messaging (FR-014)

Every violation message MUST identify three things:

1. **The offending file** (module-relative path).
2. **The offending entity** (function name, basename, or import path).
3. **The rule that was broken** (rule `name`).

A reviewer who is unfamiliar with the codebase MUST be able to read a single
line of output and locate the problem without consulting any other resource
(SC-008).

## 7. Backwards-compatibility window

For one release after the audit tool is wired in, `scripts/audit-constitution.sh`
remains available as a wrapper around `tools/archcheck` with a `--legacy-shim`
flag, so contributors who have memorised the old command keep working. After
that release, the shell script is deleted and only `make audit-constitution`
is supported.
