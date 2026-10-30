package task01solid

import "testing"

func TestStandardOption(t *testing.T) {
	store := NewMemoryStore()
	receipt, err := NewService(store).Book(Booking{
		ID: "42", Items: []Item{{Name: "Час съёмки", Price: 1000, Quantity: 2}},
		BackdropKind: "color", Room: "Студия B",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt != "Бронь 42: 2200 руб." {
		t.Fatalf("unexpected receipt: %q", receipt)
	}
	if _, err := store.Load("42"); err != nil {
		t.Fatalf("record was not saved: %v", err)
	}
}

func TestFreeOption(t *testing.T) {
	receipt, err := NewService(NewMemoryStore()).Book(Booking{
		ID: "43", Items: []Item{{Name: "Час съёмки", Price: 1000, Quantity: 1}},
		BackdropKind: "plain",
	})
	if err != nil || receipt != "Бронь 43: 1000 руб." {
		t.Fatalf("receipt=%q, err=%v", receipt, err)
	}
}

func TestInvalidQuantityIsRejected(t *testing.T) {
	store := NewMemoryStore()
	_, err := NewService(store).Book(Booking{
		ID: "44", Items: []Item{{Name: "Час съёмки", Price: 1000, Quantity: 0}},
		BackdropKind: "plain",
	})
	if err == nil || store.Count() != 0 {
		t.Fatal("invalid quantity must be rejected before saving")
	}
}
