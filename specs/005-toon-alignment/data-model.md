# Data Model: TOON Alignment & Handler Refactoring

**Feature**: 005-toon-alignment | **Date**: 2026-02-06 | **Phase**: 1 (Design)

---

## 1. Existing Entity Model (No Changes)

The domain entities in `internal/core/entities/` are already well-structured and require **no structural changes**. Only `toon` struct tags are added for TOON serialization.

### 1.1 Entity Hierarchy

```
Project
├── Name, Description, Version
├── Config: ProjectConfig
├── Systems: map[string]*System
│   ├── Name, Description, Tags, External
│   ├── Technology: PrimaryLanguage, Framework, Database
│   ├── Relationships: Responsibilities, Dependencies, KeyUsers, ExternalSystems
│   ├── Containers: map[string]*Container
│   │   ├── Name, Description, Technology, Tags
│   │   ├── Components: map[string]*Component
│   │   │   ├── Name, Description, Technology, Tags
│   │   │   ├── Relationships: map[string]string
│   │   │   ├── CodeAnnotations: map[string]string
│   │   │   └── Dependencies: []string
│   │   └── Diagram
│   └── Diagram
└── Metadata: map[string]any
```

### 1.2 TOON Struct Tag Additions

Tags are added to entities for TOON serialization. These are non-breaking — they don't affect existing JSON serialization or entity behavior.

**Project** (`entities/project.go`):
```
Name        string            `json:"name"        toon:"name"`
Description string            `json:"description" toon:"description,omitempty"`
Version     string            `json:"version"     toon:"version,omitempty"`
Systems     map[string]*System `json:"systems"    toon:"systems"`
```

**System** (`entities/system.go`):
```
ID              string              `json:"id"          toon:"id"`
Name            string              `json:"name"        toon:"name"`
Description     string              `json:"description" toon:"description,omitempty"`
Tags            []string            `json:"tags"        toon:"tags,omitempty"`
PrimaryLanguage string              `json:"language"    toon:"language,omitempty"`
Framework       string              `json:"framework"   toon:"framework,omitempty"`
Database        string              `json:"database"    toon:"database,omitempty"`
External        bool                `json:"external"    toon:"external,omitempty"`
Containers      map[string]*Container `json:"containers" toon:"containers"`
```

**Container** (`entities/container.go`):
```
ID          string                `json:"id"          toon:"id"`
Name        string                `json:"name"        toon:"name"`
Description string                `json:"description" toon:"description,omitempty"`
Technology  string                `json:"technology"  toon:"technology,omitempty"`
Tags        []string              `json:"tags"        toon:"tags,omitempty"`
Components  map[string]*Component `json:"components"  toon:"components"`
```

**Component** (`entities/component.go`):
```
ID            string            `json:"id"           toon:"id"`
Name          string            `json:"name"         toon:"name"`
Description   string            `json:"description"  toon:"description,omitempty"`
Technology    string            `json:"technology"    toon:"technology,omitempty"`
Tags          []string          `json:"tags"          toon:"tags,omitempty"`
Relationships map[string]string `json:"relationships" toon:"relationships,omitempty"`
Dependencies  []string          `json:"dependencies"  toon:"dependencies,omitempty"`
```

---

## 2. New Port Interfaces

Added to `internal/core/usecases/ports.go`.

### 2.1 DiagramGenerator

Generates D2 diagram source code from domain entities. This is a domain service that knows C4 conventions.

```
DiagramGenerator interface {
    GenerateSystemContextDiagram(system *entities.System) (string, error)
    GenerateContainerDiagram(system *entities.System) (string, error)
    GenerateComponentDiagram(container *entities.Container) (string, error)
}
```

**Adapter**: `internal/adapters/d2/generator.go` (moved from `cmd/d2_generator.go`)

### 2.2 UserPrompter

