# Схема данных

Здесь описаны таблицы курса: что хранят, из каких полей состоят и какие правила
закреплены в самой схеме. Таблицы появляются по мере работ, в каждом разделе
указано, в какой.

Готовых миграций мы не даём — их вы пишете сами. `CREATE TABLE` ниже приведены,
чтобы не гадать про типы и ограничения: перенесите их в свои миграции, добавьте
`down` и, если считаете нужным, поправьте под своё решение. Имена таблиц и
колонок менять нельзя, на них завязаны следующие работы.

---

## `trips` — поездка

Появляется в работе 1. Колонка `last_position_at` добавляется в работе 3.

```sql
CREATE TABLE trips (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL,
    driver_id       UUID NOT NULL,

    start_latitude  DOUBLE PRECISION NOT NULL CHECK (start_latitude BETWEEN -90 AND 90),
    start_longitude DOUBLE PRECISION NOT NULL CHECK (start_longitude BETWEEN -180 AND 180),
    end_latitude    DOUBLE PRECISION NOT NULL CHECK (end_latitude BETWEEN -90 AND 90),
    end_longitude   DOUBLE PRECISION NOT NULL CHECK (end_longitude BETWEEN -180 AND 180),

    price           BIGINT NOT NULL CHECK (price >= 0),
    status          TEXT NOT NULL CHECK (status IN ('active', 'completed')),

    started_at      TIMESTAMPTZ NOT NULL,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (
        (status = 'active'    AND finished_at IS NULL) OR
        (status = 'completed' AND finished_at IS NOT NULL)
    )
);

CREATE INDEX trips_status_started_at_idx ON trips (status, started_at);
```

| Поле | Что хранит | Кто заполняет |
|---|---|---|
| `id` | идентификатор поездки | сервис при создании, из запроса не принимается |
| `user_id` | пассажир | приходит в запросе |
| `driver_id` | водитель | приходит в запросе |
| `start_*`, `end_*` | точки подачи и назначения, градусы | приходит в запросе |
| `price` | стоимость в целых рублях | приходит в запросе, см. упрощение в `domain.md` |
| `status` | `active` или `completed` | сервис |
| `started_at` | когда поездка создана | сервис, не клиент |
| `finished_at` | когда завершена, пусто пока активна | сервис при завершении |
| `last_position_at` | время последней координаты | сервис при сохранении координаты, работа 3 |
| `created_at`, `updated_at` | техполя | сервис |

Почему `price` целым числом: деньги во `float` не хранят, потеряете копейки на
округлении. У нас это ещё и целые рубли — учебное упрощение, в проде хранили бы
копейки.

Последний `CHECK` держит связку статуса и времени завершения: завершённая
поездка без `finished_at` или активная с ним в базу не попадут.

**Ограничение, которое надо добавить самим:** у одного водителя не может быть
двух активных поездок одновременно. Оно должно держаться при одновременных
запросах — два параллельных создания на одного водителя дают одну поездку и одну
ошибку. Проверка `SELECT` перед `INSERT` этого не даёт: оба запроса пройдут
проверку раньше, чем любой из них вставит строку.

В работе 3 к таблице добавляется колонка:

```sql
ALTER TABLE trips ADD COLUMN last_position_at TIMESTAMPTZ;
```

---

## `trip_status_history` — журнал переходов статуса

Появляется в работе 1. Нужен, чтобы создание поездки писало в две таблицы в одной
транзакции, — на этом отрабатывается менеджер транзакций.

```sql
CREATE TABLE trip_status_history (
    id          BIGSERIAL PRIMARY KEY,
    trip_id     UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status   TEXT NOT NULL CHECK (to_status IN ('active', 'completed')),
    reason      TEXT,
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX trip_status_history_trip_changed_idx
    ON trip_status_history (trip_id, changed_at);
```

`from_status` пуст при создании поездки: прежнего статуса нет. `reason` — текст
для человека, чем вызван переход.

---

## `trip_positions` — точки маршрута

Появляется в работе 3.

```sql
CREATE TABLE trip_positions (
    id          BIGSERIAL PRIMARY KEY,
    trip_id     UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    latitude    DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude   DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX trip_positions_trip_recorded_idx
    ON trip_positions (trip_id, recorded_at, id);
```

`recorded_at` — когда координата зафиксирована на устройстве, `created_at` —
когда её принял сервис. Это разные моменты: устройство копит точки в офлайне и
шлёт их пачкой позже. Маршрут строим по `recorded_at`, поэтому индекс по нему.

Индекс под выборку поездок для опроса вы добавляете сами — это отдельный пункт
чек-листа работы 3.

Таблицы для работ 4 и 5 — `processed_commands` и `outbox_events` — появятся
здесь вместе с самими работами.
