package task01solid

import "testing"

func TestStandardOption(t *testing.T) {
	store := NewMemoryStore()
	receipt, err := NewService(store).PlaceOrder(Order{
		ID: "42", Items: []Item{{Name: "Латте", Price: 250, Quantity: 2}},
		ServingKind: "table", Area: "Зал",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt != "Заказ 42: 600 руб." {
		t.Fatalf("unexpected receipt: %q", receipt)
	}
	if _, err := store.Load("42"); err != nil {
		t.Fatalf("record was not saved: %v", err)
	}
}

func TestFreeOption(t *testing.T) {
	receipt, err := NewService(NewMemoryStore()).PlaceOrder(Order{
		ID: "43", Items: []Item{{Name: "Латте", Price: 250, Quantity: 1}},
		ServingKind: "takeaway",
	})
	if err != nil || receipt != "Заказ 43: 250 руб." {
		t.Fatalf("receipt=%q, err=%v", receipt, err)
	}
}

func TestInvalidQuantityIsRejected(t *testing.T) {
	store := NewMemoryStore()
	_, err := NewService(store).PlaceOrder(Order{
		ID: "44", Items: []Item{{Name: "Латте", Price: 250, Quantity: 0}},
		ServingKind: "takeaway",
	})
	if err == nil || store.Count() != 0 {
		t.Fatal("invalid quantity must be rejected before saving")
	}
}
