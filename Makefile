dev:
	GO_ENV=dev go run main.go

populate:
	psql postgresql://postgres:mysecretpassword@localhost:5432/postgres -f ./bin/populate_db.sql

reset-db:
	psql postgresql://postgres:mysecretpassword@localhost:5432/postgres -f ./bin/clear_db.sql
	psql postgresql://postgres:mysecretpassword@localhost:5432/postgres -f ./bin/populate_db.sql

doc:
	swag init