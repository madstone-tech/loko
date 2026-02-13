# D2 Diagramming Guide for LLMs

This document provides the essential D2 syntax and patterns needed to create architecture diagrams with loko.

## What is D2?

D2 is a modern, declarative diagramming language created by Terrastruct. It's designed specifically for software architecture diagrams with clean syntax, automatic layout, and beautiful output.

## Basic Syntax

### Nodes (Shapes)

```d2
# Simple node
MyService

# Node with label
MyService: "My Service Name"

# Node with description
MyService: "My Service" {
  description: "Handles user authentication"
}

# Node with technology
API: "REST API" {
  technology: "Go 1.21"
}
```

### Connections (Edges)

```d2
# Basic connection
A -> B

# Connection with label
A -> B: "Uses"

# Bidirectional
A <-> B: "Syncs with"

# Multiple targets
A -> B
A -> C

# Chain
A -> B -> C
```

### Nested Containers

```d2
# System containing containers
OrderSystem: "Order System" {
  API: "API Service"
  Database: "PostgreSQL"

  API -> Database: "Queries"
}

# Deep nesting
System: {
  Container: {
    Component: "Component"
  }
}
```

## Styling

### Colors and Fills

```d2
# Fill color
MyNode: {
  style: {
    fill: "#E1F5FF"
  }
}

# Stroke color and width
MyNode: {
  style: {
    stroke: "#01579B"
    stroke-width: 2
  }
}

# Common C4 colors
Person: {
  style.fill: "#FFF3E0"     # Orange tint for users
}
System: {
  style.fill: "#E1F5FF"     # Blue tint for systems
}
ExternalSystem: {
  style.fill: "#F5F5F5"     # Gray for external
}
Database: {
  style.fill: "#E8F5E9"     # Green tint for data stores
}
```

### Connection Styles

```d2
# Dashed line (async/event-driven)
A -> B: "Publishes events" {
  style.stroke-dash: 5
}

# Thick line (primary flow)
A -> B: {
  style.stroke-width: 3
}

# Colored connection
A -> B: {
  style.stroke: "#FF5722"
}
```

### Icons

```d2
# Using icons
User: "User" {
  icon: "https://icons.terrastruct.com/essentials/087-user.svg"
}

# AWS icons
Lambda: "Lambda Function" {
  icon: "https://icons.terrastruct.com/aws/_Group%20Icons/Compute.svg"
}

# Database icon
DB: "Database" {
  shape: cylinder
}
```

### Shapes

```d2
# Available shapes
circle: { shape: circle }
rectangle: { shape: rectangle }
cylinder: { shape: cylinder }      # Databases
queue: { shape: queue }            # Message queues
hexagon: { shape: hexagon }        # External services
cloud: { shape: cloud }            # Cloud services
person: { shape: person }          # Users/actors
```

## Layout

### Direction

```d2
# Left to right (default for loko)
direction: right

# Top to bottom
direction: down

# Other options: left, up
```

### Positioning

D2 uses automatic layout (ELK engine in loko). For best results:

1. **Group related items**: Use nesting
2. **Order declarations**: Items declared first tend to appear first
3. **Use direction**: Set appropriate flow direction

## C4 Model Patterns

### System Context Diagram (Level 1)

```d2
direction: right

# Users
User: "End User" {
  shape: person
  style.fill: "#FFF3E0"
}

# Main system
MySystem: "My Application" {
  style.fill: "#E1F5FF"
  style.stroke: "#01579B"
}

# External systems
PaymentGateway: "Payment Gateway" {
  style.fill: "#F5F5F5"
}
EmailService: "Email Service" {
  style.fill: "#F5F5F5"
}

# Relationships
User -> MySystem: "Uses"
MySystem -> PaymentGateway: "Processes payments"
MySystem -> EmailService: "Sends notifications"
```

### Container Diagram (Level 2)

