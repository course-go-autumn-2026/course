package task01solid

import "testing"

func TestStandardOption(t *testing.T) {
	store := NewMemoryStore()
	receipt, err := NewService(store).Book(Booking{
		ID: "42", Items: []Item{{Name: "Билет", Price: 400, Quantity: 2}},
		SeatKind: "comfort", Hall: "Малый зал",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt != "Бронь 42: 1000 руб." {
		t.Fatalf("unexpected receipt: %q", receipt)
	}
	if _, err := store.Load("42"); err != nil {
		t.Fatalf("record was not saved: %v", err)
	}
}

func TestFreeOption(t *testing.T) {
	receipt, err := NewService(NewMemoryStore()).Book(Booking{
		ID: "43", Items: []Item{{Name: "Билет", Price: 400, Quantity: 1}},
		SeatKind: "standard",
	})
	if err != nil || receipt != "Бронь 43: 400 руб." {
		t.Fatalf("receipt=%q, err=%v", receipt, err)
	}
}

func TestInvalidQuantityIsRejected(t *testing.T) {
	store := NewMemoryStore()
	_, err := NewService(store).Book(Booking{
		ID: "44", Items: []Item{{Name: "Билет", Price: 400, Quantity: 0}},
		SeatKind: "standard",
	})
	if err == nil || store.Count() != 0 {
		t.Fatal("invalid quantity must be rejected before saving")
	}
}
