---
name: "{{ContainerName}}"
description: "{{Description}}"
technology: "{{Technology}}"
---

# {{ContainerName}}

{{Description}}

## Purpose

This container is responsible for {{Purpose}}.

## Technology Stack

- **Primary**: {{Technology}}
- **Runtime**: {{Runtime}}
- **Database**: {{Database}}

## Interfaces

### Inbound

- REST API endpoints
- gRPC services
- Message queue consumers

### Outbound

- Database connections
- External service calls
- Cache operations

## Components

The container consists of the following components:

- **Handler**: HTTP request handling
- **Service**: Business logic
- **Repository**: Data access layer

## Deployment

- **Container Type**: {{ContainerType}}
- **Port**: {{Port}}
- **Environment**: {{Environment}}

## Monitoring

- Health checks: `/health`
- Metrics: Prometheus format
- Logs: Structured JSON
