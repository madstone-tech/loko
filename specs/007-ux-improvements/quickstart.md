# Developer Quickstart: Contributing to UX Improvements

**Feature**: 007-ux-improvements  
**Target Audience**: Contributors implementing relationship parsing, template selection, and drift detection  
**Prerequisites**: Go 1.25+, familiarity with Clean Architecture

---

## Getting Started

### Clone and Setup

```bash
# Clone repository
git clone https://github.com/madstone-tech/loko.git
cd loko

# Checkout feature branch
git checkout 007-ux-improvements

# Install dependencies
go mod download

# Run tests to verify setup
go test ./...

# Build binary
make build
```

---

## Project Structure Overview

```
internal/
├── core/                   # Business logic (ZERO external deps)
│   ├── entities/          # Domain objects + validation
│   │   ├── template_selector.go  # NEW - Template pattern matching
│   │   └── drift_issue.go        # NEW - Drift detection types
│   └── usecases/          # Application logic
│       ├── ports.go       # Interface definitions (add D2Parser, TemplateRegistry)
│       └── detect_drift.go       # NEW - Drift detection use case
│
├── adapters/              # External dependency implementations
│   ├── d2/
│   │   └── parser.go      # NEW - D2 relationship parser
│   └── ason/
│       └── template_registry.go  # NEW - Template file resolver
│
└── mcp/                   # MCP server (thin layer)
    └── tools/             # Tool handlers (<30 lines each)
```

**Key Principle**: Core has zero external dependencies. D2 parsing library lives in adapter layer.

---

## Task 1: Implementing D2 Parser

### Step 1: Define Port Interface

Edit `internal/core/usecases/ports.go`:

```go
// D2Parser parses D2 diagram syntax to extract relationships.
type D2Parser interface {
    ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error)
}
```

### Step 2: Create D2Relationship Entity

Create `internal/core/entities/d2_relationship.go`:

```go
package entities

import (
    "errors"
    "fmt"
)

// D2Relationship represents a relationship extracted from D2 syntax.
type D2Relationship struct {
    Source string
    Target string
    Label  string
}

// NewD2Relationship creates a validated D2Relationship.
func NewD2Relationship(source, target, label string) (*D2Relationship, error) {
    if source == "" {
        return nil, errors.New("source cannot be empty")
    }
    if target == "" {
        return nil, errors.New("target cannot be empty")
    }
    return &D2Relationship{
        Source: source,
        Target: target,
        Label:  label,
    }, nil
}

// Key returns unique identifier for deduplication.
func (r *D2Relationship) Key() string {
    return fmt.Sprintf("%s->%s:%s", r.Source, r.Target, r.Label)
}
```

### Step 3: Write Tests First (TDD)

Create `internal/adapters/d2/parser_test.go`:

