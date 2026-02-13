# Quickstart: Serverless Template

## Create a Serverless Project

```bash
# Initialize a new project
loko init my-serverless-app
cd my-serverless-app

# Create a system using the serverless template
loko new system "Order Processing API" -template serverless

# Add function group containers
loko new container "API Handlers" -parent order-processing-api -template serverless
loko new container "Event Processors" -parent order-processing-api -template serverless
loko new container "Scheduled Tasks" -parent order-processing-api -template serverless

# Add individual Lambda functions as components
loko new component "Create Order" -parent api-handlers -template serverless
loko new component "Process Payment" -parent event-processors -template serverless
loko new component "Generate Report" -parent scheduled-tasks -template serverless

# Build documentation
loko build

# Validate
loko validate
```

## Expected Output Structure

```
my-serverless-app/
├── loko.toml
├── src/
│   └── order-processing-api/
│       ├── system.md            # Serverless system overview
│       ├── system.d2            # API Gateway + Lambda context diagram
│       ├── api-handlers/
│       │   ├── container.md     # API handler function group
│       │   ├── container.d2     # Event flow diagram
│       │   └── create-order/
│       │       ├── component.md # Lambda function details
│       │       └── component.d2 # Function diagram
│       ├── event-processors/
│       │   ├── container.md
│       │   ├── container.d2
│       │   └── process-payment/
│       │       ├── component.md
│       │       └── component.d2
│       └── scheduled-tasks/
│           ├── container.md
│           ├── container.d2
│           └── generate-report/
│               ├── component.md
│               └── component.d2
└── dist/                        # Generated documentation
```

## Key Differences from Standard Template

| Aspect | standard-3layer | serverless |
|--------|----------------|------------|
| System sections | Containers, Technology Stack, Dependencies | Event Sources, Functions, External Integrations |
| Container sections | Interfaces (REST/gRPC), Components (Handler/Service/Repo), Deployment (Port/Type) | Trigger Type, Functions List, IAM Permissions |
| Component sections | Public Methods, Key Classes, Data Structures, Testing, Performance | Handler, Trigger, Runtime, Memory, Timeout, Environment |
| D2 diagram style | Solid lines, server icons | Dashed lines for async, cloud service icons |
