-- +goose Up
CREATE TABLE trip_idempotency_keys (
    key           UUID PRIMARY KEY,
    request_hash  BYTEA NOT NULL CHECK (octet_length(request_hash) = 32),
    trip_id       UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    trip_snapshot JSONB NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX trip_idempotency_keys_expires_at_idx ON trip_idempotency_keys (expires_at);

-- +goose Down
DROP TABLE trip_idempotency_keys;
