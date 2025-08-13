# Database

## Migrations

You can use `migrate` tool to do database migrations.

```bash
# Export the database url
POSTGRESQL_URL=postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable

# Run migrations with migrate
migrate -database ${POSTGRESQL_URL} -path db/migrations up
```
