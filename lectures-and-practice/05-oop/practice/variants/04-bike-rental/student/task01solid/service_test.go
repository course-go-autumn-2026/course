package task01solid

import "testing"

func TestStandardOption(t *testing.T) {
	store := NewMemoryStore()
	receipt, err := NewService(store).Rent(Rental{
		ID: "42", Items: []Item{{Name: "Велосипед на час", Price: 300, Quantity: 2}},
		ReturnKind: "other", Station: "Парк",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt != "Аренда 42: 800 руб." {
		t.Fatalf("unexpected receipt: %q", receipt)
	}
	if _, err := store.Load("42"); err != nil {
		t.Fatalf("record was not saved: %v", err)
	}
}

func TestFreeOption(t *testing.T) {
	receipt, err := NewService(NewMemoryStore()).Rent(Rental{
		ID: "43", Items: []Item{{Name: "Велосипед на час", Price: 300, Quantity: 1}},
		ReturnKind: "same",
	})
	if err != nil || receipt != "Аренда 43: 300 руб." {
		t.Fatalf("receipt=%q, err=%v", receipt, err)
	}
}

func TestInvalidQuantityIsRejected(t *testing.T) {
	store := NewMemoryStore()
	_, err := NewService(store).Rent(Rental{
		ID: "44", Items: []Item{{Name: "Велосипед на час", Price: 300, Quantity: 0}},
		ReturnKind: "same",
	})
	if err == nil || store.Count() != 0 {
		t.Fatal("invalid quantity must be rejected before saving")
	}
}
