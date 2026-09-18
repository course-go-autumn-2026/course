-- +goose Up
CREATE INDEX idx_film_title ON films (title);

-- +goose Down
DROP INDEX idx_film_title;
