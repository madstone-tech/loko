# loko Architecture Patterns Reference

This document provides common architecture patterns and how to document them using loko's C4 model approach with D2 diagrams.

## Overview

Architecture patterns are reusable solutions to common design problems. This guide shows how to represent these patterns in loko using the C4 hierarchy (System → Container → Component) and D2 diagramming.

## Pattern Categories

1. **Application Patterns**: Monolith, Microservices, Serverless
2. **Communication Patterns**: Sync, Async, Event-Driven
3. **Data Patterns**: CQRS, Event Sourcing, Saga
4. **Infrastructure Patterns**: API Gateway, Service Mesh, Sidecar

---

## Application Patterns

### Three-Tier Monolith

**When to use**: Traditional web applications, MVPs, small teams

**C4 Structure**:
```
System: WebApplication
├── Container: WebServer (Presentation + Business Logic)
│   ├── Component: Controllers
│   ├── Component: Services
│   └── Component: Repositories
└── Container: Database
```

**D2 Pattern**:
```d2
direction: down

web: Web Server {
  controllers: Controllers
  services: Services
  repos: Repositories

  controllers -> services
  services -> repos
}

db: Database {
  shape: cylinder
}

web.repos -> db: "SQL"
```

**loko Commands**:
```bash
loko new system WebApplication
loko new container WebServer --parent WebApplication
loko new container Database --parent WebApplication
loko new component Controllers --parent WebApplication/WebServer
loko new component Services --parent WebApplication/WebServer
loko new component Repositories --parent WebApplication/WebServer
```

---

### Microservices

**When to use**: Large teams, independent deployment, scaling needs

**C4 Structure**:
```
System: ECommercePlatform
├── Container: OrderService
├── Container: PaymentService
├── Container: InventoryService
├── Container: NotificationService
└── Container: MessageBroker
```

**D2 Pattern**:
```d2
direction: right

orders: Order Service
payments: Payment Service
inventory: Inventory Service
notifications: Notification Service

broker: Message Broker {
  shape: queue
}

orders -> broker: "OrderCreated"
broker -> payments: "ProcessPayment"
broker -> inventory: "ReserveStock"
broker -> notifications: "SendConfirmation"
```

**Key Decisions to Document**:
- Service boundaries (bounded contexts)
- Communication style (sync vs async)
- Data ownership per service
- Shared infrastructure

---

### Serverless / Event-Driven

**When to use**: Variable load, pay-per-use, rapid scaling

**C4 Structure**:
```
System: ImageProcessor
├── Container: UploadAPI (API Gateway + Lambda)
├── Container: ProcessorFunctions (Lambda Group)
├── Container: StorageBucket (S3)
└── Container: ResultsQueue (SQS)
```

**D2 Pattern**:
```d2
direction: right

api: API Gateway {
  icon: https://icons.terrastruct.com/aws%2FNetworking%20&%20Content%20Delivery%2FAmazon-API-Gateway.svg
}

upload: Upload Handler {
  icon: https://icons.terrastruct.com/aws%2FCompute%2FAWS-Lambda.svg
}

bucket: S3 Bucket {
  shape: cylinder
  icon: https://icons.terrastruct.com/aws%2FStorage%2FAmazon-Simple-Storage-Service-S3.svg
}

processor: Image Processor {
  icon: https://icons.terrastruct.com/aws%2FCompute%2FAWS-Lambda.svg
}

queue: Results Queue {
  shape: queue
  icon: https://icons.terrastruct.com/aws%2FApp%20Integration%2FAmazon-Simple-Queue-Service-SQS.svg
}

api -> upload: "POST /upload"
upload -> bucket: "Store"
bucket -> processor: "S3 Event" {style.stroke-dash: 5}
processor -> queue: "Results" {style.stroke-dash: 5}
```

**Serverless Documentation Tips**:
- Use dashed lines for async/event flows
- Document triggers for each function
- Include IAM permissions in container metadata
- Note cold start implications

---

## Communication Patterns

### Synchronous Request/Response

**D2 Pattern**:
```d2
client: Client
server: Server

client -> server: "HTTP Request"
server -> client: "HTTP Response" {style.stroke: "#666"}
```

**When to document**:
- API contracts (OpenAPI specs)
- Timeout configurations
- Retry policies
- Circuit breaker settings

---

### Asynchronous Messaging

**D2 Pattern**:
```d2
producer: Producer
queue: Message Queue {shape: queue}
consumer: Consumer

producer -> queue: "Publish" {style.stroke-dash: 5}
queue -> consumer: "Subscribe" {style.stroke-dash: 5}
```

**When to document**:
- Message schemas
- Dead letter queues
- Retry policies
- Ordering guarantees

---

### Event-Driven (Pub/Sub)

**D2 Pattern**:
```d2
publisher: Publisher
topic: Event Topic {shape: queue}
sub1: Subscriber A
sub2: Subscriber B
sub3: Subscriber C

publisher -> topic: "Publish" {style.stroke-dash: 5}
topic -> sub1: "Notify" {style.stroke-dash: 5}
topic -> sub2: "Notify" {style.stroke-dash: 5}
topic -> sub3: "Notify" {style.stroke-dash: 5}
```

**Event Documentation Template**:
```yaml
Event: OrderCreated
Publisher: OrderService
Subscribers:
  - PaymentService: Process payment
  - InventoryService: Reserve stock
  - NotificationService: Send confirmation
Schema: events/order-created.json
```

---

## Data Patterns

### CQRS (Command Query Responsibility Segregation)

