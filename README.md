# Trip Service

Сервис управления поездками. Выполнен в рамках лабораторной работы 1 курса Go.

## Требования

- Go 1.27 или новее;
- Docker Desktop;
- `tripgoctl`, доступный в `PATH`.

## Запуск

1. Поднять локальное окружение с PostgreSQL:

   ```bash
   tripgoctl cluster start
   tripgoctl environment start
   tripgoctl connect
   ```

2. Убедиться, что в `.env` есть все переменные из `.env.example`. Значение
   `DATABASE_URL`, созданное `tripgoctl`, заменять не нужно.

3. Применить миграции и запустить сервис:

   ```bash
   make migrate
   make run
   ```

По умолчанию сервис слушает `http://localhost:8080`.

## Команды

| Команда | Назначение |
| --- | --- |
| `make generate` | сгенерировать код из OpenAPI-контракта |
| `make migrate` | применить миграции PostgreSQL |
| `make run` | запустить сервис |
| `make test` | запустить тесты с race detector |

Дополнительно проект можно собрать командой:

```bash
go build ./...
```

## Переменные окружения

| Переменная | Назначение |
| --- | --- |
| `HTTP_ADDR` | адрес HTTP-сервера |
| `LOG_LEVEL` | уровень логирования |
| `SHUTDOWN_TIMEOUT` | максимальное время graceful shutdown |
| `HTTP_READ_TIMEOUT` | таймаут чтения запроса |
| `HTTP_READ_HEADER_TIMEOUT` | таймаут чтения заголовков |
| `HTTP_WRITE_TIMEOUT` | таймаут записи ответа |
| `HTTP_IDLE_TIMEOUT` | таймаут простоя соединения |
| `DATABASE_URL` | строка подключения к PostgreSQL |
| `DATABASE_MAX_CONNS` | максимальное число соединений в пуле |
| `DATABASE_MIN_CONNS` | минимальное число соединений в пуле |
| `DATABASE_MAX_CONN_LIFETIME` | максимальное время жизни соединения |
| `DATABASE_CONNECT_TIMEOUT` | таймаут подключения и начального Ping |
| `DATABASE_QUERY_TIMEOUT` | таймаут SQL-запроса |

Пример безопасных значений находится в `.env.example`. Локальный `.env` не
коммитится.

## HTTP API

- `GET /health` — проверка, что процесс запущен;
- `GET /ready` — проверка доступности PostgreSQL;
- `POST /api/v1/trips` — создать поездку;
- `GET /api/v1/trips/{tripId}` — получить поездку;
- `POST /api/v1/trips/{tripId}/finish` — завершить поездку.

Контракт API находится в `contracts/openapi/trip-service.openapi.yaml`. Ошибки
бизнес-операций возвращаются в формате `application/problem+json`.

## Решения

### Уровень изоляции транзакций

Используется `Read Committed` — стандартный уровень изоляции PostgreSQL. Для
создания поездки конкурентность защищена уникальным частичным индексом. Для
завершения поездки используется условный `UPDATE` только для строк со статусом
`active`, поэтому два одновременных запроса не смогут завершить одну поездку
дважды.

### Менеджер транзакций

`TransactionManager.Do` открывает транзакцию и сохраняет её в `context`.
Репозитории берут исполнителя из контекста: внутри `Do` это транзакция, вне неё
— пул соединений. Вложенный вызов `Do` использует уже существующую транзакцию.
При ошибке или panic выполняется rollback, при успешном выполнении — commit.

Создание и завершение поездки записывают изменения в `trips` и
`trip_status_history` в одной транзакции.

### Одна активная поездка у водителя

В PostgreSQL создан уникальный частичный индекс
`trips_driver_active_unique_idx` по `driver_id` для строк со статусом `active`.
Нарушение этого индекса определяется по коду PostgreSQL `23505` и превращается
в HTTP-ошибку `409 driver_busy`.
