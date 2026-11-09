package routes

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []Delivery{{ID: 7, Address: "A", DistanceMeters: 10}}
	want := Route{OrderID: 7, Description: "A"}
	got := legacyBatch(items, func(item Delivery) Route { return Route{OrderID: item.ID, Description: item.Address} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan Delivery, 2)
	in <- Delivery{ID: 7, Address: "A"}
	in <- Delivery{ID: 8, Address: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item Delivery) Route { return Route{OrderID: item.ID, Description: item.Address} })
	if len(got) != 2 || got[0] != (Route{OrderID: 7, Description: "A"}) || got[1] != (Route{OrderID: 8, Description: "B"}) {
		t.Fatal(got)
	}
}
func TestLegacyClock(t *testing.T) {
	ready := make(chan Route, 1)
	want := Route{OrderID: 7, Description: "ready"}
	ready <- want
	if got := legacyClock(ready); got != want {
		t.Fatal(got)
	}
}
