package task02grasp

import "testing"

func sampleProduct() CareSet {
	return CareSet{
		Name:    "Полный уход",
		Washing: Procedure{Name: "Мытьё", Price: 200},
		Haircut: Procedure{Name: "Стрижка", Price: 400},
		Drying:  Procedure{Name: "Сушка", Price: 100},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Полный уход × 2: 1400 руб." {
		t.Fatalf("unexpected receipt: %q", got)
	}
}

func TestHTMLReceipt(t *testing.T) {
	product := sampleProduct()
	product.Name = "Набор <Стандарт> & дополнения"
	got := product.HTMLReceipt(2)
	if got != "<p>Набор &lt;Стандарт&gt; &amp; дополнения × 2: 1400 руб.</p>" {
		t.Fatalf("unexpected receipt: %q", got)
	}
}
