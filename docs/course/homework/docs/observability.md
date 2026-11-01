# Каталог телеметрии

Список того, что сервис должен отдавать. Имена фиксированные, так как по ним
собраны дашборды и по ним проверяет ментор. Свои дополнительные сигналы добавлять
можно.

Метрики и трейсы пишутся через OpenTelemetry SDK и уходят по OTLP в коллектор
(см. [`environment.md`](environment.md)).

## 1. Атрибуты ресурса

Проставляются один раз при инициализации SDK:

| Атрибут | Значение |
|---|---|
| `service.name` | `trip-service` |
| `service.version` | версия сборки |
| `service.namespace` | `tripgo` |
| `deployment.environment` | `local` |

## 2. Метрики — лабораторная работа 2

Имена HTTP- и db-метрик взяты из семантических конвенций OpenTelemetry: так их
понимают готовые дашборды и не приходится придумывать своё.

| Метрика | Тип | Единица | Лейблы | Что показывает |
|---|---|---|---|---|
| `http.server.request.duration` | histogram | секунды | `http.request.method`, `http.route`, `http.response.status_code` | сколько занял ответ; отсюда перцентили, RPS и доля ошибок |
| `http.server.active_requests` | up-down counter | штуки | `http.request.method`, `http.route` | сколько запросов обрабатывается прямо сейчас |
| `db.client.operation.duration` | histogram | секунды | `db.operation.name`, `db.collection.name`, `error.type` | сколько занял запрос в бд |
| `tripgo.trips.created` | counter | штуки | — | созданные поездки |
| `tripgo.trips.completed` | counter | штуки | — | завершённые поездки |
| `tripgo.trips.active` | gauge | штуки | — | сколько поездок сейчас в статусе `active` |

Типы: **counter** только растёт, по нему считают скорость; **up-down counter**
растёт и убывает, показывает текущее значение; **gauge** — то же, но значение
измеряется в момент сбора; **histogram** раскладывает значения по корзинам,
из него считают перцентили.

В `http.route` кладём шаблон маршрута (`/api/v1/trips/{tripId}`), а не конкретный
URL: иначе на каждую поездку заведётся своя временная серия, и метрика перестанет
агрегироваться.

Отдельных счётчиков на бизнес-ошибки не заводим — они видны по лейблу
`http.response.status_code`.

## 3. Метрики — лабораторная работа 3

| Метрика | Тип | Единица | Лейблы | Что показывает |
|---|---|---|---|---|
| `tripgo.positions.saved` | counter | штуки | — | сохранённые координаты |
| `tripgo.stale_checks` | counter | штуки | `result` = `queued` \| `duplicate` \| `dropped` | что сканер сделал с поездкой на этом тике |
| `tripgo.position_requests.queue.size` | gauge | штуки | — | сколько задач сейчас в очереди |
| `tripgo.position_requests.workers.active` | gauge | штуки | — | сколько воркеров сейчас заняты |
| `tripgo.position_requests` | counter | штуки | `result` = `success` \| `error` \| `dropped` | чем закончилась попытка запросить координату |
| `tripgo.position_requests.duration` | histogram | секунды | `result` | сколько занял вызов Push Service |

Что читать по этим метрикам:

- `dropped` растёт — очередь переполнена, часть поездок остаётся без опроса. Это
  первое, что смотрят, когда пропали пуши;
- `workers.active` упёрся в `POSITION_WORKERS`, а `queue.size` растёт — сосед
  отвечает медленнее, чем мы успеваем разгребать;
- `duplicate` в `stale_checks` — нормально: поездка уже стоит в очереди, значит
  дедупликация работает.

## 4. Метрики — лабораторная работа 4

| Метрика | Тип | Единица | Лейблы | Что показывает |
|---|---|---|---|---|
| `rpc.server.duration` | histogram | секунды | `rpc.method`, `rpc.grpc.status_code` | сколько занял входящий gRPC-вызов |
| `rpc.client.duration` | histogram | секунды | `rpc.method`, `rpc.grpc.status_code` | сколько занял наш вызов соседа |
| `messaging.publish.duration` | histogram | секунды | `messaging.destination.name` | сколько занял publish в топик |
| `tripgo.events.published` | counter | штуки | `event_type`, `result` = `ok` \| `error` | опубликованные события; `error` означает, что событие потеряно |
| `tripgo.commands.consumed` | counter | штуки | `result` = `ok` \| `duplicate` \| `dlq` | чем закончилась обработка команды |
| `tripgo.consumer.lag` | gauge | штуки | — | сколько сообщений в топике ещё не прочитано |

