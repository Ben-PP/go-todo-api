# Deploy

Some day here will be a short guide on how to deploy. For now I will only tell
that config file must be found at `~/.config/go-todo/config.yaml`. You will also
need to create `/var/log/go-todo` directory and give write permission for the
user running the application.

# Development

## Presequisites

- Install Go

## 1. Set up the database

Firstly set up a PostgreSQL database for the development. Optionally you can
populate it using the `.sql` files in `bin`.

### Migrations

You can use `migrate` tool to do database migrations.

```bash
# Export the database url
POSTGRESQL_URL=postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable

# Run migrations with migrate
migrate -database ${POSTGRESQL_URL} -path db/migrations up
```

## 2. Set up config file

For development use `dev-config.yaml` file in the project root. You can create
this by copying the `config-example.yaml`.

```bash
cp config-example.yaml dev-config.yaml
```

## 3. Run the application

Run the application with Make.

```bash
make dev
```


Logs can be found in the `app.log` file when `GO_ENV=dev` and in
`/var/log/go-todo/app.log` else.

> With this command you can look at the logs in pretty format: `tail -f app.log | jq --indent 4`

