package library

// Задание 1. Factory: реализуйте продукты и фабрику.
func (StandardLoan) Name() string { return "" }
func (StandardLoan) Price() int   { return 0 }
func (ExtendedLoan) Name() string { return "" }
func (ExtendedLoan) Price() int   { return 0 }

func NewLoanPlan(kind string) (LoanPlan, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithScanning) Name() string        { return "" }
func (p WithScanning) Price() int          { return 0 }
func (p WithProtectiveCover) Name() string { return "" }
func (p WithProtectiveCover) Price() int   { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func DeskPickup(base int) int       { return 0 }
func HomeDelivery(base int) int     { return 0 }
func PriorityDelivery(base int) int { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff LoanPrice) int { return base }
