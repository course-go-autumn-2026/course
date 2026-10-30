package library

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckCopies(o LoanRequest) error { return ErrNotRefactored }
func CheckTotal(o LoanRequest) error  { return ErrNotRefactored }
func CheckLimit(o LoanRequest) error  { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o LoanRequest) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o LoanRequest) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewLibrary(options ...Option) (*Library, error) { return nil, ErrNotRefactored }
func WithMaxCopies(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithDeliveryFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