```go
package d2_test

import (
    "context"
    "testing"
    
    "github.com/madstone-tech/loko/internal/adapters/d2"
    "github.com/madstone-tech/loko/internal/core/entities"
)

func TestParser_ParseRelationships(t *testing.T) {
    tests := []struct {
        name        string
        d2Source    string
        want        []entities.D2Relationship
        wantErr     bool
    }{
        {
            name: "single arrow",
            d2Source: `a -> b: "label"`,
            want: []entities.D2Relationship{
                {Source: "a", Target: "b", Label: "label"},
            },
            wantErr: false,
        },
        {
            name: "empty source",
            d2Source: "",
            want: []entities.D2Relationship{},
            wantErr: false,
        },
        {
            name: "invalid syntax",
            d2Source: `invalid {{{{`,
            want: nil,
            wantErr: true,
        },
        // Add more test cases...
    }
    
    parser := d2.NewParser()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parser.ParseRelationships(context.Background(), tt.d2Source)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseRelationships() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !equalRelationships(got, tt.want) {
                t.Errorf("ParseRelationships() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Step 4: Implement Adapter

Create `internal/adapters/d2/parser.go`:

```go
package d2

import (
    "context"
    "fmt"
    
    "oss.terrastruct.com/d2/d2lib"
    "oss.terrastruct.com/d2/d2graph"
    
    "github.com/madstone-tech/loko/internal/core/entities"
)

// Parser implements D2Parser using official D2 libraries.
type Parser struct{}

// NewParser creates a new D2 parser.
func NewParser() *Parser {
    return &Parser{}
}

// ParseRelationships extracts relationship arrows from D2 source.
func (p *Parser) ParseRelationships(ctx context.Context, d2Source string) ([]entities.D2Relationship, error) {
    if d2Source == "" {
        return []entities.D2Relationship{}, nil
    }

    graph, err := d2lib.Parse(ctx, d2Source, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to parse D2 source: %w", err)
    }

    var relationships []entities.D2Relationship
    for _, edge := range graph.Edges {
        rel, err := entities.NewD2Relationship(
            edge.Src.AbsID(),
            edge.Dst.AbsID(),
            edge.Label.Value,
        )
        if err != nil {
            // Skip invalid relationships (defensive programming)
            continue
        }
        relationships = append(relationships, *rel)
    }

    return relationships, nil
}
```

### Step 5: Wire to Use Case

Update `internal/core/usecases/build_architecture_graph.go`:

```go
type BuildArchitectureGraph struct {
    projectRepo ProjectRepository
    d2Parser    D2Parser  // NEW
}

func (uc *BuildArchitectureGraph) Execute(ctx context.Context) (*entities.Graph, error) {
    // 1. Load components from filesystem
    components := uc.projectRepo.LoadAllComponents()
    
    // 2. Parse relationships from frontmatter
    frontmatterRels := extractFrontmatterRelationships(components)
    
    // 3. Parse relationships from D2 files (NEW)
    d2Rels := []entities.D2Relationship{}
    for _, comp := range components {
        d2Path := filepath.Join(comp.Path, comp.ID+".d2")
        d2Source, err := os.ReadFile(d2Path)
        if err != nil {
            continue // Skip missing D2 files
        }
        
        rels, err := uc.d2Parser.ParseRelationships(ctx, string(d2Source))
        if err != nil {
            // Graceful degradation: log warning, continue
            log.Warn("Failed to parse D2 file %s: %v", d2Path, err)
            continue
        }
        d2Rels = append(d2Rels, rels...)
    }
    
    // 4. Union merge (deduplicate)
    graph := unionMerge(frontmatterRels, d2Rels)
    
    return graph, nil
}
```

### Step 6: Run Tests

```bash
# Run unit tests
go test ./internal/adapters/d2/...

# Run integration tests
go test ./tests/integration/...

# Check coverage
go test -cover ./internal/core/...
```

---

## Task 2: Implementing Template Selection

### Step 1: Create TemplateSelector Entity

Create `internal/core/entities/template_selector.go`:

```go
package entities

import "strings"

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

func (t TemplateType) String() string {
    switch t {
    case TemplateCompute:
        return "compute"
    case TemplateDatastore:
        return "datastore"
    case TemplateMessaging:
        return "messaging"
    case TemplateAPI:
        return "api"
    case TemplateEvent:
        return "event"
    case TemplateStorage:
        return "storage"
    default:
        return "generic"
    }
}

var TechnologyPatterns = map[TemplateType][]string{
    TemplateCompute:    {"lambda", "function", "fargate", "ecs task"},
    TemplateDatastore:  {"dynamodb", "database", "table", "rds", "aurora"},
    TemplateMessaging:  {"sqs", "queue", "sns", "topic", "kinesis"},
    TemplateAPI:        {"api gateway", "rest", "graphql", "endpoint"},
    TemplateEvent:      {"eventbridge", "event", "step functions"},
    TemplateStorage:    {"s3", "bucket", "efs"},
}

// SelectTemplate selects template type based on technology string.
func SelectTemplate(technology string, override *TemplateType) TemplateType {
    if override != nil {
        return *override
    }
    
    tech := strings.ToLower(technology)
    for tmplType, patterns := range TechnologyPatterns {
        for _, pattern := range patterns {
            if strings.Contains(tech, pattern) {
                return tmplType
            }
        }
    }
    return TemplateGeneric
}
```

### Step 2: Write Template Files

Create `templates/component/compute.md`:

```markdown
---
id: {component-id}
name: "{component-name}"
description: "{component-description}"
technology: "{component-technology}"
tags: {component-tags}
---

# {component-name}

{component-description}

## Configuration
- **Trigger**: (Event source - API Gateway, SQS, EventBridge)
- **Runtime**: (e.g., Node.js 20, Python 3.12)
- **Timeout**: (seconds)
- **Memory**: (MB)
- **Environment Variables**: (list key variables)

## Implementation
- **Handler**: (entry point function)
- **Dependencies**: (runtime dependencies)
- **Layers**: (Lambda layers used)

## Error Handling
- **Retry Policy**: (max retries, backoff strategy)
- **Dead Letter Queue**: (DLQ configuration)
- **Logging**: (CloudWatch log group)

## Performance Considerations
- **Cold Start**: (optimization strategies)
- **Concurrency**: (reserved/provisioned concurrency)
```

Repeat for `datastore.md`, `messaging.md`, `api.md`, `event.md`, `storage.md`, `generic.md`.

### Step 3: Test Template Selection

Create `internal/core/entities/template_selector_test.go`:

```go
func TestSelectTemplate(t *testing.T) {
    tests := []struct {
        technology string
        override   *TemplateType
        want       TemplateType
    }{
        {"AWS Lambda", nil, TemplateCompute},
        {"DynamoDB Table", nil, TemplateDatastore},
        {"SQS Queue", nil, TemplateMessaging},
        {"Unknown Tech", nil, TemplateGeneric},
        {"Lambda", &TemplateDatastore, TemplateDatastore}, // Override
    }
    
    for _, tt := range tests {
        t.Run(tt.technology, func(t *testing.T) {
            got := SelectTemplate(tt.technology, tt.override)
            if got != tt.want {
                t.Errorf("SelectTemplate(%q) = %v, want %v", tt.technology, got, tt.want)
            }
        })
    }
}
```

---

## Task 3: Adding Custom Technology Patterns

### How to Add New Patterns

Edit `internal/core/entities/template_selector.go`:

```go
var TechnologyPatterns = map[TemplateType][]string{
    TemplateCompute:    {"lambda", "function", "fargate", "ecs task", "cloud run"},  // Added "cloud run"
    TemplateDatastore:  {"dynamodb", "database", "table", "rds", "aurora", "mongodb"},  // Added "mongodb"
    // ...
}
```

**Testing New Patterns**:
```go
func TestSelectTemplate_CustomPattern(t *testing.T) {
    got := SelectTemplate("Google Cloud Run", nil)
    if got != TemplateCompute {
        t.Errorf("Cloud Run should match Compute template, got %v", got)
    }
}
```

---

## Task 4: Creating Custom Templates

### Step 1: Create Template File

Create `templates/component/my-custom-template.md`:

```markdown
---
id: {component-id}
name: "{component-name}"
description: "{component-description}"
technology: "{component-technology}"
---

# {component-name}

## Custom Section 1
(Your custom content here)

## Custom Section 2
(Your custom content here)
```

### Step 2: Add to TemplateType Enum

Edit `internal/core/entities/template_selector.go`:

```go
const (
    TemplateCompute TemplateType = iota
    TemplateDatastore
    // ...
    TemplateMyCustom  // NEW
)

func (t TemplateType) String() string {
    switch t {
    // ...
    case TemplateMyCustom:
        return "my-custom-template"
    }
}

var TechnologyPatterns = map[TemplateType][]string{
    // ...
    TemplateMyCustom: {"my-tech", "custom-service"},
}
```

### Step 3: Test Custom Template

```bash
loko new component test-comp --technology "my-tech"
# Should create component with my-custom-template.md
```

---

## Common Development Tasks

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/adapters/d2/...

# With coverage
go test -cover ./internal/core/...

# Verbose
go test -v ./tests/integration/...

# Single test
go test -run TestParser_ParseRelationships ./internal/adapters/d2/...
```

### Running Linter

```bash
task lint

# Auto-fix issues
task fmt
```

### Building Binary

```bash
make build

# Or
go build -o loko .
```

### Testing Changes Locally

```bash
# Create test project
mkdir test-project
cd test-project

# Initialize loko project
../loko init notification-service

# Create component with new template
../loko new component email-queue --technology "SQS"

# Verify template used
cat src/notification-service/message-queue/email-queue/component.md
```

---

## Debugging Tips

### D2 Parsing Issues

```bash
# Enable debug logging
export LOKO_LOG_LEVEL=debug

# Run command
loko build

# Check D2 parser output
# Look for warnings like "Failed to parse D2 file..."
```

### Template Selection Issues

```go
// Add debug logging in SelectTemplate
func SelectTemplate(technology string, override *TemplateType) TemplateType {
    log.Debug("SelectTemplate called with technology: %s", technology)
    
    if override != nil {
        log.Debug("Override provided: %v", override)
        return *override
    }
    
    tech := strings.ToLower(technology)
    for tmplType, patterns := range TechnologyPatterns {
        for _, pattern := range patterns {
            if strings.Contains(tech, pattern) {
                log.Debug("Matched pattern %q for type %v", pattern, tmplType)
                return tmplType
            }
        }
    }
    
    log.Debug("No pattern matched, falling back to Generic")
    return TemplateGeneric
}
```

---

## Performance Benchmarking

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Specific benchmark
go test -bench=BenchmarkParse100Components ./internal/adapters/d2/...

# With memory profiling
go test -bench=. -benchmem ./...
```

### Adding New Benchmarks

Create `internal/adapters/d2/parser_bench_test.go`:

```go
func BenchmarkParseRelationships(b *testing.B) {
    parser := d2.NewParser()
    d2Source := `a -> b: "label"`
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = parser.ParseRelationships(ctx, d2Source)
    }
}
```

---

## Reference Documentation

- **Feature Spec**: `specs/007-ux-improvements/spec.md`
- **Implementation Plan**: `specs/007-ux-improvements/plan.md`
- **Data Model**: `specs/007-ux-improvements/data-model.md`
- **Contracts**: `specs/007-ux-improvements/contracts/`
- **Constitution**: `.specify/memory/constitution.md` v1.0.0

---

## Getting Help

- **Questions**: Open discussion in GitHub Discussions
- **Bugs**: File issue with template or label `007-ux-improvements`
- **Pull Requests**: Reference this spec and ensure all tests pass

---

**Happy Contributing!** 🚀
