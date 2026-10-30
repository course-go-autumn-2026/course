package warehouse

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckPlaces(o Transfer) error { return ErrNotRefactored }
func CheckTotal(o Transfer) error  { return ErrNotRefactored }
func CheckLimit(o Transfer) error  { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Transfer) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o Transfer) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewWarehouse(options ...Option) (*Warehouse, error) { return nil, ErrNotRefactored }
func WithMaxPlaces(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithHandlingFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