Interactive prompts for CLI — optional dependency (MCP handlers don't need this).

```
UserPrompter interface {
    PromptString(prompt string, defaultValue string) (string, error)
    PromptStringMulti(prompt string) ([]string, error)
}
```

**Adapter**: `internal/adapters/cli/prompts.go` (extracted from `cmd/new.go`)

### 2.3 ReportFormatter

Formats validation reports for human-readable output.

```
ReportFormatter interface {
    PrintValidationReport(errors []ValidationError)
    PrintBuildReport(stats BuildStats)
}
```

**Adapter**: `internal/adapters/cli/report_formatter.go` (extracted from `cmd/validate.go`)

---

## 3. Use Case Input/Output Models

### 3.1 CreateContainerUseCase

```
Input:
    ProjectRoot string    // filesystem path
    SystemName  string    // parent system ID
    Name        string    // container display name
    Description string    // optional
    Technology  string    // optional
    Tags        []string  // optional

Output:
    Container   *entities.Container  // created entity
    FilesCreated []string            // paths of created files
```

### 3.2 CreateComponentUseCase

```
Input:
    ProjectRoot   string    // filesystem path
    SystemName    string    // grandparent system ID
    ContainerName string    // parent container ID
    Name          string    // component display name
    Description   string    // optional
    Technology    string    // optional
    Tags          []string  // optional

Output:
    Component    *entities.Component  // created entity
    FilesCreated []string             // paths of created files
```

### 3.3 ScaffoldEntityUseCase

Orchestrates: create entity + generate D2 diagram + update parent diagram + render template.

```
Input:
    ProjectRoot string              // filesystem path
    EntityType  string              // "system" | "container" | "component"
    ParentPath  []string            // [] for system, [system] for container, [system, container] for component
    Name        string              // entity display name
    Description string              // optional
    Technology  string              // optional
    Tags        []string            // optional
    Template    string              // template name (default: from config)

Output:
    EntityID     string             // created entity ID
    FilesCreated []string           // all created/modified files
    DiagramPath  string             // path to generated D2 diagram
```

**Dependencies**: ProjectRepository, TemplateEngine, DiagramGenerator

### 3.4 UpdateDiagramUseCase

```
Input:
    ProjectRoot string    // filesystem path
    DiagramPath string    // relative path to .d2 file
    D2Source    string    // D2 source code to write

Output:
    Written     bool      // success
    FilePath    string    // absolute path written
```

**Dependencies**: ProjectRepository

---

## 4. TOON Entity Mapping

### 4.1 Architecture Summary (detail: "summary")

```
TOON Output (~200 tokens):

project "My Architecture"
  version "1.0"
  systems[3]{name,containers,components}:
    PaymentService,2,5
    OrderService,3,8
    UserService,1,3
```

Maps from: `entities.Project` → summary stats

### 4.2 Architecture Structure (detail: "structure")

```
TOON Output (~500 tokens):

project "My Architecture"
  systems[3]:
    system "PaymentService"
      description "Handles payment processing"
      containers[2]{name,technology}:
        API,"Go + Fiber"
        Database,"PostgreSQL 15"
    system "OrderService"
      containers[3]{name,technology}:
        API,"Python + FastAPI"
        Queue,"RabbitMQ"
        Database,"PostgreSQL 15"
```

Maps from: `entities.Project` + `[]entities.System` with containers

### 4.3 Full Detail (detail: "full", target: specific system)

```
TOON Output (variable tokens):

system "PaymentService"
  description "Handles payment processing for all order types"
  language "Go"
  framework "Fiber"
  tags: payments, core
  containers[2]:
    container "API"
      description "REST API for payment operations"
      technology "Go + Fiber"
      components[3]{name,technology}:
        PaymentController,"Go handler"
        PaymentService,"Go package"
        StripeAdapter,"Go + stripe-go"
    container "Database"
      description "Payment data store"
      technology "PostgreSQL 15"
```

Maps from: `entities.System` + `[]entities.Container` + `[]entities.Component`

### 4.4 Tabular Array Optimization

TOON's tabular arrays provide the biggest token savings for uniform data:

**JSON** (~120 tokens):
```json
{"containers":[{"name":"API","technology":"Go"},{"name":"DB","technology":"PostgreSQL"},{"name":"Queue","technology":"RabbitMQ"}]}
```

**TOON** (~50 tokens):
```
containers[3]{name,technology}:
  API,"Go"
  DB,"PostgreSQL"
  Queue,"RabbitMQ"
```

Entities that benefit most from tabular arrays:
- `[]System` within a Project
- `[]Container` within a System
- `[]Component` within a Container

---

## 5. Encoding Adapter Changes

### 5.1 Current State

`internal/adapters/encoding/toon.go`:
- Custom format with reflection-based encoding
- Abbreviated keys (`n`, `d`, `t`, etc.)
- No decoder (returns error)
- Custom helper types: `ArchitectureSummary`, `ArchitectureStructure`, etc.

### 5.2 Target State

`internal/adapters/encoding/toon.go`:
- `toon.Marshal(value, toon.WithLengthMarkers(true))` for encoding
- `toon.Unmarshal(data, value)` for decoding
- Entity struct tags control field names
- Remove: `keyAbbreviations`, `encodeTOONValue`, `isSimpleString`, `abbreviateKey`, `isEmptyValue`
- Keep: `ArchitectureSummary` and related types (add `toon` struct tags)
- Keep: `FormatArchitectureTOON` and `FormatStructureTOON` (refactor to use toon.Marshal)

### 5.3 Backward Compatibility

- `EncodeJSON` / `DecodeJSON`: Unchanged
- `EncodeTOON`: Output format changes (custom → TOON v3.0)
- `DecodeTOON`: Now works (was returning error)
- Old custom format: Deprecated as `--format compact` with warning message
