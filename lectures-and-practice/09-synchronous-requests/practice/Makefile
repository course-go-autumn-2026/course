.PHONY: prepare tools generate check-1 check-2 check-3 check build server client
PROTOC ?= protoc
TASK ?= 1
DELAY ?= 2s
prepare:
	go mod download
	go build ./...
tools:
	GOBIN="$(CURDIR)/.bin" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.9
	GOBIN="$(CURDIR)/.bin" go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
generate:
	PATH="$(CURDIR)/.bin:$$PATH" $(PROTOC) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative task01/api/delivery.proto
check-1:
	go test -count=1 -timeout=20s ./task01
check-2:
	go test -count=1 -timeout=20s ./task02
check-3:
	go test -count=1 -timeout=20s ./task03
check:
	go test -count=1 -timeout=30s ./...
build:
	go build -o .bin/server ./cmd/server
	go build -o .bin/client ./cmd/client
server:
	go run ./cmd/server -task=$(TASK) -delay=$(DELAY)
client:
	go run ./cmd/client -task=$(TASK)
