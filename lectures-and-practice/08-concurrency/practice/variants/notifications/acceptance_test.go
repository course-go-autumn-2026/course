package notifications

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testTimeout = 3 * time.Second

func batchResult(t *testing.T, orders []Notification, bake Sender) []DeliveryResult {
	t.Helper()
	done := make(chan []DeliveryResult, 1)
	go func() { done <- SendAll(orders, bake) }()
	select {
	case result := <-done:
		return result
	case <-time.After(testTimeout):
		t.Fatal("SendAll не завершилась: проверьте WaitGroup и сбор результатов")
		return nil
	}
}

func TestAcceptanceBatchEmpty(t *testing.T) {
	for _, orders := range [][]Notification{nil, {}} {
		var calls atomic.Int32
		got := batchResult(t, orders, func(Notification) DeliveryResult {
			calls.Add(1)
			return DeliveryResult{}
		})
		if len(got) != 0 || calls.Load() != 0 {
			t.Fatalf("пустая партия: result=%v, calls=%d", got, calls.Load())
		}
	}
}

func TestAcceptanceBatchValues(t *testing.T) {
	orders := []Notification{{ID: 42, Channel: "Маргарита", OrderID: 450}}
	before := append([]Notification(nil), orders...)
	want := DeliveryResult{MessageID: 42, Status: "отправлено: Маргарита"}
	var calls atomic.Int32
	var received Notification
	got := batchResult(t, orders, func(order Notification) DeliveryResult {
		calls.Add(1)
		received = order
		return want
	})
	if calls.Load() != 1 || received != before[0] || !reflect.DeepEqual(orders, before) || !reflect.DeepEqual(got, []DeliveryResult{want}) {
		t.Fatalf("потерян заказ, результат или изменён вход: calls=%d, received=%+v, orders=%v, result=%v", calls.Load(), received, orders, got)
	}
}

func TestAcceptanceBatchConcurrent(t *testing.T) {
	// ID намеренно совпадают: восстановление порядка по ID неверно.
	orders := []Notification{{ID: 7, Channel: "A", OrderID: 450}, {ID: 7, Channel: "B", OrderID: 550}, {ID: 7, Channel: "C", OrderID: 650}}
	before := append([]Notification(nil), orders...)
	started := make(chan int, len(orders))
	finished := make(chan int, len(orders))
	release := []chan struct{}{make(chan struct{}), make(chan struct{}), make(chan struct{})}
	opened := make([]bool, len(orders))
	defer func() {
		for i, ch := range release {
			if !opened[i] {
				close(ch)
			}
		}
	}()
	var calls [3]atomic.Int32
	done := make(chan []DeliveryResult, 1)
	go func() {
		done <- SendAll(orders, func(o Notification) DeliveryResult {
			i := int(o.Channel[0] - 'A')
			calls[i].Add(1)
			started <- i
			<-release[i]
			finished <- i
			return DeliveryResult{MessageID: o.ID, Status: "отправлено: " + o.Channel}
		})
	}()
	seen := make(map[int]bool)
	deadline := time.NewTimer(testTimeout)
	defer deadline.Stop()
	for len(seen) < len(orders) {
		select {
		case i := <-started:
			if seen[i] {
				t.Fatal("Sender вызван повторно для одного заказа")
			}
			seen[i] = true
		case <-deadline.C:
			t.Fatal("все пиццы должны начать готовиться до завершения первой; последовательный цикл не подходит")
		}
	}
	select {
	case <-done:
		t.Fatal("SendAll вернулась до готовности пицц")
	default:
	}
	// Управляем порядком готовности каналами, а не сравнением времени исполнения.
	for i := len(orders) - 1; i >= 0; i-- {
		close(release[i])
		opened[i] = true
		select {
		case index := <-finished:
			if index != i {
				t.Fatalf("неожиданный готовый заказ: %d, want %d", index, i)
			}
		case <-time.After(testTimeout):
			t.Fatal("приготовление не завершилось")
		}
	}
	select {
	case got := <-done:
		want := []DeliveryResult{{MessageID: 7, Status: "отправлено: A"}, {MessageID: 7, Status: "отправлено: B"}, {MessageID: 7, Status: "отправлено: C"}}
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(orders, before) {
			t.Fatalf("порядок или значения нарушены: got %v, want %v; orders %v", got, want, orders)
		}
	case <-time.After(testTimeout):
		t.Fatal("результаты не собраны; возможно, не завершилось ожидание работников")
	}
	for i := range calls {
		if calls[i].Load() != 1 {
			t.Fatalf("заказ %d приготовлен %d раз", i, calls[i].Load())
		}
	}
}

