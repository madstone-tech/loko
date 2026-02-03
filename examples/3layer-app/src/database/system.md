---
name: Database
description: PostgreSQL database for persistent data storage
tags:
  - database
  - postgresql
  - persistence
responsibilities:
  - Store user data
  - Store product catalog
  - Store order history
  - Ensure data integrity
---

# Database System

The Database system uses PostgreSQL to store all persistent data for the e-commerce platform.

## Schema

### Users Table
- id (UUID, PK)
- email (VARCHAR, UNIQUE)
- password_hash (VARCHAR)
- created_at (TIMESTAMP)

### Products Table
- id (UUID, PK)
- name (VARCHAR)
- description (TEXT)
- price (DECIMAL)
- stock (INTEGER)

### Orders Table
- id (UUID, PK)
- user_id (UUID, FK)
- status (VARCHAR)
- total (DECIMAL)
- created_at (TIMESTAMP)

## Technology

- PostgreSQL 15
- Automatic backups
- Read replicas for scaling
