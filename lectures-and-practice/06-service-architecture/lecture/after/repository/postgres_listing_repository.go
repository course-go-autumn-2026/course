// Package repository - адаптеры хранилища. Здесь и только здесь
// появляется SQL. Если завтра PostgreSQL сменится на другую БД,
// меняется этот пакет - domain и service не почувствуют разницы,
// потому что зависят от интерфейса service.ListingRepository.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
)

type PostgresListingRepository struct {
	db *sql.DB
}

func NewPostgresListingRepository(db *sql.DB) *PostgresListingRepository {
	return &PostgresListingRepository{db: db}
}

func (r *PostgresListingRepository) Create(ctx context.Context, l *domain.Listing) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO listings (title, price, photo_url, seller_id, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		l.Title, l.Price, l.PhotoURL, l.SellerID, l.Status, l.CreatedAt,
	).Scan(&id)
	return id, err
}

func (r *PostgresListingRepository) UpdateStatus(ctx context.Context, id int64, status domain.Status) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE listings SET status = $1 WHERE id = $2`, status, id)
	return err
}

// Get - то самое место, где sql.ErrNoRows превращается в доменную
// ошибку. Сравните с before/handler.go: там errors.Is(err, sql.ErrNoRows)
// проверялся прямо в хендлере вперемешку с HTTP-кодом ответа.
func (r *PostgresListingRepository) Get(ctx context.Context, id int64) (*domain.Listing, error) {
	var l domain.Listing
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, price, photo_url, seller_id, status, created_at
		 FROM listings WHERE id = $1`, id,
	).Scan(&l.ID, &l.Title, &l.Price, &l.PhotoURL, &l.SellerID, &l.Status, &l.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}
