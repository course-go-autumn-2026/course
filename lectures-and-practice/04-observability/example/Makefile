SHELL := /bin/sh
.DEFAULT_GOAL := build

URL ?= http://127.0.0.1:8080/orders
RPS ?= 5
MODE ?= normal
export URL RPS MODE

.PHONY: build run-orders run-couriers infra-up infra-down dashboard load mode check

build:
	mkdir -p bin
	go build -o bin/orders ./orders
	go build -o bin/couriers ./couriers
	go build -o bin/load ./load

run-orders: build
	@set -eu; \
	[ -f .env ] || { echo 'Run make infra-up first: .env is missing' >&2; exit 1; }; \
	set -a; . ./.env; set +a; \
	OTEL_SERVICE_NAME=order-service exec ./bin/orders

run-couriers: build
	@set -eu; \
	[ -f .env ] || { echo 'Run make infra-up first: .env is missing' >&2; exit 1; }; \
	set -a; . ./.env; set +a; \
	OTEL_SERVICE_NAME=courier-service exec ./bin/couriers

infra-up:
	tripgoctl cluster start
	tripgoctl environment start
	tripgoctl connect

infra-down:
	tripgoctl environment stop

dashboard:
	@sh import-dashboard.sh

load:
	@sh load.sh "$$URL" "$$RPS"

mode:
	@case "$$MODE" in normal|slow|error|timeout) ;; *) echo 'MODE must be normal, slow, error or timeout' >&2; exit 2;; esac; \
	curl --fail --silent --show-error --max-time 5 -X POST "http://127.0.0.1:8081/admin/mode/$$MODE"

check:
	go test -race ./...
	go vet ./...
