.PHONY: run build tidy migrate-up migrate-down

## run: start the BPP server (loads .env automatically)
run:
	go run ./cmd/server/...

## build: compile the server binary to ./bin/bpp
build:
	go build -o bin/bpp ./cmd/server/...

## tidy: sync go.mod / go.sum
tidy:
	go mod tidy

## migrate-up: apply all pending Tern migrations
migrate-up:
	cd db && tern migrate --config migrations/tern.conf

## migrate-down: roll back the last migration
migrate-down:
	cd db && tern migrate --destination -1 --config migrations/tern.conf
