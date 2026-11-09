package kitchen

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []Order{{ID: 7, Pizza: "A", PriceRub: 10}}
	want := BakedPizza{OrderID: 7, Label: "A"}
	got := legacyBakeAll(items, func(item Order) BakedPizza { return BakedPizza{OrderID: item.ID, Label: item.Pizza} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan Order, 2)
	in <- Order{ID: 7, Pizza: "A"}
	in <- Order{ID: 8, Pizza: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item Order) BakedPizza { return BakedPizza{OrderID: item.ID, Label: item.Pizza} })
	if len(got) != 2 || got[0] != (BakedPizza{OrderID: 7, Label: "A"}) || got[1] != (BakedPizza{OrderID: 8, Label: "B"}) {
		t.Fatal(got)
	}
}
func TestLegacyHeartbeat(t *testing.T) {
	calls := 0
	legacyHeartbeat(3, 0, func() { calls++ })
	if calls != 3 {
		t.Fatalf("report called %d times, want 3", calls)
	}
}
