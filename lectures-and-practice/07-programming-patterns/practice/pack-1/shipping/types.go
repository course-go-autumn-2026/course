package shipping

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Shipment interface {
	Name() string
	Price() int
}

type Envelope struct{}
type Parcel struct{}

type WithInsurance struct{ Base Shipment }
type WithFragileHandling struct{ Base Shipment }

type ShippingPrice func(base int) int

var ErrUnknownShipment = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
