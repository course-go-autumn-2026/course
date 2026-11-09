package notifications

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []Notification{{ID: 7, Channel: "A", OrderID: 10}}
	want := DeliveryResult{MessageID: 7, Status: "A"}
	got := legacyBatch(items, func(item Notification) DeliveryResult {
		return DeliveryResult{MessageID: item.ID, Status: item.Channel}
	})
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan Notification, 2)
	in <- Notification{ID: 7, Channel: "A"}
	in <- Notification{ID: 8, Channel: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item Notification) DeliveryResult {
		return DeliveryResult{MessageID: item.ID, Status: item.Channel}
	})
	if len(got) != 2 || got[0] != (DeliveryResult{MessageID: 7, Status: "A"}) || got[1] != (DeliveryResult{MessageID: 8, Status: "B"}) {
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