**C4 Structure**:
```
System: OrderSystem
├── Container: CommandAPI
├── Container: QueryAPI
├── Container: WriteDatabase
├── Container: ReadDatabase
└── Container: Synchronizer
```

**D2 Pattern**:
```d2
direction: down

commands: Command API
queries: Query API

write_db: Write DB {shape: cylinder}
read_db: Read DB {shape: cylinder}

sync: Synchronizer

commands -> write_db: "Write"
write_db -> sync: "Changes" {style.stroke-dash: 5}
sync -> read_db: "Sync" {style.stroke-dash: 5}
queries -> read_db: "Read"
```

---

### Saga Pattern (Distributed Transactions)

**D2 Pattern**:
```d2
direction: right

orchestrator: Saga Orchestrator

order: Order Service
payment: Payment Service
inventory: Inventory Service

orchestrator -> order: "1. Create Order"
orchestrator -> payment: "2. Process Payment"
orchestrator -> inventory: "3. Reserve Stock"

# Compensation flows (rollback)
orchestrator <- inventory: "3b. Release Stock" {style.stroke: red; style.stroke-dash: 5}
orchestrator <- payment: "2b. Refund" {style.stroke: red; style.stroke-dash: 5}
orchestrator <- order: "1b. Cancel Order" {style.stroke: red; style.stroke-dash: 5}
```

**Documentation Requirements**:
- Happy path sequence
- Compensation (rollback) steps
- Timeout handling
- Idempotency guarantees

---

## Infrastructure Patterns

### API Gateway

**D2 Pattern**:
```d2
direction: right

clients: External Clients

gateway: API Gateway {
  auth: Authentication
  rate: Rate Limiting
  route: Routing
}

services: Backend Services {
  svc1: Service A
  svc2: Service B
  svc3: Service C
}

clients -> gateway
gateway.route -> services.svc1
gateway.route -> services.svc2
gateway.route -> services.svc3
```

**Gateway Documentation**:
- Authentication methods
- Rate limit configurations
- Route mappings
- Request/response transformations

---

### Sidecar Pattern

**D2 Pattern**:
```d2
pod: Kubernetes Pod {
  app: Application Container
  sidecar: Sidecar Proxy {
    style.stroke-dash: 3
  }

  app -> sidecar: "localhost"
}

mesh: Service Mesh Control Plane

pod.sidecar -> mesh: "Config/Telemetry" {style.stroke-dash: 5}
```

**Use Cases to Document**:
- Service mesh (Istio, Linkerd)
- Log aggregation
- Secret injection
- TLS termination

---

## Pattern Selection Guide

| Pattern | Team Size | Complexity | Scaling | Best For |
|---------|-----------|------------|---------|----------|
| Monolith | Small | Low | Vertical | MVPs, startups |
| Microservices | Large | High | Horizontal | Enterprise, multiple teams |
| Serverless | Any | Medium | Auto | Variable workloads, events |
| CQRS | Medium+ | High | Independent | Read-heavy, complex queries |

---

## Documentation Best Practices

### 1. Start at the Right Level

- **New project**: Start with Context (Level 1) and Container (Level 2)
- **Detailed design**: Add Component (Level 3) for complex containers
- **Code documentation**: Level 4 only for critical algorithms

### 2. Document Decisions, Not Just Structure

Include Architecture Decision Records (ADRs):
```markdown
## Decision: Use Event Sourcing for Order History

**Context**: Need complete audit trail of order changes
**Decision**: Implement event sourcing for OrderService
**Consequences**:
- (+) Complete history, easy replay
- (-) Increased storage, eventual consistency
```

### 3. Keep Diagrams Focused

- One diagram per concern
- 5-7 elements maximum per diagram
- Use consistent styling across diagrams
- Include legends for non-obvious notation

### 4. Version Your Architecture

```bash
# Tag architecture at release points
git tag -a "arch-v1.0" -m "Initial architecture"

# Reference in loko
loko build --version v1.0
```

---

## MCP Workflow for Patterns

### Creating a New Pattern-Based System

```
1. query_project                          # Check existing structure
2. create_system(name: "OrderPlatform")   # Create system
3. # For microservices pattern:
   create_container(name: "OrderService", parent: "OrderPlatform")
   create_container(name: "PaymentService", parent: "OrderPlatform")
   create_container(name: "MessageBroker", parent: "OrderPlatform")
4. update_diagram                          # Add async connections
5. validate                                # Check consistency
6. build_docs                              # Generate output
```

### Querying Pattern Information

```
1. query_architecture(detail: "structure")  # See hierarchy
2. query_dependencies(entity_id: "OrderService", direction: "both")
3. analyze_coupling(source: "OrderService") # Check coupling metrics
```

---

## Common Anti-Patterns to Avoid

### 1. Distributed Monolith
**Symptom**: Microservices that must be deployed together
**Detection**: High coupling score in `analyze_coupling`
**Fix**: Merge tightly coupled services or properly decouple

### 2. Over-Engineering
**Symptom**: CQRS/Event Sourcing for simple CRUD
**Detection**: Complex patterns with <100 users
**Fix**: Start simple, evolve when needed

### 3. Undocumented Async
**Symptom**: Solid lines for message queues
**Detection**: Visual review of D2 diagrams
**Fix**: Use dashed lines, document event schemas

### 4. Missing Boundaries
**Symptom**: Components calling across system boundaries
**Detection**: `query_dependencies` shows cross-system calls
**Fix**: Add proper API containers at boundaries

