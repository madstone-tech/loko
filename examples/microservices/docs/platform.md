# Shared platform

Infrastructure that every service depends on but no single service owns.

## Event bus

Kafka carries domain events between services. Producers publish after committing their own
state; consumers are idempotent, so redelivery is safe.

| Topic prefix | Producer | Consumers |
|---|---|---|
| `user.*` | User Service | Order Service, Notification Service |
| `order.*` | Order Service | Notification Service |
