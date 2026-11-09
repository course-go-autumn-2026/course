package packing

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []PackRequest{{ID: 7, Size: "A", PriceRub: 10}}
	want := PackedOrder{OrderID: 7, Description: "A"}
	got := legacyBatch(items, func(item PackRequest) PackedOrder { return PackedOrder{OrderID: item.ID, Description: item.Size} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan PackRequest, 2)
	in <- PackRequest{ID: 7, Size: "A"}
	in <- PackRequest{ID: 8, Size: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item PackRequest) PackedOrder { return PackedOrder{OrderID: item.ID, Description: item.Size} })
	if len(got) != 2 || got[0] != (PackedOrder{OrderID: 7, Description: "A"}) || got[1] != (PackedOrder{OrderID: 8, Description: "B"}) {
		t.Fatal(got)
	}
}
func TestLegacyClock(t *testing.T) {
	calls := 0
	legacyClock(3, 0, func() { calls++ })
	if calls != 3 {
		t.Fatal(calls)
	}
}
