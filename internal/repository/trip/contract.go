package trip

import (
	"context"

	"github.com/DaleCoopTP/grishin_lab1/internal/infra/postgres/transaction"
)

type executorProvider interface {
	Executor(ctx context.Context) transaction.Executor
}