func callFanOut(t *testing.T, in <-chan Notification, workers int, work Sender) []<-chan DeliveryResult {
	t.Helper()
	done := make(chan []<-chan DeliveryResult, 1)
	go func() { done <- FanOut(in, workers, work) }()
	select {
	case streams := <-done:
		return streams
	case <-time.After(testTimeout):
		t.Fatal("FanOut должна вернуть потоки, не дожидаясь закрытия входа")
		return nil
	}
}
func callFanIn(t *testing.T, streams ...<-chan DeliveryResult) <-chan DeliveryResult {
	t.Helper()
	done := make(chan (<-chan DeliveryResult), 1)
	go func() { done <- FanIn(streams...) }()
	select {
	case out := <-done:
		return out
	case <-time.After(testTimeout):
		t.Fatal("FanIn должна вернуть канал, не дожидаясь входных сообщений")
		return nil
	}
}
func collectFan(t *testing.T, out <-chan DeliveryResult) []DeliveryResult {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	values := []DeliveryResult{}
	for {
		select {
		case v, ok := <-out:
			if !ok {
				return values
			}
			values = append(values, v)
			if len(values) > 100 {
				t.Fatal("слишком много результатов")
			}
		case <-timer.C:
			t.Fatal("канал результатов не закрыт или поток заблокирован")
			return nil
		}
	}
}
func TestAcceptanceFanOutInvalid(t *testing.T) {
	for _, workers := range []int{0, -1} {
		in := make(chan Notification, 1)
		in <- Notification{ID: 7, Channel: "A", OrderID: 10}
		var calls atomic.Int32
		streams := callFanOut(t, in, workers, func(Notification) DeliveryResult { calls.Add(1); return DeliveryResult{} })
		if len(streams) != 0 || len(in) != 1 || calls.Load() != 0 {
			t.Fatalf("workers=%d: streams=%d, queued=%d, calls=%d", workers, len(streams), len(in), calls.Load())
		}
		close(in)
	}
}
func TestAcceptanceFanOutEmpty(t *testing.T) {
	in := make(chan Notification)
	close(in)
	var calls atomic.Int32
	streams := callFanOut(t, in, 3, func(Notification) DeliveryResult { calls.Add(1); return DeliveryResult{} })
	if len(streams) != 3 {
		t.Fatalf("нужны три потока, got %d", len(streams))
	}
	seen := make(map[<-chan DeliveryResult]bool)
	for _, stream := range streams {
		if stream == nil || seen[stream] {
			t.Fatal("у каждого воркера должен быть собственный ненулевой канал")
		}
		seen[stream] = true
		if got := collectFan(t, stream); len(got) != 0 {
			t.Fatal(got)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("пустой вход не должен вызывать callback")
	}
}
func TestAcceptanceFanInEmpty(t *testing.T) {
	if got := collectFan(t, callFanIn(t)); len(got) != 0 {
		t.Fatal(got)
	}
}
func TestAcceptanceFanInConcurrent(t *testing.T) {
	first := make(chan DeliveryResult)
	closed := false
	defer func() {
		if !closed {
			close(first)
		}
	}()
	second := make(chan DeliveryResult, 1)
	want := DeliveryResult{MessageID: 8, Status: "second"}
	second <- want
	close(second)
	out := callFanIn(t, first, second)
	select {
	case got, ok := <-out:
		if !ok || got != want {
			t.Fatalf("got %v %v", got, ok)
		}
	case <-time.After(testTimeout):
		t.Fatal("пустой незакрытый первый поток не должен блокировать готовый второй")
	}
	last := DeliveryResult{MessageID: 9, Status: "first"}
	select {
	case first <- last:
	case <-time.After(testTimeout):
		t.Fatal("FanIn перестала читать открытый поток")
	}
	close(first)
	closed = true
	if got := collectFan(t, out); !reflect.DeepEqual(got, []DeliveryResult{last}) {
		t.Fatalf("FanIn закрыта слишком рано или потеряла сообщение: %v", got)
	}
}
func TestAcceptanceFanPipeline(t *testing.T) {
	const workers, total = 3, 6
	in := make(chan Notification)
	abort := make(chan struct{})
	defer close(abort)
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	started := make(chan int, total)
	var active atomic.Int32
	var overLimit atomic.Bool
	var calls [total]atomic.Int32
	go func() {
		defer close(in)
		for i := 0; i < total; i++ {
			item := Notification{ID: 7, Channel: string(rune('A' + i)), OrderID: i + 10}
			select {
			case in <- item:
			case <-abort:
				return
			}
		}
	}()
	streams := callFanOut(t, in, workers, func(item Notification) DeliveryResult {
		index := int(item.Channel[0] - 'A')
		calls[index].Add(1)
		if active.Add(1) > workers {
			overLimit.Store(true)
		}
		defer active.Add(-1)
		started <- index
		<-release
		return DeliveryResult{MessageID: item.ID + 100, Status: item.Channel + "!"}
	})
	if len(streams) != workers {
		t.Fatalf("got %d streams, want %d", len(streams), workers)
	}
	seen := make(map[<-chan DeliveryResult]bool)
	for _, stream := range streams {
		if stream == nil || seen[stream] {
			t.Fatal("нужны отдельные каналы воркеров")
		}
		seen[stream] = true
	}
	for i := 0; i < workers; i++ {
		select {
		case <-started:
		case <-time.After(testTimeout):
			t.Fatal("все воркеры должны начать работу до завершения первого")
		}
	}
	// Пока три callback удерживаются барьером, четвёртый начинать нельзя.
	select {
	case <-started:
		t.Fatal("число одновременно работающих callback превысило workers")
	case <-time.After(20 * time.Millisecond):
	}
	out := callFanIn(t, streams...)
	releaseOnce.Do(func() { close(release) })
	values := collectFan(t, out)
	if len(values) != total || overLimit.Load() {
		t.Fatalf("results=%d overLimit=%v", len(values), overLimit.Load())
	}
	counts := make(map[DeliveryResult]int)
	for _, value := range values {
		counts[value]++
	}
	for i := 0; i < total; i++ {
		want := DeliveryResult{MessageID: 107, Status: string(rune('A'+i)) + "!"}
		if counts[want] != 1 || calls[i].Load() != 1 {
			t.Fatalf("item %d: count=%d calls=%d", i, counts[want], calls[i].Load())
		}
	}
}
func clockResult(t *testing.T, ctx context.Context, interval time.Duration, report func()) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- RunNotificationReport(ctx, interval, report) }()
	select {
	case err := <-done:
		return err
	case <-time.After(testTimeout):
		t.Fatal("RunNotificationReport не завершилась после отмены")
		return nil
	}
}

