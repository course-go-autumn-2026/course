.PHONY: run generate

generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml contracts/openapi/trip-service.openapi.yaml

run:
	@set -a; . ./.env; set +a; go run ./cmd/trip-service
