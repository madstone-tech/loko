# Developer Quickstart: Phase 1 Completion

**Feature**: 006-phase-1-completion  
**Target**: v0.2.0 Release  
**Date**: 2026-02-13

---

## Overview

This guide helps developers contribute to the Phase 1 completion release. It covers:

1. **Development setup** - Getting started with the codebase
2. **Architecture principles** - Clean Architecture and constitution rules
3. **Implementation patterns** - How to add features correctly
4. **Testing requirements** - Coverage and validation
5. **Common tasks** - Handler refactoring, adding MCP tools, etc.

---

## Prerequisites

**Required**:
- Go 1.25+
- d2 CLI (`brew install d2` or https://d2lang.com)
- Git

**Optional**:
- veve-cli (for PDF generation)
- Docker (for testing CI examples)

---

## Development Setup

### 1. Clone and Build

```bash
# Clone repository
git clone https://github.com/madstone-tech/loko.git
cd loko

# Checkout feature branch
git checkout 006-phase-1-completion

# Install dependencies
go mod download

# Build binary
make build
# or
go build -o loko .

# Verify build
./loko --version
```

### 2. Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
task coverage
# or
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test ./internal/core/usecases/...

# Run with verbose output
go test -v ./...
```

### 3. Development Workflow

```bash
# Watch for changes and rebuild (manual)
while true; do
    make build && echo "✓ Built successfully"
    sleep 2
done

# Or use entr for file watching
ls **/*.go | entr -r make build
```

---

## Architecture Overview

### Layer Structure

```
internal/
├── core/                     # ZERO external dependencies
│   ├── entities/             # Domain objects + validation
│   └── usecases/             # Business logic + ports.go
├── adapters/                 # Infrastructure (implements ports)
│   ├── filesystem/           # ProjectRepository
│   ├── d2/                   # DiagramRenderer
│   ├── encoding/             # OutputEncoder (JSON, TOON)
│   └── ...
├── mcp/                      # MCP server (thin layer)
├── api/                      # HTTP API (thin layer)
└── ui/                       # Lipgloss styles

cmd/                          # CLI commands (thin layer)
```

### Dependency Rules

**✅ Allowed**:
- `entities/` → stdlib only
- `usecases/` → `entities/` + stdlib
- `adapters/` → `usecases/` interfaces + `entities/`
- `mcp/`, `api/`, `cmd/` → `usecases/` + `adapters/`

**❌ Forbidden**:
- `core/` → `adapters/`, `mcp/`, `api/`, `cmd/`
- `adapters/` → `mcp/`, `api/`, `cmd/`
- Circular dependencies

---

## Constitution Principles

### 1. Thin Handlers

**CLI commands**: < 50 lines (excluding imports/comments)  
**MCP tools**: < 30 lines (excluding imports/comments)  
**API handlers**: < 50 lines (excluding imports/comments)

**Pattern**:
```go
func HandlerFunction(args Args) error {
    // 1. Parse input (5-10 lines)
    req := parseToRequestStruct(args)
    
    // 2. Call use case (1 line)
    result, err := useCase.Execute(ctx, req)
    if err != nil {
        return formatError(err)
    }
    
    // 3. Format output (5-10 lines)
    fmt.Printf("Success: %v\n", result)
    return nil
}
```

### 2. Use Cases Contain Business Logic

**Never in handlers**:
- Entity creation
- Validation
- File I/O orchestration
- Complex transformations

**Always in use cases**:
```go
// internal/core/usecases/my_feature.go
type MyFeatureRequest struct {
    Field1 string
    Field2 int
}

type MyFeatureResponse struct {
    ResultID string
    Success  bool
}

type MyFeatureUseCase struct {
    repo ProjectRepository  // Port interface
}

func (uc *MyFeatureUseCase) Execute(ctx context.Context, req MyFeatureRequest) (*MyFeatureResponse, error) {
    // Business logic here
    // Validation, orchestration, error handling
}
```

### 3. Interfaces in ports.go

**All external dependencies behind interfaces**:
```go
// internal/core/usecases/ports.go
type DiagramRenderer interface {
    Render(ctx context.Context, d2Code string) ([]byte, error)
}

type ProjectRepository interface {
    LoadProject(ctx context.Context, root string) (*entities.Project, error)
    SaveProject(ctx context.Context, project *entities.Project) error
}
```

### 4. Entity Validation

**Validation in entity constructors**:
```go
// internal/core/entities/system.go
func NewSystem(name, description string) (*System, error) {
    if name == "" {
        return nil, errors.New("system name required")
    }
    // Validation logic here
    return &System{Name: name, Description: description}, nil
}
```

---

## Common Tasks

### Task 1: Add New MCP Tool

**Example**: Add `search_elements` tool

**Step 1**: Create use case (if needed)
```go
// internal/core/usecases/search_elements.go
type SearchElementsRequest struct {
    ProjectRoot string
    Query       string
    Type        string
    Tag         string
    Technology  string
    Limit       int
}

type SearchElementsResponse struct {
    Results      []ElementResult
    TotalMatched int
    QueryTimeMs  int64
}

type SearchElementsUseCase struct {
    repo ProjectRepository
}

func (uc *SearchElementsUseCase) Execute(ctx context.Context, req SearchElementsRequest) (*SearchElementsResponse, error) {
    // Implementation
}
```

**Step 2**: Create MCP tool handler (< 30 lines)
```go
// internal/mcp/tools/search_elements.go
package tools

type SearchElementsTool struct {
    useCase *usecases.SearchElementsUseCase
}

func NewSearchElementsTool(repo usecases.ProjectRepository) *SearchElementsTool {
    return &SearchElementsTool{
        useCase: usecases.NewSearchElementsUseCase(repo),
    }
}

func (t *SearchElementsTool) Name() string {
    return "search_elements"
}

func (t *SearchElementsTool) Description() string {
    return "Search architecture elements by name, technology, or tags"
}

func (t *SearchElementsTool) InputSchema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "query": map[string]any{
                "type": "string",
                "description": "Search query (supports glob patterns)",
            },
            // ... other properties
        },
        "required": []string{"query"},
    }
}

func (t *SearchElementsTool) Call(ctx context.Context, args map[string]any) (any, error) {
    // Parse args (5 lines)
    req := usecases.SearchElementsRequest{
        Query: args["query"].(string),
        // ... parse other args
    }
    
    // Call use case (1 line)
    result, err := t.useCase.Execute(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Format response (5 lines)
    return map[string]any{
        "results": result.Results,
        "total_matched": result.TotalMatched,
    }, nil
}
```

**Step 3**: Register tool
```go
// cmd/mcp.go
func registerTools(server *mcp.Server, repo *filesystem.ProjectRepository) error {
    toolList := []mcp.Tool{
        // Existing tools...
        tools.NewSearchElementsTool(repo),  // Add here
    }
    
    for _, tool := range toolList {
        if err := server.RegisterTool(tool); err != nil {
            return err
        }
    }
    return nil
}
```

**Step 4**: Write tests
```go
// internal/mcp/tools/search_elements_test.go
func TestSearchElementsTool_Execute(t *testing.T) {
    // Table-driven test
    tests := []struct {
        name string
        args map[string]any
        want int  // Expected result count
    }{
        {"search by name", map[string]any{"query": "payment"}, 1},
        {"empty results", map[string]any{"query": "nonexistent"}, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tool := setupTestTool(t)
            result, err := tool.Call(ctx, tt.args)
            require.NoError(t, err)
            assert.Len(t, result.(map[string]any)["results"], tt.want)
        })
    }
}
```

---

### Task 2: Refactor Oversized Handler

**Example**: Refactor `cmd/new.go` (504 lines → < 50 lines)

**Step 1**: Analyze current handler
```bash
# Count lines excluding imports/comments
grep -v '^import\|^/\|^$\|^package' cmd/new.go | wc -l
```

**Step 2**: Identify business logic
- Entity creation
- Template loading
- D2 generation
- File writing
- Success messaging

**Step 3**: Extract to use case
```go
// internal/core/usecases/scaffold_entity.go
type ScaffoldEntityRequest struct {
    ProjectRoot string
    EntityType  string  // "system", "container", "component"
    EntityName  string
    Description string
    ParentID    string  // For containers/components
    Template    string
}

type ScaffoldEntityResponse struct {
    EntityID      string
    FilesCreated  []string
}

type ScaffoldEntityUseCase struct {
    repo      ProjectRepository
    renderer  DiagramRenderer
    templates TemplateEngine
}

func (uc *ScaffoldEntityUseCase) Execute(ctx context.Context, req ScaffoldEntityRequest) (*ScaffoldEntityResponse, error) {
    // ALL business logic moves here
}
```

**Step 4**: Thin handler delegates
```go
// cmd/new.go (now < 50 lines)
func NewEntityCommand(entityType string, args []string) error {
    req := usecases.ScaffoldEntityRequest{
        ProjectRoot: getProjectRoot(),
        EntityType:  entityType,
        EntityName:  args[0],
        Description: getFlag("description"),
        ParentID:    getFlag("parent"),
        Template:    getFlag("template"),
    }
    
    result, err := scaffoldUseCase.Execute(ctx, req)
    if err != nil {
        return formatError(err)
    }
    
    fmt.Printf("✓ Created %s: %s\n", entityType, result.EntityID)
    fmt.Printf("  Files: %s\n", strings.Join(result.FilesCreated, ", "))
    return nil
}
```

**Step 5**: Verify line count
```bash
# Should be < 50
grep -v '^import\|^/\|^$\|^package' cmd/new.go | wc -l
```

---

### Task 3: Add Configuration Option

**Example**: Add `[api] rate_limit` setting

**Step 1**: Extend entity
```go
// internal/core/entities/project.go
type ProjectConfig struct {
    // Existing fields...
    API *APIConfig `toml:"api"`
}

type APIConfig struct {
    RateLimit      int      `toml:"rate_limit"`
    AllowedOrigins []string `toml:"allowed_origins"`
    Timeout        string   `toml:"timeout"`
}
```

**Step 2**: Update loader
```go
// internal/adapters/config/loader.go
func (l *Loader) Load(path string) (*entities.ProjectConfig, error) {
    var config entities.ProjectConfig
    
    // Parse TOML
    if err := toml.DecodeFile(path, &config); err != nil {
        return nil, err
    }
    
    // Set defaults
    if config.API == nil {
        config.API = &entities.APIConfig{
            RateLimit: 100,
            AllowedOrigins: []string{"http://localhost:*"},
            Timeout: "30s",
        }
    }
    
    return &config, nil
}
```

**Step 3**: Use in API server
```go
// internal/api/server.go
func NewServer(config entities.ProjectConfig, repo ProjectRepository) *Server {
    // Apply rate limiting if configured
    if config.API != nil && config.API.RateLimit > 0 {
        handler = rateLimit(handler, config.API.RateLimit)
    }
    return &Server{...}
}
```

**Step 4**: Test configuration
```go
// internal/adapters/config/loader_test.go
func TestLoader_API_Config(t *testing.T) {
    config, err := loader.Load("testdata/loko.toml")
    require.NoError(t, err)
    
    assert.NotNil(t, config.API)
    assert.Equal(t, 100, config.API.RateLimit)
}
```

---

## Testing Requirements

### Coverage Targets

- `internal/core/`: > 80%
- All use cases: 100%
- Critical paths: 100%

### Test Types

**Unit Tests** (internal/core/):
```go
// Test use cases with mock repositories
func TestSearchElements_Success(t *testing.T) {
    mockRepo := &MockProjectRepository{
        Systems: []*entities.System{...},
    }
    uc := usecases.NewSearchElementsUseCase(mockRepo)
    
    result, err := uc.Execute(ctx, request)
    require.NoError(t, err)
    assert.NotEmpty(t, result.Results)
}
```

**Integration Tests** (tests/integration/):
```go
// Test with real filesystem
func TestIntegration_SearchMCPTool(t *testing.T) {
    repo := filesystem.NewProjectRepository()
    tool := tools.NewSearchElementsTool(repo)
    
    result, err := tool.Call(ctx, args)
    require.NoError(t, err)
}
```

**Benchmark Tests** (tests/benchmarks/):
```go
func BenchmarkSearchElements(b *testing.B) {
    // Setup
    repo := setupLargeProject(100, 500, 1000) // systems, containers, components
    uc := usecases.NewSearchElementsUseCase(repo)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = uc.Execute(ctx, request)
    }
}
```

---

## Constitution Audit

### Running Audit

```bash
# Run audit script
.specify/scripts/bash/constitution-audit.sh

# Expected output if passing:
✓ CLI handlers within limits
✓ MCP tools within limits
✓ Constitution audit passed

# Expected output if failing:
❌ VIOLATION: cmd/new.go has 504 lines (limit: 50)
❌ VIOLATION: internal/mcp/tools/tools.go has 1084 lines (limit: 30)
```

### Adding to CI

```yaml
# .github/workflows/constitution-audit.yml
name: Constitution Audit
on: [push, pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Constitution Audit
        run: .specify/scripts/bash/constitution-audit.sh
```

---

## Performance Guidelines

### Targets

- Search tools: < 200ms
- Watch mode rebuild: < 500ms
- MCP tool responses: < 100ms
- Build (10-system project): < 5 seconds

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Trace analysis
go test -trace=trace.out -bench=.
go tool trace trace.out
```

---

## Debugging Tips

### Enable Verbose Logging

```bash
# Set debug environment variable
export LOKO_DEBUG=true
./loko build

# Or inline
LOKO_DEBUG=true ./loko build
```

### MCP Tool Testing

```bash
# Start MCP server
./loko mcp

# Send JSON-RPC request (in another terminal)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search_elements","arguments":{"query":"payment"}}}' | ./loko mcp
```

### Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/core/usecases/

# Set breakpoint
(dlv) break usecases.SearchElements.Execute
(dlv) continue
```

---

## Troubleshooting

### Issue: "Cannot find d2 binary"

**Solution**: Install d2 CLI
```bash
brew install d2
# or
curl -fsSL https://d2lang.com/install.sh | sh -s --
```

### Issue: "Test coverage below 80%"

**Solution**: Add missing tests
```bash
# Find untested functions
go test -cover -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v '100.0%'
```

### Issue: "Handler exceeds line limit"

**Solution**: Extract to use case (see Task 2 above)

---

## Getting Help

- **Documentation**: `docs/` directory
- **Examples**: `examples/` directory
- **ADRs**: `docs/adr/` (architecture decisions)
- **Constitution**: `.specify/memory/constitution.md`
- **Issues**: https://github.com/madstone-tech/loko/issues

---

## Next Steps

1. Read `CONTRIBUTING.md` for contribution guidelines
2. Review existing use cases in `internal/core/usecases/`
3. Check `specs/006-phase-1-completion/` for detailed feature specs
4. Run tests to verify setup: `go test ./...`
5. Pick a task from `tasks.md` (generated via `/speckit.tasks`)

Happy coding! 🚀
