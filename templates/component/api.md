---
name: "{{ComponentName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ComponentName}}

{{Description}}

## Responsibility

This API component is responsible for {{Responsibility}}.

## Endpoint Configuration

- **Technology**: {{Technology}}
- **Stage**: {{Stage}}
- **Base URL**: {{BaseURL}}

## Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| {{Method1}} | {{Path1}} | {{Path1Description}} | {{Auth1}} |
| {{Method2}} | {{Path2}} | {{Path2Description}} | {{Auth2}} |

## Request / Response

### Request
```json
{
  "{{RequestField1}}": "{{RequestType1}}"
}
```

### Response
```json
{
  "{{ResponseField1}}": "{{ResponseType1}}"
}
```

## Authentication & Authorization

- **Mechanism**: {{AuthMechanism}}
- **Scopes**: {{Scopes}}

## Rate Limiting

- **Requests/second**: {{RPS}}
- **Burst limit**: {{BurstLimit}}

## Error Codes

| Code | Meaning |
|------|---------|
| 400 | Bad Request |
| 401 | Unauthorized |
| 500 | Internal Server Error |
