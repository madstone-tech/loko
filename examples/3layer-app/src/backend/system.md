---
name: Backend API
description: RESTful API service providing business logic and data access
tags:
  - backend
  - api
  - go
responsibilities:
  - User authentication and authorization
  - Product catalog management
  - Order processing
  - Payment integration
dependencies:
  - Database
---

# Backend API System

The Backend API is a Go-based REST service that implements all business logic for the e-commerce platform.

## API Endpoints

### Products
- `GET /api/v1/products` - List products
- `GET /api/v1/products/{id}` - Get product details
- `POST /api/v1/products` - Create product (admin)

### Orders
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders/{id}` - Get order status
- `GET /api/v1/users/{id}/orders` - List user orders

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh token

## Technology

- Go 1.21
- Chi router
- JWT authentication
- PostgreSQL driver
