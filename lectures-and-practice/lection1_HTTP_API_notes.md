# HTTP API на Go

Конспект первой лекции модуля 1. Здесь все, что было на паре, плюс код демо — можно повторить у себя. Вопросы — в чат курса.

## 1. Как устроен Go-проект

Официального стандарта раскладки нет, но есть конвенция, которую вы встретите почти в любом сервисе — `golang-standards/project-layout`:

```text
my-service/
├── cmd/
│   └── server/
│       └── main.go         # точка входа, только сборка и запуск
├── internal/
│   ├── handler/            # HTTP-слой: разбор запроса, сборка ответа
│   ├── service/            # бизнес-логика
│   └── repository/         # работа с хранилищем
├── api/
│   └── openapi.yaml        # контракт API
├── go.mod
└── Makefile
```

### Зачем так

- `cmd/` — точки входа. Их может быть несколько: сервер, воркер, миграции.
- `internal/` — код, который нельзя импортировать снаружи модуля. Это не соглашение, а правило компилятора: пакет из чужого `internal/` просто не соберется.
- `pkg/` иногда используют для кода, который наоборот можно переиспользовать. Спорная папка, многие живут без нее.

В `main.go` не пишут логику — там только чтение конфига, сборка зависимостей и запуск.

Разделение `handler/service/repository` подробно разберем в лекции про архитектуру (модуль 2), пока договоримся: хендлер не ходит в базу напрямую.

## 2. net/http: поднимаем сервер

### Handler — это интерфейс из одного метода

Весь HTTP-стек Go стоит на одном интерфейсе:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Помните duck typing из модуля 0? Любой тип с таким методом — уже хендлер. А чтобы не заводить тип ради каждой ручки, есть адаптер `http.HandlerFunc` — он превращает обычную функцию в `Handler`.

### Минимальный сервер

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
)

type PingResponse struct {
    Status string `json:"status"`
}

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(PingResponse{Status: "ok"})
    })

    mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        w.Write([]byte("item " + id))
    })

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Это рабочий сервер. Обратите внимание на паттерны роутов: с Go 1.22 стандартный `ServeMux` умеет метод в паттерне (`GET /ping`) и path-параметры (`{id}` + `r.PathValue`). До 1.22 ничего этого не было, и роутеры брали в первую очередь ради этого.

### Запрос и ответ

- Тело запроса — `r.Body`, это `io.ReadCloser`. Для JSON: `json.NewDecoder(r.Body).Decode(&req)`.
- Статус ставится через `w.WriteHeader(http.StatusCreated)` — строго до записи тела. Если не позвать, при первой записи тела уйдет 200.
- Заголовки — `w.Header().Set(...)`, тоже до записи тела.

Типичный порядок в хендлере: заголовки, статус, тело. Перепутаете — получите предупреждение `superfluous response.WriteHeader call` в логах или молча уехавший 200.

### Graceful shutdown

`http.ListenAndServe` нельзя остановить корректно — при выкатке новой версии сервис просто оборвет живые запросы. Поэтому в проде поднимают `http.Server` и гасят его через `Shutdown`:

```go
func main() {
    ctx, stop := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT,
        syscall.SIGTERM,
    )
    defer stop()

    srv := &http.Server{Addr: ":8080", Handler: mux}

    go func() {
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Fatal(err)
        }
    }()

    <-ctx.Done() // ждем SIGTERM

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    srv.Shutdown(shutdownCtx) // перестает принимать новые, дожидается текущих
}
```

`Shutdown` перестает принимать новые соединения и ждет, пока дообработаются текущие запросы (не дольше таймаута). Kubernetes при выкатке шлет `SIGTERM` и ждет — ровно под этот сценарий.

## 3. Роутинг: chi и альтернативы

### Чего не хватает стандартному mux

После 1.22 стандартный mux закрывает базу. Не хватает трех вещей: цепочек middleware, групп роутов с общими префиксами и вложенных роутеров. Как только ручек становится больше пяти, это начинает болеть.

### chi

`go-chi/chi` — самый идиоматичный роутер: он не изобретает свой фреймворк, а остается обычным `http.Handler`.

