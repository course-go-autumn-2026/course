package shipping

// Задание 1. Factory: реализуйте продукты и фабрику.
func (Envelope) Name() string { return "" }
func (Envelope) Price() int   { return 0 }
func (Parcel) Name() string   { return "" }
func (Parcel) Price() int     { return 0 }

func NewShipment(kind string) (Shipment, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithInsurance) Name() string       { return "" }
func (p WithInsurance) Price() int         { return 0 }
func (p WithFragileHandling) Name() string { return "" }
func (p WithFragileHandling) Price() int   { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func DropOff(base int) int       { return 0 }
func CourierPickup(base int) int { return 0 }
func ExpressPickup(base int) int { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff ShippingPrice) int { return base }
