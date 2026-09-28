.PHONY: run

run:
	@set -a; . ./.env; set +a; go run ./cmd/trip-service
