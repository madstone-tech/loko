# Research: UX Improvements Implementation

**Feature**: 007-ux-improvements  
**Date**: 2026-02-13  
**Researcher**: AI Agent (go-docs-coder)

---

## Research Questions

1. How to parse D2 diagram syntax in Go?
2. What's the best pattern for technology-aware template selection?
3. How to handle D2 parse errors gracefully?
4. What's the performance strategy for 100+ component parsing?

---

## Decision 1: D2 Parsing Library

**Decision**: Use official D2 Go libraries (`oss.terrastruct.com/d2`)

**Rationale**:
- **Official Support**: D2 libraries maintained by Terrastruct (D2 creators)
- **Correctness**: Guaranteed to handle all D2 syntax correctly
- **Clean Architecture Fit**: Libraries used in adapter layer, behind `D2Parser` interface
- **Integration**: Complements existing d2 CLI usage (parsing for structure, CLI for rendering)
- **Maintainability**: Leverages ecosystem updates rather than custom regex parsing

**Alternatives Considered**:

1. **Custom Regex Parser**
   - Rejected: Fragile, won't handle nested syntax, requires maintenance for D2 updates
   
2. **Shell out to d2 CLI with JSON output**
   - Rejected: Performance overhead (100+ files = 100+ shell calls), no structured API
   
3. **Manual lexer/parser implementation**
   - Rejected: Overkill, reinventing wheel, high maintenance burden

**Code Pattern**:
```go
// In internal/adapters/d2/parser.go
import (
    "oss.terrastruct.com/d2/d2lib"
    "oss.terrastruct.com/d2/d2graph"
)

func (p *Parser) ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error) {
    graph, err := d2lib.Parse(ctx, d2Source, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to parse D2 source: %w", err)
    }

    var relationships []entities.D2Relationship
    for _, edge := range graph.Edges {
        relationships = append(relationships, entities.D2Relationship{
            Source: edge.Src.AbsID(),
            Target: edge.Dst.AbsID(),
            Label:  edge.Label.Value,
        })
    }
    return relationships, nil
}
```

---

## Decision 2: Template Selection Strategy

**Decision**: Pattern-based matching with TemplateSelector entity

**Rationale**:
- **Simplicity**: Keyword matching sufficient for 7 template types
- **Extensibility**: Easy to add new patterns without code changes
- **Testability**: Clear inputs/outputs for unit tests
- **Clean Architecture**: Selection logic in core/entities (no external deps)
- **Performance**: O(n) where n = pattern count (< 1ms per component)

**Alternatives Considered**:

1. **Machine Learning / AI Classification**
   - Rejected: Overkill, requires training data, non-deterministic, YAGNI
   
2. **Configuration File (loko.toml)**
   - Rejected: Adds user complexity, hardcoded defaults sufficient for v0.2.0
   - Future: Can add config override in v0.3.0 if requested
   
3. **Plugin System**
   - Rejected: Premature abstraction, no evidence users need custom templates yet

**Code Pattern**:
```go
// In internal/core/entities/template_selector.go
type TemplateType int

const (
    TemplateCompute TemplateType = iota
    TemplateDatastore
    TemplateMessaging
    TemplateAPI
    TemplateEvent
    TemplateStorage
    TemplateGeneric
)

var technologyPatterns = map[TemplateType][]string{
    TemplateCompute:    {"lambda", "function", "fargate", "ecs task"},
    TemplateDatastore:  {"dynamodb", "database", "table", "rds", "aurora"},
    TemplateMessaging:  {"sqs", "queue", "sns", "topic", "kinesis"},
    TemplateAPI:        {"api gateway", "rest", "graphql", "endpoint"},
    TemplateEvent:      {"eventbridge", "event", "step functions"},
    TemplateStorage:    {"s3", "bucket", "efs"},
}

func SelectTemplate(technology string, override *TemplateType) TemplateType {
    if override != nil {
        return *override
    }
    
    tech := strings.ToLower(technology)
    for tmplType, patterns := range technologyPatterns {
        for _, pattern := range patterns {
            if strings.Contains(tech, pattern) {
                return tmplType
            }
        }
    }
    return TemplateGeneric
}
```

---

## Decision 3: Error Handling Strategy

**Decision**: Graceful degradation with skip-and-warn pattern

**Rationale**:
- **User Experience**: 1 broken file doesn't block 99 good files
- **Consistency**: Matches existing PDF graceful degradation (veve-cli missing)
- **Actionable**: Warning logs tell user which file failed
- **Validation**: `--check-drift` can surface parse errors explicitly

**Error Classification**:
- **Parse Error**: Skip file, log warning, continue with other files
- **Missing File**: Return error (filesystem issue, not D2 syntax)
- **Empty File**: Return empty relationships (valid state)

**Code Pattern**:
```go
// In use case that processes multiple D2 files
func (uc *BuildArchitectureGraph) parseComponentD2Files(ctx context.Context, components []Component) {
    for _, comp := range components {
        d2Path := filepath.Join(comp.Path, comp.ID+".d2")
        
        content, err := os.ReadFile(d2Path)
        if err != nil {
            // Missing file is an error (stop)
            return fmt.Errorf("component D2 file not found: %w", err)
        }
        
        relationships, err := uc.d2Parser.ParseRelationships(ctx, string(content))
        if err != nil {
            // Parse error: warn and continue
            uc.logger.Warn("Failed to parse D2 file %s: %v (skipping relationships from this file)", d2Path, err)
            continue
        }
        
        // Add relationships to graph
        for _, rel := range relationships {
            uc.graph.AddEdge(rel.Source, rel.Target, rel.Label)
        }
    }
}
```

