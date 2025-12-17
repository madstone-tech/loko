# loko Implementation Plan

> Generated: 2024-12-17
> Based on: spec.md, tasks.md, constitution.md

## Phase Overview

| Phase | Focus | Issues | Complexity | Status |
|-------|-------|--------|------------|--------|
| 1 | Foundation | #1, #2, #3 | Low | 🟡 In Progress |
| 2 | First Use Case | #4, #5, #5b, #6 | Medium | 🔲 Not Started |
| 3 | Build Pipeline | #7, #8, #9, #10 | Medium | 🔲 Not Started |
| 4 | MCP Integration | #11, #12, #13 | High | 🔲 Not Started |
| 5 | v0.2.0 Features | #14, #15 | Medium | 🔲 Not Started |

## Complexity Analysis

### Low Complexity (1-2 days each)

| Issue | Why Low |
|-------|---------|
| #1 Project Setup | ✅ Done - Boilerplate |
| #2 Entities | ✅ Done - Pure Go structs |
| #3 Ports | Interface definitions only, no logic |

### Medium Complexity (2-3 days each)

| Issue | Why Medium |
|-------|------------|
| #4 CreateSystem UC | First use case, patterns established |
| #5 FileSystem Adapter | I/O, TOML parsing, YAML frontmatter |
| #5b ason Adapter | External library integration |
| #6 CLI Wiring | DI setup, Cobra integration |
| #7 D2 Adapter | Shell out, caching logic |
| #8 BuildDocs UC | Orchestration, parallelism |
| #9 HTML Builder | Templates, navigation |
| #10 CLI Commands | Multiple commands, each simple |
| #14 TOON Format | New encoding, benchmarking |
| #15 HTTP API | REST, auth middleware |

### High Complexity (3-5 days each)

| Issue | Why High |
|-------|----------|
| #11 Token Queries | Algorithm design, optimization |
| #12 MCP Server | Protocol implementation, 8 tools |
| #13 Documentation | Writing, examples, testing |

## Risk Assessment

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| D2 CLI behavior varies | Medium | Medium | Test on all platforms, pin version |
| MCP protocol changes | Low | High | Use stable MCP SDK |
| ason API changes | Low | Medium | Pin version, owned library |
| Token estimation wrong | Medium | Low | Build metrics, iterate |

### Schedule Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Scope creep | High | Medium | Strict v0.1.0 scope, defer to v0.2.0 |
| Integration issues | Medium | Medium | Build vertical slices early |
| Documentation debt | High | Low | Write docs with features |

## Vertical Slices

Build working features end-to-end before expanding:

### Slice 1: Create System (Issues #3-#6)

```
CLI → CreateSystem UC → FileSystem Adapter → Files on Disk
                     → ason Adapter → Template Rendering
```

**Milestone:** `loko new system PaymentService` works

### Slice 2: Build Docs (Issues #7-#10)

```
CLI → BuildDocs UC → D2 Adapter → SVG Files
                  → HTML Builder → Static Site
```

**Milestone:** `loko build && loko serve` shows docs

### Slice 3: MCP (Issues #11-#12)

```
MCP Server → QueryArchitecture UC → Token-Efficient Response
          → CreateSystem UC → (reuse from Slice 1)
          → BuildDocs UC → (reuse from Slice 2)
```

**Milestone:** Claude can design architecture via MCP

## Implementation Order

```
Week 1: #3 (ports) → enables all adapters
Week 2: #4 (CreateSystem) + #5 (filesystem) + #5b (ason) + #6 (CLI)
Week 3: #7 (D2) + #8 (BuildDocs) + #9 (HTML) + #10 (CLI)
Week 4: #11 (queries) + #12 (MCP) + #13 (docs)
Future: #14 (TOON) + #15 (API)
```

## Definition of Done

### For Each Issue

- [ ] Code complete and tested
- [ ] Tests pass (`go test ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Documentation updated (if applicable)
- [ ] PR reviewed and merged

### For Each Phase

- [ ] All issues in phase complete
- [ ] Integration tests pass
- [ ] Milestone demo works
- [ ] README updated with new features

### For v0.1.0 Release

- [ ] Phases 1-4 complete
- [ ] All success criteria met (SC-001 through SC-008)
- [ ] Documentation complete
- [ ] Examples work in CI
- [ ] Docker image builds
- [ ] goreleaser creates binaries

## Next Steps

1. **Now:** Implement Issue #3 (ports.go interfaces)
2. **Then:** Issue #4 (CreateSystem use case)
3. **Parallel:** Issue #5 (filesystem) and #5b (ason)
4. **Complete Slice 1:** Issue #6 (CLI wiring)

## Notes for Claude Code

When working on this project:

1. **Always read first:**
   - `.specify/memory/constitution.md` - coding standards
   - `CLAUDE.md` - quick reference

2. **Check progress:**
   - `.specify/memory/tasks.md` - what's done, what's next

3. **Follow patterns:**
   - Entities: pure structs with validation
   - Use cases: orchestration with ports
   - Adapters: implement port interfaces
   - CLI/MCP: thin wrappers (<50 lines)

4. **Test strategy:**
   - Core: unit tests with mocks
   - Adapters: integration tests
   - E2E: full stack tests
