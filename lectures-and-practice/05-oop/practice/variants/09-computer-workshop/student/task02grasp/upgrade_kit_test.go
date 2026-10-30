package task02grasp

import "testing"

func sampleProduct() UpgradeKit {
	return UpgradeKit{
		Name:         "Обновление компьютера",
		Memory:       Part{Name: "Память", Price: 2000},
		Storage:      Part{Name: "Накопитель", Price: 3000},
		Installation: Part{Name: "Установка", Price: 500},
	}
}

func TestTextReceipt(t *testing.T) {
	got := (Receipt{}).Text(sampleProduct(), 2)
	if got != "Обновление компьютера × 2: 11000 руб." {
		t.Fatalf("unexpected receipt: %q", got)
	}
}

func TestHTMLReceipt(t *testing.T) {
	product := sampleProduct()
	product.Name = "Набор <Стандарт> & дополнения"
	got := product.HTMLReceipt(2)
	if got != "<p>Набор &lt;Стандарт&gt; &amp; дополнения × 2: 11000 руб.</p>" {
		t.Fatalf("unexpected receipt: %q", got)
	}
}
