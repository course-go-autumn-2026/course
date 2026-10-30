package task01solid

import "testing"

func TestStandardOption(t *testing.T) {
	store := NewMemoryStore()
	receipt, err := NewService(store).Accept(Request{
		ID: "42", Items: []Item{{Name: "Установка памяти", Price: 500, Quantity: 2}},
		DiagnosticsKind: "standard", Device: "Компьютер",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt != "Заявка 42: 1200 руб." {
		t.Fatalf("unexpected receipt: %q", receipt)
	}
	if _, err := store.Load("42"); err != nil {
		t.Fatalf("record was not saved: %v", err)
	}
}

func TestFreeOption(t *testing.T) {
	receipt, err := NewService(NewMemoryStore()).Accept(Request{
		ID: "43", Items: []Item{{Name: "Установка памяти", Price: 500, Quantity: 1}},
		DiagnosticsKind: "none",
	})
	if err != nil || receipt != "Заявка 43: 500 руб." {
		t.Fatalf("receipt=%q, err=%v", receipt, err)
	}
}

func TestInvalidQuantityIsRejected(t *testing.T) {
	store := NewMemoryStore()
	_, err := NewService(store).Accept(Request{
		ID: "44", Items: []Item{{Name: "Установка памяти", Price: 500, Quantity: 0}},
		DiagnosticsKind: "none",
	})
	if err == nil || store.Count() != 0 {
		t.Fatal("invalid quantity must be rejected before saving")
	}
}
