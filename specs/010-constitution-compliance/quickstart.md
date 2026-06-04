# Quickstart: Constitution Compliance Refactor

**Feature**: 010-constitution-compliance
**Audience**: a contributor who has just checked out branch `010-constitution-compliance` and wants to (a) run the compliance check locally, (b) fix one violation end-to-end, and (c) verify the CI gate before opening a PR.

## Prerequisites

- Go 1.25+ installed (`go version` should report `go1.25` or higher).
- [Task](https://taskfile.dev) (`task` CLI) and `git` on the `$PATH`.
- The repository's standard toolchain installed (`golangci-lint`, etc.) — the existing `task setup` target handles this.

```bash
git checkout 010-constitution-compliance
task setup            # installs golangci-lint, downloads modules
```

## 1. Build and run the compliance check

```bash
task audit-constitution
```

What this does:
1. Builds `tools/archcheck/archcheck` (cached — subsequent runs are instant if the tool source hasn't changed).
2. Invokes `archcheck` with `specs/010-constitution-compliance/contracts/structural-rules.yaml` and `.archcheck-suppressions.yaml`.
3. Prints violations to stdout (text format).
4. Exits non-zero if any non-suppressed violation is found.

**Expected output on a clean working tree (after the refactor lands):**

```text
---
0 violations across 0 files. 3 suppressions applied (expire 2026-06-30, 2026-07-15, 2026-08-15).
```

**Expected output if you intentionally regress a handler size budget:**

```text
cmd/build.go:42 [cli-handler-func-size] function `runBuild` is 58 effective lines, limit is 50
---
1 violation across 1 file.
```

Exit code: `1`.

## 2. Fix one violation end-to-end (worked example)

Suppose the audit fails with:

```text
cmd/new.go:147 [cli-handler-func-size] function `runNewProject` is 68 effective lines, limit is 50
```

**Step A — Locate.** Open `cmd/new.go` at line 147. Identify what `runNewProject` is doing beyond its three legal jobs (parse input, call a use case, render output).

**Step B — Extract.** Move the surplus logic into `internal/core/usecases/scaffold_project.go`. If a use case for "scaffold a new project" does not exist yet, create one:

```go
// internal/core/usecases/scaffold_project.go
package usecases

import "github.com/madstone-io/loko/internal/core/entities"

func ScaffoldProject(repo ProjectRepository, tmpl TemplateEngine, in ScaffoldProjectInput) (*entities.Project, error) {
    // … the extracted logic …
}
```

**Step C — Slim the handler.** `runNewProject` should now look like:

```go
func runNewProject(cmd *cobra.Command, args []string) error {
    in, err := parseScaffoldProjectFlags(cmd, args)   // small helper in cmd/new_input.go
    if err != nil { return err }

    proj, err := usecases.ScaffoldProject(repo, tmpl, in)
    if err != nil { return err }

    return ui.RenderScaffoldResult(cmd.OutOrStdout(), proj)
}
```

**Step D — Test.** Add (or update) a unit test for `ScaffoldProject` using concrete mock ports under `internal/core/usecases/scaffold_project_test.go`. Run:

```bash
go test ./internal/core/usecases/... ./cmd/...
```

**Step E — Re-audit.** Run `task audit-constitution` again — the violation should be gone, and total effective lines for `runNewProject` should be well under 50.

## 3. Verify the CI gate locally

The CI pipeline runs three gating steps in order: `Test`, `Lint`, `Audit`. Reproduce all three locally before opening a PR:

```bash
task test                  # unit + integration tests
task lint                  # golangci-lint with depguard layer rules
task audit-constitution    # archcheck — the gate this feature installs
```

If all three exit `0`, your branch will pass the CI gate.

## 4. Suppression workflow (last resort)

If a pre-existing violation outside this feature's scope blocks the gate, **first** try to fix it in a small, surgical commit. If that is not feasible inside this PR, add a suppression entry to `.archcheck-suppressions.yaml`:

```yaml
- rule: outer-no-entities
  file: internal/mcp/tools/some_legacy_tool.go
  owner: "@andhi"
  expires_on: "2026-08-15"
  reason: |
    Legacy tool will be replaced under feature 011; tracking issue #789.
  notes: "Tracking: #789"
```

Constraints (the tool enforces these at load time):
- `expires_on` must be ≤ 90 days from today (or the tool exits 4).
- `rule` must reference a real rule/budget name.
- `file` glob must match at least one existing file.
- Per-function rules require a `function` field.

Renewal: when a suppression nears expiry, decide between **fix-and-remove** (preferred) or **renew** (only with a fresh `reason` referencing the new blocker).

## 5. End-to-end smoke check (before opening the PR)

Verify that the refactor preserves observable behaviour:

```bash
# Capture golden files from main BEFORE you start (one-time, by your reviewer or by you on a clean main checkout):
git worktree add ../loko-main main
(cd ../loko-main && task build && ./bin/loko new project tmp-golden --dry-run > /tmp/loko-golden.txt)

# After your refactor:
task build
./bin/loko new project tmp-test --dry-run > /tmp/loko-test.txt
diff /tmp/loko-golden.txt /tmp/loko-test.txt        # MUST be empty
```

The MCP smoke fixture lives at `specs/010-constitution-compliance/mcp-smoke.md` (replays a curated set of JSON-RPC calls and diffs the responses).

## 6. Open the PR

Title prefix: `feat(010):`. PR description must include:

- Which user stories the PR addresses (the spec has four; one PR per story is the recommended cadence).
- The output of `task audit-constitution` showing `0 violations` (or the new suppression entries with `expires_on` dates).
- Per-package coverage delta vs. the merge-base (see research.md R6).
- A note if the constitution amendment 1.1.0 → 1.2.0 is included in the PR (it ships with the US3 PR by convention).

CI will run the same three steps as your local run. If `Audit` fails, the merge button is greyed out.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `archcheck` exits 2 with "unknown rule: foo" | A suppression entry references a deleted/renamed rule | Update the suppression entry's `rule` field or remove the entry |
| `archcheck` reports a file that "should" be exempt | The file does not match a categorical-exemption pattern | Either rename to match an exemption (`schemas.go`, `*_cobra.go`, etc.) or open an ADR to add a new exemption |
| `task audit-constitution` is slow (> 30 s) | First run after `go clean -cache` | Subsequent runs are fast; verify the binary is being cached at `tools/archcheck/archcheck` |
| `depguard` and `archcheck` disagree | Bug — they should never disagree on layer rules | File an issue; treat `archcheck` as authoritative for the duration of the disagreement |
