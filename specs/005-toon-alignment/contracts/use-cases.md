# Use Case Contracts: TOON Alignment & Handler Refactoring

**Feature**: 005-toon-alignment | **Date**: 2026-02-06 | **Phase**: 1 (Design)

---

## Existing Use Cases (Enhanced)

### BuildDocsUseCase

**File**: `internal/core/usecases/build_docs.go` (exists)

**Enhancement**: Move build orchestration logic from `cmd/build.go` into this use case.

```
Dependencies:
    projectRepo    ProjectRepository
    diagramRenderer DiagramRenderer
    siteBuilder    SiteBuilder
    logger         Logger

Input:
    ProjectRoot    string   // filesystem path to project
    Format         string   // "html" | "markdown" | "pdf"
    Parallel       bool     // enable parallel D2 rendering
    MaxWorkers     int      // worker count for parallel rendering

Output:
    FilesGenerated []string // list of generated file paths
    DiagramCount   int      // number of diagrams rendered
    Duration       time.Duration // total build time
    Errors         []error  // non-fatal errors (e.g., diagram warnings)

Preconditions:
    - ProjectRoot contains a valid loko.toml
    - At least one system exists

Postconditions:
    - Output directory contains generated documentation
    - All D2 diagrams rendered to SVG
    - Site index generated (if HTML format)

Error Cases:
    - ErrProjectNotFound: no loko.toml at ProjectRoot
    - ErrD2NotAvailable: d2 binary not installed
    - ErrInvalidFormat: unknown format string
```

### ValidateArchitectureUseCase

**File**: `internal/core/usecases/validate_architecture.go` (exists)

**Enhancement**: Ensure validation logic from `cmd/validate.go` is in this use case, not the handler.

```
Dependencies:
    projectRepo ProjectRepository
    validator   Validator
    logger      Logger

Input:
    ProjectRoot string  // filesystem path to project

Output:
    Errors  []ValidationError  // all validation issues found
    IsValid bool               // true if no errors
    Stats   ValidationStats    // counts by severity

Preconditions:
    - ProjectRoot contains a valid loko.toml

Postconditions:
    - All systems, containers, components checked
    - Orphaned references identified
    - Missing required files identified
    - C4 hierarchy violations identified

Error Cases:
    - ErrProjectNotFound: no loko.toml at ProjectRoot
```

---

## New Use Cases

### CreateContainerUseCase

**File**: `internal/core/usecases/create_container.go` (new)

```
Dependencies:
    projectRepo ProjectRepository
    logger      Logger

Input:
    ProjectRoot   string   // filesystem path to project
    SystemName    string   // parent system name/ID
    Name          string   // container display name
    Description   string   // optional description
    Technology    string   // optional technology string
    Tags          []string // optional tags

Output:
    Container *entities.Container  // created entity
    Error     error                // nil on success

Preconditions:
    - Project exists at ProjectRoot
    - System with SystemName exists
    - No container with same normalized name exists in system

Postconditions:
    - Container entity created and validated
    - Container saved to project via ProjectRepository

Error Cases:
    - ErrProjectNotFound: project doesn't exist
    - ErrSystemNotFound: system doesn't exist
    - ErrDuplicateEntity: container already exists
    - ErrValidation: invalid name or fields

Notes:
    - Does NOT generate D2 diagrams (that's ScaffoldEntityUseCase)
    - Does NOT render templates (that's ScaffoldEntityUseCase)
    - Pure entity creation + persistence
```

### CreateComponentUseCase

**File**: `internal/core/usecases/create_component.go` (new)

```
Dependencies:
    projectRepo ProjectRepository
    logger      Logger

Input:
    ProjectRoot   string   // filesystem path to project
    SystemName    string   // grandparent system name/ID
    ContainerName string   // parent container name/ID
    Name          string   // component display name
    Description   string   // optional description
    Technology    string   // optional technology string
    Tags          []string // optional tags

Output:
    Component *entities.Component  // created entity
    Error     error                // nil on success

Preconditions:
    - Project exists at ProjectRoot
    - System with SystemName exists
    - Container with ContainerName exists in system
    - No component with same normalized name exists in container

Postconditions:
    - Component entity created and validated
    - Component saved to project via ProjectRepository

Error Cases:
    - ErrProjectNotFound: project doesn't exist
    - ErrSystemNotFound: system doesn't exist
    - ErrContainerNotFound: container doesn't exist
    - ErrDuplicateEntity: component already exists
    - ErrValidation: invalid name or fields
```

