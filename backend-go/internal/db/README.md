# Database Package Placeholder

This package will hold Postgres connection, migrations, and generated query code.

Planned order:

1. Add database connection lifecycle.
2. Add migrations for users, host requests, game templates, and game runs.
3. Add card/call/claim tables once the deterministic game APIs start.
4. Introduce `sqlc` after the first schema is stable enough to generate types.
