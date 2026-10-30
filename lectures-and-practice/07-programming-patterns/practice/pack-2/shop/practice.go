package shop

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckItems(o Purchase) error { return ErrNotRefactored }
func CheckTotal(o Purchase) error { return ErrNotRefactored }
func CheckLimit(o Purchase) error { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Purchase) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o Purchase) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewShop(options ...Option) (*Shop, error) { return nil, ErrNotRefactored }
func WithMaxItems(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithDeliveryFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
