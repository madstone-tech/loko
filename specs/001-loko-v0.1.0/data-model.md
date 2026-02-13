# Data Model: loko v0.1.0

**Created**: 2025-12-17  
**Status**: Ready for Implementation

---

## Entity Relationship Diagram

```
Project (root)
├── Systems (ordered collection)
│   ├── System
│   │   ├── Containers (ordered collection)
│   │   │   ├── Container
│   │   │   │   └── Components (ordered collection)
│   │   │   │       └── Component
│   │   │   └── Diagrams (D2 rendering)
│   │   └── Diagrams (D2 rendering)
│   └── Metadata (YAML frontmatter)
├── Configuration (loko.toml)
├── Templates (global + project)
└── BuildArtifacts (dist/ directory)
```

---

## Core Entities

### Project

**File**: `internal/core/entities/project.go`

```go
type Project struct {
    // Identification
    Name        string              // Project name (from loko.toml)
    Description string              // Brief description
    Version     string              // Documentation version
    
    // Content
    Systems     map[string]*System  // Systems keyed by name
    Metadata    map[string]any      // Custom metadata
    
    // Configuration
    Config      *ProjectConfig      // Parsed loko.toml
    Path        string              // Root filesystem path
    
    // Timestamps
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type ProjectConfig struct {
    // Paths
    SourceDir   string  // Default: "./src"
    OutputDir   string  // Default: "./dist"
    
    // D2 Configuration
    D2Theme     string  // Default: "neutral-default"
    D2Layout    string  // Default: "elk"
    D2Cache     bool    // Default: true
    
    // Output Configuration
    Outputs     map[string]OutputConfig
    
    // Performance
    ParallelRenders int  // Default: 4
}

type OutputConfig struct {
    Enabled bool
    Format  string  // "html", "markdown", "pdf"
}
```

**Validation**:
- Name: 1-64 alphanumeric + hyphen, no spaces
- Version: SemVer format (X.Y.Z)
- Paths: Absolute or relative (resolved to absolute)
- Config: Required, valid TOML

---

### System

**File**: `internal/core/entities/system.go`

