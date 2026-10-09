# User Service

The User Service handles all user-related operations, including authentication, registration and
profile management.

## Responsibilities

- User registration and login
- Profile management
- Password reset
- Session management

## gRPC methods

- `CreateUser(CreateUserRequest)` - Register a new user
- `GetUser(GetUserRequest)` - Get a user by ID
- `UpdateUser(UpdateUserRequest)` - Update a user profile
- `AuthenticateUser(AuthRequest)` - Validate credentials
- `RefreshToken(RefreshRequest)` - Refresh a JWT

## Events published

- `user.created` - New user registered
- `user.updated` - User profile updated
- `user.deleted` - User account deleted

## Technology

- Go with gRPC
- PostgreSQL for user data
- Redis for the session cache
- Kafka for events
