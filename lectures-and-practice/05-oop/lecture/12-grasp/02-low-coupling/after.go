package main

type Pizza struct {
	ingredients []Ingredient
}

func NewPizza(ingredients ...Ingredient) Pizza {
	return Pizza{ingredients: append([]Ingredient(nil), ingredients...)}
}

func (p Pizza) Price() int {
	total := 0
	for _, ingredient := range p.ingredients {
		total += ingredient.Price
	}

	return total
}

// Чек зависит от операции Price, а не от полей каждого ингредиента.
// Добавление ингредиента не меняет код чека. Интерфейс для этого не обязателен.
func ReceiptTotal(pizza Pizza) int {
	return pizza.Price()
}