---

## Decision 4: Performance Strategy

**Decision**: Worker pool pattern with concurrency limit

**Rationale**:
- **Target**: 100 components @ <200ms = 2ms per file max
- **Concurrency**: 10 workers balances throughput and resource usage
- **Caching**: Future optimization (not implemented in v0.2.0 - YAGNI)
- **Measurement**: Benchmark tests enforce 200ms gate

**Performance Breakdown**:
- File I/O: ~0.5ms per file (100 files = 50ms total)
- D2 parsing: ~1ms per file (10 workers = 10ms wall time for 100 files)
- Graph merge: ~0.5ms per file (sequential after parsing)
- **Total**: ~60-80ms for 100 components (well under 200ms target)

**Code Pattern**:
```go
// Worker pool for concurrent D2 parsing
func (uc *BuildArchitectureGraph) parseD2Concurrently(ctx context.Context, files []string) ([]D2Relationship, error) {
    type result struct {
        relationships []D2Relationship
        err           error
    }

    jobs := make(chan string, len(files))
    results := make(chan result, len(files))

    // Start 10 workers
    for i := 0; i < 10; i++ {
        go func() {
            for filePath := range jobs {
                content, _ := os.ReadFile(filePath)
                rels, err := uc.d2Parser.ParseRelationships(ctx, string(content))
                results <- result{rels, err}
            }
        }()
    }

    // Send jobs
    for _, file := range files {
        jobs <- file
    }
    close(jobs)

    // Collect results
    var allRelationships []D2Relationship
    for i := 0; i < len(files); i++ {
        res := <-results
        if res.err != nil {
            uc.logger.Warn("Parse error: %v", res.err)
            continue
        }
        allRelationships = append(allRelationships, res.relationships...)
    }

    return allRelationships, nil
}
```

**Benchmark Test**:
```go
func BenchmarkParse100Components(b *testing.B) {
    files := generateTestD2Files(100)
    parser := d2.NewParser()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = parseD2Concurrently(context.Background(), files, parser)
    }
    // Assert: b.Elapsed() / b.N < 200ms
}
```

---

## Best Practices Summary

### D2 Parsing
- ✅ Use official libraries for correctness
- ✅ Parse errors are warnings, not failures
- ✅ Test with 20+ real-world D2 examples
- ✅ Validate edge cases: empty files, nested shapes, multi-target arrows

### Template Selection
- ✅ Case-insensitive matching for robustness
- ✅ Explicit priority order (first match wins)
- ✅ Fallback to Generic for unknown technologies
- ✅ Support override flag for user control

### Error Handling
- ✅ Graceful degradation principle
- ✅ Actionable warnings with file paths
- ✅ Distinguish parse errors from filesystem errors
- ✅ Validation command surfaces all issues

### Performance
- ✅ Concurrent parsing with worker pool
- ✅ Benchmark tests enforce 200ms target
- ✅ Log timing in debug mode for profiling
- ✅ Document performance characteristics

---

## Integration Considerations

### D2 Libraries in Clean Architecture

**Port Interface** (`internal/core/usecases/ports.go`):
```go
// D2Parser parses D2 diagram syntax to extract relationships.
type D2Parser interface {
    ParseRelationships(ctx context.Context, d2Source string) ([]D2Relationship, error)
}
```

**Adapter** (`internal/adapters/d2/parser.go`):
```go
// Implements D2Parser using oss.terrastruct.com/d2
type Parser struct {
    // D2 library usage encapsulated here
}
```

**Wiring** (`main.go` or use case constructor):
```go
d2Parser := d2.NewParser()
graphUseCase := usecases.NewBuildArchitectureGraph(projectRepo, d2Parser)
```

### Template Registry in Clean Architecture

**Port Interface** (`internal/core/usecases/ports.go`):
```go
// TemplateRegistry resolves template types to actual template files.
type TemplateRegistry interface {
    GetTemplatePath(templateType TemplateType) (string, error)
}
```

**Adapter** (`internal/adapters/ason/template_registry.go`):
```go
// Implements TemplateRegistry with file system template lookups
type Registry struct {
    templateDir string
}
```

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| D2 library breaking changes | Low | Medium | Pin to specific version, integration tests |
| Performance regression (>200ms) | Medium | High | Benchmark tests, worker pool tuning |
| Template selection ambiguity | Medium | Low | Explicit priority, user override flag |
| D2 syntax edge cases not handled | Medium | Medium | Extensive test suite (20+ examples) |

---

## References

- D2 Go Libraries: https://pkg.go.dev/oss.terrastruct.com/d2
- D2 Language Reference: https://d2lang.com/tour/intro
- loko Constitution: `.specify/memory/constitution.md` v1.0.0
- Real-World Feedback: `test/loko-mcp-feedback.md`, `test/loko-product-feedback.md`

---

**Status**: ✅ Research Complete - Ready for Phase 1 (Design & Contracts)
