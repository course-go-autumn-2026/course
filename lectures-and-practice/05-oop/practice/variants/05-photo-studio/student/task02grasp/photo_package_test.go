package task02grasp

import "testing"

func sampleProduct() PhotoPackage {
	return PhotoPackage{
		Name:     "Семейная съёмка",
		Shooting: ServicePart{Name: "Съёмка", Price: 1000},
		Retouch:  ServicePart{Name: "Ретушь", Price: 400},
		Prints:   ServicePart{Name: "Печать фото", Price: 100},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Семейная съёмка × 2: 3000 руб." {
		t.Fatalf("unexpected receipt: %q", got)
	}
}

func TestHTMLReceipt(t *testing.T) {
	product := sampleProduct()
	product.Name = "Набор <Стандарт> & дополнения"
	got := product.HTMLReceipt(2)
	if got != "<p>Набор &lt;Стандарт&gt; &amp; дополнения × 2: 3000 руб.</p>" {
		t.Fatalf("unexpected receipt: %q", got)
	}
}
