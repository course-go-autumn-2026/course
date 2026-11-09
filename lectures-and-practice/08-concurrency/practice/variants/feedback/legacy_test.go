package feedback

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []ReviewRequest{{ID: 7, Product: "A", Stars: 10}}
	want := Review{ReviewID: 7, Text: "A"}
	got := legacyBatch(items, func(item ReviewRequest) Review { return Review{ReviewID: item.ID, Text: item.Product} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan ReviewRequest, 2)
	in <- ReviewRequest{ID: 7, Product: "A"}
	in <- ReviewRequest{ID: 8, Product: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item ReviewRequest) Review { return Review{ReviewID: item.ID, Text: item.Product} })
	if len(got) != 2 || got[0] != (Review{ReviewID: 7, Text: "A"}) || got[1] != (Review{ReviewID: 8, Text: "B"}) {
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
