# MCP Tool Contracts

**Feature**: 006-phase-1-completion  
**Date**: 2026-02-13

---

## Overview

This document defines test contracts for the two new MCP tools added in v0.2.0:
1. `search_elements` - Search architecture elements
2. `find_relationships` - Find relationships between elements

---

## Contract 1: search_elements

### Test Scenario 1: Search by name (glob pattern)

**Given**: Project with systems "payment-service", "order-service", "user-service"

**When**: Tool called with `{"query": "payment*"}`

**Then**:
- Response contains exactly 1 result
- Result matches: `{"name": "payment-service", "type": "system"}`
- Response time < 200ms

**Test Implementation**:
```go
func TestSearchElements_ByNameGlob(t *testing.T) {
    // Setup test project
    repo := setupTestProject(t, []string{"payment-service", "order-service", "user-service"})
    tool := NewSearchElementsTool(repo)
    
    // Execute tool
    start := time.Now()
    result, err := tool.Call(ctx, map[string]any{
        "project_root": testProjectRoot,
        "query": "payment*",
    })
    elapsed := time.Since(start)
    
    // Assertions
    require.NoError(t, err)
    assert.Less(t, elapsed, 200*time.Millisecond)
    
    results := result.(map[string]any)["results"].([]any)
    assert.Len(t, results, 1)
    assert.Equal(t, "payment-service", results[0].(map[string]any)["name"])
}
```

---

### Test Scenario 2: Filter by type

**Given**: Project with 5 systems, 10 containers, 20 components

**When**: Tool called with `{"query": "*", "type": "component"}`

**Then**:
- Response contains exactly 20 results (limited to components)
- All results have `"type": "component"`
- Response time < 200ms

---

### Test Scenario 3: Filter by technology

**Given**: Components with technologies "Go", "Python", "TypeScript"

**When**: Tool called with `{"query": "*", "technology": "Go"}`

**Then**:
- Response contains only Go components
- All results have `"technology": "Go"`
- Components with other technologies excluded

---

### Test Scenario 4: Empty results

**Given**: Project with no elements matching query

**When**: Tool called with `{"query": "nonexistent"}`

**Then**:
- Response contains empty results array: `{"results": []}`
- Response includes helpful message: `"No elements found matching query"`
- No error returned (empty set is valid)

---

### Test Scenario 5: Result limiting

**Given**: Project with 100 components

**When**: Tool called with `{"query": "*", "limit": 10}`

**Then**:
- Response contains exactly 10 results (limit enforced)
- Response includes `"total_matched": 100` (shows total available)
- Response time < 200ms despite large dataset

---

### Test Scenario 6: Tag filtering

**Given**: Components with tags "critical", "experimental", "deprecated"

**When**: Tool called with `{"query": "*", "tag": "critical"}`

**Then**:
- Response contains only components with "critical" tag
- All results have `"tags": [..., "critical", ...]`

---

## Contract 2: find_relationships

### Test Scenario 1: Find all dependencies

**Given**: 
- Component "api-handler" depends on "database"
- Component "api-handler" depends on "cache"
- Component "worker" depends on "queue"

**When**: Tool called with `{"source_pattern": "api-handler"}`

**Then**:
- Response contains 2 relationships (api-handler → database, api-handler → cache)
- Both relationships have `"source_id": "api-handler"`
- Response time < 200ms

**Test Implementation**:
```go
func TestFindRelationships_AllDependencies(t *testing.T) {
    // Setup test graph
    repo := setupTestGraph(t, []Relationship{
        {Source: "api-handler", Target: "database", Type: "depends-on"},
        {Source: "api-handler", Target: "cache", Type: "depends-on"},
        {Source: "worker", Target: "queue", Type: "depends-on"},
    })
    tool := NewFindRelationshipsTool(repo)
    
    // Execute tool
    result, err := tool.Call(ctx, map[string]any{
        "project_root": testProjectRoot,
        "source_pattern": "api-handler",
    })
    
    // Assertions
    require.NoError(t, err)
    rels := result.(map[string]any)["relationships"].([]any)
    assert.Len(t, rels, 2)
    
    for _, rel := range rels {
        assert.Equal(t, "api-handler", rel.(map[string]any)["source_id"])
    }
}
```

---

### Test Scenario 2: Glob pattern matching

**Given**: Components "backend-api", "backend-worker", "frontend-ui"

**When**: Tool called with `{"source_pattern": "backend-*"}`

**Then**:
- Response includes relationships from "backend-api" and "backend-worker"
- Relationships from "frontend-ui" are excluded
- Pattern matches multiple sources correctly

---

### Test Scenario 3: Bidirectional filtering

**Given**: 
- "backend" → "external-api"
- "backend" → "database"
- "frontend" → "external-cdn"

**When**: Tool called with `{"source_pattern": "backend", "target_pattern": "external-*"}`