### ScaffoldEntityUseCase

**File**: `internal/core/usecases/scaffold_entity.go` (new)

**Purpose**: Orchestrates the full entity creation workflow shared by CLI and MCP.

```
Dependencies:
    projectRepo      ProjectRepository
    templateEngine   TemplateEngine
    diagramGenerator DiagramGenerator
    logger           Logger

Input:
    ProjectRoot string    // filesystem path to project
    EntityType  string    // "system" | "container" | "component"
    ParentPath  []string  // hierarchy path: [] for system, [system] for container, etc.
    Name        string    // entity display name
    Description string    // optional description
    Technology  string    // optional technology string
    Tags        []string  // optional tags
    Template    string    // template name (empty = use project default)

Output:
    EntityID     string   // normalized ID of created entity
    FilesCreated []string // all files created/modified
    DiagramPath  string   // path to generated D2 diagram (empty if no diagram)
    Error        error    // nil on success

Preconditions:
    - Project exists at ProjectRoot
    - Parent entities exist (for container/component)
    - No duplicate entity at same level

Postconditions:
    - Entity created and persisted
    - Template files rendered and written
    - D2 diagram generated for entity
    - Parent diagram updated to include new entity

Orchestration Flow:
    1. Resolve parent entities from ParentPath
    2. Create entity (delegates to CreateSystem/Container/Component)
    3. Generate D2 diagram (delegates to DiagramGenerator)
    4. Write D2 file to project
    5. Update parent diagram to include new entity
    6. Render template files (delegates to TemplateEngine)
    7. Return created file paths

Error Cases:
    - ErrProjectNotFound: project doesn't exist
    - ErrParentNotFound: parent entity in path doesn't exist
    - ErrDuplicateEntity: entity already exists
    - ErrTemplateNotFound: requested template doesn't exist
    - ErrValidation: invalid entity name or fields
```

### UpdateDiagramUseCase

**File**: `internal/core/usecases/update_diagram.go` (new)

```
Dependencies:
    projectRepo ProjectRepository
    logger      Logger

Input:
    ProjectRoot string // filesystem path to project
    DiagramPath string // relative path to .d2 file within project
    D2Source    string // D2 source code to write

Output:
    FilePath string // absolute path of written file
    Error    error  // nil on success

Preconditions:
    - Project exists at ProjectRoot
    - DiagramPath is within project source directory
    - D2Source is non-empty

Postconditions:
    - D2 file written at DiagramPath
    - File contains provided D2Source content

Error Cases:
    - ErrProjectNotFound: project doesn't exist
    - ErrInvalidPath: path outside project or invalid extension
    - ErrEmptyContent: D2Source is empty
```

---

## Port Interface Contracts

### DiagramGenerator

**File**: `internal/core/usecases/ports.go` (add to existing)

```
DiagramGenerator interface {
    GenerateSystemContextDiagram(system *entities.System) (string, error)
    GenerateContainerDiagram(system *entities.System) (string, error)
    GenerateComponentDiagram(container *entities.Container) (string, error)
}

Contracts:
    GenerateSystemContextDiagram:
        - Input: System with optional ExternalSystems, KeyUsers
        - Output: Valid D2 source code showing system in context
        - Must include: system box, external system boxes, user actors, relationships
        - Error: only if system is nil or invalid

    GenerateContainerDiagram:
        - Input: System with Containers populated
        - Output: Valid D2 source code showing containers within system
        - Must include: system boundary, container boxes with technology labels
        - Error: only if system is nil or has no containers

    GenerateComponentDiagram:
        - Input: Container with Components populated
        - Output: Valid D2 source code showing components within container
        - Must include: container boundary, component boxes, relationship arrows
        - Error: only if container is nil or has no components
```

### UserPrompter

**File**: `internal/core/usecases/ports.go` (add to existing)

