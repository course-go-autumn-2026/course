.PHONY: run generate migrate migrate-down

generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml contracts/openapi/trip-service.openapi.yaml

migrate:
	@set -a; . ./.env; set +a; go tool goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	@set -a; . ./.env; set +a; go tool goose -dir migrations postgres "$$DATABASE_URL" down

run:
	@set -a; . ./.env.example; . ./.env; set +a; go run ./cmd/trip-service
