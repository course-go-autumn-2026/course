package processing

import (
	"context"
	"errors"
	"time"
)

var ErrPermanent = errors.New("permanent processing error")

func Wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return ctx.Err()
	}
}
