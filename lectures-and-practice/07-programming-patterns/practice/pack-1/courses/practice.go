package courses

// Задание 1. Factory: реализуйте продукты и фабрику.
func (RecordedCourse) Name() string { return "" }
func (RecordedCourse) Price() int   { return 0 }
func (LiveCourse) Name() string     { return "" }
func (LiveCourse) Price() int       { return 0 }

func NewCourse(kind string) (Course, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithMentoring) Name() string   { return "" }
func (p WithMentoring) Price() int     { return 0 }
func (p WithCertificate) Name() string { return "" }
func (p WithCertificate) Price() int   { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func SelfEnrollment(base int) int     { return 0 }
func AssistedEnrollment(base int) int { return 0 }
func PriorityEnrollment(base int) int { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff EnrollmentPrice) int { return base }
