# Загружаем переменные из .env файла
include .env
export $(shell sed 's/=.*//' .env)

GOOSE_DRIVER = postgres
GOOSE_DBSTRING = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)
MIGRATIONS_DIR = migrations

.PHONY: migrate-up migrate-up-to migrate-down migrate-status psql psql-docker

# Применение миграций
migrate-up:
	@echo "Applying migrations from $(MIGRATIONS_DIR)..."
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) up

# Применение миграций до указанной версии включительно
migrate-up-to:
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) up-to $(VERSION)

# Откат миграций
migrate-down:
	@echo "Rolling back migrations..."
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) down

# Просмотр статуса миграций
migrate-status:
	@echo "Migration status in $(MIGRATIONS_DIR):"
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) status

# Подключение с хоста; также поддерживает make psql < path/to/query.sql
psql:
	@PGPASSWORD="$(POSTGRES_PASSWORD)" psql -X -v ON_ERROR_STOP=1 -h "$(POSTGRES_HOST)" -p "$(POSTGRES_PORT)" -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)"

# Подключение через psql внутри контейнера (не требует локальной установки psql)
psql-docker:
	docker compose exec postgres psql -X -v ON_ERROR_STOP=1 -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)"

