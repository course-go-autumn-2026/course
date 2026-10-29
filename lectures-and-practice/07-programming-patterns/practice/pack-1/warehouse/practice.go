package warehouse

// Задание 1. Factory: реализуйте продукты и фабрику.
func (SmallCell) Name() string { return "" }
func (SmallCell) Price() int   { return 0 }
func (LargeCell) Name() string { return "" }
func (LargeCell) Price() int   { return 0 }

func NewStorageCell(kind string) (StorageCell, error) { return nil, ErrNotRefactored }

// Задание 2. Decorator: делегируйте Base и добавляйте услугу.
func (p WithClimateControl) Name() string { return "" }
func (p WithClimateControl) Price() int   { return 0 }
func (p WithInsurance) Name() string      { return "" }
func (p WithInsurance) Price() int        { return 0 }

// Задание 3. Strategy: перенесите тарифы в функции.
func SelfHandling(base int) int     { return 0 }
func StandardHandling(base int) int { return 0 }
func ExpressHandling(base int) int  { return 0 }

// TODO: вызовите переданный тариф ровно один раз.
func Checkout(base int, tariff HandlingPrice) int { return base }
