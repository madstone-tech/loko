# Process Payment

Lambda function that charges new orders through the payment provider.

## Handler

```go
func Handler(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
    var batchItemFailures []events.SQSBatchItemFailure

    for _, record := range event.Records {
        // Parse order from message
        // Retrieve the provider API key from Secrets Manager
        // Create a payment intent
        // Update order status in DynamoDB
        // Handle failures
    }

    return events.SQSEventResponse{BatchItemFailures: batchItemFailures}, nil
}
```

- **Entry point**: `bootstrap` (Go `provided.al2` runtime)
- **Function name**: `order-processing-process-payment`

## Trigger

| Property | Value |
|----------|-------|
| Type | SQS |
| Queue | order-processing-queue |
| Batch size | 10 |
| Invocation | Asynchronous |
| Partial batch response | Enabled |

## Runtime configuration

| Property | Value |
|----------|-------|
| Runtime | provided.al2 (Go) |
| Architecture | arm64 |
| Memory | 512 MB |
| Timeout | 60 seconds |
| Reserved concurrency | 10 |

## Input (SQS message body)

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

## Error handling

- **Retries**: 3 automatic retries via the SQS visibility timeout
- **DLQ**: failed messages go to the dead-letter queue after the last retry
- **Idempotency**: `order_id` is the idempotency key sent to the payment provider

## Observability

- **Logs**: CloudWatch Logs `/aws/lambda/order-processing-process-payment`
- **Traces**: AWS X-Ray with payment provider subsegments
- **Metrics**: PaymentsProcessed, PaymentsFailed, PaymentAmount (custom)
