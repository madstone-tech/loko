# Contract: D2Parser Interface

**Feature**: 007-ux-improvements  
**Interface**: `D2Parser`  
**Location**: `internal/core/usecases/ports.go`  
**Adapter**: `internal/adapters/d2/parser.go`

---

## Interface Definition

```go
// D2Parser parses D2 diagram syntax to extract relationships.
type D2Parser interface {
    // ParseRelationships extracts relationship arrows from D2 source code.
    // Returns a slice of D2Relationship or an error if parsing fails.
    ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error)
}
```

---

## Method: ParseRelationships

### Signature

```go
ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error)
```

### Parameters

| Parameter | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| `ctx` | `context.Context` | Context for cancellation and timeout | Required, may be `context.Background()` |
| `d2Source` | `string` | D2 diagram source code | Required, can be empty string (returns empty slice) |

### Returns

| Type | Description |
|------|-------------|
| `[]entities.D2Relationship` | Slice of relationships extracted from D2 source |
| `error` | Error if parsing fails catastrophically |

### Behavior

**Success Cases**:
1. **Valid D2 with relationships**: Returns slice of `D2Relationship` objects
2. **Valid D2 without relationships**: Returns empty slice, no error
3. **Empty string**: Returns empty slice, no error

**Error Cases**:
1. **Invalid D2 syntax**: Returns error with descriptive message
2. **Context cancelled**: Returns error indicating cancellation

### Error Handling Contract

**Graceful Degradation**:
- Partial parse success: Return relationships successfully parsed + log warnings (not exposed in return)
- Invalid arrow syntax: Skip that arrow, continue with others
- Only return error for catastrophic failures (e.g., completely malformed file)

**Error Format**:
```go
fmt.Errorf("failed to parse D2 source: %w", underlyingError)
```

---

## Examples

### Example 1: Valid D2 with Relationships

**Input**:
```go
d2Source := `
email-queue: {
  shape: rectangle
  tooltip: "SQS queue"
}

email-queue -> email-sender: "triggers"
notification-api -> email-queue: "publishes to"
`

relationships, err := parser.ParseRelationships(ctx, d2Source)
```

**Expected Output**:
```go
relationships == []entities.D2Relationship{
    {Source: "email-queue", Target: "email-sender", Label: "triggers"},
    {Source: "notification-api", Target: "email-queue", Label: "publishes to"},
}
err == nil
```

### Example 2: Valid D2 without Relationships

**Input**:
```go
d2Source := `
email-queue: {
  shape: rectangle
  tooltip: "SQS queue"
}
`

relationships, err := parser.ParseRelationships(ctx, d2Source)
```

**Expected Output**:
```go
relationships == []entities.D2Relationship{} // empty slice
err == nil
```

### Example 3: Empty String

**Input**:
```go
d2Source := ""
relationships, err := parser.ParseRelationships(ctx, d2Source)
```

**Expected Output**:
```go
relationships == []entities.D2Relationship{} // empty slice
err == nil
```

### Example 4: Invalid D2 Syntax

**Input**:
```go
d2Source := `
email-queue -> 
invalid syntax {{{{
`

relationships, err := parser.ParseRelationships(ctx, d2Source)
```

**Expected Output**:
```go
relationships == nil
err != nil
err.Error() contains "failed to parse D2 source"
```

### Example 5: Unlabeled Arrow

**Input**:
```go
d2Source := `
email-queue -> email-sender
`

relationships, err := parser.ParseRelationships(ctx, d2Source)
```

**Expected Output**:
```go
relationships == []entities.D2Relationship{
    {Source: "email-queue", Target: "email-sender", Label: ""},
}
err == nil
```

### Example 6: Multi-target Arrow

**Input**:
```go
d2Source := `
router -> email-sender: "email"
router -> sms-sender: "sms"
`

relationships, err := parser.ParseRelationships(ctx, d2Source)
```

**Expected Output**:
```go
relationships == []entities.D2Relationship{
    {Source: "router", Target: "email-sender", Label: "email"},
    {Source: "router", Target: "sms-sender", Label: "sms"},
}
err == nil
```

---

## Test Cases

### Unit Tests (Adapter Layer)

1. **Valid D2 with single arrow**: Verify 1 relationship extracted
2. **Valid D2 with multiple arrows**: Verify all relationships extracted
3. **Empty D2**: Verify empty slice returned
4. **Invalid syntax**: Verify error returned with descriptive message
5. **Unlabeled arrow**: Verify empty label in relationship
6. **Nested shapes**: Verify arrows between nested shapes handled correctly
7. **Context cancellation**: Verify error returned when context cancelled

### Integration Tests (Use Case Layer)

1. **Parse 100 component D2 files**: Verify <200ms total time
2. **Parse file with mixed valid/invalid arrows**: Verify graceful degradation
3. **Union merge with frontmatter**: Verify deduplication works correctly

---

## Performance Requirements

| Scenario | Requirement | Measurement |
|----------|------------|-------------|
| Single file (1KB) | <2ms | Benchmark test |
| 100 files (concurrent) | <20ms wall time | Integration test |
| Large file (100KB) | <50ms | Benchmark test |

**Enforcement**: Benchmark tests in `internal/adapters/d2/parser_bench_test.go`

---

## Dependencies

**Adapter Implementation Uses**:
- `oss.terrastruct.com/d2/d2lib` - Official D2 parsing library
- `oss.terrastruct.com/d2/d2graph` - D2 graph structures

**Version Pinning**:
- Pin to specific version in `go.mod` to prevent breaking changes
- Update via explicit dependency upgrade, not automatic

---

## Mock Implementation (for Testing)

```go
// MockD2Parser implements D2Parser for testing.
type MockD2Parser struct {
    ParseRelationshipsFunc func(ctx context.Context, d2Source string) ([]entities.D2Relationship, error)
}

func (m *MockD2Parser) ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error) {
    if m.ParseRelationshipsFunc != nil {
        return m.ParseRelationshipsFunc(ctx, d2Source)
    }
    return nil, nil
}
```

**Usage in Tests**:
```go
mockParser := &MockD2Parser{
    ParseRelationshipsFunc: func(ctx context.Context, d2Source string) ([]entities.D2Relationship, error) {
        return []entities.D2Relationship{
            {Source: "a", Target: "b", Label: "test"},
        }, nil
    },
}

useCase := NewBuildArchitectureGraph(projectRepo, mockParser)
graph, err := useCase.Execute(ctx)
// Verify graph has expected edges from mock relationships
```

---

## Validation Checklist

- [ ] Returns empty slice (not nil) for valid D2 without relationships
- [ ] Returns empty slice (not nil) for empty string input
- [ ] Returns error for invalid D2 syntax with descriptive message
- [ ] Handles context cancellation correctly
- [ ] Extracts source, target, label correctly from arrows
- [ ] Handles unlabeled arrows (empty label)
- [ ] Handles multi-target arrows (multiple edges from one source)
- [ ] Performance: <2ms for typical component D2 file
- [ ] Thread-safe (can be called concurrently)

---

**Status**: ✅ Contract complete - Ready for implementation
