# Order Service

The Order Service manages the complete order lifecycle, from creation to fulfillment.

## Responsibilities

- Order creation and validation
- Order status management
- Payment processing coordination
- Inventory reservation

## gRPC methods

- `CreateOrder(CreateOrderRequest)` - Create a new order
- `GetOrder(GetOrderRequest)` - Get an order by ID
- `UpdateOrderStatus(UpdateStatusRequest)` - Update order status
- `CancelOrder(CancelRequest)` - Cancel a pending order
- `ListUserOrders(ListRequest)` - Get a user's orders

## Events

### Published

- `order.created` - New order placed
- `order.paid` - Payment confirmed
- `order.shipped` - Order shipped
- `order.delivered` - Order delivered

### Consumed

- `user.deleted` - Anonymize the user's orders

## Technology

- Go with gRPC
- PostgreSQL for order data
- Kafka for event streaming
