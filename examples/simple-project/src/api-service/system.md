---
name: API Service
description: RESTful API service for the application
tags:
  - api
  - backend
  - rest
responsibilities:
  - Handle HTTP requests
  - Validate input data
  - Coordinate with downstream services
  - Return JSON responses
dependencies:
  - Database Service
  - Cache Service
---

# API Service

The API Service is the main entry point for client applications. It exposes a RESTful interface for all application functionality.

## Overview

This service handles:
- User authentication and authorization
- CRUD operations for resources
- Request validation and error handling
- Response formatting

## Architecture

The service follows a clean architecture pattern with clear separation of concerns:

1. **Handlers** - HTTP request/response handling
2. **Services** - Business logic
3. **Repositories** - Data access

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/users` | GET | List all users |
| `/api/v1/users/{id}` | GET | Get user by ID |
| `/api/v1/users` | POST | Create new user |
| `/api/v1/users/{id}` | PUT | Update user |
| `/api/v1/users/{id}` | DELETE | Delete user |

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: Standard library `net/http`
- **Database**: PostgreSQL 15
- **Cache**: Redis 7

## Getting Started

```bash
# Run the service
go run cmd/api/main.go

# Run tests
go test ./...
```
