package shop

// Задание 1. Factory: реализуйте продукты и фабрику.
func (Keyboard) Name() string { return "" }
func (Keyboard) Price() int   { return 0 }
func (Mouse) Name() string    { return "" }
func (Mouse) Price() int      { return 0 }

func NewProduct(kind string) (Product, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithWarranty) Name() string { return "" }
func (p WithWarranty) Price() int   { return 0 }
func (p WithGiftWrap) Name() string { return "" }
func (p WithGiftWrap) Price() int   { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func StorePickup(base int) int { return 0 }
func Courier(base int) int     { return 0 }
func Express(base int) int     { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff FulfillmentPrice) int { return base }
