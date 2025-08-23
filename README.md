# Database

## Migrations

You can use `migrate` tool to do database migrations.

```bash
# Export the database url
POSTGRESQL_URL=postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable

# Run migrations with migrate
migrate -database ${POSTGRESQL_URL} -path db/migrations up
```

# Development

1. Set up a postgresql database with database 'postgres' and postgres password 'mysecretpassword'.
2. Run the migrations against the db. Optionally populate the db with test data.

```bash
make populate
```

3. Download dependencies.

```bash
go get .
```

4. Run the backend.

```bash
make dev
```

With this command you can look at the logs in pretty format.

```bash
tail -f app.log | jq --indent 4
```
