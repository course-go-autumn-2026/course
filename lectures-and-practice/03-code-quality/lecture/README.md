# Примеры к лекции «Качество кода»

Проект повторяет порядок тем из презентации и состоит из независимых небольших примеров. Команды
выполняются из этого каталога.

## Содержание

| Каталог | Тема лекции | Что показать |
|---|---|---|
| `01-testing` | Пакет `testing` | Соглашения `_test.go` и `TestXxx`, `Errorf`, `Log` |
| `02-table-testify` | Табличные тесты, `testify` | `t.Run`, `assert`, `require`, проверка ошибок |
| `03-httptest` | `httptest` | Табличный тест HTTP-обработчика без сетевого порта |
| `04-mocks` | Моки и dependency injection | Ручной stub, GoMock, `go:generate`, ожидания вызовов |
| `05-integration` | Интеграционные тесты | PostgreSQL, build tag `integration`, `suite.Suite`, setup/teardown |
| `06-test-quality/parallel` | Параллельные тесты | `t.Parallel()` и флаг `-parallel` |
| `06-test-quality/race` | Состояние гонки | Небезопасный и защищённый счётчики, `go test -race` |
| `06-test-quality/goleak` | Утечки горутин | Намеренная утечка и корректная остановка по context |
| `06-test-quality/flaky` | Flaky-тесты | Неопределённый выбор при равных значениях и `-count` |
| `06-test-quality/coverage` | Code coverage | Непокрытые ветви и HTML-отчёт |
| `06-test-quality/loadtest` | Нагрузочное тестирование | Конкурентные запросы, число ошибок, average и p95 |
| `07-linters` | Линтеры, форматтеры, style guide | Конфиг golangci-lint, намеренные нарушения под build tag |

Из прошлогоднего `l6-examples` сохранены небольшие учебные сюжеты про стандартные тесты,
`testify`, `httptest`, PostgreSQL, parallel/race/goleak/flaky и coverage. Из
`l11-examples` взят только относящийся к презентации паттерн «consumer-side interface +
сгенерированный GoMock». CRUD-сервис, CI, Dockerfile и прочие темы той лекции сюда не переносились.

## Быстрый запуск

Нужен Go 1.25 или новее.

```bash
go test ./...
```

Обычный прогон зелёный. Намеренно падающие демонстрации включаются отдельно и перечислены ниже.

### Пакет testing, таблицы, testify и httptest

```bash
go test -v ./01-testing
go test -v ./02-table-testify
go test -v ./03-httptest
```

Запуск небольшого HTTP-сервера для ручной проверки:

```bash
make server
curl 'http://localhost:8080/greet?name=Alice'
```

### Моки

Интерфейс `UserRepository` объявлен у потребителя — сервиса. В каталоге есть два теста:
с простым stub, написанным вручную, и со сгенерированным GoMock.

```bash
go test -v ./04-mocks
make generate
git diff -- 04-mocks/mocks/user_repository_mock.go
```

Директива `go:generate` фиксирует версию генератора, поэтому глобальная установка `mockgen`
не требуется. Сгенерированный файл хранится в репозитории: слушателям достаточно обычного
`go test`.

### Интеграционный тест

```bash
docker compose up -d --wait
go test -v -tags=integration ./05-integration
docker compose down -v
```

Или короче: `make integration`, затем `make integration-down`. Compose поднимает PostgreSQL
на порту 5434 и применяет SQL из `05-integration/migrations`. Другой сервер можно указать через
`TEST_DATABASE_URL`.

В `RepositoryTestSuite`:

- `SetupSuite` открывает реальное соединение;
- `SetupTest` создаёт данные для каждого теста;
- `TearDownTest` очищает их;
- `TearDownSuite` закрывает пул.

Без тега `integration` этот тест не компилируется и не входит в быстрый unit-прогон.

## Качество тестов

### Параллельное выполнение

Сравните продолжительность:

```bash
go test -count=1 -v -parallel=1 ./06-test-quality/parallel
go test -count=1 -v -parallel=2 ./06-test-quality/parallel
```

### Race detector

```bash
go test ./06-test-quality/race
go test -race -run '^TestCounterRace$' ./06-test-quality/race
go test -race -run '^TestSafeCounter$' ./06-test-quality/race
```

Вторая команда должна завершиться ошибкой и напечатать `DATA RACE`; третья должна пройти.

### Утечки горутин

```bash
go test ./06-test-quality/goleak
RUN_LEAK_DEMO=1 go test ./06-test-quality/goleak
```

Первая команда проверяет корректно останавливаемый worker. Во второй намеренно включается
падающий тест: `goleak` находит горутину, которая навсегда осталась в `select {}`.

### Flaky-тест

```bash
RUN_FLAKY_DEMO=1 go test -count=100 ./06-test-quality/flaky
```

У двух ключей одинаковый максимум, а порядок обхода map не определён. Тест ошибочно требует
конкретный ключ и при многократном запуске падает. Исправления для обсуждения: определить
tie-breaker в контракте либо проверять множество допустимых результатов.

### Покрытие

```bash
go test -coverprofile=coverage.out ./06-test-quality/coverage
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Откройте отчёт и добавьте случаи `weightKG <= 0`, тяжёлой и express-посылки. Это удобный способ
найти непроверенные ветви. При этом высокий процент сам по себе не доказывает, что тесты содержат
сильные проверки результата.

### Минимальный генератор нагрузки

В первом терминале:

```bash
make server
```

Во втором:

```bash
go run ./06-test-quality/loadtest -requests=1000 -concurrency=20
```

Пример печатает количество запросов и ошибок, среднее время и p95. Это учебный генератор для
демонстрации терминов из слайда, а не замена k6/JMeter и мониторингу CPU, памяти и базы данных.

## Линтеры и форматтеры

Корневой `.golangci.yml` использует формат v2 и включает небольшой базовый набор проверок.

```bash
golangci-lint run ./...
golangci-lint fmt ./...
```

Намеренно проблемный `07-linters/problems.go` скрыт build tag, поэтому обычная проверка проекта
остаётся зелёной. Для демонстрации диагностики:

```bash
golangci-lint run --build-tags=lintdemo ./07-linters
```

Для показа разницы форматирования:

```bash
gofmt -d 07-linters/unformatted.go.txt
```

В `style.go` также видны правила из Go Uber Style Guide: ранний возврат, оборачивание ошибки
через `%w` и заранее заданный `cap` слайса.

## Удобные make-цели

```bash
make test
make coverage
make generate
make integration
make test-race     # ожидаемо падает
make flaky         # ожидаемо падает
make leak          # ожидаемо падает
```
