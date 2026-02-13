# C4 Model Guide for LLMs

This document provides the essential C4 model knowledge needed to work with loko effectively.

## What is the C4 Model?

The C4 model is a hierarchical approach to software architecture documentation created by Simon Brown. It uses four levels of abstraction to describe a software system, from high-level context down to implementation details.

## The Four Levels

### Level 1: System Context

**Purpose**: Shows how the system fits into the world around it.

**Contains**:
- The software system being documented (center)
- Users/actors who interact with it
- External systems it integrates with

**Key Questions**:
- Who uses this system?
- What other systems does it interact with?
- What is the system's boundary?

**loko Entity**: Not explicitly modeled (project root represents context)

**Example D2**:
```d2
User -> MySystem: "Uses"
MySystem -> PaymentProvider: "Processes payments via"
MySystem -> EmailService: "Sends notifications via"
```

### Level 2: Container

**Purpose**: Shows the high-level technology choices and how responsibilities are distributed.

**Contains**:
- Applications (web apps, mobile apps, desktop apps)
- Data stores (databases, file systems)
- Services (APIs, microservices)
- Message queues, caches, etc.

**Key Questions**:
- What are the major deployable/runnable units?
- What technologies are used?
- How do containers communicate?

**loko Entity**: `System` (contains containers)

**loko Command**: `loko new system <name>`

**Example D2**:
```d2
WebApp: "Web Application" {
  technology: "React"
}
API: "API Service" {
  technology: "Go"
}
Database: "PostgreSQL" {
  technology: "PostgreSQL 15"
}

WebApp -> API: "REST/JSON"
API -> Database: "SQL"
```

### Level 3: Component

**Purpose**: Shows how a container is made up of components and their interactions.

**Contains**:
- Logical groupings of code (modules, packages, namespaces)
- Controllers, services, repositories
- Major classes or interfaces

**Key Questions**:
- What are the major structural building blocks?
- How are responsibilities divided?
- What are the key abstractions?

**loko Entity**: `Container` (contains components)

**loko Command**: `loko new container <name> --parent <system>`

**Example D2**:
```d2
API: "API Container" {
  AuthController: "Auth Controller"
  UserService: "User Service"
  UserRepository: "User Repository"

  AuthController -> UserService
  UserService -> UserRepository
}
```

### Level 4: Code

**Purpose**: Shows implementation details at the class/function level.

**Note**: loko focuses on Levels 1-3. Level 4 is typically handled by IDE tools and code documentation.

**loko Entity**: `Component`

**loko Command**: `loko new component <name> --parent <container>`

## Hierarchy Rules

```
Project (Context)
  └── System (Level 2)
       └── Container (Level 3)
            └── Component (Level 4)
```

**Constraints**:
- A System MUST belong to exactly one Project
- A Container MUST belong to exactly one System
- A Component MUST belong to exactly one Container
- Names must be unique within their parent scope

## File Structure Convention

```
my-project/
├── loko.toml              # Project configuration
├── src/
│   ├── context.md         # Optional: Project-level context
│   ├── context.d2         # Optional: Context diagram
│   └── SystemName/
│       ├── system.md      # System documentation
│       ├── system.d2      # System/container diagram
│       └── ContainerName/
│           ├── container.md   # Container documentation
│           ├── container.d2   # Component diagram
│           └── ComponentName/
│               └── component.md  # Component documentation
```

## Best Practices for LLMs

### When Creating Architecture

1. **Start at the right level**: Begin with Systems before diving into Containers
2. **Name meaningfully**: Use domain terminology, not technical jargon
3. **Describe purpose, not implementation**: Focus on WHAT, not HOW
4. **Keep descriptions concise**: 1-2 sentences per entity

### When Querying Architecture

1. **Use progressive detail**: Start with `summary`, drill down to `structure` or `full`
2. **Target specific entities**: Query one system at a time for large projects
3. **Consider token budget**: Use TOON format for large architectures

### Common Mistakes to Avoid

1. **Mixing levels**: Don't put databases directly in a System (they're Containers)
2. **Over-decomposition**: Not everything needs to be a Component
3. **Implementation leakage**: Avoid putting class names in System/Container descriptions
4. **Missing relationships**: Always document how elements communicate

## C4 Notation in D2

| C4 Concept | D2 Representation |
|------------|-------------------|
| System | Box with description and style |
| Container | Nested box with technology label |
| Component | Innermost box |
| Person/User | Box with user icon |
| External System | Box with different fill color |
| Relationship | Arrow with label |
| Async Communication | Dashed arrow |

## Example: E-Commerce Architecture

```
# Systems
OrderService: Handles order lifecycle
PaymentService: Processes payments
NotificationService: Sends emails and SMS

# OrderService Containers
OrderService/
  ├── API: REST API for order operations (Go)
  ├── Worker: Background job processor (Go)
  ├── Database: Order data store (PostgreSQL)
  └── Cache: Order lookup cache (Redis)

# API Container Components
API/
  ├── OrderController: HTTP handlers for orders
  ├── OrderService: Business logic
  ├── OrderRepository: Data access
  └── PaymentClient: Integration with PaymentService
```

## References

- [C4 Model Official Site](https://c4model.com/)
- [C4 Model FAQ](https://c4model.com/#faq)
- [Simon Brown's Blog](https://www.codingthearchitecture.com/)