```go
type System struct {
    // Identification
    ID          string              // Unique ID within project
    Name        string              // Display name
    Description string              // System purpose
    
    // Content
    Containers  map[string]*Container  // Containers in this system
    Diagram     *Diagram            // system.d2 diagram
    
    // Metadata
    Metadata    map[string]any
    Tags        []string            // For categorization
    
    // File References
    MarkdownPath string              // system.md file path
    DiagramPath string               // system.d2 file path
    
    // Timestamps
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Validation**:
- ID: Unique within project, alphanumeric + hyphen
- Name: Required, 1-100 chars
- At least one container (or can be empty for context level)

---

### Container

**File**: `internal/core/entities/container.go`

```go
type Container struct {
    // Identification
    ID          string              // Unique ID within system
    Name        string              // Display name
    Description string              // Container purpose
    Type        string              // "Web", "API", "Database", "Message Queue", etc.
    
    // Content
    Components  map[string]*Component  // Components in this container
    Diagram     *Diagram            // container.d2 diagram
    
    // Relationships
    SystemID    string              // Parent system
    
    // Metadata
    Metadata    map[string]any
    Tags        []string
    
    // Technology
    Technology  string              // "Node.js", "PostgreSQL", etc.
    
    // File References
    MarkdownPath string              // container.md file path
    DiagramPath string               // container.d2 file path
    
    // Timestamps
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Validation**:
- ID: Unique within system
- Name: Required, 1-100 chars
- Type: One of known types or custom
- SystemID: Must reference valid system

---

### Component

**File**: `internal/core/entities/component.go`

```go
type Component struct {
    // Identification
    ID          string              // Unique ID within container
    Name        string              // Display name
    Description string              // Component purpose
    
    // Relationships
    ContainerID string              // Parent container
    
    // Metadata
    Metadata    map[string]any
    Tags        []string
    
    // Technology
    Technology  string              // Implementation technology
    
    // File References
    MarkdownPath string              // component.md file path (optional)
    
    // Timestamps
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Validation**:
- ID: Unique within container
- Name: Required, 1-100 chars
- ContainerID: Must reference valid container

---

### Diagram

**File**: `internal/core/entities/diagram.go`

```go
type Diagram struct {
    // Source
    Source      string              // D2 source code
    SourcePath  string              // File path to .d2 file
    
    // Rendered Outputs
    SVGPath     string              // Path to rendered SVG
    PNGPath     string              // Path to rendered PNG (optional)
    
    // Metadata
    Format      string              // "d2"
    Theme       string              // D2 theme used
    Layout      string              // D2 layout used (elk, dagre, etc.)
    
    // Caching
    ContentHash string              // SHA256(Source) for cache key
    RenderedAt  time.Time
    
    // Timestamps
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Validation**:
- Source: Valid D2 syntax
- ContentHash: Computed on save

---

### Template

**File**: `internal/core/entities/template.go`

```go
type Template struct {
    // Identification
    Name        string              // e.g., "standard-3layer"
    Version     string              // Template version
    Description string              // What this template creates
    
    // Definition
    Files       map[string]string   // Filename → Template content
    Variables   []TemplateVariable  // Required/optional variables
    
    // Metadata
    Type        string              // "system", "container", "component"
    Location    string              // "global" or "project"
    Path        string              // File system path
    
    // Timestamps
    CreatedAt   time.Time
}

type TemplateVariable struct {
    Name        string              // Variable name
    Description string              // What it's used for
    Default     string              // Default value (if any)
    Required    bool                // Whether user must provide
    Validation  string              // Regex pattern (if any)
}
```

**Validation**:
- Name: Alphanumeric + hyphen
- Files: Must contain at least one template
- Variables: Used in {{VarName}} syntax

---

## Relationships

### Project ↔ System (One-to-Many)

- Project contains multiple Systems
- System belongs to one Project
- File structure: `src/{SystemName}/`

### System ↔ Container (One-to-Many)

- System contains multiple Containers
- Container belongs to one System
- File structure: `src/{SystemName}/{ContainerName}/`

### Container ↔ Component (One-to-Many)

- Container contains multiple Components
- Component belongs to one Container
- File structure: Stored in container markdown/metadata

### System/Container/Component ↔ Diagram (One-to-One)

- Each can have zero or one diagram
- Diagrams are optional (can add later)

---

## File System Mapping

### Directory Structure

```
myproject/
├── loko.toml                          # Project config
├── src/                               # Source directory
│   ├── context.md                     # Optional context level
│   ├── context.d2
│   ├── PaymentService/                # System directory
│   │   ├── system.md                  # Required
│   │   ├── system.d2                  # Optional
│   │   ├── API/                       # Container directory
│   │   │   ├── container.md           # Required
│   │   │   ├── container.d2           # Optional
│   │   │   └── components.md          # Optional (list components)
│   │   ├── Database/
│   │   │   ├── container.md
│   │   │   └── container.d2
│   │   └── PubSub/
│   │       └── container.md
│   └── AuthService/                   # Another system
│       ├── system.md
│       └── ...
├── .loko/                             # Project templates (optional)
│   └── templates/
│       └── custom-system/
│           └── system.md.tmpl
└── dist/                              # Generated output
    ├── index.html
    ├── diagrams/
    │   ├── system-PaymentService.svg
    │   └── container-API.svg
    ├── styles/
    │   └── main.css
    ├── search.json
    └── README.md
```

---

## Markdown Format

### System Markdown (system.md)

```markdown
---
name: "Payment Service"
description: "Handles payment processing and transactions"
type: "external-system"  # optional
---

# Payment Service

Detailed description of the system.

## Responsibilities

- Process payments
- Validate transactions
- Report fraud

## Technology Stack

- Node.js
- PostgreSQL
```

### Container Markdown (container.md)

```markdown
---
name: "API Service"
description: "REST API for payment operations"
technology: "Node.js + Express"
---

# API Service

REST API for external integrations.

## Endpoints

- POST /api/payments - Create payment
- GET /api/payments/{id} - Get payment details
```

### Component Markdown (components.md)

```markdown
---
container: "API Service"
---

# Components

## PaymentProcessor

Handles payment validation and processing.

Technology: Node.js

## TransactionLogger

Logs all payment transactions for auditing.

Technology: Node.js
```

---

## YAML Frontmatter Specification

All markdown files support YAML frontmatter:

```yaml
---
name: "Entity Name"           # Required: Display name
description: "Purpose"        # Optional: Description
technology: "Tech stack"      # Optional: Technology used
type: "system|container|..."  # Optional: C4 level
tags:                         # Optional: Tags for categorization
  - tag1
  - tag2
metadata:                     # Optional: Custom key-value
  owner: "team-name"
  status: "active"
---
```

---

## Configuration Format (loko.toml)

```toml
[project]
name = "payment-service"
description = "Payment processing system"
version = "0.1.0"

[paths]
source = "./src"        # Relative to project root
output = "./dist"

[d2]
theme = "neutral-default"
layout = "elk"
cache = true

[outputs.html]
enabled = true
navigation = "sidebar"
search = true

[outputs.markdown]
enabled = false

[outputs.pdf]
enabled = false

[build]
parallel = true
max_workers = 4
```

---

## Validation Rules

### Name Validation

- **System/Container/Component names**: 1-100 characters, alphanumeric + spaces
- **IDs**: Alphanumeric + hyphen, no spaces, unique within parent

### Hierarchy Validation

- **C4 Compliance**: 
  - Context → System → Container → Component (levels must be followed)
  - No skipping levels
  - Each level must have valid children

- **Orphan Detection**:
  - Container references non-existent system → ERROR
  - Component references non-existent container → ERROR
  - File exists but not in metadata → WARNING

- **Uniqueness**:
  - System IDs unique within project
  - Container IDs unique within system
  - Component IDs unique within container

### File Validation

- **Required files**: system.md for systems, container.md for containers
- **Optional files**: .d2 diagrams, component.md
- **Invalid**: Duplicate files, corrupted frontmatter

---

## State Management

### Loading Order

1. Load loko.toml → ProjectConfig
2. Scan src/ directory → Systems list
3. For each system: Load system.md → System metadata
4. For each system: Scan subdirectories → Containers list
5. For each container: Load container.md → Container metadata
6. For each diagram (.d2): Load source → Diagram object
7. Validate complete hierarchy

### Saving Order (When Creating New Entity)

1. Create directory if needed
2. Write template-rendered files (system.md, system.d2)
3. Update parent metadata
4. Return created entity

---

## Example: Payment Service System

### File System

```
src/
├── PaymentService/
│   ├── system.md
│   ├── system.d2
│   ├── API/
│   │   ├── container.md
│   │   ├── container.d2
│   │   └── components.md
│   └── Database/
│       └── container.md
```

### Data Model Instance

```
Project{
  Name: "payment-service",
  Systems: {
    "PaymentService": System{
      ID: "PaymentService",
      Name: "Payment Service",
      Containers: {
        "API": Container{
          ID: "API",
          Name: "API Service",
          Type: "Web Service",
          Components: {
            "PaymentProcessor": Component{...},
            "TransactionLogger": Component{...}
          }
        },
        "Database": Container{...}
      }
    }
  }
}
```

---

## Extensibility

### Custom Metadata

All entities support arbitrary metadata:

```go
System.Metadata["owner"] = "payment-team"
System.Metadata["sla"] = "99.9%"
System.Metadata["cost-per-month"] = 5000
```

### Custom Tags

Used for categorization:

```markdown
---
tags:
  - "production"
  - "payment-critical"
  - "team-platform"
---
```

### Custom Fields in Frontmatter

Any YAML field is preserved:

```yaml
---
name: "API"
custom_field: "custom_value"
nested:
  key: "value"
---
```

---
