# API Handlers

Lambda functions handling synchronous API Gateway requests for order management: creation,
retrieval, listing and cancellation.

## Trigger

- **Primary trigger**: API Gateway (REST API)
- **Invocation**: synchronous
- **Concurrency**: on demand, with reserved concurrency for critical operations

## Functions

| Function | Route | Description |
|----------|-------|-------------|
| create-order | `POST /orders` | Creates a new order |
| get-order | `GET /orders/{id}` | Retrieves an order by ID |
| list-orders | `GET /orders` | Lists orders with pagination |
| cancel-order | `DELETE /orders/{id}` | Cancels an existing order |

## IAM permissions

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

## Environment variables

| Variable | Description | Source |
|----------|-------------|--------|
| `ORDERS_TABLE` | DynamoDB table name | CloudFormation |
| `PROCESSING_QUEUE_URL` | SQS queue URL for order processing | CloudFormation |
| `LOG_LEVEL` | Logging verbosity (DEBUG, INFO, WARN) | Static |
| `STAGE` | Deployment stage (dev, staging, prod) | CloudFormation |

## Runtime

- **Runtime**: Go (`provided.al2`)
- **Memory**: 256 MB
- **Timeout**: 30 seconds
- **Architecture**: arm64

## Error handling

- API Gateway error responses with proper HTTP status codes
- DynamoDB conditional write failures return 409 Conflict
- Input validation errors return 400 Bad Request
- Structured error logging with correlation IDs
