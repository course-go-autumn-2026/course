package routes

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

func batchResult(t *testing.T, orders []Delivery, bake RouteEstimator) []Route {
	t.Helper()
	done := make(chan []Route, 1)
	go func() { done <- EstimateAll(orders, bake) }()
	select {
	case result := <-done:
		return result
	case <-time.After(testTimeout):
		t.Fatal("EstimateAll не завершилась: проверьте WaitGroup и сбор результатов")
		return nil
	}
}

func TestAcceptanceBatchEmpty(t *testing.T) {
	for _, orders := range [][]Delivery{nil, {}} {
		var calls atomic.Int32
		got := batchResult(t, orders, func(Delivery) Route {
			calls.Add(1)
			return Route{}
		})
		if len(got) != 0 || calls.Load() != 0 {
			t.Fatalf("пустая партия: result=%v, calls=%d", got, calls.Load())
		}
	}
}

func TestAcceptanceBatchValues(t *testing.T) {
	orders := []Delivery{{ID: 42, Address: "Маргарита", DistanceMeters: 450}}
	before := append([]Delivery(nil), orders...)
	want := Route{OrderID: 42, Description: "маршрут: Маргарита"}
	var calls atomic.Int32
	var received Delivery
	got := batchResult(t, orders, func(order Delivery) Route {
		calls.Add(1)
		received = order
		return want
	})
	if calls.Load() != 1 || received != before[0] || !reflect.DeepEqual(orders, before) || !reflect.DeepEqual(got, []Route{want}) {
		t.Fatalf("потерян заказ, результат или изменён вход: calls=%d, received=%+v, orders=%v, result=%v", calls.Load(), received, orders, got)
	}
}

