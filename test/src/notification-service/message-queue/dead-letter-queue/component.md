---
id: dead-letter-queue
name: "Dead Letter Queue"
description: "Shared DLQ for failed email and SMS messages. 14-day retention, CloudWatch alarm on depth > 0"
technology: "Amazon SQS Standard"
tags:
  - "queue"
  - "dlq"
  - "retry"
---

# Dead Letter Queue

Shared DLQ for failed email and SMS messages. 14-day retention, CloudWatch alarm on depth > 0

## Context

This is a **C4 Level 3 - Component** representing code-level abstractions within a container.

## Responsibility

This component is responsible for Shared DLQ for failed email and SMS messages. 14-day retention, CloudWatch alarm on depth > 0.

## Technology

- **Language**: Amazon SQS Standard
- **Framework**: (specify framework)
- **Pattern**: (e.g., MVC, CQRS, Event-Sourcing)

## Interfaces

### Public Methods

- `Method1()` - Description of method 1
- `Method2()` - Description of method 2

### Dependencies

- (List external dependencies like libraries, frameworks)

## Implementation Details

### Key Classes/Functions

- `Class1` - Description
- `Class2` - Description

### Data Structures

- (List important data structures)

## Testing

- Unit tests: (specify framework)
- Integration tests: (specify framework)
- Coverage: (target %)

## Performance Considerations

- (Note any performance-critical aspects)
- (Document caching strategies)
- (List optimization opportunities)

