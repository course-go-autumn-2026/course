package couriers

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []CourierTask{{ID: 7, Name: "A", DistanceMeters: 10}}
	want := CourierCard{CourierID: 7, Description: "A"}
	got := legacyBatch(items, func(item CourierTask) CourierCard { return CourierCard{CourierID: item.ID, Description: item.Name} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan CourierTask, 2)
	in <- CourierTask{ID: 7, Name: "A"}
	in <- CourierTask{ID: 8, Name: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item CourierTask) CourierCard { return CourierCard{CourierID: item.ID, Description: item.Name} })
	if len(got) != 2 || got[0] != (CourierCard{CourierID: 7, Description: "A"}) || got[1] != (CourierCard{CourierID: 8, Description: "B"}) {
		t.Fatal(got)
	}
}
func TestLegacyClock(t *testing.T) {
	ready := make(chan CourierCard, 1)
	want := CourierCard{CourierID: 7, Description: "ready"}
	ready <- want
	if got := legacyClock(ready); got != want {
		t.Fatal(got)
	}
}
