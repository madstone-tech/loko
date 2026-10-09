# Quickstart: validating MCP HCL Authoring

**Feature**: `015-mcp-hcl-authoring` | **Plan**: [plan.md](plan.md)

Each scenario names the requirement it proves. Tool shapes are in
[contracts/mcp-tools.md](contracts/mcp-tools.md).

## Prerequisites

```bash
task build && export PATH="$PWD:$PATH"   # task build writes ./loko
FIX=testdata/projects/two-systems
```

## 1. Read the architecture (US1, FR-001..FR-009)

```bash
go test ./internal/mcp/tools/ -run 'TestDescribeTool|TestQueryTool|TestValidateTool' -count=1
loko query dependents container.orders_db -p $FIX --transitive
loko query path person.customer external.bank -p $FIX
loko query coupling -p $FIX --limit 5
```

**Expect**:
- `dependents` lists `container.api` and `component.repo`, which is reached through the
  descendant rule.
- `path` runs customer → web → api → gateway → bank.
- A broken copy of the fixture returns diagnostics with positions instead of answers.

## 2. Build an architecture from nothing (US2, SC-001)

```bash
go test ./internal/mcp/ -run TestBuildArchitectureEndToEnd -count=1 -v
```

**Expect**: driving only `apply_edit` and `move` over the MCP protocol, starting from a `project`
block, produces a project that compiles with zero errors. Its export, with ranges stripped, equals
`testdata/projects/two-systems`'s export.

## 3. Hand-written source survives (US3, SC-002, SC-003)

```bash
go test ./internal/adapters/hclsource/ -run 'TestPlan|TestRoundTripProperty' -count=1
go test ./internal/adapters/hclsource/ -fuzz FuzzApplyEdits -fuzztime 60s
```

**Expect**: zero divergences in 2,000 seeded sequences, and every byte outside the edited
declarations unchanged in every file.

## 4. Safety refusals (US3, FR-012, FR-016, FR-018, SC-004)

```bash
go test ./internal/core/usecases/ -run 'TestApply|TestRemove' -count=1
go test ./internal/mcp/tools/ -run 'TestRefusedWritesChangeNothing|TestStaleRevisionIsPerFile' -count=1
```

**Expect**: every refusal leaves the disk byte-identical (asserted by hashing the tree), and
applying after a preview equals the preview.

## 5. Rename (US4, FR-022..FR-025)

```bash
go test ./internal/adapters/hclsource/ -run 'TestRename|TestRemoveDropsTheElementsHistory' -count=1
go test ./internal/mcp/tools/ -run TestMove -count=1
go test ./internal/core/usecases/ -run 'TestValidateMoved' -count=1
```

**Expect**: references in every file are rewritten, a `moved` block is appended, a kind change
with a parent set in the same batch compiles, and the three `moved_*` diagnostics fire on their
fixtures.

## 6. Nothing but HCL (FR-026, SC-005)

```bash
go test ./internal/mcp/tools/ -run TestToolsWriteOnlyHCL -count=1
task audit-constitution
```

**Expect**: every tool, driven against a fixture, creates or modifies only `*.loko.hcl` files, and
archcheck reports 0 violations.

## 7. CLI and MCP agree (US5, SC-006)

```bash
go test ./cmd/ -run TestQueryMatchesMCP -count=1
```

## 8. Performance (SC-007, SC-008)

```bash
go test ./cmd/ -run TestAuthoringPerformance -count=1 -v
```

**Expect**: reads under 1 s and an edit with its compile under 2 s on 1,000 elements, with a
TOON summary of at most 300 tokens.

## 9. Full gate

```bash
task lint && task test && task audit-constitution
```
