# Кинотеатр: PostgreSQL из Go

Стартовый репозиторий для [практики](./Практика.md). Вы добавите таблицу
бронирований, индекс, динамический SQL-запрос и транзакционное бронирование.

## Что нужно

- Go 1.23 или новее;
- Docker с плагином Compose;
- make;
- curl;
- goose v3.

Установить goose можно командой:

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
```

Проверьте доступность утилит:

```bash
goose -version
curl --version
```

Если `goose` не найден, добавьте в `PATH` каталог `$(go env GOPATH)/bin`
или каталог `GOBIN`, если он задан при установке.

## Быстрый старт

```bash
docker compose up -d --wait

make migrate-up
make run
```

Сервер слушает `http://localhost:8080`. Адрес можно изменить переменной `HTTP_ADDR`.
После изменения Go-кода остановите сервер сочетанием `Ctrl+C` и снова выполните
`make run`: автоматической перезагрузки нет.

## Структура

```text
cmd/app/                    точка входа HTTP-сервера и demo-booking
internal/cinema/            handler, service, repository и transaction manager
migrations/00001_...sql     готовая схема и данные screenings
Практика.md                 задания и команды проверки
```

В стартовом коде оставлены два `TODO` для заданий 3 и 4. Миграции из заданий 1
и 2 вы создадите сами. Тесты стартового проекта проверяют разбор фильтров;
успешный `go test ./...` не заменяет проверки SQL и бронирования из практики.

## Миграции

`Makefile` задаёт `DATABASE_URL` для учебной БД из `compose.yaml`:
`postgres://app:app@localhost:5432/cinema?sslmode=disable`.
Настраивать переменную окружения не нужно — она доступна и миграциям, и `make run`.

```bash
make migrate-up       # применить все новые миграции
make migrate-down     # откатить последнюю миграцию
make migrate-status   # показать состояние миграций
make bookings-schema  # показать структуру bookings в БД из Docker Compose
```

### Другая база данных

Для другой БД задайте адрес в каждом терминале, где запускаете миграции или приложение:

```bash
export DATABASE_URL='postgres://user:password@localhost:5432/database?sslmode=disable'
```

Команды миграций `make migrate-up`, `make migrate-down`, `make migrate-status` и
`make run` используют этот адрес. Если вместо `export` передать
`DATABASE_URL='...'` аргументом одной команды `make`, настройка подействует
только на этот вызов.

Команды `docker compose exec ... psql` и `make bookings-schema` не используют
`DATABASE_URL`: они обращаются к БД `cinema` с пользователем `app` в контейнере.
Для другой БД установите локальный psql по инструкции ниже и используйте:

```bash
psql "$DATABASE_URL"
psql "$DATABASE_URL" -c '\d bookings'
```

Первая команда заменяет подключение через `docker compose exec ... psql`, вторая
заменяет `make bookings-schema`. SQL-блоки из практики выполняйте в этом подключении;
для одиночных запросов используйте `psql "$DATABASE_URL" -c 'SQL-запрос'`.

## Полезные команды

```bash
go test ./...
go vet ./...
gofmt -w ./cmd ./internal
docker compose down
```

## Шпаргалка по psql

`psql` — консольный клиент PostgreSQL. Сначала запустите БД и примените миграции
из раздела «Быстрый старт». Команды терминала выполняйте из каталога проекта.

### Установка psql (необязательно)

При подключении через контейнер устанавливать `psql` на компьютер не нужно —
он уже есть внутри. Для подключения с компьютера установите клиент своей ОС.
Локальный сервер PostgreSQL запускать не нужно: база работает в Docker.

**Ubuntu / Debian:**

```bash
sudo apt update
sudo apt install postgresql-client
```

**macOS** (нужен [Homebrew](https://brew.sh/)):

```bash
brew install libpq
export PATH="$(brew --prefix libpq)/bin:$PATH"
```
Проверка установки:

```bash
psql --version
```

### Подключение

Если `psql` установлен на компьютере:

```bash
psql -h localhost -p 5432 -U app -d cinema
```

`-h` — адрес сервера, `-p` — порт, `-U` — пользователь, `-d` — база данных.
Пароль для учебной БД — `app`; при вводе символы не отображаются.

Без установки `psql` на компьютер — через контейнер:

```bash
docker compose exec postgres psql -U app -d cinema
```

После подключения появится приглашение `cinema=#`. Дальше вводите команды в нём,
а не в обычном терминале.

### Полезные команды внутри psql

| Команда | Что делает |
| --- | --- |
| `\conninfo` | Показывает текущее подключение |
| `\l` | Список баз данных |
| `\c cinema` | Переключает на базу `cinema` |
| `\dt` | Список таблиц |
| `\d screenings` | Структура таблицы: столбцы, типы, индексы и ограничения |
| `\x` | Включает/выключает вертикальный вывод — удобно для широких строк |
| `\?` | Справка по командам psql |
| `\h SELECT` | Справка по SQL-команде `SELECT` |
| `\q` | Выход в терминал |

Пример SQL-запроса:

```sql
SELECT * FROM screenings LIMIT 5;
```

- SQL-запрос заканчивается `;`, команды с `\` — без точки с запятой.
- Приглашение `cinema-#` означает, что ввод запроса продолжается: допишите `;`
  и нажмите Enter. `Ctrl+C` сбрасывает незавершённый ввод или отменяет текущий запрос.
- Если длинный вывод открылся в просмотрщике, нажмите `q`, чтобы вернуться к вводу.

### Один запрос без интерактивного режима

В обычном терминале:

```bash
docker compose exec postgres psql -U app -d cinema -c 'SELECT * FROM screenings LIMIT 5;'
```
