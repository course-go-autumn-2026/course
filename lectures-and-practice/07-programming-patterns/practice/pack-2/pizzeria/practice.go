package pizzeria

// Задание 1. Chain of Responsibility.
// TODO: извлеките три независимые проверки из legacyValidate.
func CheckPizzas(o Order) error { return ErrNotRefactored }
func CheckTotal(o Order) error  { return ErrNotRefactored }
func CheckLimit(o Order) error  { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Order) error { return ErrNotRefactored }
}

// TODO: используйте Chain с прежним порядком проверок.
func Validate(o Order) error { return legacyValidate(o) }

// Задание 2. Observer.
// TODO: храните подписчиков в Notifier; Publish не знает их конкретных ролей.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options.
// TODO: значения по умолчанию — 5 пицц, доставка 100 рублей.
// Опции применяются по порядку, до первой ошибки.
func NewPizzeria(options ...Option) (*Pizzeria, error) { return nil, ErrNotRefactored }
func WithMaxPizzas(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithDeliveryFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
