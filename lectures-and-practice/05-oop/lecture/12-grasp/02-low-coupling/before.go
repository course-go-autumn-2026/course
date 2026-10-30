package main

type Ingredient struct {
	Price int
}

type DetailedPizza struct {
	Dough  Ingredient
	Cheese Ingredient
	Sauce  Ingredient
	// Позже добавили ещё один платный ингредиент.
	Topping Ingredient
}

// Потребитель знает внутреннюю структуру пиццы и сам считает её цену.
// При добавлении Topping этот код забыли обновить: намеренная ошибка примера.
func BeforeReceiptTotal(pizza DetailedPizza) int {
	return pizza.Dough.Price + pizza.Cheese.Price + pizza.Sauce.Price
}
