.PHONY: help test generate migrate run

help:
	@echo "Available commands:"
	@echo "  make generate  Generate API code from OpenAPI"
	@echo "  make migrate   Apply database migrations"
	@echo "  make run       Run the Trip Service"
	@echo "  make test      Run tests with race detector"

test:
	go test -race ./...

generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml contracts/openapi/trip-service.openapi.yaml

migrate:
	set -a; . ./.env; set +a; go tool goose -dir migrations postgres "$$DATABASE_URL" up

run:
	set -a; . ./.env; set +a; go run ./cmd/trip-service