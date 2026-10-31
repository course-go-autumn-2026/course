# Ручная проверка варианта 3

Запускай команды из каталога `practice/variant-3`. Нужны Go версии из
`go.mod`, `curl` и Python 3 для локальной заглушки. Примеры рассчитаны на
Bash и последовательное выполнение с пустым хранилищем.

## 1. Проверить сборку и тесты

- [ ] Все команды завершаются без ошибок:

```bash
go build -o /tmp/variant3-order .
go vet ./...
go test ./... -v
```

Хранилище работает в памяти. При перезапуске сервера заказы исчезают,
а нумерация начинается с 1. В ответах GET и PATCH сохранены исходные
имена полей: `ID`, `CustomerID`, `Items`, `Total`, `ConfirmedQty`,
`Status`, `CreatedAt`; у позиций — `SKU`, `Quantity`, `Price`.

## 2. Запустить управляемые заглушки

- [ ] В первом терминале запустить заглушку склада и антифрода:

```bash
mkdir -p /tmp/variant3-manual-check
printf '200' > /tmp/variant3-manual-check/stock
printf '200' > /tmp/variant3-manual-check/fraud

python3 -u - <<'PY'
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

state = Path('/tmp/variant3-manual-check')

class ChecksHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path not in ('/stock', '/fraud'):
            self.send_response(404)
            self.end_headers()
            return
        code = int((state / self.path.removeprefix('/')).read_text())
        self.send_response(code)
        self.end_headers()

ThreadingHTTPServer(('127.0.0.1', 8081), ChecksHandler).serve_forever()
PY
```

- [ ] Во втором терминале запустить сервис:

```bash
STOCK_CHECK_URL=http://127.0.0.1:8081/stock \
PAYMENT_FRAUD_URL=http://127.0.0.1:8081/fraud \
go run .
```

Ожидается `listening on :8080`. Все следующие команды выполняй в третьем
терминале. Ответами заглушки можно управлять через файлы в `/tmp` без
перезапуска сервиса и потери заказов.

Без переменных окружения используются адреса из исходного задания:
`http://stock-check.internal/check` и
`http://payment-fraud.internal/check`.

## 3. Создать и получить подтверждённый заказ

- [ ] Создать первый заказ:

```bash
curl -i -X POST http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":2,"price":500}]}'
```

Ожидается `201`, тело `{"id":1,"status":"confirmed"}`. В логе заглушки
сначала появляется `POST /fraud`, затем `POST /stock`.

- [ ] Получить сохранённый заказ:

```bash
curl -i http://localhost:8080/orders/1
```

Ожидается `200`, `Content-Type: application/json`, `CustomerID: 1`,
`Total: 1000`, `ConfirmedQty: 2`, `Status: "confirmed"`, исходные позиции
и заполненный `CreatedAt`.

## 4. Проверить изменение состава и статусов

- [ ] Такое же суммарное количество сохраняет `confirmed`, даже при смене SKU:

```bash
curl -i -X PATCH http://localhost:8080/orders/1/items \
  -d '{"items":[{"sku":"book","quantity":1,"price":300},{"sku":"chair","quantity":1,"price":700}]}'
```

Ожидается `200`, `Total: 1000`, `Status: "confirmed"`, `ConfirmedQty: 2`.

- [ ] Уменьшение количества сохраняет подтверждение:

```bash
curl -i -X PATCH http://localhost:8080/orders/1/items \
  -d '{"items":[{"sku":"lamp","quantity":1,"price":500}]}'
```

Ожидается `200`, `Total: 500`, `Status: "confirmed"`, `ConfirmedQty: 2`.

- [ ] Превышение подтверждённого количества требует новой проверки склада:

```bash
curl -i -X PATCH http://localhost:8080/orders/1/items \
  -d '{"items":[{"sku":"lamp","quantity":3,"price":500}]}'
```

Ожидается `200`, `Total: 1500`, `Status: "pending_stock"`,
`ConfirmedQty: 2`. В логе заглушки новых запросов нет: PATCH сам не
вызывает склад и антифрод.

- [ ] Обратное уменьшение не подтверждает заказ автоматически:

```bash
curl -i -X PATCH http://localhost:8080/orders/1/items \
  -d '{"items":[{"sku":"lamp","quantity":2,"price":500}]}'
curl -i http://localhost:8080/orders/1
```

В обоих ответах ожидается `200`, `Total: 1000`, `Status: "pending_stock"`,
`ConfirmedQty: 2`. GET подтверждает, что изменения сохранены.

## 5. Проверить валидацию и отсутствующие заказы

- [ ] Создание без позиций возвращает `400` и `order must have at least one item`:

```bash
curl -i -X POST http://localhost:8080/orders -d '{"customer_id":1,"items":[]}'
```

- [ ] Нулевая quantity возвращает `400` и `item quantity must be positive`:

```bash
curl -i -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":0,"price":500}]}'
```

- [ ] Нулевая price возвращает `400` и `item price must be positive`:

```bash
curl -i -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":1,"price":0}]}'
```

Повтори две проверки с отрицательными значениями: результат тот же.
Невалидный POST не должен обращаться к складу или антифроду.

- [ ] Проверить ошибки разбора, 404 и порядок проверок PATCH:

```bash
curl -i -X POST http://localhost:8080/orders -d '{'
curl -i http://localhost:8080/orders/abc
curl -i http://localhost:8080/orders/999999
curl -i -X PATCH http://localhost:8080/orders/999999/items -d '{"items":[]}'
curl -i -X PATCH http://localhost:8080/orders/999999/items \
  -d '{"items":[{"sku":"lamp","quantity":1,"price":500}]}'
```

