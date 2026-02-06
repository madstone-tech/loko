---
name: "Event Processors"
description: "Lambda functions processing asynchronous events from SQS queues"
technology: "Go"
---

# Event Processors

Lambda functions processing asynchronous events from SQS queues for order fulfillment.

## Purpose

This function group handles all asynchronous order processing tasks triggered by SQS messages, including payment processing, inventory updates, and customer notifications.

## Trigger Type

- **Primary Trigger**: SQS (Order Processing Queue)
- **Invocation**: Asynchronous (event-driven)
- **Concurrency**: On-demand with reserved concurrency limits

## Functions List

| Function | Trigger | Description |
|----------|---------|-------------|
| process-payment | SQS | Processes payment via Stripe |
| update-inventory | SQS | Updates inventory counts |
| send-notification | SQS | Sends email notifications via SendGrid |

## IAM Permissions

Required permissions for event processor functions:

```yaml
- Effect: Allow
  Action:
    - dynamodb:GetItem
    - dynamodb:UpdateItem
  Resource:
    - !Sub "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/orders"

- Effect: Allow
  Action:
    - sqs:ReceiveMessage
    - sqs:DeleteMessage
    - sqs:GetQueueAttributes
  Resource: !Sub "arn:aws:sqs:${AWS::Region}:${AWS::AccountId}:order-processing-queue"

- Effect: Allow
  Action:
    - sqs:SendMessage
  Resource: !Sub "arn:aws:sqs:${AWS::Region}:${AWS::AccountId}:order-dlq"

- Effect: Allow
  Action:
    - secretsmanager:GetSecretValue
  Resource:
    - !Sub "arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:stripe-api-key*"
    - !Sub "arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:sendgrid-api-key*"
```

## Environment Variables

| Variable | Description | Source |
|----------|-------------|--------|
| `ORDERS_TABLE` | DynamoDB table name | CloudFormation |
| `DLQ_URL` | Dead Letter Queue URL | CloudFormation |
| `STRIPE_SECRET_NAME` | Secrets Manager secret name | Static |
| `SENDGRID_SECRET_NAME` | Secrets Manager secret name | Static |
| `LOG_LEVEL` | Logging verbosity | Static |

## Technology Stack

- **Runtime**: Go 1.x (provided.al2)
- **Memory**: 512 MB (payment processing needs more memory)
- **Timeout**: 60 seconds
- **Architecture**: arm64

## Error Handling

- Failed messages sent to Dead Letter Queue after 3 retries
- Partial batch failures reported for SQS
- Idempotency keys prevent duplicate payment processing
- Structured error logging with correlation IDs
