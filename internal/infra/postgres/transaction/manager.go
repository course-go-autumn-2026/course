package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type (
	Manager    struct{ pool *pgxpool.Pool }
	contextKey struct{}
)

func New(pool *pgxpool.Pool) *Manager { return &Manager{pool: pool} }

func (manager *Manager) Do(ctx context.Context, operation func(context.Context) error) (err error) {
	if _, ok := ctx.Value(contextKey{}).(pgx.Tx); ok {
		return operation(ctx)
	}
	tx, err := manager.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			panic(recovered)
		}
		if err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			return
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit transaction: %w", commitErr)
		}
	}()
	return operation(context.WithValue(ctx, contextKey{}, tx))
}

func FromContext(ctx context.Context, pool *pgxpool.Pool) Executor {
	if tx, ok := ctx.Value(contextKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}
