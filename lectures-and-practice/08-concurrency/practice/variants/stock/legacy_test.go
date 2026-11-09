package stock

import "testing"

func TestLegacyBatch(t *testing.T) {
	items := []Ingredient{{ID: 7, Name: "A", RequiredGrams: 10}}
	want := StockStatus{IngredientID: 7, Description: "A"}
	got := legacyBatch(items, func(item Ingredient) StockStatus { return StockStatus{IngredientID: item.ID, Description: item.Name} })
	if len(got) != 1 || got[0] != want {
		t.Fatal(got)
	}
}
func TestLegacyFanPipeline(t *testing.T) {
	in := make(chan Ingredient, 2)
	in <- Ingredient{ID: 7, Name: "A"}
	in <- Ingredient{ID: 8, Name: "B"}
	close(in)
	got := legacyFanPipeline(in, func(item Ingredient) StockStatus { return StockStatus{IngredientID: item.ID, Description: item.Name} })
	if len(got) != 2 || got[0] != (StockStatus{IngredientID: 7, Description: "A"}) || got[1] != (StockStatus{IngredientID: 8, Description: "B"}) {
		t.Fatal(got)
	}
}
func TestLegacyClock(t *testing.T) {
	ready := make(chan StockStatus, 1)
	want := StockStatus{IngredientID: 7, Description: "ready"}
	ready <- want
	if got := legacyClock(ready); got != want {
		t.Fatal(got)
	}
}
