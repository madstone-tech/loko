---
name: "{{ComponentName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ComponentName}}

{{Description}}

## Handler

```
exports.handler = async (event, context) => {
  // Function implementation
}
```

- **Entry Point**: `src/handlers/{{ComponentName}}/index.handler`
- **Function Name**: `{{ComponentName}}`

## Trigger

| Property | Value |
|----------|-------|
| **Type** | API Gateway / SQS / SNS / EventBridge / S3 |
| **Source** | Resource ARN or endpoint |
| **Batch Size** | 1-10 (for queue triggers) |
| **Invocation** | Synchronous / Asynchronous |

## Runtime Configuration

| Property | Value |
|----------|-------|
| **Runtime** | {{Technology}} |
| **Architecture** | arm64 |
| **Memory** | 256 MB |
| **Timeout** | 30 seconds |
| **Reserved Concurrency** | None |
| **Provisioned Concurrency** | None |

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `TABLE_NAME` | DynamoDB table name | Yes |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR) | No |
| `REGION` | AWS region | Yes |

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
      - dynamodb:PutItem
    Resource: !Sub "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/${TableName}"
```

## Input/Output

### Input Event Schema

```json
{
  "type": "object",
  "properties": {
    "id": { "type": "string" },
    "data": { "type": "object" }
  },
  "required": ["id"]
}
```

### Output Response

```json
{
  "statusCode": 200,
  "body": {
    "message": "Success",
    "data": {}
  }
}
```

## Error Handling

- **Retries**: 2 automatic retries on failure
- **DLQ**: Failed events sent to Dead Letter Queue
- **Timeout**: Returns 504 if function times out

## Testing

- Unit tests: Jest / pytest / go test
- Integration tests: LocalStack / SAM Local
- Load tests: Artillery / k6

## Observability

- **Logs**: CloudWatch Logs `/aws/lambda/{{ComponentName}}`
- **Traces**: AWS X-Ray enabled
- **Metrics**: Invocations, Duration, Errors, Throttles
