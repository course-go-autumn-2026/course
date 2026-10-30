package task02grasp

import "testing"

func sampleProduct() PrintSet {
	return PrintSet{
		Name:    "Учебный комплект",
		Pages:   ServicePart{Name: "Страницы", Price: 300},
		Cover:   ServicePart{Name: "Обложка", Price: 200},
		Binding: ServicePart{Name: "Переплёт", Price: 100},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Учебный комплект × 2: 1200 руб." {
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
