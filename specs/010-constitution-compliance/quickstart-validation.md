# Quickstart Validation — Constitution Compliance Refactor

**Date**: 2026-06-03
**Branch**: `010-constitution-compliance`
**Validator**: automated run via `tools/archcheck` against the live repo + a synthetic violating fixture tree.

This file records the outcome of the behaviour-preservation and diagnostic-legibility
checks called for by `tasks.md` (T060), the archcheck wall-clock benchmark (T065), and
the branch-009 rules-file retirement decision (T062).

---

## 1. Whole-repo audit (behaviour preservation, SC-001..SC-008)

```
$ task audit-constitution   # → tools/archcheck against the repo with defaults
archcheck: 240 file(s) scanned, 1534 function(s) scanned, 0 violation(s) found.
exit 0
```

```
$ go test ./...
ok  (all packages — see baseline-toolchain.txt; no failures)
```

**Result**: ✅ The refactor is byte-clean against every structural budget and layer rule;
the full test suite passes. `cmd/new.go`, `cmd/build.go`, every `internal/mcp/tools/*`
handler, and every `internal/core/*` file are within budget.

---

## 2. Diagnostic-legibility check (SC-010, T060)

A synthetic fixture tree (`module example.com/fixture`) was created with one deliberate
violation of each kind and audited with the repo's canonical `tools/archcheck/rules.yaml`.
SC-010 requires each diagnostic to be actionable by a contributor unfamiliar with the
project, naming: (a) the repo-relative file path, (b) the offending entity, (c) the
rule/budget name, and (d) measured-vs-limit for size kinds.

### Sample diagnostics (verbatim)

**`layer_import`** — MCP tool importing `internal/core/entities` directly (forbidden v1.2.0):

```
internal/mcp/tools/leak.go:3: layer 'mcp' may not import 'example.com/fixture/internal/core/entities' (rule: MCP server may import core/usecases, adapters, and its own sub-packages. MUST NOT import internal/core/entities directly — obtain entity types via use-case return values or adapter outputs (Constitution v1.2.0).)
```

✅ Names file+line, the offending import path, the layer, and the rationale.

**`file_size`** — entity file over the 300-line budget:

```
internal/core/entities/big.go:1: big.go exceeds entity-file-size (322 > 300)
```

✅ Names file, rule name (`entity-file-size`), and measured-vs-limit (`322 > 300`).

**`function_size`** — CLI handler over the 50-line budget:

```
cmd/big.go:3: runBig exceeds cli-handler-func-size (74 > 50)
```

✅ Names file+line, the offending function (`runBig`), the rule (`cli-handler-func-size`),
and measured-vs-limit (`74 > 50`).

**`over-90-day suppression`** — rejected at load:

```
archcheck: suppression file validation errors:
  - entry #1: expires_on 2027-01-01 is more than 90 days from today; longer-lived suppressions require an ADR
exit 4
```

✅ Names the offending entry, the date, the 90-day rule, and the remedy (ADR). Distinct
exit code (4) separates "bad suppression file" from "violations found" (1).

### Suppression lifecycle (verified)

| Scenario | Behaviour | Exit |
|----------|-----------|------|
| In-date suppression of a real violation | Violation removed from the report (3 → 2) | 1 (other violations remain) |
| Expired suppression (`expires_on` in the past) | Underlying violation **re-fires** as a normal failure (3 → 3) | 1 |
| `expires_on` > 90 days out | Rejected at load with a validation error | 4 |

**Known gap (honest note)**: an *expired* suppression silently re-fires the underlying
violation but does **not** emit a dedicated `expired_suppression` / stale-suppression
diagnostic line in text output (`main.go` currently discards `staleSuppressions` with a
`// future` comment). The 30-days-past-expiry stale warning described in T009/FR is not
yet surfaced. This is a minor legibility shortfall — the violation is still reported, so
the gate still blocks — but the *reason* ("your suppression lapsed") is not spelled out.
Tracked as follow-up; does not affect the pass/fail behaviour of the gate.

**Result**: ✅ PASS for the three structural violation kinds and the over-90-day rejection.
⚠️ Partial for expired-suppression legibility (re-fires correctly, but without a named
"expired" diagnostic).

---

## 3. archcheck wall-clock benchmark (T065, SC-009 / R6)

Five warm runs of `tools/archcheck --format=json > /dev/null` on the full repo:

```
0.07  0.07  0.07  0.07  0.07   (real seconds)
```

| Metric | Value |
|--------|-------|
| min | 0.07 s |
| median | 0.07 s |
| max | 0.07 s |
| SC-009 budget (< 30 s) | ✅ PASS (430× margin) |
| R6 internal target (< 10 s) | ✅ PASS (soft target met) |

**Result**: ✅ Well within budget.

---

## 4. Branch-009 rules-file retirement (T062)

`specs/009-constitution-compliance/contracts/structural-rules.yaml` — the canonical rule
set now lives at `tools/archcheck/rules.yaml` (placed in T013) and the constitution
governance footer points there. **Decision**: leave the 009 copy in place for now — branch
009 is not being formally retired, and `scripts/check-rules-sync.sh` validates the live
`tools/archcheck/rules.yaml` against the constitution prose, not the 009 archive. No action
taken. Revisit when 009 is archived.