Что читать: растущий `lag` означает, что консьюмер не успевает; всплеск `dlq` —
что в топик поехало что-то, чего сервис не понимает; ненулевой `error` в
`events.published` — что события теряются, и это тот случай, который чинит outbox
из работы 5.

## 5. Метрики — лабораторная работа 5

Обязательные:

| Метрика | Тип | Единица | Лейблы | Что показывает |
|---|---|---|---|---|
| `tripgo.retry.attempts` | counter | штуки | `operation`, `outcome` = `retry` \| `success` \| `exhausted` | сколько было повторов и чем кончились; `exhausted` — попытки исчерпаны |
| `tripgo.ratelimit.decisions` | counter | штуки | `operation`, `result` = `allowed` \| `throttled` | сколько вызовов пропустил лимитер, сколько придержал |

Вариант A, outbox:

| Метрика | Тип | Единица | Лейблы | Что показывает |
|---|---|---|---|---|
| `tripgo.outbox.pending` | gauge | штуки | — | сколько событий ждёт публикации |
| `tripgo.outbox.published` | counter | штуки | `result` = `ok` \| `error` | опубликованные события |
| `tripgo.outbox.age` | histogram | секунды | — | возраст самого старого неопубликованного события; растёт — publisher не справляется |

Вариант B, кеш:

| Метрика | Тип | Единица | Лейблы | Что показывает |
|---|---|---|---|---|
| `tripgo.cache.requests` | counter | штуки | `result` = `hit` \| `miss` | попадания и промахи; из них считают hit rate |
| `tripgo.cache.size` | gauge | штуки | — | сколько записей в кеше |
| `tripgo.cache.evictions` | counter | штуки | `reason` = `ttl` \| `size` | что вытеснили и почему |

Вариант C, WebSocket:

| Метрика | Тип | Единица | Лейблы | Что показывает |
|---|---|---|---|---|
| `tripgo.ws.connections` | gauge | штуки | — | сколько соединений открыто сейчас |
| `tripgo.ws.messages.sent` | counter | штуки | `result` = `ok` \| `dropped` | отправленные сообщения |
| `tripgo.ws.slow_clients` | counter | штуки | — | сколько соединений закрыли из-за переполнения буфера |

## 6. Трейсы

| Работа | Что обязано попадать в трейс |
|---|---|
| Лабораторная работа 2 | входящий HTTP-запрос (корневой спан), каждый запрос в БД |
| Лабораторная работа 3 | вызов Push Service; спан обработки задачи воркером связан с породившей проверкой через span link |
| Лабораторная работа 4 | входящий gRPC, исходящий gRPC, publish и consume Kafka |

Контекст пробрасываем через `W3C TraceContext`: в HTTP заголовками, в gRPC
метаданными, в Kafka заголовками сообщения. Ручной проброс `trace_id` строкой в
теле сообщения не засчитывается.

Результат, который ждём в лабораторной работе 4: в Jaeger виден один трейс, где
есть и публикация события, и его обработка консьюмером.

## 7. Логи

JSON, stdout, `log/slog`. Обязательные поля:

```json
{
  "time": "2026-08-24T12:00:00.123Z",
  "level": "INFO",
  "msg": "trip completed",
  "service": "trip-service",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "trip_id": "5cb72c04-..."
}
```

Обязательные события уровня `Info`: создание поездки, завершение поездки,
старт и остановка сервиса и каждого фонового компонента.

Обязательные события уровня `Warn`: ретрай исходящего вызова, срабатывание
rate limiter, дроп задачи из-за переполненной очереди, недоступность Push
Service.

Обязательные события уровня `Error`: невозможность сохранить данные, отправка
сообщения в DLQ, паника (восстановленная middleware).
