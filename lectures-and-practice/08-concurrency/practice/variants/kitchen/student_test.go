package kitchen

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStudentBatch(t *testing.T) {
	if got := BakeAll(nil, func(Order) BakedPizza { return BakedPizza{} }); len(got) != 0 {
		t.Fatal(got)
	}

	items := []Order{{ID: 7, Pizza: "A", PriceRub: 10}, {ID: 7, Pizza: "B", PriceRub: 20}}
	want := []BakedPizza{{OrderID: 7, Label: "A"}, {OrderID: 7, Label: "B"}}
	got := BakeAll(items, func(item Order) BakedPizza { return BakedPizza{OrderID: item.ID, Label: item.Pizza} })
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatal(got)
	}

	const total = 3
	orders := make([]Order, total)
	for i := range orders {
		orders[i] = Order{ID: i, Pizza: "P", PriceRub: i}
	}
	started := make(chan struct{}, total)
	release := make(chan struct{})
	done := make(chan []BakedPizza, 1)
	go func() {
		done <- BakeAll(orders, func(order Order) BakedPizza {
			started <- struct{}{}
			<-release
			return BakedPizza{OrderID: order.ID, Label: order.Pizza}
		})
	}()
	for range total {
		<-started
	}
	close(release)
	res := <-done
	for i, pizza := range res {
		if pizza != (BakedPizza{OrderID: i, Label: "P"}) {
			t.Fatal(res)
		}
	}
}

func TestStudentFan(t *testing.T) {
	if FanOut(nil, 0, nil) != nil {
		t.Fatal("workers <= 0 должен вернуть nil")
	}

	in := make(chan Order, 3)
	for i := range 3 {
		in <- Order{ID: i, Pizza: "P", PriceRub: i}
	}
	close(in)
	streams := FanOut(in, 2, func(item Order) BakedPizza { return BakedPizza{OrderID: item.ID, Label: item.Pizza} })
	if len(streams) != 2 {
		t.Fatalf("got %d streams, want 2", len(streams))
	}
	for _, stream := range streams {
		if stream == nil {
			t.Fatal("nil stream")
		}
	}

	counts := make(map[BakedPizza]int)
	for pizza := range FanIn(streams...) {
		counts[pizza]++
	}
	if len(counts) != 3 {
		t.Fatal(counts)
	}
	for pizza, count := range counts {
		if count != 1 || pizza.Label != "P" {
			t.Fatal(counts)
		}
	}
}

func TestStudentHeartbeat(t *testing.T) {
	if err := RunHeartbeat(context.Background(), 0, func() {}); !errors.Is(err, ErrInvalidInterval) {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	err := RunHeartbeat(ctx, time.Millisecond, func() {
		calls++
		if calls == 3 {
			cancel()
		}
	})
	if calls < 3 || !errors.Is(err, context.Canceled) {
		t.Fatalf("report called %d times, err=%v", calls, err)
	}
}
