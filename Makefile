export $(shell cat .env | xargs)

MIGRATIONS_DIR = internal/infrastructure/postgres/migrations

migrate-up:
	migrate -database ${DATABASE_URL} -path ${MIGRATIONS_DIR} up

migrate-new:
	@echo "Usage: make migrate-new name=create_users_table"
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
