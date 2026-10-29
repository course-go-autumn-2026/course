package fitness

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckAttendees(o Visit) error { return ErrNotRefactored }
func CheckTotal(o Visit) error     { return ErrNotRefactored }
func CheckLimit(o Visit) error     { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Visit) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o Visit) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewClub(options ...Option) (*Club, error) { return nil, ErrNotRefactored }
func WithMaxAttendees(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithRegistrationFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