Ожидания по порядку: `400` с пустым телом; `400` с пустым телом;
`404` с пустым телом; `400 order must have at least one item`;
`404` с пустым телом. У сообщений, записанных через `http.Error`,
в конце тела есть перевод строки.

## 6. Проверить лимиты клиентов

- [ ] Повторить запрос, подставляя значения из таблицы:

```bash
curl -i -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"sofa","quantity":1,"price":100000}]}'
```

| customer_id | quantity | price | Ожидаемый результат |
|---|---|---|---|
| 1 | 1 | 100000 | 201, confirmed |
| 1 | 1 | 100001 | 400 |
| 1 | 2 | 60000 | 400, лимит считается по сумме |
| 2 | 1 | 300000 | 201, confirmed |
| 2 | 2 | 500000 | 201, confirmed |
| 2 | 1 | 1000001 | 400 |
| 999 | 1 | 100000 | 201, confirmed |
| 999 | 1 | 100001 | 400 |

Для превышения лимита точный текст — `order total exceeds limit for this customer`.
Неизвестный клиент не даёт 404: он считается `new` с лимитом 100000.

- [ ] Проверить, что PATCH не применяет лимит создания:

```bash
curl -i -X PATCH http://localhost:8080/orders/1/items \
  -d '{"items":[{"sku":"sofa","quantity":1,"price":2000000}]}'
```

Ожидается `200`, `Total: 2000000`, `Status: "pending_stock"`.
PATCH сохраняет поведение исходника: не повторяет проверки положительных
quantity/price и лимита суммы. Эти различия с POST также закреплены в тестах.

## 7. Проверить блокировку антифрода

- [ ] Получить ID успешного заказа перед блокировкой:

```bash
previous_id=$(curl -sS -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":1,"price":500}]}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
printf '%s\n' "$previous_id"
```

- [ ] Включить блокировку и отправить следующий заказ:

```bash
printf '403' > /tmp/variant3-manual-check/fraud
curl -i -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":1,"price":500}]}'
curl -i "http://localhost:8080/orders/$((previous_id + 1))"
```

Ожидается `403 payment looks fraudulent`, затем `404` с пустым телом.
В логе заглушки есть `/fraud`, но нет нового `/stock`.

- [ ] Снять блокировку и убедиться, что отклонённый запрос не занял ID:

```bash
printf '200' > /tmp/variant3-manual-check/fraud
curl -i -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":1,"price":500}]}'
```

Ожидается `201 confirmed`; ID равен `previous_id + 1`.

## 8. Проверить отказ склада и сбои внешних сервисов

- [ ] Отказ склада переводит новый заказ в ожидание:

```bash
printf '503' > /tmp/variant3-manual-check/stock
curl -i -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":2,"price":500}]}'
```

Ожидается `201 pending_stock`. При GET по возвращённому ID будет
`ConfirmedQty: 0`. Любой ответ склада кроме 200 даёт такой результат.

- [ ] Ответ 500 антифрода не блокирует создание:

```bash
printf '200' > /tmp/variant3-manual-check/stock
printf '500' > /tmp/variant3-manual-check/fraud
curl -i -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":2,"price":500}]}'
```

Ожидается `201 confirmed`. Антифрод блокирует только ответом 403.

- [ ] Остановить заглушку через Ctrl+C в первом терминале, оставив сервис
  работающим, и повторить запрос:

```bash
curl -i --max-time 5 -X POST http://localhost:8080/orders \
  -d '{"customer_id":1,"items":[{"sku":"lamp","quantity":2,"price":500}]}'
```

Ожидается `201 pending_stock`: сетевая ошибка антифрода пропускается,
а сетевая ошибка склада переводит заказ в ожидание.

## 9. Проверить недоступные через публичный API состояния

Отдельной ручки отмены нет; состояние `created` существует между
сохранением и ответом склада. Эти случаи проверяются подготовленными
состояниями заказа в тестах, без добавления служебных HTTP-ручек.

- [ ] Отменённый заказ возвращает HTTP 409 с нужным текстом:

```bash
go test ./handler -run '^TestCancelledOrderHTTP$' -v
```

- [ ] Отменённый заказ не изменяется, а `created` после PATCH сохраняет статус:

```bash
go test ./service -run '^TestUpdateItems/(created|cancelled)$' -v
```

- [ ] Разная обработка сбоев и порядок «антифрод → сохранение → склад»:

```bash
go test ./service -run '^TestCreateOrder_ExternalChecks$' -v
```

## 10. Проверить архитектуру перед сдачей

- [ ] В [main.go](main.go) только создание зависимостей, маршруты и запуск.
- [ ] В [domain/order.go](domain/order.go) правила создания, пересчёт суммы
  и переходы состояний; в [domain/customer.go](domain/customer.go) — лимит клиента.
- [ ] В [service/ports.go](service/ports.go) отдельные порты заказов,
  клиентов, склада и антифрода. Реакция на ошибки внешних проверок находится
  в [service/order_service.go](service/order_service.go).
- [ ] Хранилище и HTTP-вызовы находятся в `repository`; изменения состава
  явно сохраняются, а чтение не раскрывает изменяемые данные хранилища.
- [ ] В [handler/order_handler.go](handler/order_handler.go) только разбор
  запросов, вызовы сервиса, JSON и перевод доменных ошибок в HTTP.
- [ ] В Go-файлах нет комментариев.
- [ ] После проверки остановить сервис через Ctrl+C. В описании PR отметить
  закрытые пункты: слои, богатые модели, ошибки, порты и адаптеры, DDD.
