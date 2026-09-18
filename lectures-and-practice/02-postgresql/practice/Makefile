export DATABASE_URL ?= postgres://app:app@localhost:5432/cinema?sslmode=disable

.PHONY: migrate-up migrate-down migrate-status bookings-schema run

migrate-up:
	goose -dir ./migrations postgres "$$DATABASE_URL" up

migrate-down:
	goose -dir ./migrations postgres "$$DATABASE_URL" down

migrate-status:
	goose -dir ./migrations postgres "$$DATABASE_URL" status

# Task 1
bookings-schema:
	docker compose exec postgres psql -U app -d cinema -c '\d bookings'

run:
	go run ./cmd/app $(ARGS)
