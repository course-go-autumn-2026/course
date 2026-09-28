# Trip Service — лабораторная работа 1

Сервис создаёт поездки, возвращает их по ID и завершает. HTTP API задан в
[`contracts/openapi/trip-service.openapi.yaml`](contracts/openapi/trip-service.openapi.yaml),
данные хранятся в PostgreSQL. Материалы курса сохранены в [`homework/`](homework/)
и [`lectures-and-practice/`](lectures-and-practice/).

## Запуск

Нужны Go **1.25.7+** и установленный [`tripgoctl`](https://github.com/course-go-autumn-2026/course-infra).
Для ЛР1 окружение поднимает только PostgreSQL. Из корня репозитория выполните:

```bash
tripgoctl cluster start          # один раз на машине
tripgoctl environment start
tripgoctl connect                # проверить адреса окружения
make migrate
make run
```

`tripgoctl environment start` создаёт `.env` с адресом БД. `make run` сначала
загружает безопасные значения из `.env.example`, затем значения из `.env`.
Пароли и локальные файлы `.env` и `.tripgo/` в Git не попадают. При отсутствии
обязательной переменной сервис завершается с ошибкой. Для остановки отправьте
`SIGINT` (`Ctrl+C`) или `SIGTERM`.

После запуска проверьте:

```bash
curl -i localhost:8080/health
curl -i localhost:8080/ready
```

Адрес HTTP берите из `HTTP_ADDR` в `.env` или `.env.example`, если он переопределён.
`/health` отвечает `200`, пока работает процесс, и не обращается к БД. `/ready`
проверяет PostgreSQL и отвечает `503`, если соединиться не удалось.

## Переменные окружения

Полный пример — [`.env.example`](.env.example). Все значения ниже задаются через
окружение; указаны значения примера, а `DATABASE_URL` берите из `.env`, созданного
`tripgoctl`.

| Переменная | Назначение | Пример |
|---|---|---|
| `HTTP_ADDR` | адрес HTTP-сервера | `:8080` |
| `HTTP_READ_TIMEOUT` | чтение запроса | `15s` |
| `HTTP_READ_HEADER_TIMEOUT` | чтение заголовков | `5s` |
| `HTTP_WRITE_TIMEOUT` | запись ответа | `15s` |
| `HTTP_IDLE_TIMEOUT` | простой соединения | `60s` |
| `LOG_LEVEL` | уровень JSON-логов | `info` |
| `SHUTDOWN_TIMEOUT` | общий бюджет остановки | `10s` |
| `DATABASE_URL` | адрес и учётные данные PostgreSQL | из `.env` |
| `DATABASE_MAX_CONNS` | максимум соединений пула | `10` |
| `DATABASE_MIN_CONNS` | минимум соединений пула | `2` |
| `DATABASE_MAX_CONN_LIFETIME` | срок жизни соединения | `30m` |
| `DATABASE_CONNECT_TIMEOUT` | подключение и стартовый `Ping` | `5s` |
| `DATABASE_QUERY_TIMEOUT` | запросы и транзакции | `3s` |

## Команды и проверка

```bash
go build ./...
make generate       # заново сгенерировать API-код из OpenAPI
make migrate        # применить миграции
make migrate-down   # откатить одну миграцию; повторить для второй
make test           # go test -race ./...
```

Интеграционные тесты используют отдельную переменную `TEST_DATABASE_URL`. Без неё
они пропускаются; для полного прогона подготовьте **отдельную тестовую БД**,
примените к ней миграции, задайте `TEST_DATABASE_URL` с её адресом и запустите
`make test`. Тесты создают поездки и временный триггер, затем удаляют их. Команда
`make generate` должна оставлять `internal/generated/api.gen.go` без изменений.

Для ручной проверки отправьте `POST /api/v1/trips` с телом из
[условия ЛР1](homework/tasks/01-http-and-postgres/README.md), затем запросите
`GET /api/v1/trips/{tripId}` и `POST /api/v1/trips/{tripId}/finish`.
Повторное создание поездки для занятого водителя возвращает `409 driver_busy`,
повторное завершение — `409 trip_completed`.

## Решения

- Транзакции открывает `TxManager.Do`. Он помещает `pgx.Tx` в контекст, при
  успехе выполняет `COMMIT`, при ошибке или панике — `ROLLBACK`. Репозиторий
  выбирает транзакцию из контекста, а вне `Do` использует пул. Вложенный `Do`
  переиспользует текущую транзакцию. Создание поездки и запись истории статуса
  выполняются в одном `Do`; завершение и его запись в историю — тоже.
- Уровень изоляции — **Read Committed**. Этого достаточно для этих операций:
  частичный уникальный индекс по `driver_id` запрещает две активные поездки
  даже при одновременных вставках, а завершение обновляет только строку со
  статусом `active`. Второй конкурентный вызов видит уже завершённую поездку.
- Нарушение уникального индекса распознаётся по коду PostgreSQL `23505` и
  превращается в `driver_busy`. SQL в репозитории собирается через `squirrel`
  с плейсхолдерами `$N`.
- HTTP-типы и серверный интерфейс генерируются из неизменённого OpenAPI.
  Для ЛР1 выбраны пять операций: `createTrip`, `getTrip`, `finishTrip`, `health`,
  `ready`. Ошибки поездок имеют формат `application/problem+json`.

Сервис реализует обязательную часть ЛР1. Дополнительные задания на
`Idempotency-Key` и Dockerfile не включены.
