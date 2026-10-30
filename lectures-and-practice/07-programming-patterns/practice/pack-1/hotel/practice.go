package hotel

// Задание 1. Factory: реализуйте продукты и фабрику.
func (StandardRoom) Name() string { return "" }
func (StandardRoom) Price() int   { return 0 }
func (SuiteRoom) Name() string    { return "" }
func (SuiteRoom) Price() int      { return 0 }

func NewRoom(kind string) (Room, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithBreakfast) Name() string { return "" }
func (p WithBreakfast) Price() int   { return 0 }
func (p WithParking) Name() string   { return "" }
func (p WithParking) Price() int     { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func DirectBooking(base int) int   { return 0 }
func AgentBooking(base int) int    { return 0 }
func PriorityBooking(base int) int { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff BookingPrice) int { return base }
