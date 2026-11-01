# Trip Service — лабораторная работа 1

Сервис создаёт поездки, возвращает их по ID и завершает. HTTP API задан в
[`contracts/openapi/trip-service.openapi.yaml`](contracts/openapi/trip-service.openapi.yaml),
данные хранятся в PostgreSQL. Текущая архитектура описана в
[`docs/architecture.md`](docs/architecture.md). Материалы курса сохранены в
[`docs/course/homework/`](docs/course/homework/) и
[`docs/course/lectures-and-practice/`](docs/course/lectures-and-practice/).

## Структура репозитория

```text
cmd/                  точка входа сервиса
internal/             реализация сервиса и тесты
migrations/           SQL-миграции goose
api/                  конфигурация генерации HTTP API
contracts/            контракты, используемые сервисом
docs/
  architecture.md     описание текущей архитектуры
  course/
    homework/         задания, требования и справочные материалы курса
    lectures-and-practice/  лекции и практические задания
go.mod, go.sum        модуль Go и зависимости
Makefile              запуск, миграции, генерация и тесты
Dockerfile            сборка контейнера сервиса
environment.toml      состав локального окружения
.env.example          пример конфигурации
README.md             описание и инструкция запуска
```

Все лабораторные последовательно развивают один сервис. Копия контрактов в
`docs/course/homework/contracts/` относится к материалам курса; генерация кода
сервиса использует корневой `contracts/`.

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
make generate       # сгенерировать API-код из OpenAPI и моки
make migrate        # применить миграции
make migrate-down   # откатить одну миграцию; повторить для второй
make test           # go test -race ./...
```

Доменные и HTTP unit-тесты запускаются без БД. HTTP-тесты используют `httptest`
и сгенерированные `gomock`-моки зависимостей. `make generate` обновляет моки через
`go:generate`; сгенерированные `mocks_test.go` хранятся рядом с потребителями.

Интеграционные тесты используют отдельную переменную `TEST_DATABASE_URL`. Без неё
они пропускаются; для полного прогона подготовьте **отдельную тестовую БД**,
примените к ней миграции, задайте `TEST_DATABASE_URL` с её адресом и запустите
`make test`. Тесты создают поездки и временный триггер, затем удаляют их. Команда
`make generate` должна оставлять `internal/generated/api.gen.go` без изменений.

Для ручной проверки отправьте `POST /api/v1/trips` с телом из
[условия ЛР1](docs/course/homework/tasks/01-http-and-postgres/README.md), затем запросите
`GET /api/v1/trips/{tripId}` и `POST /api/v1/trips/{tripId}/finish`.
Повторное создание поездки для занятого водителя возвращает `409 driver_busy`,
повторное завершение — `409 trip_completed`.

## Идемпотентное создание

Необязательный заголовок `Idempotency-Key` должен содержать UUID, созданный
клиентом. Первый `POST /api/v1/trips` с ключом возвращает `201`, повтор с теми же
данными — `200`, исходное тело ответа и тот же `Location`. Повтор с другим
содержимым возвращает `409 idempotency_conflict`. Без ключа создание работает
как обычно.

Ключ действует **24 часа** после успешного создания. Сравниваются проверенные
значения полей запроса, поэтому порядок полей и пробелы в JSON не влияют на
результат. По истечении срока ключ можно использовать заново; запись с этим
ключом удаляется при следующем обращении. Для очистки остальных просроченных
записей можно периодически выполнять:

```sql
DELETE FROM trip_idempotency_keys WHERE expires_at <= now();
```

Запросы с одним ключом последовательно проходят через транзакционную блокировку
PostgreSQL. Поездка, история статуса и снимок первого ответа сохраняются в одной
транзакции. Миграция `00003` создаёт таблицу ключей; `make migrate-down` откатывает
по одной миграции за вызов.

## Запуск в Docker

`Dockerfile` собирает бинарь в образе Go и переносит его в
`distroless/static-debian12:nonroot`. В итоговом образе нет исходников, тестов и
Go toolchain; процесс работает с UID/GID `65532:65532`. Размер образа при
проверке для `linux/arm64` — **4 267 012 байт** (около 4,1 МиБ).

Пример отдельного локального запуска сервиса рядом с PostgreSQL в Docker:

```bash
docker build -t trip-service:lab1 .
docker network create trip-lab
docker run -d --name trip-postgres --network trip-lab -p 55432:5432 \
  -e POSTGRES_USER=tripgo -e POSTGRES_PASSWORD=tripgo -e POSTGRES_DB=tripgo \
  postgres:17-alpine
until docker exec trip-postgres pg_isready -U tripgo -d tripgo; do sleep 1; done

docker_db_url='postgres://tripgo:tripgo@127.0.0.1:55432/tripgo?sslmode=disable'
go tool goose -dir migrations postgres "$docker_db_url" up
docker run --rm --name trip-service --network trip-lab -p 8080:8080 \
  --env-file .env.example \
  -e DATABASE_URL='postgres://tripgo:tripgo@trip-postgres:5432/tripgo?sslmode=disable' \
  trip-service:lab1
```

Во втором терминале проверьте `curl -i localhost:8080/ready`; ожидается `200`.
После остановки сервиса удалите проверочное окружение командами
`docker rm -f trip-postgres` и `docker network rm trip-lab`.

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

Сервис реализует обязательную часть ЛР1 и оба дополнительных задания.