```go
r := chi.NewRouter()

r.Use(middleware.Logger) // на все роуты
r.Use(middleware.Recoverer)

r.Route("/api/v1", func(r chi.Router) {
    r.Get("/items", listItems)
    r.Post("/items", createItem)

    r.Route("/items/{id}", func(r chi.Router) {
        r.Get("/", getItem) // GET /api/v1/items/{id}
        r.Delete("/", deleteItem)
    })
})

http.ListenAndServe(":8080", r)
```

Path-параметр достается через `chi.URLParam(r, "id")`.

### Что у chi внутри

Две вещи, ради которых мы вообще заглядываем в чужой код:

1. Роуты хранятся в radix tree — префиксном дереве. Поиск хендлера по пути не перебирает все роуты подряд, а спускается по дереву посимвольно. Поэтому скорость роутинга почти не зависит от количества ручек.
2. На каждый запрос chi нужен объект `RouteContext` — туда складываются распарсенные path-параметры. Вместо того чтобы аллоцировать его каждый раз, chi держит их в `sync.Pool` — помните его из модуля 0? Взяли из пула, обнулили, положили обратно. На нагруженном сервисе это заметно снимает давление на GC.

Живой пример того, зачем пул вообще нужен: `chi/mux.go`.

### Альтернативы

| Роутер | Подход | Цена |
|---|---|---|
| chi | обычный `http.Handler`, ничего своего | — |
| gorilla/mux | тоже `http.Handler`, старейший | медленнее chi, был в архиве, сейчас ожил |
| gin | свой `gin.Context` вместо `(w, r)` | быстрый, но хендлеры несовместимы со стандартной библиотекой |
| echo | свой `echo.Context` | то же самое |

Главный trade-off: gin и echo дают удобный контекст с хелперами, но ваш код прирастает к фреймворку. Хендлер на chi можно повесить куда угодно, где ждут `http.Handler` — в тесты, в другой роутер, за любой middleware. В Авито поэтому в основном chi и стандартная библиотека.

## 4. Middleware

Middleware — функция, которая оборачивает хендлер и возвращает новый хендлер:

```go
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s took %s", r.Method, r.URL.Path, time.Since(start))
    })
}
```

До `next.ServeHTTP` — код «на входе», после — «на выходе». Цепочка работает как луковица: первый middleware в списке видит запрос первым, а ответ — последним.

### Ловушка: как узнать статус ответа

В `http.ResponseWriter` нет метода «прочитать статус» — он write-only. Логирующему middleware приходится подсовывать хендлеру обертку:

```go
type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (rec *statusRecorder) WriteHeader(code int) {
    rec.status = code
    rec.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rec := &statusRecorder{
            ResponseWriter: w,
            status:         http.StatusOK,
        }

        next.ServeHTTP(rec, r)

        log.Printf("%s %s -> %d", r.Method, r.URL.Path, rec.status)
    })
}
```

Здесь работает встраивание структур из модуля 0: `statusRecorder` наследует все методы `ResponseWriter`, а `WriteHeader` переопределяет и запоминает статус. Дефолт 200 нужен потому, что хендлер может вообще не позвать `WriteHeader`.

Это ровно та задача, что будет на практике.

### Что обычно живет в middleware

Recovery (ловит панику хендлера и возвращает 500 вместо падения всего сервера), таймауты, аутентификация, rate limiting, метрики и трейсинг. В сервисах Авито цепочка стандартная: recovery, логи, метрики, трейсинг, auth — про метрики и трейсинг будет отдельная лекция Observability.

## 5. REST как подход

REST — соглашение о том, как HTTP-методы и пути отражают операции над ресурсами:

| Операция | Метод и путь | Успех | Идемпотентна |
|---|---|---:|---|
| список | `GET /items` | 200 | да |
| один | `GET /items/{id}` | 200 | да |
| создать | `POST /items` | 201 | нет |
| заменить | `PUT /items/{id}` | 200 | да |
| изменить часть | `PATCH /items/{id}` | 200 | нет |
| удалить | `DELETE /items/{id}` | 204 | да |

Идемпотентность — повтор запроса не меняет результат. Это не теория: ретраи и балансировщики повторяют запросы, и неидемпотентный POST при ретрае создаст дубль. Подробнее столкнемся в модуле 3.

