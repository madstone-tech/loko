---
name: "{{ComponentName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ComponentName}}

{{Description}}

## Responsibility

This storage component is responsible for {{Responsibility}}.

## Bucket / Volume Configuration

- **Technology**: {{Technology}}
- **Name**: {{BucketName}}
- **Region**: {{Region}}

## Object Structure

```
{{BucketName}}/
  {{Prefix1}}/         # {{Prefix1Description}}
  {{Prefix2}}/         # {{Prefix2Description}}
```

## Access Patterns

- **Read**: {{ReadDescription}}
- **Write**: {{WriteDescription}}

## Lifecycle Rules

| Rule | Prefix | Transition | Expiry |
|------|--------|-----------|--------|
| {{Rule1}} | {{Rule1Prefix}} | {{Rule1Transition}} | {{Rule1Expiry}} |

## Security

- **Bucket policy**: {{BucketPolicy}}
- **Server-side encryption**: {{SSEType}}
- **Public access**: {{PublicAccess}}

## Versioning

- **Enabled**: {{VersioningEnabled}}
- **MFA delete**: {{MFADelete}}
