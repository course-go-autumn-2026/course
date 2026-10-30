package autoservice

// Задание 1. Factory: реализуйте продукты и фабрику.
func (Inspection) Name() string     { return "" }
func (Inspection) Price() int       { return 0 }
func (WheelAlignment) Name() string { return "" }
func (WheelAlignment) Price() int   { return 0 }

func NewService(kind string) (Service, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithCarWash) Name() string      { return "" }
func (p WithCarWash) Price() int        { return 0 }
func (p WithPrioritySlot) Name() string { return "" }
func (p WithPrioritySlot) Price() int   { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func Scheduled(base int) int  { return 0 }
func SameDay(base int) int    { return 0 }
func NightShift(base int) int { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff ServicePrice) int { return base }
