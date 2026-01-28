---
name: Notification Service
description: Handles all user notifications via multiple channels
tags:
  - notifications
  - email
  - sms
  - push
responsibilities:
  - Send email notifications
  - Send SMS messages
  - Send push notifications
  - Manage notification preferences
---

# Notification Service

The Notification Service handles all outbound communications to users across multiple channels.

## Channels

- **Email** - Transactional and marketing emails via SendGrid
- **SMS** - Text messages via Twilio
- **Push** - Mobile push notifications via Firebase

## Events Consumed

- `order.created` - Send order confirmation
- `order.shipped` - Send shipping notification
- `order.delivered` - Send delivery confirmation
- `user.created` - Send welcome email

## API

### gRPC Methods

- `SendNotification(NotificationRequest)` - Send immediate notification
- `GetPreferences(PreferencesRequest)` - Get user notification preferences
- `UpdatePreferences(UpdatePreferencesRequest)` - Update preferences

## Technology

- Node.js with gRPC
- MongoDB for templates and preferences
- Kafka consumer for events
- SendGrid, Twilio, Firebase SDKs
