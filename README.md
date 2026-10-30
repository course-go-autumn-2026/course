# Trip Service

Сервис управления поездками. Выполнен в рамках лабораторных работ 1 и 2 курса Go.

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
| `make test-integration` | запустить интеграционные тесты PostgreSQL |
| `make lint` | проверить код конфигурацией курса |

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
| `OTEL_SERVICE_NAME` | имя сервиса в telemetry |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | адрес OTLP collector |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | протокол экспорта OTLP |
| `OTEL_RESOURCE_ATTRIBUTES` | дополнительные атрибуты ресурса |
| `OTEL_TRACES_SAMPLER` | стратегия сэмплирования трейсов |

Пример безопасных значений находится в `.env.example`. Локальный `.env` не
коммитится.

`HTTP_READ_TIMEOUT`, `HTTP_READ_HEADER_TIMEOUT`, `HTTP_WRITE_TIMEOUT` и
`HTTP_IDLE_TIMEOUT` можно переопределить через окружение. Если они не заданы,
используются безопасные значения по умолчанию: 5s, 2s, 10s и 60s.

## HTTP API

- `GET /health` — проверка, что процесс запущен;
- `GET /ready` — проверка доступности PostgreSQL;
- `POST /api/v1/trips` — создать поездку;
- `GET /api/v1/trips/{tripId}` — получить поездку;
- `POST /api/v1/trips/{tripId}/finish` — завершить поездку.

Контракт API находится в `contracts/openapi/trip-service.openapi.yaml`. Ошибки
бизнес-операций возвращаются в формате `application/problem+json`.

## Архитектура

Зависимости направлены внутрь:

```text
cmd → container → http → usecase → domain
                      ↓
                 repository → infra
```

- `cmd/trip-service` отвечает только за жизненный цикл процесса и graceful shutdown;
- `internal/container` — composition root: создаёт пул, telemetry, репозитории,
  use cases, HTTP-сервер и middleware;
- `internal/domain/trip` содержит сущность поездки, валидацию и переход статуса;
- `internal/usecase/trip` реализует операции создания, получения и завершения;
- `internal/repository/trip` содержит SQL на pgx и squirrel;
- `internal/infra/postgres` создаёт инструментированный pgxpool и менеджер
  транзакций;
- `internal/http/trip` содержит отдельные обработчики операций и преобразование
  между доменной моделью и типами, сгенерированными из OpenAPI.

Интерфейсы зависимостей находятся у потребителей: use case не зависит от
конкретного репозитория, а HTTP-обработчик — от конкретного use case. Точка
сборки с конкретными реализациями одна — `internal/container`.

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

## Наблюдаемость

Сервис отправляет трейсы и метрики по OTLP/HTTP в окружение `tripgoctl`.
HTTP middleware подключены в порядке `Recover → Tracing → Logging → Metrics →
router`.

- Логи пишутся в stdout в JSON. Для записей внутри спана добавляются `trace_id`
  и `span_id`.
- Для каждого HTTP-запроса записываются метод, шаблон маршрута, статус и
  длительность. Идентификатор поездки не попадает в `http.route`.
- В Jaeger видны HTTP-спан и вложенные спаны pgx для `SELECT`, `INSERT`,
  `UPDATE`, `BEGIN` и `COMMIT`.
- В Prometheus доступны `http.server.request.duration`,
  `http.server.active_requests`, `tripgo.trips.created` и
  `tripgo.trips.completed`.

Перед запуском integration-тестов и проверкой telemetry нужно поднять окружение:

```bash
tripgoctl cluster start
tripgoctl environment start
make migrate
make test-integration
make run
```

После запуска `tripgoctl connect` выводит локальные адреса Grafana, Jaeger и
Prometheus.
