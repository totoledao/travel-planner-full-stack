# Load environment variables from .env file
include .env

.PHONY: migrate

migrate:
	tern migrate --host=$(POSTGRES_HOST) --port=$(POSTGRES_PORT) --database=$(POSTGRES_DB) --user=$(POSTGRES_USER) --password=$(POSTGRES_PASSWORD) --migrations=./internal/pgstore/migrations