```d2
direction: right

# External user
User: "User" {
  shape: person
  style.fill: "#FFF3E0"
}

# System boundary
OrderSystem: "Order System" {
  style.fill: "#E1F5FF"

  WebApp: "Web Application" {
    technology: "React"
    style.fill: "#E3F2FD"
  }

  API: "API Service" {
    technology: "Go"
    style.fill: "#E3F2FD"
  }

  Database: "Order Database" {
    technology: "PostgreSQL"
    shape: cylinder
    style.fill: "#E8F5E9"
  }

  Cache: "Cache" {
    technology: "Redis"
    shape: cylinder
    style.fill: "#FFF8E1"
  }

  # Internal relationships
  WebApp -> API: "REST/JSON"
  API -> Database: "SQL"
  API -> Cache: "Read/Write"
}

# External relationships
User -> OrderSystem.WebApp: "HTTPS"
```

### Component Diagram (Level 3)

```d2
direction: right

API: "API Service" {
  style.fill: "#E3F2FD"

  OrderController: "Order Controller" {
    description: "HTTP handlers"
  }

  OrderService: "Order Service" {
    description: "Business logic"
  }

  OrderRepository: "Order Repository" {
    description: "Data access"
  }

  PaymentClient: "Payment Client" {
    description: "External integration"
  }

  # Internal flow
  OrderController -> OrderService: "Calls"
  OrderService -> OrderRepository: "Queries"
  OrderService -> PaymentClient: "Requests payment"
}

# External dependency
PaymentGateway: "Payment Gateway" {
  style.fill: "#F5F5F5"
}

API.PaymentClient -> PaymentGateway: "HTTPS"
```

### Event-Driven / Serverless Pattern

```d2
direction: right

# Event sources
APIGateway: "API Gateway" {
  style.fill: "#FFE0B2"
}

EventBridge: "EventBridge" {
  shape: queue
  style.fill: "#E1BEE7"
}

# Lambda functions
OrderFunction: "Order Lambda" {
  style.fill: "#B3E5FC"
}

NotifyFunction: "Notify Lambda" {
  style.fill: "#B3E5FC"
}

# Data stores
DynamoDB: "DynamoDB" {
  shape: cylinder
  style.fill: "#E8F5E9"
}

# Async flows (dashed lines)
APIGateway -> OrderFunction: "Invokes"
OrderFunction -> DynamoDB: "Writes"
OrderFunction -> EventBridge: "Publishes" {
  style.stroke-dash: 5
}
EventBridge -> NotifyFunction: "Triggers" {
  style.stroke-dash: 5
}
```

## Best Practices for LLMs

### Do

1. **Use descriptive labels**: `"Order Service"` not `"svc1"`
2. **Include technology**: Add technology info to containers
3. **Style consistently**: Use C4 color conventions
4. **Show flow direction**: Label relationships with verbs
5. **Group logically**: Use nesting for system boundaries

### Don't

1. **Over-complicate**: Keep diagrams focused on one level
2. **Mix levels**: Don't show components in a context diagram
3. **Forget styling**: Unstyled diagrams are hard to read
4. **Use abbreviations**: Write full, clear names
5. **Overcrowd**: If too many elements, split into multiple diagrams

### Validation

Before generating D2, validate with loko:

```
Use the validate_diagram MCP tool to check syntax before saving
```

Common errors:
- Missing quotes around labels with spaces
- Unclosed braces
- Invalid shape names
- Referencing undefined nodes

## D2 vs Other Tools

| Feature | D2 | Mermaid | PlantUML |
|---------|----|---------| ---------|
| Syntax | Clean, minimal | Verbose | Very verbose |
| Layout | ELK (excellent) | Basic | Good |
| Styling | Flexible | Limited | Complex |
| Nesting | Native | Limited | Supported |
| C4 Support | Manual styling | Plugin | Native C4 |

## References

- [D2 Language Documentation](https://d2lang.com/tour/intro)
- [D2 Style Reference](https://d2lang.com/tour/style)
- [D2 Icons](https://icons.terrastruct.com/)
- [ELK Layout Algorithm](https://www.eclipse.org/elk/)
