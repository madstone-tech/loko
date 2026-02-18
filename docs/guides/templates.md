# Templates Guide

This guide explains how loko's technology-aware template system works, how to override the default template, and how to create custom templates.

## Table of Contents

- [Overview](#overview)
- [Technology-to-Template Mapping](#technology-to-template-mapping)
- [Using Templates](#using-templates)
- [The --template Override Flag](#the---template-override-flag)
- [Custom Templates](#custom-templates)
- [Template Placeholders](#template-placeholders)
- [Available Templates](#available-templates)

---

## Overview

When you create a new component with `loko new component`, loko automatically selects the most appropriate template based on the component's **technology** field.

For example:
- `Lambda` → `compute` template (includes Trigger, Runtime, Memory sections)
- `DynamoDB` → `datastore` template (includes Schema, Access Patterns, Capacity sections)
- `SQS` → `messaging` template (includes Queue config, Producers, Consumers sections)

If no technology match is found, the `generic` template is used as a fallback.

---

## Technology-to-Template Mapping

| Technology Keywords | Template Category | Template File |
|--------------------|------------------|---------------|
| Lambda, Function, Functions, Serverless, Step Functions | `compute` | `compute.md` |
| DynamoDB, RDS, Aurora, PostgreSQL, MySQL, Redis, ElastiCache, Elasticsearch | `datastore` | `datastore.md` |
| SQS, SNS, Kinesis, EventBridge, Kafka, RabbitMQ, NATS | `messaging` | `messaging.md` |
| API Gateway, REST, GraphQL, gRPC, HTTP, FastAPI, Express, Gin | `api` | `api.md` |
| EventBridge, Step Functions, Scheduler, Cron | `event` | `event.md` |
| S3, EFS, EBS, Storage, Blob, GCS, CloudFront | `storage` | `storage.md` |
| (anything else) | `generic` | `generic.md` |

The matching is case-insensitive and uses partial matching (e.g., "Lambda" matches "AWS Lambda Function").

---

## Using Templates

Templates are selected automatically when you create a component:

```bash
# Auto-select template based on technology
loko new component --name "Payment Processor" \
                   --technology "AWS Lambda" \
                   --container api-gateway \
                   --system payment-service
```

loko will:
1. Detect "AWS Lambda" → `compute` category
2. Render `templates/component/compute.md` with your component's variables
3. Write the rendered file to `src/payment-service/api-gateway/payment-processor.md`

---

## The --template Override Flag

Override automatic selection with the `--template` flag:

```bash
# Force the datastore template for a custom database wrapper
loko new component --name "Cache Manager" \
                   --technology "Redis Cluster" \
                   --template datastore \
                   --container backend \
                   --system my-service
```

Valid template names: `compute`, `datastore`, `messaging`, `api`, `event`, `storage`, `generic`

---

## Custom Templates

You can create custom templates in your project or globally.

### Project-Local Templates

Place custom templates in `.loko/templates/component/`:

```
.loko/
└── templates/
    └── component/
        └── my-custom.md   ← Custom template
```

Use with:
```bash
loko new component --name "My Thing" --template my-custom
```

### Global Templates

Place templates in `~/.config/loko/templates/component/` to share across projects.

### Template Search Order

1. Project-local: `.loko/templates/component/`
2. Project templates directory: `templates/component/`
3. Global: `~/.config/loko/templates/component/`
4. Built-in: embedded in the loko binary

---

## Template Placeholders

All templates support these standard placeholders:

| Placeholder | Description |
|-------------|-------------|
| `{{ComponentName}}` | Component display name |
| `{{Description}}` | Component description |
| `{{Technology}}` | Technology stack |
| `{{ContainerName}}` | Parent container name |
| `{{SystemName}}` | Parent system name |
| `{{Date}}` | Current date (YYYY-MM-DD) |
| `{{component_table}}` | Auto-generated component table (in container templates) |
| `{{container_table}}` | Auto-generated container table (in system templates) |

---

## Available Templates

### compute.md

For serverless functions, Lambda, Azure Functions, Cloud Run.

**Sections**: Purpose, Trigger, Runtime Configuration, Memory & Timeout, IAM Permissions, Environment Variables, Error Handling, Monitoring

### datastore.md

For databases, caches, search indices.

**Sections**: Purpose, Schema Definition, Access Patterns, Capacity & Performance, Backup & Recovery, Security

### messaging.md

For queues, topics, event streams.

**Sections**: Purpose, Queue/Topic Configuration, Producers, Consumers, Dead Letter Queue, Monitoring

### api.md

For REST APIs, GraphQL endpoints, gRPC services.

**Sections**: Purpose, Endpoints, Authentication, Rate Limiting, Error Responses, Versioning

### event.md

For event buses, schedulers, workflow engines.

**Sections**: Purpose, Event Schema, Rules & Filters, Targets, Error Handling

### storage.md

For object storage, file systems, CDN.

**Sections**: Purpose, Bucket/Container Configuration, Access Control, Lifecycle Policies, Encryption

### generic.md

Fallback for any technology not matched above.

**Sections**: Purpose, Implementation, Dependencies, Monitoring
