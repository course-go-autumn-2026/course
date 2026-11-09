package promos

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []PromoRequest{{ID: 7, Code: "A", BaseRub: 10}}
	want := Quote{OrderID: 7, Description: "A"}
	got := legacyBatch(items, func(item PromoRequest) Quote { return Quote{OrderID: item.ID, Description: item.Code} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan PromoRequest, 2)
	in <- PromoRequest{ID: 7, Code: "A"}
	in <- PromoRequest{ID: 8, Code: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item PromoRequest) Quote { return Quote{OrderID: item.ID, Description: item.Code} })
	if len(got) != 2 || got[0] != (Quote{OrderID: 7, Description: "A"}) || got[1] != (Quote{OrderID: 8, Description: "B"}) {
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
