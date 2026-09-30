SERVICE=trip-service

.PHONY: run migrate generate test build docker-build

run:
	set -a; \
	[ ! -f .env.example ] || . ./.env.example; \
	[ ! -f .env ] || . ./.env; \
	set +a; \
	go run ./cmd/trip-service

migrate:
	set -a; \
	[ ! -f .env ] || . ./.env; \
	set +a; \
	go tool goose -dir migrations postgres "$$DATABASE_URL" up

generate:
	go tool oapi-codegen -config config/oapi-codegen.yaml contracts/openapi/trip-service.openapi.yaml

test:
	go test -race ./...

build:
	go build -o bin/$(SERVICE) ./cmd/trip-service

docker-build:
	docker build -f deploy/Dockerfile -t trip-service:lab1 .