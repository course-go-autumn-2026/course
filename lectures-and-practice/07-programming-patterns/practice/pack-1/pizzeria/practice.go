package pizzeria

// Задание 1. Factory.
// TODO: опишите два продукта через общий интерфейс Pizza.
func (Margherita) Name() string { return "" }
func (Margherita) Price() int   { return 0 }
func (Pepperoni) Name() string  { return "" }
func (Pepperoni) Price() int    { return 0 }

// TODO: фабрика создаёт Margherita или Pepperoni; неизвестный вид — nil и ошибка.
func NewPizza(kind string) (Pizza, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator.
// TODO: каждая обёртка делегирует базовой Pizza и добавляет своё название/цену.
func (p WithCheese) Name() string { return "" }
func (p WithCheese) Price() int   { return 0 }
func (p WithBox) Name() string    { return "" }
func (p WithBox) Price() int      { return 0 }

// Задание 3. Strategy.
// TODO: перенесите тарифы из legacyOrder в отдельные функции.
func Pickup(base int) int  { return 0 }
func Courier(base int) int { return 0 }
func Express(base int) int { return 0 }

// TODO: стоимость определяет переданная функция, а не switch внутри Checkout.
func Checkout(base int, delivery DeliveryPrice) int { return base }
