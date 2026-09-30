-- +goose Up

CREATE TABLE trip_idempotency_keys (
    idempotency_key UUID PRIMARY KEY,
    request_hash TEXT NOT NULL,
    trip_id UUID REFERENCES trips(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '24 hours'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX trip_idempotency_keys_expires_idx
    ON trip_idempotency_keys (expires_at);

-- +goose Down

DROP TABLE trip_idempotency_keys;