MIGRATION_PATH = ./cmd/migrate/migrations

migrate:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose up

migrate-up1:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose up 1

migrate-down:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose down

migrate-down1:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose down 1
