package task03

import (
	"context"
	"time"
)

// Wait имитирует дорогой расчёт. Именно эту работу нужно остановить при отмене.
func Wait(ctx context.Context, d time.Duration) error {
	// TODO: замените Sleep на таймер и select по ctx.Done(). Освободите таймер.
	time.Sleep(d)
	return nil
}
