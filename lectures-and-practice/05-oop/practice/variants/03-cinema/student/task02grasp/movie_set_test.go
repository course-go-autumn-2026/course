package task02grasp

import "testing"

func sampleProduct() MovieSet {
	return MovieSet{
		Name:    "Киновечер",
		Ticket:  Component{Name: "Билет", Price: 300},
		Popcorn: Component{Name: "Попкорн", Price: 200},
		Glasses: Component{Name: "3D-очки", Price: 100},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Киновечер × 2: 1200 руб." {
		t.Fatalf("unexpected receipt: %q", got)
	}
}

func TestHTMLReceipt(t *testing.T) {
	product := sampleProduct()
	product.Name = "Набор <Стандарт> & дополнения"
	got := product.HTMLReceipt(2)
	if got != "<p>Набор &lt;Стандарт&gt; &amp; дополнения × 2: 1200 руб.</p>" {
		t.Fatalf("unexpected receipt: %q", got)
	}
}
