package receipts

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []ReceiptRequest{{ID: 7, Customer: "A", AmountRub: 10}}
	want := Receipt{OrderID: 7, Text: "A"}
	got := legacyBatch(items, func(item ReceiptRequest) Receipt { return Receipt{OrderID: item.ID, Text: item.Customer} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan ReceiptRequest, 2)
	in <- ReceiptRequest{ID: 7, Customer: "A"}
	in <- ReceiptRequest{ID: 8, Customer: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item ReceiptRequest) Receipt { return Receipt{OrderID: item.ID, Text: item.Customer} })
	if len(got) != 2 || got[0] != (Receipt{OrderID: 7, Text: "A"}) || got[1] != (Receipt{OrderID: 8, Text: "B"}) {
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
