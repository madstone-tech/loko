---
name: "API Gateway"
description: "REST API entry point for sending notifications and querying delivery status"
technology: "Amazon API Gateway REST"
---

# API Gateway

REST API entry point for sending notifications and querying delivery status

## Context

This is a **C4 Level 2 - Container** representing a deployable unit within the system.

## Purpose

This container is responsible for REST API entry point for sending notifications and querying delivery status.

## Technology Stack

- **Primary**: Amazon API Gateway REST
- **Runtime**: (e.g., Docker, JVM, Node.js)
- **Database**: (e.g., PostgreSQL, Redis)

## Components

This container is composed of the following components:

| Component | Description | Technology |
|-----------|-------------|------------|
| (Add your components here) | | |

## Interfaces

### Inbound

- REST API endpoints
- gRPC services
- Message queue consumers

### Outbound

- Database connections
- External service calls
- Cache operations

## Deployment

- **Container Type**: (e.g., Docker, Pod)
- **Port**: (e.g., 8080)
- **Environment**: (e.g., dev, staging, prod)

## Monitoring

- Health checks: `/health`
- Metrics: Prometheus format
- Logs: Structured JSON

