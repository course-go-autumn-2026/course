# Практика 10. Асинхронное взаимодействие на Go

## Requirements

- Go 1.25+
- Docker с Compose v2
- Apache Kafka 4.1.2 (образ указан в `compose.yaml`)

1. Скачайте репозиторий и откройте каталог практики с `go.mod`.
2. Выполните `make prepare`.
3. Выполните `make up`.
4. Выполните `make build`.

<details>
<summary>Команды без make, в том числе для Windows</summary>

В терминале IDE или PowerShell выполните по очереди:

```sh
go mod download
docker compose up -d --wait
go run ./cmd/lab -mode=setup
go build ./...
```

Kafka доступна по `localhost:19092`. Команда setup создаёт топики `orders.task1`, `orders.task2`, `orders.task3` и `orders.dlq`, по три партиции и одной реплике.

Остановка стенда: `docker compose down`. Данные остаются в volume; `down -v` удаляет историю.

</details>

## Контракт события

JSON содержит непустые `id`, `order_id`, `address` и `version=1`. Ключ сообщения равен `order_id`. При повторной отправке того же события сохраняются ID и исходные данные. Для нового события используются новые ID.

Обработчик пишет `handled` в лог. Дедупликации и внешнего бизнес-эффекта в стенде нет. Задания используют отдельные топики; producer для заданий 2 и 3 уже готов.

Для команд проверки ниже предполагается свежий стенд. При повторном запуске учитывайте историю: новая группа читает сохранённые события с начала. Незавершённые consumer той же группы могут забрать сообщение.

## Задание 1. Обмен и группы — 0,6 балла

Producer должен проверять контракт, отправлять событие в JSON с ключом `order_id` и возвращать результат отправки. Consumer должен фиксировать offset только после успешной обработки. Ошибки чтения, обработки и commit возвращаются вызывающему коду.

Проверка: `go test -count=1 -timeout=30s ./task01` и обмен между процессами:

```sh
go run ./cmd/lab -mode=send -task=1 -id=e1 -order=o1
go run ./cmd/lab -mode=consume -task=1 -group=team1-task1 -count=1
```

В логах ожидаются `event_id`, ключ, партиция и offset. Два consumer с одной группой должны делить партиции, две разные группы должны независимо прочитать события. Для проверки используйте несколько `order_id` и consumer с `-count=0` в отдельных терминалах.

## Задание 2. Retry и commit — 0,6 балла

Временная ошибка допускает три попытки обработки всего с паузами 100 и 200 мс. После успеха фиксируется offset. При исчерпании попыток возвращается последняя ошибка без commit. Постоянная ошибка (`ErrPermanent`, включая обёрнутую) и отмена завершают обработку без повторов и commit. Ожидание между попытками должно поддерживать отмену; ошибки ожидания и commit возвращаются вызывающему коду. DLQ в этом задании не используется.

Проверка: `go test -count=1 -timeout=30s ./task02` и сценарии:

```sh
go run ./cmd/lab -mode=send -task=2 -id=e2 -order=o2
go run ./cmd/lab -mode=consume -task=2 -group=team1-task2 -scenario=transient
```

Ожидаем успех на третьей попытке и commit. Для нового события при `-scenario=always-fail` ожидаем три ошибки без commit; запуск той же группы с `-scenario=ok` должен обработать это событие.

Проверка сбоя перед commit:

```sh
go run ./cmd/lab -mode=send -task=2 -id=e3 -order=o3
go run ./cmd/lab -mode=consume -task=2 -group=team1-task2 -crash-before-commit
go run ./cmd/lab -mode=consume -task=2 -group=team1-task2
```

В обоих запусках ожидается `handled` с одинаковыми `event_id`, partition и offset. Лимит попыток действует на один вызов обработки и сбрасывается после рестарта.

## Задание 3. DLQ и восстановление — 0,8 балла

После одной неуспешной попытки обработки сообщение должно попадать в DLQ. Запись содержит исходные `Topic`, `Partition`, `Offset`, `Key`, `Value`, `Headers`, число попыток `Attempts=1` и причину `Reason`. Исходный offset фиксируется только после подтверждения записи в DLQ. При успешной обработке достаточно commit.

Ошибка DLQ возвращается без commit. Отмена и превышение deadline не приводят ни к DLQ, ни к commit. Ошибка commit возвращается вызывающему коду.

Проверка: `go test -count=1 -timeout=30s ./task03 ./internal/replay` и сценарий восстановления:

```sh
go run ./cmd/lab -mode=send -task=3 -id=e4 -order=o4
go run ./cmd/lab -mode=consume -task=3 -group=team1-task3 -scenario=permanent
go run ./cmd/lab -mode=dlq -group=team1-inspect -count=1
```

Готовый режим replay должен повторно отправить выбранную запись после устранения причины ошибки. Подставьте `P` и `O` из строки `DLQ partition=P offset=O` (координаты записи DLQ):

```sh
go run ./cmd/lab -mode=replay -partition=P -offset=O
go run ./cmd/lab -mode=consume -task=3 -group=team1-task3 -scenario=ok
```

Ожидаем исходный `event_id` и новые координаты публикации. Переключение с `permanent` на `ok` моделирует исправление обработчика. Key, value и headers сохраняются, запись DLQ остаётся в журнале.

Для нового события режим `-scenario=permanent -dlq-fail` должен завершиться без commit; следующий запуск той же группы без `-dlq-fail` повторно получит событие. Сбой `-crash-before-commit` после записи DLQ допускает дубль в DLQ. Повторный replay тех же координат создаёт ещё одну публикацию и не восстанавливает прежний порядок событий.

Итоговая проверка: `go test -count=1 -timeout=30s ./...`. Остановка стенда: `make down`.
