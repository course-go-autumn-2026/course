package packing

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

func batchResult(t *testing.T, orders []PackRequest, bake Packer) []PackedOrder {
	t.Helper()
	done := make(chan []PackedOrder, 1)
	go func() { done <- PackAll(orders, bake) }()
	select {
	case result := <-done:
		return result
	case <-time.After(testTimeout):
		t.Fatal("PackAll не завершилась: проверьте WaitGroup и сбор результатов")
		return nil
	}
}

func TestAcceptanceBatchEmpty(t *testing.T) {
	for _, orders := range [][]PackRequest{nil, {}} {
		var calls atomic.Int32
		got := batchResult(t, orders, func(PackRequest) PackedOrder {
			calls.Add(1)
			return PackedOrder{}
		})
		if len(got) != 0 || calls.Load() != 0 {
			t.Fatalf("пустая партия: result=%v, calls=%d", got, calls.Load())
		}
	}
}

func TestAcceptanceBatchValues(t *testing.T) {
	orders := []PackRequest{{ID: 42, Size: "Маргарита", PriceRub: 450}}
	before := append([]PackRequest(nil), orders...)
	want := PackedOrder{OrderID: 42, Description: "упакован: Маргарита"}
	var calls atomic.Int32
	var received PackRequest
	got := batchResult(t, orders, func(order PackRequest) PackedOrder {
		calls.Add(1)
		received = order
		return want
	})
	if calls.Load() != 1 || received != before[0] || !reflect.DeepEqual(orders, before) || !reflect.DeepEqual(got, []PackedOrder{want}) {
		t.Fatalf("потерян заказ, результат или изменён вход: calls=%d, received=%+v, orders=%v, result=%v", calls.Load(), received, orders, got)
	}
}

func TestAcceptanceBatchConcurrent(t *testing.T) {
	// ID намеренно совпадают: восстановление порядка по ID неверно.
	orders := []PackRequest{{ID: 7, Size: "A", PriceRub: 450}, {ID: 7, Size: "B", PriceRub: 550}, {ID: 7, Size: "C", PriceRub: 650}}
	before := append([]PackRequest(nil), orders...)
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
	done := make(chan []PackedOrder, 1)
	go func() {
		done <- PackAll(orders, func(o PackRequest) PackedOrder {
			i := int(o.Size[0] - 'A')
			calls[i].Add(1)
			started <- i
			<-release[i]
			finished <- i
			return PackedOrder{OrderID: o.ID, Description: "упакован: " + o.Size}
		})
	}()
	seen := make(map[int]bool)
	deadline := time.NewTimer(testTimeout)
	defer deadline.Stop()
	for len(seen) < len(orders) {
		select {
		case i := <-started:
			if seen[i] {
				t.Fatal("Packer вызван повторно для одного заказа")
			}
			seen[i] = true
		case <-deadline.C:
			t.Fatal("все пиццы должны начать готовиться до завершения первой; последовательный цикл не подходит")
		}
	}
	select {
	case <-done:
		t.Fatal("PackAll вернулась до готовности пицц")
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
		want := []PackedOrder{{OrderID: 7, Description: "упакован: A"}, {OrderID: 7, Description: "упакован: B"}, {OrderID: 7, Description: "упакован: C"}}
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

func callFanOut(t *testing.T, in <-chan PackRequest, workers int, work Packer) []<-chan PackedOrder {
	t.Helper()
	done := make(chan []<-chan PackedOrder, 1)
	go func() { done <- FanOut(in, workers, work) }()
	select {
	case streams := <-done:
		return streams
	case <-time.After(testTimeout):
		t.Fatal("FanOut должна вернуть потоки, не дожидаясь закрытия входа")
		return nil
	}
}
func callFanIn(t *testing.T, streams ...<-chan PackedOrder) <-chan PackedOrder {
	t.Helper()
	done := make(chan (<-chan PackedOrder), 1)
	go func() { done <- FanIn(streams...) }()
	select {
	case out := <-done:
		return out
	case <-time.After(testTimeout):
		t.Fatal("FanIn должна вернуть канал, не дожидаясь входных сообщений")
		return nil
	}
}
func collectFan(t *testing.T, out <-chan PackedOrder) []PackedOrder {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	values := []PackedOrder{}
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
		in := make(chan PackRequest, 1)
		in <- PackRequest{ID: 7, Size: "A", PriceRub: 10}
		var calls atomic.Int32
		streams := callFanOut(t, in, workers, func(PackRequest) PackedOrder { calls.Add(1); return PackedOrder{} })
		if len(streams) != 0 || len(in) != 1 || calls.Load() != 0 {
			t.Fatalf("workers=%d: streams=%d, queued=%d, calls=%d", workers, len(streams), len(in), calls.Load())
		}
		close(in)
	}
}
func TestAcceptanceFanOutEmpty(t *testing.T) {
	in := make(chan PackRequest)
	close(in)
	var calls atomic.Int32
	streams := callFanOut(t, in, 3, func(PackRequest) PackedOrder { calls.Add(1); return PackedOrder{} })
	if len(streams) != 3 {
		t.Fatalf("нужны три потока, got %d", len(streams))
	}
	seen := make(map[<-chan PackedOrder]bool)
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
	first := make(chan PackedOrder)
	closed := false
	defer func() {
		if !closed {
			close(first)
		}
	}()
	second := make(chan PackedOrder, 1)
	want := PackedOrder{OrderID: 8, Description: "second"}
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
	last := PackedOrder{OrderID: 9, Description: "first"}
	select {
	case first <- last:
	case <-time.After(testTimeout):
		t.Fatal("FanIn перестала читать открытый поток")
	}
	close(first)
	closed = true
	if got := collectFan(t, out); !reflect.DeepEqual(got, []PackedOrder{last}) {
		t.Fatalf("FanIn закрыта слишком рано или потеряла сообщение: %v", got)
	}
}
func TestAcceptanceFanPipeline(t *testing.T) {
	const workers, total = 3, 6
	in := make(chan PackRequest)
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
			item := PackRequest{ID: 7, Size: string(rune('A' + i)), PriceRub: i + 10}
			select {
			case in <- item:
			case <-abort:
				return
			}
		}
	}()
	streams := callFanOut(t, in, workers, func(item PackRequest) PackedOrder {
		index := int(item.Size[0] - 'A')
		calls[index].Add(1)
		if active.Add(1) > workers {
			overLimit.Store(true)
		}
		defer active.Add(-1)
		started <- index
		<-release
		return PackedOrder{OrderID: item.ID + 100, Description: item.Size + "!"}
	})
	if len(streams) != workers {
		t.Fatalf("got %d streams, want %d", len(streams), workers)
	}
	seen := make(map[<-chan PackedOrder]bool)
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
	counts := make(map[PackedOrder]int)
	for _, value := range values {
		counts[value]++
	}
	for i := 0; i < total; i++ {
		want := PackedOrder{OrderID: 107, Description: string(rune('A'+i)) + "!"}
		if counts[want] != 1 || calls[i].Load() != 1 {
			t.Fatalf("item %d: count=%d calls=%d", i, counts[want], calls[i].Load())
		}
	}
}
func clockResult(t *testing.T, ctx context.Context, interval time.Duration, report func()) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- RunPackingReport(ctx, interval, report) }()
	select {
	case err := <-done:
		return err
	case <-time.After(testTimeout):
		t.Fatal("RunPackingReport не завершилась после отмены")
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
		t.Fatal("после возврата RunPackingReport продолжились фоновые отчёты")
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
