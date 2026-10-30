package fitness

// Задание 1. Factory: реализуйте продукты и фабрику.
func (DaytimePass) Name() string { return "" }
func (DaytimePass) Price() int   { return 0 }
func (FullDayPass) Name() string { return "" }
func (FullDayPass) Price() int   { return 0 }

func NewMembership(kind string) (Membership, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithLocker) Name() string { return "" }
func (p WithLocker) Price() int   { return 0 }
func (p WithTowel) Name() string  { return "" }
func (p WithTowel) Price() int    { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func StartNextWeek(base int) int      { return 0 }
func StartToday(base int) int         { return 0 }
func PersonalOnboarding(base int) int { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff ActivationPrice) int { return base }
