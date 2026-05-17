# 1. Include the .env file
include .env

# 2. Export all variables so the terminal commands can see them
export

# 3. Construct the Database URL dynamically using $(VARIABLE_NAME)
DB_URL="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

# 4. Your migration commands
.PHONY: migrate-up migrate-down migrate-reset

migrate-up:
	goose -dir db/migrations postgres $(DB_URL) up

migrate-down:
	goose -dir db/migrations postgres $(DB_URL) down

migrate-reset:
	goose -dir db/migrations postgres $(DB_URL) reset