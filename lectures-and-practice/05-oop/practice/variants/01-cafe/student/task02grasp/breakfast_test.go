package task02grasp

import "testing"

func sampleProduct() Breakfast {
	return Breakfast{
		Name:     "Завтрак",
		Coffee:   Component{Name: "Кофе", Price: 150},
		Sandwich: Component{Name: "Сэндвич", Price: 250},
		Fruit:    Component{Name: "Фрукты", Price: 100},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Завтрак × 2: 1000 руб." {
		t.Fatalf("unexpected receipt: %q", got)
	}
}

func TestHTMLReceipt(t *testing.T) {
	product := sampleProduct()
	product.Name = "Набор <Стандарт> & дополнения"
	got := product.HTMLReceipt(2)
	if got != "<p>Набор &lt;Стандарт&gt; &amp; дополнения × 2: 1000 руб.</p>" {
		t.Fatalf("unexpected receipt: %q", got)
	}
}
