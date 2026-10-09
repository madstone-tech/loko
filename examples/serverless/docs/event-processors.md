# Event Processors

Lambda functions processing asynchronous order messages from SQS: payment, inventory updates and
customer notifications.

## Trigger

- **Primary trigger**: SQS (order processing queue)
- **Invocation**: asynchronous (event-driven)
- **Concurrency**: on demand, with reserved concurrency limits

## Functions

| Function | Description |
|----------|-------------|
| process-payment | Charges the order through the payment provider |
| update-inventory | Reserves stock in the inventory service |
| send-notification | Emails the customer through the email provider |

## IAM permissions

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
    - secretsmanager:GetSecretValue
  Resource:
    - !Sub "arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:payment-provider-key*"
    - !Sub "arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:email-provider-key*"
```

## Environment variables

| Variable | Description | Source |
|----------|-------------|--------|
| `ORDERS_TABLE` | DynamoDB table name | CloudFormation |
| `DLQ_URL` | Dead-letter queue URL | CloudFormation |
| `PAYMENT_SECRET_NAME` | Secrets Manager secret name | Static |
| `EMAIL_SECRET_NAME` | Secrets Manager secret name | Static |
| `LOG_LEVEL` | Logging verbosity | Static |

## Runtime

- **Runtime**: Go (`provided.al2`)
- **Memory**: 512 MB (payment processing needs more memory)
- **Timeout**: 60 seconds
- **Architecture**: arm64

## Error handling

- Failed messages go to the dead-letter queue after 3 retries
- Partial batch failures are reported back to SQS
- Idempotency keys prevent duplicate payment processing
- Structured error logging with correlation IDs
