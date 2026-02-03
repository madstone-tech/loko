---
name: User Service
description: Manages user accounts and authentication
tags:
  - users
  - authentication
  - identity
responsibilities:
  - User registration and login
  - Profile management
  - Password reset
  - Session management
---

# User Service

The User Service handles all user-related operations including authentication, registration, and profile management.

## API

### gRPC Methods

- `CreateUser(CreateUserRequest)` - Register new user
- `GetUser(GetUserRequest)` - Get user by ID
- `UpdateUser(UpdateUserRequest)` - Update user profile
- `AuthenticateUser(AuthRequest)` - Validate credentials
- `RefreshToken(RefreshRequest)` - Refresh JWT token

## Events Published

- `user.created` - New user registered
- `user.updated` - User profile updated
- `user.deleted` - User account deleted

## Technology

- Go with gRPC
- PostgreSQL for user data
- Redis for session cache
- Kafka for events
