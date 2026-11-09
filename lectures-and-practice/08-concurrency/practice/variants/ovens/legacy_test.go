package ovens

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []OvenTask{{ID: 7, Model: "A", TargetCelsius: 10}}
	want := OvenStatus{OvenID: 7, Description: "A"}
	got := legacyBatch(items, func(item OvenTask) OvenStatus { return OvenStatus{OvenID: item.ID, Description: item.Model} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan OvenTask, 2)
	in <- OvenTask{ID: 7, Model: "A"}
	in <- OvenTask{ID: 8, Model: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item OvenTask) OvenStatus { return OvenStatus{OvenID: item.ID, Description: item.Model} })
	if len(got) != 2 || got[0] != (OvenStatus{OvenID: 7, Description: "A"}) || got[1] != (OvenStatus{OvenID: 8, Description: "B"}) {
		t.Fatal(got)
	}
}
func TestLegacyClock(t *testing.T) {
	ready := make(chan OvenStatus, 1)
	want := OvenStatus{OvenID: 7, Description: "ready"}
	ready <- want
	if got := legacyClock(ready); got != want {
		t.Fatal(got)
	}
}
