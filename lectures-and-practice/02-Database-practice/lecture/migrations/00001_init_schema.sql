-- +goose Up
-- +goose StatementBegin
-- Таблица режиссёров
CREATE TABLE directors (
    id SERIAL PRIMARY KEY,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL
);

-- Таблица фильмов
CREATE TABLE films (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    release_date DATE NOT NULL,
    director_id INTEGER NOT NULL REFERENCES directors(id) ON DELETE RESTRICT,
    uuid UUID UNIQUE NOT NULL,
    rating DECIMAL(3,1) CHECK (rating >= 0 AND rating <= 10),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE films;
DROP TABLE directors;
-- +goose StatementEnd
