package shipping

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckPackages(o Dispatch) error { return ErrNotRefactored }
func CheckTotal(o Dispatch) error    { return ErrNotRefactored }
func CheckLimit(o Dispatch) error    { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Dispatch) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o Dispatch) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewDeliveryService(options ...Option) (*DeliveryService, error) { return nil, ErrNotRefactored }
func WithMaxPackages(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithPickupFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