Статусы, которых хватает в 95% случаев: `200/201/204`, `400` (невалидный запрос), `401/403` (кто ты / нельзя), `404`, `409` (конфликт), `500`. Не изобретайте экзотику — клиенту важнее предсказуемость.

Ошибки отдавайте в едином формате, всегда одном на весь сервис:

```json
{
  "error": {
    "code": "item_not_found",
    "message": "item 42 not found"
  }
}
```

Валидация входа живет на границе — в хендлере, до бизнес-логики. Невалидный JSON — 400 сразу.

Версионирование чаще всего делают префиксом пути `/api/v1/...` — грубо, зато видно в логах и легко роутится. Вариант с заголовком аккуратнее, но менее нагляден. Правило одно: ломающее изменение — новая версия, а не молчаливая правка старой.

Ловушка сериализации: `omitempty` прячет нулевые значения. Если поле `Count int` с тегом `json:",omitempty"` равно 0 — его не будет в ответе вообще. Когда ноль — легитимное значение, берите указатель `*int` или уберите `omitempty`.

## 6. API-first: OpenAPI и генерация кода

### Идея

Код-first: написали хендлеры, потом задокументировали (или нет). API-first наоборот: сначала описываем контракт в спеке, ревьюим его как код, потом генерируем типы и каркас сервера.

Что это дает:

- фронт и бэк работают параллельно — фронт мокает по спеке, не дожидаясь сервера;
- ревью контракта до написания кода: поменять поле в YAML дешевле, чем переписать хендлеры;
- типы запросов и ответов не разъезжаются с докой — они из нее сгенерированы.

В Авито так и живут: внутренний формат контрактов называется brief, из него генерируются клиенты и серверные заготовки. Про межсервисное общение будет модуль 3.

### Спека

```yaml
openapi: 3.0.3

info:
  title: Items API
  version: 1.0.0

paths:
  /items/{id}:
    get:
      operationId: getItem
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: item
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Item"
        "404":
          description: not found

components:
  schemas:
    Item:
      type: object
      required: [id, title]
      properties:
        id:
          type: integer
        title:
          type: string
```

Три кита спеки: `paths` (ручки), `components/schemas` (модели), `$ref` (переиспользование схем).

### Генерация: oapi-codegen

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

oapi-codegen -generate types,chi-server \
  -package api \
  api/openapi.yaml > internal/api/gen.go
```

Из спеки получаем структуру `Item` и интерфейс сервера:

```go
type ServerInterface interface {
    GetItem(w http.ResponseWriter, r *http.Request, id int)
}
```

Дальше пишем свой тип, реализующий этот интерфейс (снова duck typing), и регистрируем: `api.HandlerFromMux(myServer, r)`. Роутинг, разбор path-параметров и типы — сгенерированы, руками пишется только логика.

Альтернатива — ogen: генерирует больше (включая валидацию по схеме), но и магии больше. Начинать проще с `oapi-codegen`.

Команду генерации кладут в `Makefile` или `//go:generate` — сгенерированный код не правят руками, он перезаписывается.

## 7. Чем проверять API

`curl` — минимальный набор флагов:

```bash
curl -i http://localhost:8080/ping
# -i показать статус и заголовки

curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"title":"lamp"}' \
  http://localhost:8080/items

curl -v http://localhost:8080/ping
# -v весь диалог, для отладки
```

Postman или Insomnia — когда запросов много: коллекции можно сохранить в репозиторий и передать команде. А если есть OpenAPI-спека, то Swagger UI отдаст интерактивную документацию бесплатно — импортируйте YAML в `editor.swagger.io` и потыкайте.

## 8. Что дальше

На следующей паре — практика, три чекпоинта:

1. Поднять сервер и написать свой роутер.
2. Описать спеку в OpenAPI, сгенерировать код, добавить ручки со статикой.
3. Написать логирующий middleware (метод, URI, статус — вспоминайте `statusRecorder`).

Репозиторий-заготовка с тестами будет в гите курса. Тесты запускаются локально через `go test` — как на Exercism.

### Почитать

- `net/http` docs — да, стандартная дока читается.
- Routing Enhancements for Go 1.22 — блог Go про новый ServeMux.
- chi — README и исходники `mux.go`.
- `oapi-codegen`.