func TestAcceptanceClockInvalidInterval(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, ctx := range []context.Context{context.Background(), cancelled} {
		for _, interval := range []time.Duration{0, -time.Second} {
			var calls atomic.Int32
			err := clockResult(t, ctx, interval, func() { calls.Add(1) })
			if !errors.Is(err, ErrInvalidInterval) || calls.Load() != 0 {
				t.Fatalf("interval=%v, ctx.Err()=%v: err=%v, calls=%d; want ErrInvalidInterval and no reports", interval, ctx.Err(), err, calls.Load())
			}
		}
	}
}

func TestAcceptanceClockAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls atomic.Int32
	err := clockResult(t, ctx, time.Hour, func() { calls.Add(1) })
	if !errors.Is(err, context.Canceled) || calls.Load() != 0 {
		t.Fatalf("err=%v, calls=%d; уже отменённый контекст не запускает отчёты", err, calls.Load())
	}
}

func TestAcceptanceClockRepeatedAndCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	err := clockResult(t, ctx, 8*time.Millisecond, func() {
		if calls.Add(1) == 3 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) || calls.Load() < 3 {
		t.Fatalf("нужны повторные отчёты и возврат причины отмены: calls=%d, err=%v", calls.Load(), err)
	}
	// Возврат функции должен означать завершение и её callback.
	before := calls.Load()
	time.Sleep(24 * time.Millisecond)
	if calls.Load() != before {
		t.Fatal("после возврата RunNotificationReport продолжились фоновые отчёты")
	}
}

func TestAcceptanceClockDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	var calls atomic.Int32
	err := clockResult(t, ctx, time.Hour, func() { calls.Add(1) })
	if !errors.Is(err, context.DeadlineExceeded) || calls.Load() != 0 {
		t.Fatalf("ожидание первого тика должно прерываться: err=%v, calls=%d", err, calls.Load())
	}
}
