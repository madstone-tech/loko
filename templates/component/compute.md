---
name: "{{ComponentName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ComponentName}}

{{Description}}

## Responsibility

This compute component is responsible for {{Responsibility}}.

## Runtime & Configuration

- **Technology**: {{Technology}}
- **Runtime**: {{Runtime}}
- **Memory**: {{MemoryMB}} MB
- **Timeout**: {{TimeoutSeconds}}s

## Trigger

- **Type**: {{TriggerType}}
- **Source**: {{TriggerSource}}

## Dependencies

- {{Dependency1}}
- {{Dependency2}}

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| {{EnvVar1}} | {{EnvVar1Description}} | Yes |
| {{EnvVar2}} | {{EnvVar2Description}} | No |

## Error Handling

- **Retry policy**: {{RetryPolicy}}
- **Dead-letter queue**: {{DLQName}}

## Testing

- Unit tests: {{UnitTestFramework}}
- Integration tests: {{IntegrationTestFramework}}

## Performance

- **Cold start**: {{ColdStartMs}}ms (estimated)
- **P99 latency**: {{P99LatencyMs}}ms
