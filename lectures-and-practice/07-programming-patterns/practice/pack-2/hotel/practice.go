package hotel

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckNights(o Reservation) error { return ErrNotRefactored }
func CheckTotal(o Reservation) error  { return ErrNotRefactored }
func CheckLimit(o Reservation) error  { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Reservation) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o Reservation) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewHotel(options ...Option) (*Hotel, error) { return nil, ErrNotRefactored }
func WithMaxNights(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithServiceFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
