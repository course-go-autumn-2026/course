package task01solid

import "testing"

func TestStandardOption(t *testing.T) {
	store := NewMemoryStore()
	receipt, err := NewService(store).Book(Appointment{
		ID: "42", Items: []Item{{Name: "Стрижка", Price: 700, Quantity: 2}},
		CareKind: "coat", Size: "Маленький",
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt != "Запись 42: 1600 руб." {
		t.Fatalf("unexpected receipt: %q", receipt)
	}
	if _, err := store.Load("42"); err != nil {
		t.Fatalf("record was not saved: %v", err)
	}
}

func TestFreeOption(t *testing.T) {
	receipt, err := NewService(NewMemoryStore()).Book(Appointment{
		ID: "43", Items: []Item{{Name: "Стрижка", Price: 700, Quantity: 1}},
		CareKind: "basic",
	})
	if err != nil || receipt != "Запись 43: 700 руб." {
		t.Fatalf("receipt=%q, err=%v", receipt, err)
	}
}

func TestInvalidQuantityIsRejected(t *testing.T) {
	store := NewMemoryStore()
	_, err := NewService(store).Book(Appointment{
		ID: "44", Items: []Item{{Name: "Стрижка", Price: 700, Quantity: 0}},
		CareKind: "basic",
	})
	if err == nil || store.Count() != 0 {
		t.Fatal("invalid quantity must be rejected before saving")
	}
}
