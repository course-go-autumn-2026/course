package task02grasp

import "testing"

func sampleProduct() RideKit {
	return RideKit{
		Name:   "Набор для прогулки",
		Bike:   Equipment{Name: "Велосипед", Price: 500},
		Helmet: Equipment{Name: "Шлем", Price: 100},
		Lock:   Equipment{Name: "Замок", Price: 50},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Набор для прогулки × 2: 1300 руб." {
		t.Fatalf("unexpected receipt: %q", got)
	}
}

func TestHTMLReceipt(t *testing.T) {
	product := sampleProduct()
	product.Name = "Набор <Стандарт> & дополнения"
	got := product.HTMLReceipt(2)
	if got != "<p>Набор &lt;Стандарт&gt; &amp; дополнения × 2: 1300 руб.</p>" {
		t.Fatalf("unexpected receipt: %q", got)
	}
}
