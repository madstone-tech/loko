# Requirements Checklist: MCP LLM Integration Enhancement

**Spec**: 004-mcp-llm-integration
**Version**: 0.1.0
**Last Updated**: 2026-02-06

---

## Functional Requirements

### P1 - Must Have

| ID | Requirement | Status | Notes |
|----|-------------|--------|-------|
| FR-001 | All tool descriptions MUST include a "When to use" section | ⬜ Pending | |
| FR-002 | All tool descriptions MUST include C4 level context where applicable | ⬜ Pending | Applies to create_* tools |
| FR-003 | All creation tools MUST mention the query-first pattern | ⬜ Pending | |
| FR-004 | `query_architecture` description MUST explain detail levels and token estimates | ✅ Done | Already implemented |
| FR-008 | Tool descriptions MUST guide toward validation after mutations | ⬜ Pending | |

### P2 - Should Have

| ID | Requirement | Status | Notes |
|----|-------------|--------|-------|
| FR-005 | `query_architecture` description MUST explain TOON format benefits | ✅ Done | Already implemented |
| FR-006 | `update_diagram` description MUST include D2 syntax examples | ⬜ Pending | |
| FR-007 | All tools MUST include example invocations in descriptions | ⬜ Pending | |
| FR-009 | `validate_diagram` MUST explain C4 compliance checks | ⬜ Pending | |

### P3 - Nice to Have

| ID | Requirement | Status | Notes |
|----|-------------|--------|-------|
| FR-010 | llms.txt reference documentation MUST be discoverable via tool descriptions | ⬜ Pending | |

---

## Non-Functional Requirements

| ID | Requirement | Target | Status | Actual |
|----|-------------|--------|--------|--------|
| NFR-001 | Tool description length | 100-300 words | ⬜ Pending | |
| NFR-002 | Example invocation clarity | Copy-pasteable JSON | ⬜ Pending | |
| NFR-003 | C4 terminology consistency | Exact C4 terms | ⬜ Pending | |
| NFR-004 | Token overhead of descriptions | < 500 tokens/tool | ⬜ Pending | |

---

## Tools to Update

| Tool | File | Current State | Enhancement Needed |
|------|------|---------------|-------------------|
| `create_system` | tools.go | 1 line | Full template |
| `create_container` | tools.go | 1 line | Full template |
| `create_component` | tools.go | 1 line | Full template |
| `update_diagram` | tools.go | 1 line | D2 examples |
| `build_docs` | tools.go | 1 line | Workflow context |
| `validate` | tools.go | 1 line | What's validated |
| `validate_diagram` | tools.go | Good | C4 checks |
| `query_project` | query_project.go | 1 line | When to use |
| `query_architecture` | query_architecture.go | Good | Minor tweaks |
| `query_dependencies` | graph_tools.go | 1 line | Graph context |
| `query_related_components` | graph_tools.go | 1 line | Use cases |
| `analyze_coupling` | graph_tools.go | 1 line | Interpretation |

---

## Success Criteria

| ID | Criterion | Target | Status | Measured |
|----|-----------|--------|--------|----------|
| SC-001 | LLM first-attempt success rate | > 80% | ⬜ Pending | |
| SC-002 | Query-before-create pattern adoption | > 90% | ⬜ Pending | |
| SC-003 | Validation usage after mutations | > 80% | ⬜ Pending | |
| SC-004 | Token-efficient format usage | > 50% | ⬜ Pending | |
| SC-005 | Valid D2 generation rate | > 90% | ⬜ Pending | |

---

## User Stories Coverage

| Story | Priority | Status | Acceptance Scenarios |
|-------|----------|--------|---------------------|
| US-001: First-Time Architecture Creation | P1 | ⬜ Pending | 3 scenarios |
| US-002: C4 Level Understanding | P1 | ⬜ Pending | 3 scenarios |
| US-003: Query Before Mutate Pattern | P1 | ⬜ Pending | 3 scenarios |
| US-004: Token-Efficient Workflows | P2 | ⬜ Pending | 3 scenarios |
| US-005: D2 Diagram Guidance | P2 | ⬜ Pending | 3 scenarios |

---

## Testing Checklist

### Manual LLM Testing

- [ ] Test with Claude: "Create a new e-commerce system"
- [ ] Test with Claude: "Add a database to the Order system"
- [ ] Test with Claude: "Show me the architecture"
- [ ] Test with Claude: "Create a serverless image processor"
- [ ] Verify query-before-create pattern is followed
- [ ] Verify validation is called after mutations
- [ ] Verify appropriate detail level usage

### Before/After Comparison

- [ ] Document baseline LLM behavior (before changes)
- [ ] Document improved LLM behavior (after changes)
- [ ] Measure success rate improvement

---

## Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Author | | 2026-02-06 | Draft |
| Reviewer | | | Pending |
| Approver | | | Pending |
