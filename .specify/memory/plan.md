# loko Implementation Plan

> Generated: 2024-12-17
> Updated: 2026-02-06
> Based on: spec.md, tasks.md, constitution.md

## Phase Overview

| Phase | Focus | Issues | Status |
|-------|-------|--------|--------|
| 1 | Foundation | #1, #2, #3, #002, #003 | ✅ Complete |
| 2 | Handler Refactoring + TOON | #005 | 🟡 Spec Phase |
| 3 | First Use Case (Scaffolding) | #4, #5, #5b, #6 | 🔲 Not Started |
| 4 | Build Pipeline | #7, #8, #9, #10 | 🔲 Not Started |
| 5 | MCP Integration | #11, #12, #13 | 🔲 Not Started |
| 6 | v0.2.0 Features | #14, #15 | 🔲 Not Started |

## What's Changed Since Initial Plan

### Completed (not in original plan)

- **Cobra/Viper migration (#002)**: CLI now uses Cobra for commands, Viper for config hierarchy, shell completions, and aliases
- **Serverless template (#003)**: Added serverless architecture template with `-template` flag
- **Ports defined (#3)**: All 18 port interfaces defined in `usecases/ports.go`

### New: Phase 2 — Handler Refactoring + TOON Alignment (#005)

A constitution audit revealed **10 files** violating the thin handler principle. This phase was inserted before new feature work to prevent compounding debt.

**Why before Phase 3?**: Phase 3 (scaffolding use cases) will add new CLI commands and MCP tools. If we don't fix the handler pattern first, new handlers will follow the bloated pattern.

## Phase 1: Foundation (COMPLETE)

| Task | Status | Notes |
|------|--------|-------|
| T001: Initialize Go project | ✅ Done | Clean Architecture structure |
| T002: Domain entities with tests | ✅ Done | Project, System, Container, Component |
| T003: Port interfaces | ✅ Done | 18 interfaces in ports.go |
| Cobra/Viper CLI migration | ✅ Done | PR #5 merged |
| Serverless template | ✅ Done | PR #4 merged |

## Phase 2: Handler Refactoring + TOON Alignment (CURRENT)

**Goal**: Pay down handler debt, then implement spec-compliant TOON encoding.

### Part A: Handler Refactoring (P1)

Extract business logic from bloated handlers into use cases:

| File | Lines | Target | Action |
|------|-------|--------|--------|
| `cmd/new.go` | 504 | < 50 | Extract CreateSystem/Container/Component use cases |
| `cmd/d2_generator.go` | 282 | Move | Relocate to `internal/adapters/d2/` or use case |
| `cmd/build.go` | 251 | < 50 | Extract to BuildDocs use case, move reporter to adapter |
| `cmd/new_cobra.go` | 199 | < 50 | Thin Cobra wrapper over use case |
| `cmd/root.go` | 162 | Assess | May be acceptable if purely Cobra wiring |
| `cmd/watch.go` | 146 | < 50 | Extract watch orchestration to use case |
| `cmd/validate.go` | 142 | < 50 | Extract validation orchestration to use case |
| `cmd/build_cobra.go` | 107 | < 50 | Thin Cobra wrapper over use case |
| `mcp/tools/tools.go` | 1,084 | Split | One file per tool, each handler < 30 lines |
| `mcp/tools/graph_tools.go` | 348 | < 30 | Extract graph query logic to use case |

**Order**: Refactor CLI first (handlers create use cases), then MCP tools (reuse same use cases).

### Part B: TOON v3.0 Alignment (P1)

1. Replace custom encoder in `internal/adapters/encoding/toon.go` with spec-compliant implementation
2. Implement TOON decoder (round-trip support)
3. Deprecate old custom format
4. Run token efficiency benchmarks

### Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Refactoring breaks behavior | Medium | High | Run full test suite before/after each file |
| Use case interfaces don't match both CLI and MCP needs | Low | Medium | Design use case inputs/outputs to be interface-agnostic |
| toon-go library missing features | Medium | Low | Can implement spec-compliant encoder in adapter |

## Phase 3: First Use Case — Scaffolding

**Blocked by**: Phase 2 (handler refactoring creates the use cases that Phase 3 needs)

**Goal**: `loko init` and `loko new` commands work end-to-end via proper use cases.

After Phase 2, these use cases should exist:
- CreateSystem (extracted from `cmd/new.go`)
- CreateContainer (extracted from `cmd/new.go`)
- CreateComponent (extracted from `cmd/new.go`)

Phase 3 wires them up with proper adapters:
- Filesystem adapter (ProjectRepository)
- ason template engine adapter (TemplateEngine)
- CLI commands as thin wrappers

## Phase 4: Build Pipeline

**Goal**: `loko build`, `loko serve`, `loko watch` work end-to-end.

After Phase 2, BuildDocs use case should exist (extracted from `cmd/build.go`).

Phase 4 adds:
- D2 diagram renderer adapter
- HTML site builder adapter
- File watcher adapter
- Serve command (HTTP server for dist/)

## Phase 5: MCP Integration

**Goal**: Claude can design architecture via MCP end-to-end.

After Phase 2, MCP tool handlers should be thin and use cases should be shared with CLI.

Phase 5 adds:
- QueryArchitecture use case enhancements
- Token-efficient response formatting
- MCP tool descriptions with format hints

## Phase 6: v0.2.0 Features

- HTTP API server
- PDF export via veve-cli
- Advanced validation rules

## Vertical Slices

### Slice 1: Refactor + TOON (Phase 2)

```
Constitution audit → Extract use cases → Thin handlers → TOON adapter
```

**Milestone**: All handlers under line limits, TOON output validates against spec

### Slice 2: Scaffolding (Phase 3)

```
CLI → CreateSystem UC → FileSystem Adapter → Files on Disk
                     → ason Adapter → Template Rendering
```

**Milestone**: `loko new system PaymentService` works

### Slice 3: Build Docs (Phase 4)

```
CLI → BuildDocs UC → D2 Adapter → SVG Files
                  → HTML Builder → Static Site
```

**Milestone**: `loko build && loko serve` shows docs

### Slice 4: MCP (Phase 5)

```
MCP Server → QueryArchitecture UC → Token-Efficient Response
          → CreateSystem UC → (reuse from Slice 2)
          → BuildDocs UC → (reuse from Slice 3)
```

**Milestone**: Claude can design architecture via MCP

## Definition of Done

### For Each Issue

- [ ] Code complete and tested
- [ ] Tests pass (`task test`)
- [ ] Linter passes (`task lint`)
- [ ] Handler line counts verified (CLI < 50, MCP < 30)
- [ ] No new external dependencies in `internal/core/`
- [ ] Documentation updated (if applicable)
- [ ] PR reviewed and merged

### For v0.1.0 Release

- [ ] Phases 1-5 complete
- [ ] All success criteria met (SC-001 through SC-010)
- [ ] All handlers under line limits
- [ ] Documentation complete
- [ ] Docker image builds
- [ ] goreleaser creates binaries

## Next Steps

1. **Now**: Run `/speckit.clarify` on 005-toon-alignment spec
2. **Then**: Run `/speckit.plan` to generate detailed task breakdown for Phase 2
3. **Then**: Implement handler refactoring (Part A)
4. **Then**: Implement TOON alignment (Part B)
