package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionKey struct{}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionManager struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

var _ TxManager = (*TransactionManager)(nil)

func NewTxManager(pool *pgxpool.Pool, timeout time.Duration) *TransactionManager {
	return &TransactionManager{pool: pool, timeout: timeout}
}

func (m *TransactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
	if transactionFromContext(ctx) != nil {
		return fn(ctx)
	}

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Rollback is also called after commit; pgx safely ignores a closed transaction.
	// A separate timeout allows cleanup even when the request was cancelled.
	defer m.rollback(ctx, tx)

	txCtx := context.WithValue(ctx, transactionKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (m *TransactionManager) rollback(ctx context.Context, tx pgx.Tx) {
	rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), m.timeout)
	defer cancel()
	_ = tx.Rollback(rollbackCtx)
}

func transactionFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(transactionKey{}).(pgx.Tx)
	return tx
}
