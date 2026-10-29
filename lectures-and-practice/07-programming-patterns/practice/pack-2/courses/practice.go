package courses

// Задание 1. Chain: независимые проверки и общий исполнитель.
func CheckStudents(o Enrollment) error { return ErrNotRefactored }
func CheckTotal(o Enrollment) error    { return ErrNotRefactored }
func CheckLimit(o Enrollment) error    { return ErrNotRefactored }
func Chain(checks ...Check) Check {
	return func(o Enrollment) error { return ErrNotRefactored }
}

// TODO: соберите проверки через Chain в прежнем порядке.
func Validate(o Enrollment) error { return ErrNotRefactored }

// Задание 2. Observer: регистрации и синхронная публикация всем.
func (n *Notifier) Subscribe(observer Observer) {}
func (n *Notifier) Publish(event Event) error   { return ErrNotRefactored }

// Задание 3. Functional Options: defaults 5/100, порядок и первая ошибка.
func NewSchool(options ...Option) (*School, error) { return nil, ErrNotRefactored }
func WithMaxStudents(max int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
func WithEnrollmentFee(fee int) Option {
	return func(c *Config) error { return ErrNotRefactored }
}
