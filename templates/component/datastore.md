---
name: "{{ComponentName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ComponentName}}

{{Description}}

## Responsibility

This datastore component is responsible for {{Responsibility}}.

## Data Model

- **Technology**: {{Technology}}
- **Region**: {{Region}}
- **Table/Index name**: {{TableName}}

## Schema

| Field | Type | Key | Description |
|-------|------|-----|-------------|
| {{Field1}} | {{Type1}} | PK | {{Field1Description}} |
| {{Field2}} | {{Type2}} | SK | {{Field2Description}} |
| {{Field3}} | {{Type3}} | - | {{Field3Description}} |

## Access Patterns

- {{AccessPattern1}}
- {{AccessPattern2}}

## Capacity & Scaling

- **Billing mode**: {{BillingMode}}
- **Read capacity**: {{ReadCapacity}}
- **Write capacity**: {{WriteCapacity}}

## Indexes

- **GSI**: {{GSI1Name}} ({{GSI1Keys}})

## Backup & Recovery

- **Point-in-time recovery**: {{PITREnabled}}
- **Backup frequency**: {{BackupFrequency}}

## Security

- **Encryption**: {{EncryptionType}}
- **Access control**: {{IAMPolicy}}
