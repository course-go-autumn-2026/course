# EXPLAIN ANALYZE: сравнение запроса до и после индекса

```sql
ANALYZE films;

EXPLAIN (ANALYZE, BUFFERS)
SELECT id, title, release_date, director_id, uuid, rating
FROM films
WHERE title = 'Index demo film 500000';
```
