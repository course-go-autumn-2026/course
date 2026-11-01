package transaction

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestManagerTransactionLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.MaxConns = 1 // A nested Do would block if it tried to open another transaction.
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	table := pgx.Identifier{"tx_manager_test_" + strconv.FormatInt(time.Now().UnixNano(), 10)}.Sanitize()
	if _, err := pool.Exec(ctx, "CREATE TABLE "+table+" (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, "DROP TABLE "+table); err != nil {
			t.Errorf("drop test table: %v", err)
		}
	}()

	manager := NewTxManager(pool, 3*time.Second)
	insert := "INSERT INTO " + table + " (id) VALUES ($1)"
	count := func() int {
		var n int
		if err := manager.Executor(ctx).QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	if _, err := manager.Executor(ctx).Exec(ctx, insert, 1); err != nil {
		t.Fatal(err)
	}
	if err := manager.Do(ctx, func(txCtx context.Context) error {
		var isolation string
		if err := manager.Executor(txCtx).QueryRow(txCtx, "SHOW transaction_isolation").Scan(&isolation); err != nil {
			return err
		}
		if isolation != "read committed" {
			return fmt.Errorf("transaction isolation: got %q, want read committed", isolation)
		}
		if _, err := manager.Executor(txCtx).Exec(txCtx, insert, 2); err != nil {
			return err
		}
		return manager.Do(txCtx, func(nestedCtx context.Context) error {
			_, err := manager.Executor(nestedCtx).Exec(nestedCtx, insert, 3)
			return err
		})
	}); err != nil {
		t.Fatal(err)
	}
	if got := count(); got != 3 {
		t.Fatalf("after commit: got %d rows, want 3", got)
	}

	wantErr := errors.New("abort")
	if err := manager.Do(ctx, func(txCtx context.Context) error {
		if _, err := manager.Executor(txCtx).Exec(txCtx, insert, 4); err != nil {
			return err
		}
		return wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("callback error: got %v, want %v", err, wantErr)
	}
	if got := count(); got != 3 {
		t.Fatalf("after error rollback: got %d rows, want 3", got)
	}

	panicked := false
	func() {
		defer func() { panicked = recover() == "abort by panic" }()
		_ = manager.Do(ctx, func(txCtx context.Context) error {
			if _, err := manager.Executor(txCtx).Exec(txCtx, insert, 5); err != nil {
				return fmt.Errorf("insert before panic: %w", err)
			}
			panic("abort by panic")
		})
	}()
	if !panicked {
		t.Fatal("panic was not propagated")
	}
	if got := count(); got != 3 {
		t.Fatalf("after panic rollback: got %d rows, want 3", got)
	}
}
