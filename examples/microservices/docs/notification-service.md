# Notification Service

The Notification Service handles all outbound communication to users across multiple channels.

## Channels

- **Email** - Transactional and marketing email
- **SMS** - Text messages
- **Push** - Mobile push notifications

## Events consumed

- `order.created` - Send order confirmation
- `order.shipped` - Send shipping notification
- `order.delivered` - Send delivery confirmation
- `user.created` - Send welcome email

## gRPC methods

- `SendNotification(NotificationRequest)` - Send an immediate notification
- `GetPreferences(PreferencesRequest)` - Get a user's notification preferences
- `UpdatePreferences(UpdatePreferencesRequest)` - Update preferences

## Technology

- Node.js with gRPC
- MongoDB for templates and preferences
- Kafka consumer for events
- Provider SDKs for email, SMS and push