```
UserPrompter interface {
    PromptString(prompt string, defaultValue string) (string, error)
    PromptStringMulti(prompt string) ([]string, error)
}

Contracts:
    PromptString:
        - Displays prompt to user, returns their input
        - If user provides empty input, returns defaultValue
        - Error: if stdin is not a terminal (non-interactive mode)

    PromptStringMulti:
        - Displays prompt, collects multiple lines until empty line
        - Returns slice of non-empty strings
        - Error: if stdin is not a terminal
```

### ReportFormatter

**File**: `internal/core/usecases/ports.go` (add to existing)

```
ReportFormatter interface {
    PrintValidationReport(errors []ValidationError)
    PrintBuildReport(stats BuildStats)
}

BuildStats struct {
    FilesGenerated int
    DiagramCount   int
    Duration       time.Duration
    Format         string
}

Contracts:
    PrintValidationReport:
        - Formats validation errors for human display
        - Groups by severity/type
        - Uses color (lipgloss) for CLI, plain text for non-TTY

    PrintBuildReport:
        - Formats build statistics for human display
        - Shows file count, diagram count, duration
        - Uses color for CLI, plain text for non-TTY
```

---

## MCP Tool Schema Updates

After refactoring, each MCP tool becomes a thin handler calling a shared use case.

### create_system

```json
{
  "name": "create_system",
  "description": "Scaffold a new C4 system with D2 diagram and documentation",
  "inputSchema": {
    "type": "object",
    "required": ["name"],
    "properties": {
      "name": { "type": "string", "description": "System display name" },
      "description": { "type": "string" },
      "template": { "type": "string", "default": "standard-3layer" },
      "tags": { "type": "array", "items": { "type": "string" } }
    }
  }
}
```

**Handler** (~20 lines): Parse input → call `ScaffoldEntityUseCase.Execute(EntityType: "system")` → format response

### create_container

```json
{
  "name": "create_container",
  "description": "Scaffold a new C4 container within a system",
  "inputSchema": {
    "type": "object",
    "required": ["system", "name"],
    "properties": {
      "system": { "type": "string", "description": "Parent system name" },
      "name": { "type": "string", "description": "Container display name" },
      "description": { "type": "string" },
      "technology": { "type": "string" },
      "template": { "type": "string" },
      "tags": { "type": "array", "items": { "type": "string" } }
    }
  }
}
```

**Handler** (~20 lines): Parse input → call `ScaffoldEntityUseCase.Execute(EntityType: "container", ParentPath: [system])` → format response

### create_component

```json
{
  "name": "create_component",
  "description": "Scaffold a new C4 component within a container",
  "inputSchema": {
    "type": "object",
    "required": ["system", "container", "name"],
    "properties": {
      "system": { "type": "string" },
      "container": { "type": "string" },
      "name": { "type": "string" },
      "description": { "type": "string" },
      "technology": { "type": "string" },
      "tags": { "type": "array", "items": { "type": "string" } }
    }
  }
}
```

**Handler** (~20 lines): Parse input → call `ScaffoldEntityUseCase.Execute(EntityType: "component", ParentPath: [system, container])` → format response

### update_diagram

```json
{
  "name": "update_diagram",
  "description": "Write D2 diagram source code to a file",
  "inputSchema": {
    "type": "object",
    "required": ["path", "source"],
    "properties": {
      "path": { "type": "string", "description": "Relative path to .d2 file" },
      "source": { "type": "string", "description": "D2 source code" }
    }
  }
}
```

**Handler** (~15 lines): Parse input → call `UpdateDiagramUseCase.Execute()` → format response

### build_docs

```json
{
  "name": "build_docs",
  "description": "Build documentation in specified format",
  "inputSchema": {
    "type": "object",
    "properties": {
      "format": { "type": "string", "enum": ["html", "markdown", "pdf"], "default": "html" }
    }
  }
}
```

**Handler** (~15 lines): Parse input → call `BuildDocsUseCase.Execute()` → format response

### validate

```json
{
  "name": "validate",
  "description": "Validate architecture for consistency and completeness",
  "inputSchema": {
    "type": "object",
    "properties": {}
  }
}
```

**Handler** (~15 lines): Call `ValidateArchitectureUseCase.Execute()` → format response
