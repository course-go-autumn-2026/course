package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type TransactionManager struct {
	pool *pgxpool.Pool
}

type transactionContextKey struct{}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{
		pool: pool,
	}
}

func (manager *TransactionManager) Do(
	ctx context.Context,
	fn func(ctx context.Context) error,
) (err error) {
	if _, ok := TransactionFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := manager.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if panicValue := recover(); panicValue != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			panic(panicValue)
		}

		if err != nil {
			_ = tx.Rollback(context.WithoutCancel(ctx))
			return
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("commit transaction: %w", commitErr)
		}
	}()

	txCtx := context.WithValue(ctx, transactionContextKey{}, tx)
	return fn(txCtx)
}

func TransactionFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(transactionContextKey{}).(pgx.Tx)
	return tx, ok
}

func ExecutorFromContext(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx, ok := TransactionFromContext(ctx); ok {
		return tx
	}

	return pool
}
