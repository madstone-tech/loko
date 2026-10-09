# API Gateway

The API Gateway is the single entry point for all client applications. It routes requests to the
appropriate microservice.

## Responsibilities

- Route requests to the right service
- Handle authentication and authorization
- Rate limiting
- Request and response transformation

## Features

- Request routing based on path and headers
- JWT token validation
- Rate limiting per client
- Request and response logging
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
