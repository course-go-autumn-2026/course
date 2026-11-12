# Практика 9. Синхронные запросы на Go

## Requirements

- Go 1.25+
- protoc 33.0
- protoc-gen-go v1.36.9 и protoc-gen-go-grpc v1.5.1

1. Скачайте репозиторий и откройте каталог практики с `go.mod`.
2. Выполните `go mod download` и `go build ./...`.
3. Установите protoc и проверьте `protoc --version`.
4. Выполните `make tools`.
5. Выполните `make generate` и повторите `go build ./...`.

<details>
<summary>Установка и команды без make, в том числе для Windows</summary>

Для установки protoc используйте [официальную инструкцию](https://protobuf.dev/installation/).
На Windows команды можно выполнять в терминале IDE или PowerShell.

`make tools` устанавливает плагины в локальную `.bin`, а `make generate` добавляет её в `PATH` на время генерации. Если запускаете protoc вручную с этими плагинами, добавьте `.bin` в `PATH` своего терминала.

Без make установите плагины командами:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.9
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
```

При такой установке добавьте в `PATH` каталог `bin` внутри пути, который возвращает `go env GOPATH` (либо свой `GOBIN`, если он задан).

Генерация:

```sh
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative task01/api/delivery.proto
```

</details>

## Общий бизнес-контракт

`Estimate(address, weight_grams)` возвращает `price_kopecks` и `eta_minutes`.

- Адреса `Центр` и `Север` обслуживаются. Строки точные, регистр важен.
- Вес — целое число граммов от 1 до 20000 включительно.
- База: 20000 копеек для Центра и 30000 для Севера.
- За каждый начатый килограмм добавляется 5000 копеек.
- Время: 30 минут для Центра и 45 для Севера.
- Для 1500 граммов: Центр — 30000 копеек и 30 минут; Север — 40000 копеек и 45 минут.
- Пустой адрес, нулевой, отрицательный или слишком большой вес — некорректные данные.
- Любой другой непустой адрес — зона не найдена.
- Специальный адрес `Недоступно` имитирует временный отказ. Это управление учебным стендом, а не настоящий адрес.

Расчёт уже реализован в `internal/domain/estimate.go`. Не дублируйте его и не изменяйте тесты для получения зелёного результата.

## Задание 1. Контракт и сервер — 0,6 балла

Сервис `task01` должен возвращать стоимость и время доставки по бизнес-контракту. Поле ответа `eta_minutes` имеет тип `int32` и номер `2`; номера существующих полей сохраняются. Proto-контракт и сгенерированный код должны соответствовать друг другу.

Проверка: `go test -count=1 -timeout=20s ./task01` и вызов сервера из другого процесса.

```sh
# Терминал 1
go run ./cmd/server -task=1

# Терминал 2
go run ./cmd/client -task=1
go run ./cmd/client -task=1 -address=Север
```

Для веса 1500 г ожидаем `price_kopecks:30000`, `eta_minutes:30` для Центра и `price_kopecks:40000`, `eta_minutes:45` для Севера.

## Задание 2. Клиент и ошибки — 0,6 балла

Клиент `task02` должен получать расчёт и различать ошибки по контракту ниже, в том числе обёрнутые доменные ошибки. Отмена и deadline исходного контекста должны сохраняться. Остальные gRPC-ошибки возвращаются без потери статуса.

| Ошибка домена | Статус сервера | Ошибка клиента |
|---|---|---|
| `ErrInvalid` | `InvalidArgument` | `ErrInput` |
| `ErrOutside` | `NotFound` | `ErrNoDelivery` |
| `ErrUnavailable` | `Unavailable` | `ErrDependency` |
| Другая ошибка | `Internal` без внутренних подробностей | Исходная gRPC-ошибка |

Проверка: `go test -count=1 -timeout=20s ./task02` и вызовы клиента при запущенном сервере `go run ./cmd/server -task=2`:

```sh
go run ./cmd/client -task=2
go run ./cmd/client -task=2 -weight=0
go run ./cmd/client -task=2 -address=Луна
go run ./cmd/client -task=2 -address=Недоступно
```

Первый вызов возвращает расчёт, остальные завершаются с ненулевым кодом и сообщениями соответствующих ошибок клиента.

## Задание 3. Deadline и отмена — 0,8 балла

В `task03` вызов должен учитывать `budget`, более короткий deadline родителя и явную отмену. После отмены сервер прекращает ожидание. Быстрый запрос завершается успешно.

Проверка: `go test -count=1 -timeout=20s ./task03` и два сценария:

```sh
# Терминал 1
go run ./cmd/server -task=3 -delay=2s

# Терминал 2
go run ./cmd/client -task=3 -timeout=200ms
go run ./cmd/client -task=3 -timeout=3s -cancel-after=100ms
```

Ожидаем `DeadlineExceeded` и `Canceled` соответственно. Сервер пишет `calculation ended` вскоре после отмены, не дожидаясь двух секунд. При `-delay=0s` обычный запрос должен проходить.

Итоговая проверка: `go test -count=1 -timeout=30s ./...`.
