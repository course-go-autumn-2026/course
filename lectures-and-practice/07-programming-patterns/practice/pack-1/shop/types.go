package shop

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Product interface {
	Name() string
	Price() int
}

type Keyboard struct{}
type Mouse struct{}

type WithWarranty struct{ Base Product }
type WithGiftWrap struct{ Base Product }

type FulfillmentPrice func(base int) int

var ErrUnknownProduct = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
