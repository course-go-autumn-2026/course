package task02grasp

import "testing"

func sampleProduct() Bouquet {
	return Bouquet{
		Name:     "Весенний букет",
		Flowers:  Material{Name: "Цветы", Price: 400},
		Greenery: Material{Name: "Зелень", Price: 100},
		Ribbon:   Material{Name: "Лента", Price: 50},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Весенний букет × 2: 1100 руб." {
		t.Fatalf("unexpected receipt: %q", got)
	}
}

func TestHTMLReceipt(t *testing.T) {
	product := sampleProduct()
	product.Name = "Набор <Стандарт> & дополнения"
	got := product.HTMLReceipt(2)
	if got != "<p>Набор &lt;Стандарт&gt; &amp; дополнения × 2: 1100 руб.</p>" {
		t.Fatalf("unexpected receipt: %q", got)
	}
}
