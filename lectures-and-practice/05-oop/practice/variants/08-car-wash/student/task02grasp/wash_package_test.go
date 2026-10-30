package task02grasp

import "testing"

func sampleProduct() WashPackage {
	return WashPackage{
		Name:     "Комплексная мойка",
		Body:     Procedure{Name: "Кузов", Price: 300},
		Interior: Procedure{Name: "Салон", Price: 200},
		Mats:     Procedure{Name: "Коврики", Price: 100},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Комплексная мойка × 2: 1200 руб." {
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
