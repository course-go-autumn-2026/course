package task02grasp

import "testing"

func sampleProduct() GameSet {
	return GameSet{
		Name:    "Игровой набор",
		Game:    Accessory{Name: "Игра", Price: 300},
		Sleeves: Accessory{Name: "Протекторы", Price: 100},
		Insert:  Accessory{Name: "Органайзер", Price: 100},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Игровой набор × 2: 1000 руб." {
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
