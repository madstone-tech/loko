# External services

- **Payment provider** - card payments. Called by Process Payment with the order ID as the
  idempotency key.
- **Email provider** - transactional email. Called by Send Notification.
- **Inventory service** - stock levels and reservations. Called by Update Inventory.

API keys are held in Secrets Manager and read at cold start.
