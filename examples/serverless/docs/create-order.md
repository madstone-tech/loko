# Create Order

Lambda function that creates new orders.

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

- **Entry point**: `bootstrap` (Go `provided.al2` runtime)
- **Function name**: `order-processing-create-order`

## Trigger

| Property | Value |
|----------|-------|
| Type | API Gateway |
| Method | POST |
| Path | /orders |
| Authorization | Cognito User Pool |
| Invocation | Synchronous |

## Runtime configuration

| Property | Value |
|----------|-------|
| Runtime | provided.al2 (Go) |
| Architecture | arm64 |
| Memory | 256 MB |
| Timeout | 30 seconds |

## Environment variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ORDERS_TABLE` | DynamoDB orders table name | Yes |
| `PROCESSING_QUEUE_URL` | SQS queue for order processing | Yes |
| `LOG_LEVEL` | Logging level | No |

## Input

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

## Output

```json
{
  "statusCode": 201,
  "body": {
    "id": "order-uuid",
    "status": "pending",
    "customer_id": "customer-uuid",
    "items": [],
    "total": 99.99,
    "created_at": "2026-02-05T12:00:00Z"
  }
}
```

## Error handling

- **400 Bad Request**: invalid request body or missing required fields
- **409 Conflict**: an order with the same idempotency key already exists
- **500 Internal Server Error**: DynamoDB or SQS failures

## Observability

- **Logs**: CloudWatch Logs `/aws/lambda/order-processing-create-order`
- **Traces**: AWS X-Ray enabled
- **Metrics**: Invocations, Duration, Errors, OrdersCreated (custom)
