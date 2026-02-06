---
name: "Process Payment"
description: "Lambda function that processes payments via Stripe"
technology: "Go"
---

# Process Payment

Lambda function that processes payments via Stripe for new orders.

## Handler

```go
func Handler(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
    var batchItemFailures []events.SQSBatchItemFailure

    for _, record := range event.Records {
        // Parse order from message
        // Retrieve Stripe API key from Secrets Manager
        // Create Stripe PaymentIntent
        // Update order status in DynamoDB
        // Handle failures
    }

    return events.SQSEventResponse{BatchItemFailures: batchItemFailures}, nil
}
```

- **Entry Point**: `bootstrap` (Go provided.al2 runtime)
- **Function Name**: `order-processing-process-payment`

## Trigger

| Property | Value |
|----------|-------|
| **Type** | SQS |
| **Queue** | order-processing-queue |
| **Batch Size** | 10 |
| **Invocation** | Asynchronous |
| **Partial Batch Response** | Enabled |

## Runtime Configuration

| Property | Value |
|----------|-------|
| **Runtime** | provided.al2 (Go) |
| **Architecture** | arm64 |
| **Memory** | 512 MB |
| **Timeout** | 60 seconds |
| **Reserved Concurrency** | 10 |
| **Provisioned Concurrency** | None |

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `ORDERS_TABLE` | DynamoDB orders table name | Yes |
| `STRIPE_SECRET_NAME` | Secrets Manager secret name | Yes |
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
      - dynamodb:GetItem
      - dynamodb:UpdateItem
    Resource: !Sub "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/orders"
  - Effect: Allow
    Action:
      - sqs:ReceiveMessage
      - sqs:DeleteMessage
      - sqs:GetQueueAttributes
    Resource: !Sub "arn:aws:sqs:${AWS::Region}:${AWS::AccountId}:order-processing-queue"
  - Effect: Allow
    Action:
      - secretsmanager:GetSecretValue
    Resource: !Sub "arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:stripe-api-key*"
```

## Input/Output

### Input Event Schema (SQS Message Body)

```json
{
  "type": "object",
  "properties": {
    "order_id": { "type": "string" },
    "customer_id": { "type": "string" },
    "amount": { "type": "number" },
    "currency": { "type": "string" },
    "payment_method_id": { "type": "string" }
  },
  "required": ["order_id", "amount", "payment_method_id"]
}
```

### Output (SQS Batch Response)

```json
{
  "batchItemFailures": [
    { "itemIdentifier": "message-id-that-failed" }
  ]
}
```

## Error Handling

- **Retries**: 3 automatic retries via SQS visibility timeout
- **DLQ**: Failed messages sent to `order-processing-dlq` after max retries
- **Idempotency**: Uses order_id as idempotency key for Stripe

## Testing

- Unit tests: `go test ./handlers/process-payment/...`
- Integration tests: Stripe test mode with mock cards
- E2E tests: LocalStack with SQS and DynamoDB

## Observability

- **Logs**: CloudWatch Logs `/aws/lambda/order-processing-process-payment`
- **Traces**: AWS X-Ray with Stripe API subsegments
- **Metrics**: PaymentsProcessed, PaymentsFailed, PaymentAmount (custom)
