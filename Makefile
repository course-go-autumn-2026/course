.PHONY: help test test-integration generate migrate run lint

help:
	@echo "Available commands:"
	@echo "  make generate  Generate API code from OpenAPI"
	@echo "  make migrate   Apply database migrations"
	@echo "  make run       Run the Trip Service"
	@echo "  make test      Run tests with race detector"
	@echo "  make test-integration  Run PostgreSQL integration tests"
	@echo "  make lint      Run golangci-lint"

test:
	go test -race -cover ./...

test-integration:
	set -a; . ./.env; set +a; go test -tags=integration ./internal/repository/...

generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml contracts/openapi/trip-service.openapi.yaml
	go generate ./internal/...

lint:
	golangci-lint run ./...

migrate:
	set -a; . ./.env; set +a; go tool goose -dir migrations postgres "$$DATABASE_URL" up

run:
	set -a; . ./.env; set +a; go run ./cmd/trip-service
