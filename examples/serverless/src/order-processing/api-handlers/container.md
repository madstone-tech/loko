---
name: "API Handlers"
description: "Lambda functions handling synchronous API Gateway requests"
technology: "Go"
---

# API Handlers

Lambda functions handling synchronous API Gateway requests for order management.

## Purpose

This function group handles all synchronous HTTP requests from API Gateway, including order creation, retrieval, listing, and cancellation.

## Trigger Type

- **Primary Trigger**: API Gateway (REST API)
- **Invocation**: Synchronous
- **Concurrency**: On-demand with reserved concurrency for critical operations

## Functions List

| Function | Trigger | Route | Description |
|----------|---------|-------|-------------|
| create-order | API Gateway | POST /orders | Creates a new order |
| get-order | API Gateway | GET /orders/{id} | Retrieves order by ID |
| list-orders | API Gateway | GET /orders | Lists orders with pagination |
| cancel-order | API Gateway | DELETE /orders/{id} | Cancels an existing order |

## IAM Permissions

Required permissions for API handler functions:

```yaml
- Effect: Allow
  Action:
    - dynamodb:GetItem
    - dynamodb:PutItem
    - dynamodb:Query
    - dynamodb:UpdateItem
  Resource:
    - !Sub "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/orders"
    - !Sub "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/orders/index/*"

- Effect: Allow
  Action:
    - sqs:SendMessage
  Resource: !Sub "arn:aws:sqs:${AWS::Region}:${AWS::AccountId}:order-processing-queue"
```

## Environment Variables

| Variable | Description | Source |
|----------|-------------|--------|
| `ORDERS_TABLE` | DynamoDB table name | CloudFormation |
| `PROCESSING_QUEUE_URL` | SQS queue URL for order processing | CloudFormation |
| `LOG_LEVEL` | Logging verbosity (DEBUG, INFO, WARN) | Static |
| `STAGE` | Deployment stage (dev, staging, prod) | CloudFormation |

## Technology Stack

- **Runtime**: Go 1.x (provided.al2)
- **Memory**: 256 MB
- **Timeout**: 30 seconds
- **Architecture**: arm64

## Error Handling

- API Gateway error responses with proper HTTP status codes
- DynamoDB conditional write failures return 409 Conflict
- Input validation errors return 400 Bad Request
- Structured error logging with correlation IDs
