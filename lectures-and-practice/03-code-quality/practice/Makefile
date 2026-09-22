.PHONY: test coverage race leak generate lint

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./03-quality-checks
	go tool cover -func=coverage.out

race:
	go test -race ./03-quality-checks

leak:
	go test -tags=leaktest ./03-quality-checks

generate:
	go generate ./02-mocks

lint:
	golangci-lint run ./04-linters
