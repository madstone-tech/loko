---
name: "Create Order"
description: "Lambda function that creates new orders in the system"
technology: "Go"
---

# Create Order

Lambda function that creates new orders in the system.

## Handler

```go
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // Parse request body
    // Validate order data
    // Generate order ID
    // Store in DynamoDB
    // Publish to processing queue
    // Return order details
}
```

- **Entry Point**: `bootstrap` (Go provided.al2 runtime)
- **Function Name**: `order-processing-create-order`

## Trigger

| Property | Value |
|----------|-------|
| **Type** | API Gateway |
| **Method** | POST |
| **Path** | /orders |
| **Authorization** | Cognito User Pool |
| **Invocation** | Synchronous |

## Runtime Configuration

| Property | Value |
|----------|-------|
| **Runtime** | provided.al2 (Go) |
| **Architecture** | arm64 |
| **Memory** | 256 MB |
| **Timeout** | 30 seconds |
| **Reserved Concurrency** | None |
| **Provisioned Concurrency** | None |

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ORDERS_TABLE` | DynamoDB orders table name | Yes |
| `PROCESSING_QUEUE_URL` | SQS queue for order processing | Yes |
| `LOG_LEVEL` | Logging level | No |

## IAM Role

Minimum required permissions:

```yaml
Statement:
  - Effect: Allow
    Action:
      - logs:CreateLogGroup
      - logs:CreateLogStream
      - logs:PutLogEvents
    Resource: "*"
  - Effect: Allow
    Action:
      - dynamodb:PutItem
      - dynamodb:ConditionCheckItem
    Resource: !Sub "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/orders"
  - Effect: Allow
    Action:
      - sqs:SendMessage
    Resource: !Sub "arn:aws:sqs:${AWS::Region}:${AWS::AccountId}:order-processing-queue"
```

## Input/Output

### Input Event Schema (Request Body)

```json
{
  "type": "object",
  "properties": {
    "customer_id": { "type": "string" },
    "items": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "product_id": { "type": "string" },
          "quantity": { "type": "integer" }
        }
      }
    },
    "shipping_address": { "type": "object" }
  },
  "required": ["customer_id", "items"]
}
```

### Output Response

```json
{
  "statusCode": 201,
  "body": {
    "id": "order-uuid",
    "status": "pending",
    "customer_id": "customer-uuid",
    "items": [...],
    "total": 99.99,
    "created_at": "2026-02-05T12:00:00Z"
  }
}
```

## Error Handling

- **400 Bad Request**: Invalid request body or missing required fields
- **409 Conflict**: Order with same idempotency key already exists
- **500 Internal Server Error**: DynamoDB or SQS failures

## Testing

- Unit tests: `go test ./handlers/create-order/...`
- Integration tests: LocalStack with DynamoDB and SQS
- Load tests: Artillery with 100 concurrent users

## Observability

- **Logs**: CloudWatch Logs `/aws/lambda/order-processing-create-order`
- **Traces**: AWS X-Ray enabled
- **Metrics**: Invocations, Duration, Errors, OrdersCreated (custom)
