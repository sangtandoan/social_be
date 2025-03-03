MIGRATION_PATH = ./cmd/migrate/migrations
DB_ADDR = postgres://admin:secret@localhost:5432/social?sslmode=disable

# Define variables with default values and you can override it when running make 
#
# Example:
# make migrate-force VERSION=3
VERSION ?= 

migrate:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-force:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) force $(VERSION)

migrate-up:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose up

migrate-up1:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose up 1

migrate-down:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose down

migrate-down1:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) -verbose down 1

.PHONY: migrate migrate-up migrate-up1 migrate-down migrate-down1 migrate-force
