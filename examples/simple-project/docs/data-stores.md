# Data stores

- **PostgreSQL 15** holds all persistent data. Only the repositories talk to it.
- **Redis 7** caches frequent lookups. The business services read through it and fall back to the
  repositories on a miss.
