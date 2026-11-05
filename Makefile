.PHONY: run generate migrate migrate-down test test-integration lint

TEST_FLAGS ?=
GOLANGCI_LINT ?= golangci-lint

generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml contracts/openapi/trip-service.openapi.yaml
	go generate ./internal/...

migrate:
	@set -a; . ./.env; set +a; go tool goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	@set -a; . ./.env; set +a; go tool goose -dir migrations postgres "$$DATABASE_URL" down

run:
	@set -a; . ./.env.example; . ./.env; set +a; go run ./cmd/trip-service

test:
	@set -eu; profile=$$(mktemp); trap 'rm -f "$$profile"' EXIT; \
		go test -race $(TEST_FLAGS) -coverprofile="$$profile" ./...; \
		summary=$$(go tool cover -func="$$profile"); printf '%s\n' "$$summary" | tail -n 1

test-integration:
	@: "$${TEST_DATABASE_URL:?set TEST_DATABASE_URL to a migrated test database}"
	go test -race $(TEST_FLAGS) -tags=integration ./internal/integration ./internal/infra/postgres/transaction

lint:
	$(GOLANGCI_LINT) run ./...
