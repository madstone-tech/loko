---
name: API Gateway
description: Central entry point for all client requests
tags:
  - gateway
  - routing
  - authentication
responsibilities:
  - Route requests to appropriate services
  - Handle authentication/authorization
  - Rate limiting
  - Request/response transformation
dependencies:
  - User Service
  - Order Service
---

# API Gateway

The API Gateway serves as the single entry point for all client applications, routing requests to the appropriate microservices.

## Features

- Request routing based on path and headers
- JWT token validation
- Rate limiting per client
- Request/response logging
- Circuit breaker pattern

## Routes

| Path | Service | Description |
|------|---------|-------------|
| `/api/users/*` | User Service | User management |
| `/api/orders/*` | Order Service | Order management |
| `/api/notifications/*` | Notification Service | Notification preferences |

## Technology

- Kong Gateway
- Lua plugins for custom logic
- Redis for rate limiting state
