.PHONY: prepare up down setup build check check-1 check-2 check-3
prepare:
	go mod download
up:
	docker compose up -d --wait
	go run ./cmd/lab -mode=setup
down:
	docker compose down
setup:
	go run ./cmd/lab -mode=setup
build:
	go build ./...
check:
	go test -count=1 -timeout=30s ./...
check-1:
	go test -count=1 ./task01
check-2:
	go test -count=1 ./task02
check-3:
	go test -count=1 ./task03 ./internal/replay
