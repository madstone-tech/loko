---
name: "{{ComponentName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ComponentName}}

{{Description}}

## Responsibility

This messaging component is responsible for {{Responsibility}}.

## Queue / Topic Configuration

- **Technology**: {{Technology}}
- **Type**: {{QueueType}}
- **Name**: {{QueueName}}

## Message Schema

```json
{
  "{{Field1}}": "{{Type1}}",
  "{{Field2}}": "{{Type2}}"
}
```

## Producers

- {{Producer1}} — {{Producer1Description}}

## Consumers

- {{Consumer1}} — {{Consumer1Description}}

## Delivery & Retry

- **Visibility timeout**: {{VisibilityTimeoutSeconds}}s
- **Max receive count**: {{MaxReceiveCount}}
- **Dead-letter queue**: {{DLQName}}

## Throughput

- **Estimated TPS**: {{EstimatedTPS}}
- **Batch size**: {{BatchSize}}

## Security

- **Encryption at rest**: {{EncryptionAtRest}}
- **Encryption in transit**: {{EncryptionInTransit}}
