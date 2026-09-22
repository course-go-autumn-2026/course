.PHONY: test test-race flaky leak coverage integration integration-down generate lint fmt server load

test:
	go test ./...

# Команда должна завершиться ошибкой: детектор покажет гонку в Counter.
test-race:
	go test -race ./06-test-quality/race

# Эти команды должны завершиться ошибкой и тем самым показать проблему.
flaky:
	RUN_FLAKY_DEMO=1 go test -count=100 ./06-test-quality/flaky

leak:
	RUN_LEAK_DEMO=1 go test ./06-test-quality/goleak

coverage:
	go test -coverprofile=coverage.out ./06-test-quality/coverage
	go tool cover -func=coverage.out

integration:
	docker compose up -d --wait
	go test -tags=integration ./05-integration

integration-down:
	docker compose down -v

generate:
	go generate ./04-mocks

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

server:
	go run ./03-httptest/cmd/server

load:
	go run ./06-test-quality/loadtest -requests=1000 -concurrency=20