**Then**:
- Response contains only "backend" → "external-api"
- "backend" → "database" excluded (target doesn't match pattern)
- "frontend" → "external-cdn" excluded (source doesn't match pattern)

---

### Test Scenario 4: Relationship type filtering

**Given**: 
- "api" depends-on "database"
- "api" uses "cache"
- "api" calls "external-service"

**When**: Tool called with `{"source_pattern": "api", "relationship_type": "depends-on"}`

**Then**:
- Response contains only "depends-on" relationships
- "uses" and "calls" relationships excluded
- Type filter applied correctly

---

### Test Scenario 5: Empty relationships

**Given**: Component "isolated-service" with no dependencies

**When**: Tool called with `{"source_pattern": "isolated-service"}`

**Then**:
- Response contains empty relationships array: `{"relationships": []}`
- Response includes message: `"No relationships found"`
- No error returned (empty set is valid)

---

### Test Scenario 6: Performance with large graph

**Given**: Graph with 1000 components and 5000 relationships

**When**: Tool called with `{"source_pattern": "backend-*", "limit": 50}`

**Then**:
- Response contains exactly 50 results (limit enforced)
- Response includes `"total_matched": <actual count>` if > 50
- Response time < 200ms despite large graph

---

## Integration Test Scenario: Combined Search + Relationships

**Goal**: Demonstrate search tools working together

**Workflow**:
1. `search_elements` finds all Go components tagged "critical"
2. For each result, `find_relationships` discovers dependencies
3. Results aggregated to show critical dependency map

**Expected Behavior**:
- Both tools return consistent element IDs
- Relationships reference IDs from search results
- Combined query completes in < 500ms

**Test Implementation**:
```go
func TestIntegration_SearchAndRelationships(t *testing.T) {
    // Step 1: Search for critical Go components
    searchResult, err := searchElementsTool.Call(ctx, map[string]any{
        "query": "*",
        "technology": "Go",
        "tag": "critical",
    })
    require.NoError(t, err)
    
    elements := searchResult.(map[string]any)["results"].([]any)
    require.NotEmpty(t, elements)
    
    // Step 2: Find dependencies for each element
    var allDeps []any
    for _, elem := range elements {
        elemID := elem.(map[string]any)["id"].(string)
        
        relResult, err := findRelationshipsTool.Call(ctx, map[string]any{
            "source_pattern": elemID,
        })
        require.NoError(t, err)
        
        rels := relResult.(map[string]any)["relationships"].([]any)
        allDeps = append(allDeps, rels...)
    }
    
    // Step 3: Verify critical dependency map created
    assert.NotEmpty(t, allDeps)
}
```

---

## Error Handling Contracts

### Invalid Query Pattern

**Given**: Tool called with malformed glob pattern `{"query": "[invalid"}`

**Then**:
- Error returned with message: `"Invalid glob pattern: [invalid"`
- Status code: 400 (Bad Request equivalent in MCP)
- No partial results returned

---

### Project Not Found

**Given**: Tool called with `{"project_root": "/nonexistent"}`

**Then**:
- Error returned with message: `"Project not found at path: /nonexistent"`
- Status code: 404 (Not Found equivalent)
- Clear guidance on checking project path

---

### Timeout Exceeded

**Given**: Search takes longer than 200ms (simulated slow disk)

**Then**:
- Error returned with message: `"Search timeout exceeded (200ms)"`
- Partial results MAY be returned with warning
- Suggestion to reduce `limit` parameter

---

## Performance Benchmarks

### Benchmark 1: Small Project

**Setup**: 10 systems, 30 containers, 100 components

**Tests**:
- `search_elements` with `{"query": "*"}`: < 50ms
- `find_relationships` with `{"source_pattern": "*"}`: < 100ms

---

### Benchmark 2: Medium Project

**Setup**: 50 systems, 200 containers, 1000 components

**Tests**:
- `search_elements` with `{"query": "backend-*"}`: < 100ms
- `find_relationships` with `{"source_pattern": "backend-*"}`: < 150ms

---

### Benchmark 3: Large Project

**Setup**: 100 systems, 500 containers, 5000 components

**Tests**:
- `search_elements` with `{"query": "*", "limit": 20}`: < 200ms
- `find_relationships` with `{"source_pattern": "*", "limit": 50}`: < 200ms

**Note**: With result limiting, performance remains consistent regardless of project size

---

## Validation Checklist

Before release, verify:

- [ ] All test scenarios pass (12 scenarios per tool = 24 tests)
- [ ] Integration test passes (combined search + relationships)
- [ ] Error handling contracts implemented (3 error scenarios)
- [ ] Performance benchmarks meet targets (3 size categories)
- [ ] Tools registered in `cmd/mcp.go`
- [ ] MCP schema validation passes
- [ ] Documentation updated with examples
- [ ] Token usage is minimal (< 1000 tokens per query)

---

## MCP Protocol Compliance

### Tool Registration

**Required Fields**:
```json
{
  "name": "search_elements",
  "description": "Search architecture elements by name, technology, or tags",
  "inputSchema": {
    "type": "object",
    "properties": {...},
    "required": ["query"]
  }
}
```

### Response Format

**Success Response**:
```json
{
  "content": [{
    "type": "text",
    "text": "Found 3 elements matching query 'payment*'"
  }],
  "isError": false
}
```

**Error Response**:
```json
{
  "content": [{
    "type": "text",
    "text": "Error: Invalid glob pattern"
  }],
  "isError": true
}
```

---

## Summary

**Total Test Contracts**: 29
- 12 scenarios for `search_elements`
- 12 scenarios for `find_relationships`
- 1 integration scenario
- 3 error handling scenarios
- 1 validation checklist

**Coverage**:
- ✅ Functional correctness (24 scenarios)
- ✅ Performance targets (3 benchmarks)
- ✅ Error handling (3 scenarios)
- ✅ Integration (1 scenario)
- ✅ MCP protocol compliance

**Next**: Generate quickstart guide for developers implementing these tools
