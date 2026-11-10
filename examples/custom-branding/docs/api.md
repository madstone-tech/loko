# API

The service follows a clean architecture pattern with clear separation of concerns:

1. **Handlers** - HTTP request/response handling
2. **Services** - Business logic
3. **Repositories** - Data access

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/users` | GET | List all users |
| `/api/v1/users/{id}` | GET | Get user by ID |
| `/api/v1/users` | POST | Create new user |
| `/api/v1/users/{id}` | PUT | Update user |
| `/api/v1/users/{id}` | DELETE | Delete user |

## Getting started

```bash
# Run the service
go run cmd/api/main.go

# Run tests
go test ./...
```
