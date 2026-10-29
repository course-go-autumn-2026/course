package autoservice

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckJobs(o WorkOrder) error  { return ErrNotRefactored }
func CheckTotal(o WorkOrder) error { return ErrNotRefactored }
func CheckLimit(o WorkOrder) error { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o WorkOrder) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o WorkOrder) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewWorkshop(options ...Option) (*Workshop, error) { return nil, ErrNotRefactored }
func WithMaxJobs(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithCalloutFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
