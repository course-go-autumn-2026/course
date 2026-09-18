-- +goose Up
CREATE TABLE screenings(
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    film_title text NOT NULL,
    starts_at timestamptz NOT NULL,
    available_seats integer NOT NULL CHECK (available_seats >= 0)
);

INSERT INTO screenings (film_title, starts_at, available_seats) VALUES
    ('Интерстеллар', '2026-09-10 19:00:00+00', 1),
    ('Начало', '2026-09-10 21:30:00+00', 3),
    ('Криминальное чтиво', '2026-09-11 17:00:00+00', 5),
    ('Бегущий по лезвию 2049', '2026-09-11 20:00:00+00', 0);

-- +goose Down
DROP TABLE screenings;

