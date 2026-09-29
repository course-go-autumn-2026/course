-- +goose Up
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
        (status = 'active' AND finished_at IS NULL) OR
        (status = 'completed' AND finished_at IS NOT NULL)
    )
);

CREATE INDEX trips_status_started_at_idx ON trips (status, started_at);

CREATE UNIQUE INDEX trips_driver_active_uidx
    ON trips (driver_id)
    WHERE status = 'active';

-- +goose Down
DROP TABLE trips;
