# Database

PostgreSQL stores all persistent data for the e-commerce platform. Writes go to the primary;
reads are served by a replica. The primary is backed up daily to object storage.

## Responsibilities

- Store user data
- Store the product catalog
- Store order history
- Ensure data integrity

## Schema

### users

- id (UUID, PK)
- email (VARCHAR, UNIQUE)
- password_hash (VARCHAR)
- created_at (TIMESTAMP)

### products

- id (UUID, PK)
- name (VARCHAR)
- description (TEXT)
- price (DECIMAL)
- stock (INTEGER)

### orders

- id (UUID, PK)
- user_id (UUID, FK)
- status (VARCHAR)
- total (DECIMAL)
- created_at (TIMESTAMP)
