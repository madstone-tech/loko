---
name: "{{ComponentName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ComponentName}}

{{Description}}

## Responsibility

This event component is responsible for {{Responsibility}}.

## Event Bus / Rule Configuration

- **Technology**: {{Technology}}
- **Bus name**: {{BusName}}
- **Rule name**: {{RuleName}}

## Event Pattern

```json
{
  "source": ["{{EventSource}}"],
  "detail-type": ["{{DetailType}}"]
}
```

## Targets

| Target | ARN | Description |
|--------|-----|-------------|
| {{Target1}} | {{Target1ARN}} | {{Target1Description}} |

## Schedule (if applicable)

- **Expression**: {{ScheduleExpression}}
- **Timezone**: {{Timezone}}

## Retry Policy

- **Max attempts**: {{MaxAttempts}}
- **Max event age**: {{MaxEventAgeSeconds}}s

## Dead Letter Config

- **DLQ ARN**: {{DLQARN}}
