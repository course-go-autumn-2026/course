package cinema

// Задание 1. Factory: реализуйте продукты и фабрику.
func (StandardTicket) Name() string { return "" }
func (StandardTicket) Price() int   { return 0 }
func (PremiumTicket) Name() string  { return "" }
func (PremiumTicket) Price() int    { return 0 }

func NewTicket(kind string) (Ticket, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithPopcorn) Name() string { return "" }
func (p WithPopcorn) Price() int   { return 0 }
func (p WithDrink) Name() string   { return "" }
func (p WithDrink) Price() int     { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func BoxOffice(base int) int        { return 0 }
func OnlineBooking(base int) int    { return 0 }
func ConciergeBooking(base int) int { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff TicketPrice) int { return base }
