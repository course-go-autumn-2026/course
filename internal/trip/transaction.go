package trip

import "context"

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
