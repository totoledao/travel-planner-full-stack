# Load environment variables from .env file
include .env

run:
	goapi-gen --out ./internal/api/spec/spec.go ./internal/api/spec/spec.json && tern migrate --host=$(POSTGRES_HOST) --port=$(POSTGRES_PORT) --database=$(POSTGRES_DB) --user=$(POSTGRES_USER) --password=$(POSTGRES_PASSWORD) --migrations=./internal/pgstore/migrations && sqlc generate -f ./internal/pgstore/sqlc.yaml
	