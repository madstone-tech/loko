# Scheduled Tasks

Lambda functions invoked by EventBridge rules rather than by requests or messages.

| Function | Schedule | Description |
|----------|----------|-------------|
| generate-daily-report | Daily | Summarizes the previous day's orders |
| cleanup-expired-orders | Hourly | Cancels orders left unpaid past their expiry |

Both read the orders table directly. The cleanup task uses conditional updates, so it never
cancels an order that was paid while it ran.
