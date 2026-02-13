# Contract: TemplateRegistry Interface

**Feature**: 007-ux-improvements  
**Interface**: `TemplateRegistry`  
**Location**: `internal/core/usecases/ports.go`  
**Adapter**: `internal/adapters/ason/template_registry.go`

---

## Interface Definition

```go
// TemplateRegistry resolves template types to actual template file paths.
type TemplateRegistry interface {
    // GetTemplatePath returns the absolute path to the template file for the given type.
    // Returns error if template file does not exist.
    GetTemplatePath(templateType entities.TemplateType) (string, error)
    
    // ValidateTemplate checks if the template file exists and is readable.
    ValidateTemplate(templateType entities.TemplateType) error
}
```

---

## Method: GetTemplatePath

### Signature

```go
GetTemplatePath(templateType entities.TemplateType) (string, error)
```

### Parameters

| Parameter | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| `templateType` | `entities.TemplateType` | Template category enum | Required, valid enum value |

### Returns

| Type | Description |
|------|-------------|
| `string` | Absolute path to template file |
| `error` | Error if template file does not exist or is unreadable |

### Behavior

**Success Cases**:
1. Template file exists: Returns absolute path (e.g., `/path/to/templates/component/compute.md`)

**Error Cases**:
1. Template file missing: Returns error with template type and expected path
2. Invalid template type: Returns error (defensive check, shouldn't happen with enum)

### Template File Naming Convention

| TemplateType | File Name | Example Path |
|--------------|-----------|--------------|
| `TemplateCompute` | `compute.md` | `templates/component/compute.md` |
| `TemplateDatastore` | `datastore.md` | `templates/component/datastore.md` |
| `TemplateMessaging` | `messaging.md` | `templates/component/messaging.md` |
| `TemplateAPI` | `api.md` | `templates/component/api.md` |
| `TemplateEvent` | `event.md` | `templates/component/event.md` |
| `TemplateStorage` | `storage.md` | `templates/component/storage.md` |
| `TemplateGeneric` | `generic.md` | `templates/component/generic.md` |

---

## Method: ValidateTemplate

### Signature

```go
ValidateTemplate(templateType entities.TemplateType) error
```

### Parameters

| Parameter | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| `templateType` | `entities.TemplateType` | Template category enum | Required, valid enum value |

### Returns

| Type | Description |
|------|-------------|
| `error` | Error if template file missing/unreadable, nil if valid |

### Behavior

**Success Cases**:
1. Template file exists and is readable: Returns nil

**Error Cases**:
1. Template file missing: Returns error
2. Template file exists but not readable: Returns error with permission details

---

## Examples

### Example 1: Get Compute Template Path

**Input**:
```go
path, err := registry.GetTemplatePath(entities.TemplateCompute)
```

**Expected Output**:
```go
path == "/Users/user/loko/templates/component/compute.md"
err == nil
```

### Example 2: Get Template Path for Non-Existent File

**Input**:
```go
// Assume storage.md doesn't exist yet
path, err := registry.GetTemplatePath(entities.TemplateStorage)
```

**Expected Output**:
```go
path == ""
err != nil
err.Error() contains "template file not found: storage.md"
```

### Example 3: Validate Existing Template

**Input**:
```go
err := registry.ValidateTemplate(entities.TemplateDatastore)
```

**Expected Output**:
```go
err == nil
```

### Example 4: Validate Missing Template

**Input**:
```go
// Assume event.md is missing
err := registry.ValidateTemplate(entities.TemplateEvent)
```

**Expected Output**:
```go
err != nil
err.Error() contains "template file not found: event.md"
```

---

## Test Cases

### Unit Tests (Adapter Layer)

1. **GetTemplatePath for all 7 types**: Verify correct file names returned
2. **GetTemplatePath for missing file**: Verify error with expected path
3. **ValidateTemplate for existing file**: Verify nil error
4. **ValidateTemplate for missing file**: Verify descriptive error
5. **ValidateTemplate for unreadable file**: Verify permission error

### Integration Tests (Use Case Layer)

1. **CreateComponent with compute template**: Verify template loaded and rendered
2. **CreateComponent with unknown template type**: Verify falls back to generic
3. **Startup validation**: Verify all 7 templates exist during app startup

---

## Performance Requirements

| Scenario | Requirement | Measurement |
|----------|------------|-------------|
| GetTemplatePath | <0.1ms | File system stat only |
| ValidateTemplate | <0.5ms | File existence + read check |
| Validate all 7 templates | <5ms | Startup validation |

**Enforcement**: Benchmark tests in `internal/adapters/ason/template_registry_bench_test.go`

---

## Implementation Notes

### Template Directory Structure

```
templates/
└── component/
    ├── compute.md       # Lambda, Functions, ECS Task
    ├── datastore.md     # DynamoDB, RDS, Aurora
    ├── messaging.md     # SQS, SNS, Kinesis
    ├── api.md           # API Gateway, REST, GraphQL
    ├── event.md         # EventBridge, Step Functions
    ├── storage.md       # S3, EFS
    └── generic.md       # Unknown technology fallback
```

### Adapter Construction

```go
// NewRegistry creates a TemplateRegistry with the given template directory.
func NewRegistry(templateDir string) (*Registry, error) {
    if templateDir == "" {
        return nil, errors.New("template directory cannot be empty")
    }
    
    registry := &Registry{templateDir: templateDir}
    
    // Validate all templates exist on construction
    for tmplType := TemplateCompute; tmplType <= TemplateGeneric; tmplType++ {
        if err := registry.ValidateTemplate(tmplType); err != nil {
            return nil, fmt.Errorf("missing template: %w", err)
        }
    }
    
    return registry, nil
}
```

---

## Mock Implementation (for Testing)

```go
// MockTemplateRegistry implements TemplateRegistry for testing.
type MockTemplateRegistry struct {
    GetTemplatePathFunc func(templateType entities.TemplateType) (string, error)
    ValidateTemplateFunc func(templateType entities.TemplateType) error
}

func (m *MockTemplateRegistry) GetTemplatePath(templateType entities.TemplateType) (string, error) {
    if m.GetTemplatePathFunc != nil {
        return m.GetTemplatePathFunc(templateType)
    }
    return "/mock/templates/" + templateType.String() + ".md", nil
}

func (m *MockTemplateRegistry) ValidateTemplate(templateType entities.TemplateType) error {
    if m.ValidateTemplateFunc != nil {
        return m.ValidateTemplateFunc(templateType)
    }
    return nil
}
```

**Usage in Tests**:
```go
mockRegistry := &MockTemplateRegistry{
    GetTemplatePathFunc: func(tmplType entities.TemplateType) (string, error) {
        if tmplType == entities.TemplateCompute {
            return "/test/compute.md", nil
        }
        return "", fmt.Errorf("template not found: %s", tmplType.String())
    },
}

useCase := NewCreateComponent(projectRepo, templateEngine, mockRegistry)
err := useCase.Execute(ctx, componentSpec)
// Verify correct template was used
```

---

## Validation Checklist

- [ ] Returns absolute paths (not relative)
- [ ] Returns error for missing template files with expected path
- [ ] Validates all 7 template types on construction
- [ ] Template file names match enum values (.String() method)
- [ ] Thread-safe (read-only operations after construction)
- [ ] ValidateTemplate distinguishes between missing and unreadable
- [ ] GetTemplatePath fails fast (doesn't attempt to read content)
- [ ] Performance: <0.1ms per GetTemplatePath call

---

**Status**: ✅ Contract complete - Ready for implementation
