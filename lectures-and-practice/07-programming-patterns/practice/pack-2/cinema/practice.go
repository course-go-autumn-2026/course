package cinema

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckSeats(o Booking) error { return ErrNotRefactored }
func CheckTotal(o Booking) error { return ErrNotRefactored }
func CheckLimit(o Booking) error { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Booking) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o Booking) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewCinema(options ...Option) (*Cinema, error) { return nil, ErrNotRefactored }
func WithMaxSeats(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithBookingFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
