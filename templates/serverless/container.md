---
name: "{{ContainerName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ContainerName}}

{{Description}}

## Purpose

This function group handles {{Purpose}} within the serverless architecture.

## Trigger Type

- **Primary Trigger**: API Gateway / SQS / SNS / EventBridge / S3
- **Invocation**: Synchronous / Asynchronous
- **Concurrency**: Reserved / Provisioned / On-demand

## Functions List

{{component_table}}

<!-- The following Lambda functions belong to this container:

| Function | Trigger | Description |
|----------|---------|-------------|
| function-1 | API Gateway | Handles HTTP requests |
| function-2 | SQS | Processes queue messages |
| function-3 | EventBridge | Scheduled execution -->


## IAM Permissions

Required permissions for this function group:

```yaml
- Effect: Allow
  Action:
    - dynamodb:GetItem
    - dynamodb:PutItem
    - dynamodb:Query
  Resource: !Sub "arn:aws:dynamodb:${AWS::Region}:${AWS::AccountId}:table/*"

- Effect: Allow
  Action:
    - sqs:SendMessage
    - sqs:ReceiveMessage
    - sqs:DeleteMessage
  Resource: !Sub "arn:aws:sqs:${AWS::Region}:${AWS::AccountId}:*"

- Effect: Allow
  Action:
    - s3:GetObject
    - s3:PutObject
  Resource: "arn:aws:s3:::bucket-name/*"
```

## Environment Variables

| Variable | Description | Source |
|----------|-------------|--------|
| `TABLE_NAME` | DynamoDB table name | CloudFormation |
| `QUEUE_URL` | SQS queue URL | CloudFormation |
| `LOG_LEVEL` | Logging verbosity | Static |
| `API_KEY` | External API key | Secrets Manager |

## Technology Stack

- **Runtime**: {{Technology}}
- **Memory**: 256-1024 MB
- **Timeout**: 30-900 seconds
- **Architecture**: x86_64 / arm64

## Error Handling

- Dead Letter Queue (DLQ) for failed invocations
- Retry policies configured per trigger type
- CloudWatch alarms for error rate thresholds

## Monitoring

- CloudWatch Logs: `/aws/lambda/{{ContainerName}}-*`
- X-Ray tracing enabled
- Custom metrics for business KPIs
