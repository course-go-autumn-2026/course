-- +goose Up
WITH demo_director AS (
    INSERT INTO directors (last_name, first_name)
    VALUES ('Index demo', 'Director')
    RETURNING id
)
INSERT INTO films (title, release_date, director_id, uuid, rating)
SELECT
    'Index demo film ' || n,
    DATE '1980-01-01' + (n % 16000),
    demo_director.id,
    -- Отдельный диапазон UUID позволяет откатить только демонстрационные записи.
    ('f11d0000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
    (n % 101) / 10.0
FROM generate_series(1, 1000000) AS series(n)
CROSS JOIN demo_director;

-- +goose Down
WITH deleted_films AS (
    DELETE FROM films
    WHERE uuid IN (
        SELECT ('f11d0000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid
        FROM generate_series(1, 1000000) AS series(n)
    )
    RETURNING director_id
)
DELETE FROM directors
WHERE id IN (SELECT director_id FROM deleted_films)
  AND last_name = 'Index demo'
  AND first_name = 'Director';