func TestAcceptanceBatchConcurrent(t *testing.T) {
	// ID намеренно совпадают: восстановление порядка по ID неверно.
	orders := []Delivery{{ID: 7, Address: "A", DistanceMeters: 450}, {ID: 7, Address: "B", DistanceMeters: 550}, {ID: 7, Address: "C", DistanceMeters: 650}}
	before := append([]Delivery(nil), orders...)
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
	done := make(chan []Route, 1)
	go func() {
		done <- EstimateAll(orders, func(o Delivery) Route {
			i := int(o.Address[0] - 'A')
			calls[i].Add(1)
			started <- i
			<-release[i]
			finished <- i
			return Route{OrderID: o.ID, Description: "маршрут: " + o.Address}
		})
	}()
	seen := make(map[int]bool)
	deadline := time.NewTimer(testTimeout)
	defer deadline.Stop()
	for len(seen) < len(orders) {
		select {
		case i := <-started:
			if seen[i] {
				t.Fatal("RouteEstimator вызван повторно для одного заказа")
			}
			seen[i] = true
		case <-deadline.C:
			t.Fatal("все пиццы должны начать готовиться до завершения первой; последовательный цикл не подходит")
		}
	}
	select {
	case <-done:
		t.Fatal("EstimateAll вернулась до готовности пицц")
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
		want := []Route{{OrderID: 7, Description: "маршрут: A"}, {OrderID: 7, Description: "маршрут: B"}, {OrderID: 7, Description: "маршрут: C"}}
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

func callFanOut(t *testing.T, in <-chan Delivery, workers int, work RouteEstimator) []<-chan Route {
	t.Helper()
	done := make(chan []<-chan Route, 1)
	go func() { done <- FanOut(in, workers, work) }()
	select {
	case streams := <-done:
		return streams
	case <-time.After(testTimeout):
		t.Fatal("FanOut должна вернуть потоки, не дожидаясь закрытия входа")
		return nil
	}
}
func callFanIn(t *testing.T, streams ...<-chan Route) <-chan Route {
	t.Helper()
	done := make(chan (<-chan Route), 1)
	go func() { done <- FanIn(streams...) }()
	select {
	case out := <-done:
		return out
	case <-time.After(testTimeout):
		t.Fatal("FanIn должна вернуть канал, не дожидаясь входных сообщений")
		return nil
	}
}
func collectFan(t *testing.T, out <-chan Route) []Route {
	t.Helper()
	timer := time.NewTimer(testTimeout)
	defer timer.Stop()
	values := []Route{}
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
		in := make(chan Delivery, 1)
		in <- Delivery{ID: 7, Address: "A", DistanceMeters: 10}
		var calls atomic.Int32
		streams := callFanOut(t, in, workers, func(Delivery) Route { calls.Add(1); return Route{} })
		if len(streams) != 0 || len(in) != 1 || calls.Load() != 0 {
			t.Fatalf("workers=%d: streams=%d, queued=%d, calls=%d", workers, len(streams), len(in), calls.Load())
		}
		close(in)
	}
}
func TestAcceptanceFanOutEmpty(t *testing.T) {
	in := make(chan Delivery)
	close(in)
	var calls atomic.Int32
	streams := callFanOut(t, in, 3, func(Delivery) Route { calls.Add(1); return Route{} })
	if len(streams) != 3 {
		t.Fatalf("нужны три потока, got %d", len(streams))
	}
	seen := make(map[<-chan Route]bool)
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
	first := make(chan Route)
	closed := false
	defer func() {
		if !closed {
			close(first)
		}
	}()
	second := make(chan Route, 1)
	want := Route{OrderID: 8, Description: "second"}
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
	last := Route{OrderID: 9, Description: "first"}
	select {
	case first <- last:
	case <-time.After(testTimeout):
		t.Fatal("FanIn перестала читать открытый поток")
	}
	close(first)
	closed = true
	if got := collectFan(t, out); !reflect.DeepEqual(got, []Route{last}) {
		t.Fatalf("FanIn закрыта слишком рано или потеряла сообщение: %v", got)
	}
}
func TestAcceptanceFanPipeline(t *testing.T) {
	const workers, total = 3, 6
	in := make(chan Delivery)
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
			item := Delivery{ID: 7, Address: string(rune('A' + i)), DistanceMeters: i + 10}
			select {
			case in <- item:
			case <-abort:
				return
			}
		}
	}()
	streams := callFanOut(t, in, workers, func(item Delivery) Route {
		index := int(item.Address[0] - 'A')
		calls[index].Add(1)
		if active.Add(1) > workers {
			overLimit.Store(true)
		}
		defer active.Add(-1)
		started <- index
		<-release
		return Route{OrderID: item.ID + 100, Description: item.Address + "!"}
	})
	if len(streams) != workers {
		t.Fatalf("got %d streams, want %d", len(streams), workers)
	}
	seen := make(map[<-chan Route]bool)
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
	counts := make(map[Route]int)
	for _, value := range values {
		counts[value]++
	}
	for i := 0; i < total; i++ {
		want := Route{OrderID: 107, Description: string(rune('A'+i)) + "!"}
		if counts[want] != 1 || calls[i].Load() != 1 {
			t.Fatalf("item %d: count=%d calls=%d", i, counts[want], calls[i].Load())
		}
	}
}
func clockResult(t *testing.T, ctx context.Context, ready <-chan Route, timeout time.Duration) (Route, error) {
	t.Helper()
	type response struct {
		value Route
		err   error
	}
	done := make(chan response, 1)
	go func() { v, e := WaitForCourier(ctx, ready, timeout); done <- response{v, e} }()
	select {
	case r := <-done:
		return r.value, r.err
	case <-time.After(testTimeout):
		t.Fatal("waiting did not finish")
		return Route{}, nil
	}
}
func TestAcceptanceClockInvalidAndCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ready := make(chan Route, 1)
	want := Route{OrderID: 7, Description: "ready"}
	ready <- want
	for _, d := range []time.Duration{0, -time.Second} {
		v, e := clockResult(t, ctx, ready, d)
		if v != (Route{}) || !errors.Is(e, ErrInvalidTimeout) {
			t.Fatalf("invalid: %v %v", v, e)
		}
	}
	v, e := clockResult(t, ctx, ready, time.Hour)
	if v != (Route{}) || !errors.Is(e, context.Canceled) || len(ready) != 1 {
		t.Fatalf("already canceled: %v %v len=%d", v, e, len(ready))
	}
}
func TestAcceptanceClockReady(t *testing.T) {
	ready := make(chan Route, 2)
	want := Route{OrderID: 7, Description: "ready"}
	next := Route{OrderID: 8, Description: "next"}
	ready <- want
	ready <- next
	v, e := clockResult(t, context.Background(), ready, time.Hour)
	if e != nil || v != want || len(ready) != 1 {
		t.Fatalf("first result: %v %v remaining=%d", v, e, len(ready))
	}
	if got := <-ready; got != next {
		t.Fatalf("consumed or changed second result: %v", got)
	}
	// A zero result is a valid message, not evidence that the channel is closed.
	ready <- Route{}
	v, e = clockResult(t, context.Background(), ready, time.Hour)
	if e != nil || v != (Route{}) {
		t.Fatalf("zero result is valid: %v %v", v, e)
	}
	buffered := make(chan Route, 1)
	buffered <- want
	close(buffered)
	v, e = clockResult(t, context.Background(), buffered, time.Hour)
	if v != want || e != nil {
		t.Fatalf("closed buffered channel still has a message: %v %v", v, e)
	}
}
func TestAcceptanceClockTimeoutAndClosed(t *testing.T) {
	ready := make(chan Route)
	v, e := clockResult(t, context.Background(), ready, 15*time.Millisecond)
	if v != (Route{}) || !errors.Is(e, ErrTimeout) {
		t.Fatalf("timeout: %v %v", v, e)
	}
	close(ready)
	v, e = clockResult(t, context.Background(), ready, time.Hour)
	if v != (Route{}) || !errors.Is(e, ErrClosed) {
		t.Fatalf("closed: %v %v", v, e)
	}
}
func TestAcceptanceClockDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	v, e := clockResult(t, ctx, make(chan Route), time.Hour)
	if v != (Route{}) || !errors.Is(e, context.DeadlineExceeded) {
		t.Fatalf("deadline: %v %v", v, e)
	}
}
