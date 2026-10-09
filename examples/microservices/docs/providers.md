# Delivery providers

The Notification Service hands messages to one external provider per channel: email, SMS and
mobile push. Each is called over HTTPS and treated as unreliable; failed sends are retried.
