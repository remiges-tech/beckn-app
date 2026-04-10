# Load .env if it exists — makes DB_* vars available to tern and other tools.
ifneq (,$(wildcard cmd/server/.env))
  include cmd/server/.env
  export
endif

.PHONY: run build tidy migrate-up migrate-down migrate-status sqlc

## run: start the BAP server (loads .env automatically)
run:
	go run ./cmd/server/...

## build: compile the server binary to ./bin/bap
build:
	go build -o bin/bap ./cmd/server/...

## tidy: sync go.mod / go.sum
tidy:
	go mod tidy

## migrate-up: apply all pending Tern migrations
migrate-up:
	cd db/migrations && tern migrate --config tern.conf

## migrate-down: roll back the last migration
migrate-down:
	cd db/migrations && tern migrate --destination -1 --config tern.conf

## migrate-status: show current migration version
migrate-status:
	cd db/migrations && tern status --config tern.conf

## sqlc: regenerate type-safe Go code from SQL queries
sqlc:
	cd db && sqlc generate
