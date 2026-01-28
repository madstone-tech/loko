---
name: Order Service
description: Manages order lifecycle and processing
tags:
  - orders
  - transactions
  - processing
responsibilities:
  - Order creation and validation
  - Order status management
  - Payment processing coordination
  - Inventory reservation
dependencies:
  - User Service
  - Notification Service
---

# Order Service

The Order Service manages the complete order lifecycle from creation to fulfillment.

## API

### gRPC Methods

- `CreateOrder(CreateOrderRequest)` - Create new order
- `GetOrder(GetOrderRequest)` - Get order by ID
- `UpdateOrderStatus(UpdateStatusRequest)` - Update order status
- `CancelOrder(CancelRequest)` - Cancel pending order
- `ListUserOrders(ListRequest)` - Get orders for user

## Events

### Published
- `order.created` - New order placed
- `order.paid` - Payment confirmed
- `order.shipped` - Order shipped
- `order.delivered` - Order delivered

### Consumed
- `user.deleted` - Handle user deletion (anonymize orders)

## Technology

- Go with gRPC
- PostgreSQL for order data
- Kafka for event streaming
