# Database Connection String
DB_DSN="host=localhost port=5432 user=postgres password=root dbname=expense_db sslmode=disable"

# Target to apply all pending migrations
migrate-up:
	goose -dir db/migrations postgres $(DB_DSN) up

# Target to rollback the last migration
migrate-down:
	goose -dir db/migrations postgres $(DB_DSN) down

# Target to check migration status
migrate-status:
	goose -dir db/migrations postgres $(DB_DSN) status