-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_film_release_date ON films(release_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_film_release_date;
-- +goose StatementEnd
